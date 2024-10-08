package bridge

import (
	"context"
	"fmt"
	"github.com/Ja7ad/meilibridge/pkg/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"sync"

	meili "github.com/meilisearch/meilisearch-go"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
	"go.mongodb.org/mongo-driver/bson"
)

type mongo struct {
	name         string
	triggerToken string
	executor     database.MongoExecutor
	indexMap     map[config.Collection]*config.IndexConfig
	meili        meilisearch.Meilisearch
	queue        *Queue
	log          logger.Logger
}

func (m *mongo) Name() string {
	return m.name
}

func (m *mongo) Trigger() http.HandlerFunc {
	return triggerHandler(m.triggerToken, m.queue)
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
					m.name,
					s.col,
					s.index,
				)
			}
		}
	}()

	wg.Wait()
	close(statCh)
}

func (m *mongo) onDemandWorker(ctx context.Context, wg *sync.WaitGroup, taskCh <-chan task) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-taskCh:
			if !ok {
				return
			}

			if err := m.handleTask(ctx, t); err != nil {
				m.log.Error(err.Error())
			}
		}
	}
}

func (m *mongo) handleTask(ctx context.Context, t task) error {
	hasView, col, view := m.prepareCollections(t)

	if !m.meili.IsExistsIndex(ctx, t.des.IndexName) {
		if err := recreateIndex(ctx, t.des.IndexName, t.des.PrimaryKey, t.des.Settings, m.meili); err != nil {
			return err
		}
	}

	watch, err := m.executor.Watcher(ctx, col)
	if err != nil {
		return err
	}

	idx := m.meili.Index(t.des.IndexName)

	for {
		select {
		case <-ctx.Done():
			return nil
		case w, ok := <-watch:
			if !ok {
				return nil
			}
			if err := m.handleWatchEvent(ctx, idx, t, w, hasView, view); err != nil {
				m.log.Error(err.Error())
			}
		}
	}
}

func (m *mongo) prepareCollections(t task) (bool, string, string) {
	col, view := t.col.String(), ""
	hasView := false

	if t.col.HasView() {
		col, view = t.col.GetCollectionAndView()
		m.executor.AddCollection(view)
		hasView = true
	}

	m.executor.AddCollection(col)
	return hasView, col, view
}

func (m *mongo) handleWatchEvent(
	ctx context.Context,
	idx meili.IndexManager,
	t task,
	w func() (database.WatcherType, database.WatchResult),
	hasView bool,
	view string,
) error {
	wType, res := w()

	switch wType {
	case database.OnInsert:
		go m.handleInsert(ctx, idx, t, res, hasView, view)
	case database.OnUpdate:
		go m.handleUpdate(ctx, idx, t, res, view)
	case database.OnReplace:
		go m.handleReplace(ctx, idx, t, res, hasView, view)
	case database.OnDelete:
		go m.handleDelete(ctx, idx, t, res)
	default:
	}
	return nil
}

func (m *mongo) handleInsert(
	ctx context.Context,
	idx meili.IndexManager,
	t task,
	res database.WatchResult,
	hasView bool,
	view string,
) {
	m.log.InfoContext(ctx, fmt.Sprintf("add new document %s", res.DocumentId.Hex()),
		"collection", t.col, "index", t.des.IndexName)

	result := res.Document
	if hasView {
		var err error
		result, err = m.executor.FindOne(ctx, bson.D{{"_id", res.DocumentId}}, view)
		if err != nil {
			m.log.Warn(fmt.Sprintf("failed find documents in view index: %s", t.des.IndexName),
				"err", err.Error())
			return
		}
	}

	updateItemKeys([]*database.Result{&result}, t.des.Fields)
	tInfo, err := idx.AddDocuments(&result)
	if err != nil {
		m.log.Error(fmt.Sprintf("failed to add documents to index: %s", t.des.IndexName),
			"err", err.Error())
		return
	}

	if err := m.meili.WaitForTask(ctx, tInfo); err != nil {
		m.log.Error("failed to wait for complete insert task", "err", err.Error())
	}
}

func (m *mongo) handleUpdate(
	ctx context.Context,
	idx meili.IndexManager,
	t task,
	res database.WatchResult,
	view string,
) {
	m.log.InfoContext(ctx, fmt.Sprintf("updating document %s", res.DocumentId.Hex()),
		"collection", t.col, "index", t.des.IndexName)

	doc := make(database.Result)
	err := idx.GetDocument(res.DocumentId.Hex(), nil, &doc)
	if err != nil {
		doc, err = m.executor.FindOne(ctx, bson.D{{"_id", res.DocumentId}}, view)
		if err != nil {
			m.log.Warn(fmt.Sprintf("failed find documents in view index: %s", t.des.IndexName),
				"err", err.Error())
			return
		}
		updateItemKeys([]*database.Result{&doc}, t.des.Fields)
		tInfo, err := idx.AddDocuments(&doc)
		if err != nil {
			m.log.Error(fmt.Sprintf("failed to add documents to index: %s", t.des.IndexName),
				"err", err.Error())
			return
		}

		if err := m.meili.WaitForTask(ctx, tInfo); err != nil {
			m.log.Error("failed to wait for complete insert task", "err", err.Error())
		}
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
		m.log.Error(fmt.Sprintf("failed to update document to index: %s", t.des.IndexName),
			"err", err.Error())
		return
	}

	if err := m.meili.WaitForTask(ctx, tInfo); err != nil {
		m.log.Error("failed to wait for complete update task", "err", err.Error())
	}
}

func (m *mongo) handleReplace(
	ctx context.Context,
	idx meili.IndexManager,
	t task,
	res database.WatchResult,
	hasView bool,
	view string,
) {
	m.log.InfoContext(ctx, fmt.Sprintf("replace document %s", res.DocumentId.Hex()),
		"collection", t.col, "index", t.des.IndexName)

	if hasView {
		var err error
		res.Document, err = m.executor.FindOne(ctx, bson.M{"_id": res.DocumentId}, view)
		if err != nil {
			m.log.Warn(fmt.Sprintf("failed find documents in view index: %s", t.des.IndexName),
				"err", err.Error())
			return
		}
	}

	tInfo, err := idx.UpdateDocuments(&res.Document, t.des.PrimaryKey)
	if err != nil {
		m.log.Error(fmt.Sprintf("failed to replace document to index: %s", t.des.IndexName),
			"err", err.Error())
		return
	}
	if err := m.meili.WaitForTask(ctx, tInfo); err != nil {
		m.log.Error("failed to wait for complete replace task", "err", err.Error())
	}
}

func (m *mongo) handleDelete(ctx context.Context, idx meili.IndexManager, t task, res database.WatchResult) {
	m.log.InfoContext(ctx, fmt.Sprintf("remove document %s", res.DocumentId.Hex()),
		"collection", t.col, "index", t.des.IndexName)
	id := res.DocumentId.Hex()
	tInfo, err := idx.DeleteDocument(id)
	if err != nil {
		m.log.Error(fmt.Sprintf("failed to remove document to index: %s", t.des.IndexName),
			"err", err.Error())
		return
	}

	if err := m.meili.WaitForTask(ctx, tInfo); err != nil {
		m.log.Error("failed to wait for complete delete task", "err", err.Error())
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
				if !m.meili.IsExistsIndex(ctx, t.des.IndexName) {
					m.log.Fatal(fmt.Sprintf("index %s does not exist for resync", t.des.IndexName))
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

func (m *mongo) processTrigger(ctx context.Context, item types.TriggerRequestBody) (bool, error) {
	col, idx := indexConfigByUID(item.IndexUID, m.indexMap)
	if idx == nil {
		return false, fmt.Errorf("invalid index UID %s", item.IndexUID)
	}

	if !m.meili.IsExistsIndex(ctx, idx.IndexName) {
		if err := recreateIndex(ctx, idx.IndexName, idx.PrimaryKey, idx.Settings, m.meili); err != nil {
			return true, err
		}
	}

	var (
		val        any
		identifier string
	)
	val = item.Document.PrimaryValue

	identifier = fmt.Sprintf("%v", item.Document.PrimaryValue)

	obj, ok := item.Document.PrimaryValue.(string)
	if ok {
		v, err := primitive.ObjectIDFromHex(obj)
		if err == nil {
			val = v
			identifier = obj
		}
	}

	res, err := m.executor.FindOne(ctx, bson.M{item.Document.PrimaryKey: val}, col)
	if err != nil {
		return true, err
	}

	if err := processTrigger(ctx,
		m.meili.WaitForTask,
		m.meili.Index(item.IndexUID),
		item.Type,
		res,
		identifier,
	); err != nil {
		return true, err
	}

	return false, nil
}
