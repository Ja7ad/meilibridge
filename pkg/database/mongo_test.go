package database

import (
	"context"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"sync"
	"testing"
	"time"
)

const (
	testDBName         = "foo"
	testCollectionName = "bar"
)

var (
	exampleManyData = []interface{}{
		map[string]interface{}{
			"_id":  primitive.NewObjectID(),
			"name": "foo1",
			"last": "bar1",
		},
		map[string]interface{}{
			"_id":  primitive.NewObjectID(),
			"name": "foo2",
			"last": "bar2",
		},
		map[string]interface{}{
			"_id":  primitive.NewObjectID(),
			"name": "foo3",
			"last": "bar3",
		},
		map[string]interface{}{
			"_id":  primitive.NewObjectID(),
			"name": "foo4",
			"last": "bar4",
		},
		map[string]interface{}{
			"_id":  primitive.NewObjectID(),
			"name": "foo5",
			"last": "bar5",
		},
		map[string]interface{}{
			"_id":  primitive.NewObjectID(),
			"name": "foo6",
			"last": "bar6",
		},
		map[string]interface{}{
			"_id":  primitive.NewObjectID(),
			"name": "foo7",
			"last": "bar7",
		},
		map[string]interface{}{
			"_id":  primitive.NewObjectID(),
			"name": "foo8",
			"last": "bar8",
		},
		map[string]interface{}{
			"_id":  primitive.NewObjectID(),
			"name": "foo8",
			"last": "bar8",
		},
	}

	client *mongo.Client
	engine MongoExecutor
	once   sync.Once
)

func setup(t *testing.T) {
	ctx := context.TODO()

	once.Do(func() {
		e, err := newMongo(context.TODO(), os.Getenv("MONGO_URI"), testDBName, logger.DefaultLogger)
		if err != nil {
			t.Fatal(err)
		}

		engine = e
		engine.AddCollection(testCollectionName)

		cli, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO_URI")))
		if err != nil {
			t.Fatal(err)
		}

		client = cli

		col := cli.Database(testDBName).Collection(testCollectionName)
		if _, err := col.InsertMany(ctx, exampleManyData); err != nil {
			t.Fatal(err)
		}
	})

}

func cleanup() error {
	return client.Database(testDBName).Drop(context.Background())
}

func TestMongo_FindOne(t *testing.T) {
	setup(t)
	defer cleanup()

	ctx := context.TODO()

	for _, ee := range exampleManyData {
		data, ok := ee.(map[string]interface{})
		assert.True(t, ok)

		res, err := engine.FindOne(ctx, bson.M{"name": data["name"]}, testCollectionName)
		assert.Nil(t, err)
		assert.NotNil(t, res)
	}
}

func Test_Find(t *testing.T) {
	setup(t)
	defer cleanup()

	ctx := context.TODO()

	resCh := engine.Find(ctx, bson.M{}, testCollectionName)

	for {
		select {
		case res, ok := <-resCh:
			if !ok {
				goto done
			}
			assert.NotNil(t, res)
		}
	}
done:
}

func Test_FindLimit(t *testing.T) {
	setup(t)
	defer cleanup()
	limit := int64(2)

	ctx := context.TODO()

	cursor, err := engine.FindLimit(ctx, limit, testCollectionName)
	assert.Nil(t, err)

	for cursor.Next(ctx) {
		res, err := cursor.Result()
		assert.Nil(t, err)
		assert.Len(t, res, int(limit))
	}
}

func Test_Watcher(t *testing.T) {
	setup(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resCh, err := engine.Watcher(ctx, testCollectionName)
	assert.Nil(t, err)

	id := primitive.NewObjectID()

	go func() {
		for res := range resCh {
			assert.NotNil(t, res)
			tp, val := res()
			assert.NotNil(t, val)
			t.Log(tp, val)
		}
	}()

	time.Sleep(1 * time.Second)

	coll := client.Database(testDBName).Collection(testCollectionName)

	res, err := coll.InsertOne(ctx, map[string]interface{}{
		"_id":  id,
		"name": "foo9",
		"last": "bar9",
	})
	assert.Nil(t, err)
	assert.NotNil(t, res.InsertedID)

	time.Sleep(2 * time.Second)

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"_id", id}}
	update := bson.D{
		{"$set", bson.D{{"name", "foo99"}}},
		{"$unset", bson.D{{"last", ""}}},
	}

	_, err = coll.UpdateOne(ctx, filter, update, opts)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	replacement := bson.D{{"name", "foo999"}, {"last", "bar999"}}
	_, err = coll.ReplaceOne(ctx, filter, replacement)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	_, err = coll.DeleteOne(ctx, bson.M{"_id": id})
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)
}
