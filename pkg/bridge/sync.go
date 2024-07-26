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

const _bulkLimit = int64(100)

type mongo struct {
	executor database.MongoExecutor
	indexMap map[config.Collection]*config.Destination
	meili    meilisearch.Meilisearch
	log      logger.Logger
}

func (m *mongo) OnDemand() {}

func (m *mongo) Bulk(ctx context.Context, isContinue bool) {
	var wg sync.WaitGroup
	taskCh := make(chan task, len(m.indexMap))
	statCh := make(chan stat, len(m.indexMap))

	for i := 0; i < len(m.indexMap); i++ {
		wg.Add(1)
		go m.bulkWorker(ctx, &wg, taskCh, statCh, isContinue)
	}

	for col, des := range m.indexMap {
		taskCh <- task{col: col.String(), des: des}
	}
	close(taskCh)

	go func() {
		for {
			select {
			case s, ok := <-statCh:
				if !ok {
					return
				}

				if s.err != nil {
					m.log.Fatal(s.err.Error())
				}

				progressBar(
					s.total,
					s.indexed,
					s.col,
					s.index,
				)
			}
		}
	}()

	wg.Wait()
	close(statCh)
}

func (m *mongo) bulkWorker(ctx context.Context,
	wg *sync.WaitGroup,
	taskCh <-chan task,
	statCh chan<- stat,
	isContinue bool,
) {
	defer wg.Done()
	for {
		select {
		case t, ok := <-taskCh:
			if !ok {
				return
			}

			s := m.meili.Stats()
			m.executor.AddCollection(t.col)

			count, err := m.executor.Count(ctx, t.col)
			if err != nil {
				statCh <- stat{err: err}
				return
			}

			if !isContinue {
				if err := recreateIndex(ctx,
					t.des.IndexName,
					t.des.PrimaryKey,
					t.des.Settings,
					m.meili); err != nil {
					m.log.Fatal("failed to recreate index", "err", err)
				}
			} else {
				if !m.meili.IsExistsIndex(t.des.IndexName) {
					m.log.Fatal(fmt.Sprintf("index %s does not exist for resync", t.des.IndexName))
				}

				if s != nil {
					if idxStat, ok := s.Indexes[t.des.IndexName]; ok {
						if idxStat.NumberOfDocuments == count {
							m.log.InfoContext(ctx, fmt.Sprintf("index %s already synced", t.des.IndexName))
							return
						}
					}
				}
			}

			idx := m.meili.Index(t.des.IndexName)
			cur, err := m.executor.FindLimit(ctx, _bulkLimit, t.col)
			if err != nil {
				statCh <- stat{err: err}
				return
			}

			totalIndexed := int64(0)

			for cur.Next(ctx) {
				items, err := cur.Result()
				if err != nil {
					statCh <- stat{err: err}
					return
				}

				updateItemKeys(items, t.des.Fields)

				tsk, err := idx.UpdateDocuments(&items)
				if err != nil {
					statCh <- stat{err: err}
					return
				}

				if err := m.meili.WaitForTask(ctx, tsk); err != nil {
					statCh <- stat{err: err}
					return
				}

				totalIndexed += int64(len(items))

				statCh <- stat{
					col:     t.col,
					index:   t.des.IndexName,
					total:   count,
					indexed: totalIndexed,
					err:     nil,
				}
			}
		}
	}
}
