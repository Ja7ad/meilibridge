package main

import (
	"context"
	"errors"

	"github.com/Ja7ad/meilibridge/pkg/bridge"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
	meili "github.com/meilisearch/meilisearch-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func bulk(ctx context.Context) error {
	s, err := testSuite(ctx, true)
	if err != nil {
		return err
	}

	if err := s.cfg.Validate(); err != nil {
		return err
	}

	log := logger.DefaultLogger

	for _, b := range s.cfg.Bridges {
		err = database.AddEngine(
			ctx,
			b.Source.Engine,
			b.Source.URI,
			b.Source.Database,
			log,
		)
		if err != nil {
			return err
		}
	}

	m, err := meilisearch.New(ctx, s.cfg.Meilisearch.APIURL, s.cfg.Meilisearch.APIKey, log)
	if err != nil {
		return err
	}

	b := bridge.New(s.cfg.Bridges, m, log)

	if err := b.BulkSync(ctx, false); err != nil {
		return err
	}

	idx := m.Index(indexName)
	oldDocs := make([]map[string]interface{}, 0)
	if err := idx.GetDocuments(nil, &meili.DocumentsResult{
		Results: oldDocs,
	}); err != nil {
		return err
	}

	if len(oldDocs) != len(exampleBulk) {
		return errors.New("old documents count mismatch")
	}

	docs := []interface{}{
		map[string]interface{}{
			"_id":       primitive.NewObjectID(),
			"name":      "foo6",
			"last_name": "foo6",
			"age":       21,
		},
		map[string]interface{}{
			"_id":       primitive.NewObjectID(),
			"name":      "foo7",
			"last_name": "bar7",
			"age":       16,
		},
	}

	col := s.mgo.Database(s.cfg.Bridges[0].Source.Database).Collection(colName)

	if _, err := col.InsertMany(ctx, docs); err != nil {
		return err
	}

	if err := b.BulkSync(ctx, true); err != nil {
		return err
	}

	newDocs := make([]map[string]interface{}, 0)
	if err := idx.GetDocuments(nil, &meili.DocumentsResult{
		Results: newDocs,
	}); err != nil {
		return err
	}

	count := len(exampleBulk) + len(docs)

	if len(newDocs) != count {
		return errors.New("new documents count mismatch")
	}

	return nil
}
