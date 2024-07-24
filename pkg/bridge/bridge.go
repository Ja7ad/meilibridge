package bridge

import (
	"context"
	"fmt"
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
	"sync"
)

type StatFunc func(totalItems, totalIndexedItem int64)

type Bridge struct {
	meili   meilisearch.Meilisearch
	bridges []*config.Bridge
	engine  config.Engine
	log     logger.Logger
}

type stat struct {
	totalIndexed int64
	completed    bool
	err          error
}

func New(
	bridges []*config.Bridge,
	meili meilisearch.Meilisearch,
	engine config.Engine,
	log logger.Logger,
) *Bridge {
	return newBridge(bridges, meili, engine, log)
}

func newBridge(
	bridges []*config.Bridge,
	meili meilisearch.Meilisearch,
	engine config.Engine,
	log logger.Logger,
) *Bridge {
	b := &Bridge{
		meili:   meili,
		log:     log,
		bridges: bridges,
		engine:  engine,
	}

	return b
}

func (b *Bridge) Sync() {}

func (b *Bridge) BulkSync(ctx context.Context, statFunc StatFunc, isContinue bool) error {
	resultsCh := make([]<-chan stat, 0, len(b.bridges))
	totalItems := int64(0)
	totalIndexed := int64(0)

	var mu sync.Mutex
	var wg sync.WaitGroup

	b.log.InfoContext(ctx, "starting bulk sync")

	for _, bridge := range b.bridges {
		if !isContinue {
			if err := b.recreateIndex(ctx, bridge); err != nil {
				return err
			}
		} else {
			b.log.InfoContext(ctx, "continue bulk sync on exists index")
			if !b.meili.IsExistsIndex(bridge.IndexName) {
				return fmt.Errorf("index %s does not exist for resync", bridge.IndexName)
			}
			if bridge.Settings == nil || bridge.Settings != nil && bridge.Settings.PrimaryKey == "" {
				return fmt.Errorf("primary key for index %s not set", bridge.IndexName)
			}
		}

		// TODO: how to support another database engine?
		eng := database.GetEngine[database.MongoExecutor](b.engine)
		eng.AddCollection(bridge.Collection)

		c, err := eng.Count(ctx, bridge.Collection)
		if err != nil {
			return err
		}

		totalItems += c

		resultsCh = append(resultsCh, b.processMany(ctx, c, bridge, eng))

	}

	b.log.InfoContext(ctx, fmt.Sprintf("total item %d for sync", totalItems))

	for _, resCh := range resultsCh {
		wg.Add(1)
		go func(resCh <-chan stat) {
			for res := range resCh {
				if res.err != nil {
					b.log.Fatal(res.err.Error())
				}

				if res.completed {
					wg.Done()
				}

				mu.Lock()
				totalIndexed += res.totalIndexed
				mu.Unlock()

				statFunc(totalItems, totalIndexed)
			}
		}(resCh)
	}

	b.log.InfoContext(ctx, "started bulk sync...")

	wg.Wait()

	fmt.Println()
	b.log.InfoContext(ctx, "finished bulk sync")

	return nil
}

func (b *Bridge) processMany(
	ctx context.Context,
	totalItem int64,
	bridge *config.Bridge,
	eng database.MongoExecutor,
) <-chan stat {
	statCh := make(chan stat)

	limit := int64(100)

	if totalItem < limit {
		limit = totalItem
	}

	go func() {
		defer close(statCh)

		s := b.meili.Stats()
		if s != nil {
			if idxStat, ok := s.Indexes[bridge.IndexName]; ok {
				if idxStat.NumberOfDocuments == totalItem {
					statCh <- stat{
						completed: true,
					}
					b.log.InfoContext(ctx, fmt.Sprintf("index %s already synced", bridge.IndexName))
					return
				}
			}
		}

		idx := b.meili.Index(bridge.IndexName)

		cur, err := eng.FindLimit(ctx, limit, bridge.Collection)
		if err != nil {
			statCh <- stat{err: err}
			return
		}

		for cur.Next(ctx) {
			items, err := cur.Result()
			if err != nil {
				statCh <- stat{err: err}
				return
			}

			b.updateItemKeys(items, bridge.Fields)

			t, err := idx.UpdateDocuments(&items)
			if err != nil {
				statCh <- stat{err: err}
				return
			}

			if err := b.meili.WaitForTask(ctx, t); err != nil {
				statCh <- stat{err: err}
				return
			}

			statCh <- stat{
				totalIndexed: int64(len(items)),
				err:          nil,
			}
		}

		statCh <- stat{
			completed: true,
		}
	}()

	return statCh
}

func (b *Bridge) recreateIndex(ctx context.Context, bridge *config.Bridge) error {
	if b.meili.IsExistsIndex(bridge.IndexName) {
		b.log.InfoContext(ctx, "removing old index...")
		if err := b.meili.DeleteIndex(ctx, bridge.IndexName); err != nil {
			return err
		}
	}

	primaryKey := ""
	if bridge.Settings != nil {
		primaryKey = bridge.Settings.PrimaryKey
	}

	b.log.InfoContext(ctx, fmt.Sprintf("creating index %s", bridge.IndexName))
	if err := b.meili.CreateIndex(ctx, bridge.IndexName, primaryKey); err != nil {
		return err
	}

	if err := b.meili.UpdateIndexSettings(ctx, bridge.IndexName, bridge.Settings.Settings); err != nil {
		return err
	}

	b.log.InfoContext(ctx, fmt.Sprintf("created new index %s", bridge.IndexName))

	return nil
}

func (b *Bridge) updateItemKeys(results []*database.Result, fields map[string]string) {
	if fields == nil {
		return
	}

	for i := range results {
		resultMap := *results[i]

		for key := range resultMap {
			if _, exists := fields[key]; !exists {
				delete(resultMap, key)
			}
		}

		for fk, fv := range fields {
			if fv != "" {
				if value, exists := resultMap[fk]; exists {
					resultMap[fv] = value
					delete(resultMap, fk)
				}
			}
		}
	}
}
