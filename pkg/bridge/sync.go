package bridge

import (
	"context"
	"fmt"
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
	"go.mongodb.org/mongo-driver/bson"
	"sync"
)

const _bulkLimit = int64(100)

type mongo struct {
	executor database.MongoExecutor
	indexMap map[config.Collection]*config.Destination
	meili    meilisearch.Meilisearch
	log      logger.Logger
}

func (m *mongo) OnDemand(ctx context.Context) {
	var wg sync.WaitGroup
	taskCh := make(chan task, len(m.indexMap))

	for i := 0; i < len(m.indexMap); i++ {
		wg.Add(1)
		go m.onDemandWorker(ctx, &wg, taskCh)
	}

	for col, des := range m.indexMap {
		taskCh <- task{col: col, des: des}
	}
	close(taskCh)

	wg.Wait()
}

func (m *mongo) Bulk(ctx context.Context, isContinue bool) {
	var wg sync.WaitGroup
	taskCh := make(chan task, len(m.indexMap))
	statCh := make(chan stat, len(m.indexMap))

	for i := 0; i < len(m.indexMap); i++ {
		wg.Add(1)
		go m.bulkWorker(ctx, &wg, taskCh, statCh, isContinue)
	}

	for col, des := range m.indexMap {
		taskCh <- task{col: col, des: des}
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

func (m *mongo) onDemandWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	taskCh <-chan task,
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

			hasView := false
			col, view := t.col.String(), ""

			if t.col.HasView() {
				col, view = t.col.GetCollectionAndView()
				m.executor.AddCollection(view)
				hasView = true
			}

			m.executor.AddCollection(col)

			if !m.meili.IsExistsIndex(t.des.IndexName) {
				if err := recreateIndex(
					ctx,
					t.des.IndexName,
					t.des.PrimaryKey,
					t.des.Settings,
					m.meili,
				); err != nil {
					m.log.Fatal(err.Error())
				}
			}

			statCh, err := m.executor.Watcher(ctx, col)
			if err != nil {
				m.log.Fatal(err.Error())
			}

			idx := m.meili.Index(t.des.IndexName)

			for {
				select {
				case <-ctx.Done():
					return
				case s, ok := <-statCh:
					if !ok {
						return
					}
					wType, res := s()

					switch wType {
					case database.OnInsert:
						go func() {
							m.log.InfoContext(ctx,
								fmt.Sprintf("add new document %s", res.DocumentId.Hex()),
								"collection", t.col, "index", t.des.IndexName,
							)
							result := res.Document

							if hasView {
								fmt.Println(res.DocumentId.Hex())
								result, err = m.executor.FindOne(ctx, bson.D{{"_id", res.DocumentId}}, view)
								if err != nil {
									m.log.Error(
										fmt.Sprintf("failed find documents in view index: %s", t.des.IndexName),
										"err", err.Error())
									return
								}
							}

							updateItemKeys([]*database.Result{&result}, t.des.Fields)
							tInfo, err := idx.AddDocuments(&result)
							if err != nil {
								m.log.Error(
									fmt.Sprintf("failed to add documents to index: %s", t.des.IndexName),
									"err", err.Error())
								return
							}

							if err := m.meili.WaitForTask(ctx, tInfo); err != nil {
								m.log.Error("failed to wait for complete insert task", "err", err.Error())
							}
						}()
					case database.OnUpdate:
						go func() {
							m.log.InfoContext(ctx,
								fmt.Sprintf("updating document %s", res.DocumentId.Hex()),
								"collection", t.col, "index", t.des.IndexName,
							)
							doc := make(map[string]interface{})
							err := idx.GetDocument(res.DocumentId.Hex(), nil, &doc)
							if err != nil {
								m.log.Error(
									fmt.Sprintf("failed to get document %s", res.DocumentId.Hex()),
									"err", err.Error())
								return
							}

							for k, v := range res.Update.UpdateFields {
								if _, ok := doc[k]; ok {
									doc[k] = v
								}
							}

							for _, field := range res.Update.RemoveFields {
								if _, ok := doc[field]; ok {
									delete(doc, field)
								}
							}

							tInfo, err := idx.UpdateDocuments(&doc, t.des.PrimaryKey)
							if err != nil {
								m.log.Error(
									fmt.Sprintf("failed to update document to index: %s", t.des.IndexName),
									"err", err.Error())
								return
							}

							if err := m.meili.WaitForTask(ctx, tInfo); err != nil {
								m.log.Error("failed to wait for complete update task", "err", err.Error())
							}
						}()
					case database.OnReplace:
						go func() {
							m.log.InfoContext(ctx,
								fmt.Sprintf("replace document %s", res.DocumentId.Hex()),
								"collection", t.col, "index", t.des.IndexName,
							)

							if hasView {
								res.Document, err = m.executor.FindOne(ctx, bson.M{"_id": res.DocumentId}, view)
								if err != nil {
									m.log.Error(
										fmt.Sprintf("failed find documents in view index: %s", t.des.IndexName),
										"err", err.Error())
									return
								}
							}

							tInfo, err := idx.UpdateDocuments(&res.Document, t.des.PrimaryKey)
							if err != nil {
								m.log.Error(
									fmt.Sprintf("failed to replace document to index: %s", t.des.IndexName),
									"err", err.Error())
								return
							}
							if err := m.meili.WaitForTask(ctx, tInfo); err != nil {
								m.log.Error("failed to wait for complete replace task", "err", err.Error())
							}
						}()
					case database.OnDelete:
						go func() {
							m.log.InfoContext(ctx,
								fmt.Sprintf("remove document %s", res.DocumentId.Hex()),
								"collection", t.col, "index", t.des.IndexName,
							)
							id := res.DocumentId.Hex()
							tInfo, err := idx.DeleteDocument(id)
							if err != nil {
								m.log.Error(
									fmt.Sprintf("failed to remove document to index: %s", t.des.IndexName),
									"err", err.Error())
								return
							}

							if err := m.meili.WaitForTask(ctx, tInfo); err != nil {
								m.log.Error("failed to wait for complete delete task", "err", err.Error())
							}
						}()
					default:
					}
				}
			}
		}
	}
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
		case <-ctx.Done():
			return
		case t, ok := <-taskCh:
			if !ok {
				return
			}

			col := t.col.String()

			if t.col.HasView() {
				_, col = t.col.GetCollectionAndView()
			}

			s := m.meili.Stats()
			m.executor.AddCollection(col)

			count, err := m.executor.Count(ctx, col)
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
			cur, err := m.executor.FindLimit(ctx, _bulkLimit, col)
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
					col:     col,
					index:   t.des.IndexName,
					total:   count,
					indexed: totalIndexed,
					err:     nil,
				}
			}
		}
	}
}
