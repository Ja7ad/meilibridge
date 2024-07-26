package main

import (
	"context"
	"errors"
	"github.com/Ja7ad/meilibridge/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

type suite struct {
	cfg *config.Config
	mgo *mongo.Client
}

func testSuite(ctx context.Context, isBulk bool) (*suite, error) {
	s := new(suite)

	mongoURI, ok := os.LookupEnv("MONGO_URI")
	if !ok {
		return nil, errors.New("environment variable MONGO_URI not found")
	}

	mongoDB, ok := os.LookupEnv("MONGO_DB")
	if !ok {
		return nil, errors.New("environment variable MONGO_DB not found")
	}

	meiliHost, ok := os.LookupEnv("MEILI_HOST")
	if !ok {
		return nil, errors.New("environment variable MEILI_HOST not found")
	}

	meiliAPI, ok := os.LookupEnv("MEILI_API")
	if !ok {
		return nil, errors.New("environment variable MEILI_API not found")
	}

	cfg := &config.Config{
		Meilisearch: &config.Meilisearch{
			APIURL: meiliHost,
			APIKey: meiliAPI,
		},
		Bridges: []*config.Bridge{
			{
				Name: "e2e1",
				Source: &config.Source{
					Engine:   config.MONGO,
					URI:      mongoURI,
					Database: mongoDB,
				},
				IndexMap: map[config.Collection]*config.Destination{
					"col1": {
						IndexName:  "idx1",
						PrimaryKey: "_id",
						Fields: map[string]string{
							"name":      "first_name",
							"last_name": "",
							"age":       "",
						},
						Settings: &config.Settings{
							FilterableAttributes: []string{"first_name", "age"},
							SortableAttributes:   []string{"age"},
						},
					},
				},
			},
		},
	}

	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Bridges[0].Source.URI))
	if err != nil {
		return nil, err
	}

	if err := cli.Ping(ctx, nil); err != nil {
		return nil, err
	}

	if isBulk {
		col := cli.Database(cfg.Bridges[0].Source.Database).Collection("col1")
		if _, err := col.InsertMany(ctx, exampleBulk); err != nil {
			return nil, err
		}
	}

	s.cfg = cfg

	return s, nil
}

var (
	exampleBulk = []interface{}{
		map[string]interface{}{
			"_id":       primitive.NewObjectID(),
			"name":      "foo1",
			"last_name": "bar1",
			"age":       13,
		},
		map[string]interface{}{
			"_id":       primitive.NewObjectID(),
			"name":      "foo2",
			"last_name": "bar2",
			"age":       33,
		},
		map[string]interface{}{
			"_id":       primitive.NewObjectID(),
			"name":      "foo3",
			"last_name": "foo3",
			"age":       34,
		},
		map[string]interface{}{
			"_id":       primitive.NewObjectID(),
			"name":      "foo4",
			"last_name": "bar4",
			"age":       21,
		},
	}
)
