package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
)

type sql struct {
	name     string
	executor database.SQLExecutor
	indexMap map[config.Collection]*config.IndexConfig
	meili    meilisearch.Meilisearch
	log      logger.Logger
}

func (s *sql) Name() string {
	return s.name
}

func (s *sql) OnDemand(_ context.Context) {
	s.log.Warn("currently not support real-time sync for sql, " +
		"you can use bulk sync with --continue with scheduler.")
	return
}

func (s *sql) Bulk(ctx context.Context, isContinue bool) {
	var wg sync.WaitGroup
	taskCh := make(chan task, len(s.indexMap))
	statCh := make(chan stat, len(s.indexMap))

	for i := 0; i < len(s.indexMap); i++ {
		wg.Add(1)
		go s.bulkWorker(ctx, &wg, taskCh, statCh, isContinue)
	}

	for col, des := range s.indexMap {
		taskCh <- task{col: col, des: des}
	}
	close(taskCh)

	go func() {
		for {
			select {
			case st, ok := <-statCh:
				if !ok {
					return
				}

				if st.err != nil {
					s.log.Fatal(st.err.Error())
				}

				progressBar(
					st.total,
					st.indexed,
					s.name,
					st.col,
					st.index,
				)
			}
		}
	}()

	wg.Wait()
	close(statCh)
}

func (s *sql) bulkWorker(ctx context.Context,
	wg *sync.WaitGroup,
	taskCh <-chan task,
	statCh chan<- stat,
	isContinue bool,
) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-taskCh:
			if !ok {
				return
			}

			table := t.col.String()

			if t.col.HasView() {
				_, table = t.col.GetCollectionAndView()
			}

			st := s.meili.Stats()

			count, err := s.executor.Count(ctx, table)
			if err != nil {
				statCh <- stat{err: err}
				return
			}

			if !isContinue {
				if err := recreateIndex(ctx,
					t.des.IndexName,
					t.des.PrimaryKey,
					t.des.Settings,
					s.meili); err != nil {
					s.log.Fatal("failed to recreate index", "err", err)
				}
			} else {
				if !s.meili.IsExistsIndex(t.des.IndexName) {
					s.log.Fatal(fmt.Sprintf("index %s does not exist for resync", t.des.IndexName))
				}

				if s != nil {
					if idxStat, ok := st.Indexes[t.des.IndexName]; ok {
						if idxStat.NumberOfDocuments == count {
							s.log.InfoContext(ctx, fmt.Sprintf("index %s already synced", t.des.IndexName))
							return
						}
					}
				}
			}

			idx := s.meili.Index(t.des.IndexName)
			cur, err := s.executor.FindLimit(ctx, table, _bulkLimit)
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

				if err := s.meili.WaitForTask(ctx, tsk); err != nil {
					statCh <- stat{err: err}
					return
				}

				totalIndexed += int64(len(items))

				statCh <- stat{
					col:     table,
					index:   t.des.IndexName,
					total:   count,
					indexed: totalIndexed,
					err:     nil,
				}
			}
		}
	}
}
