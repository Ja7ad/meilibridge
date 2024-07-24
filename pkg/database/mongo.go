package database

import (
	"context"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
)

type Mongo struct {
	cli *mongo.Client
	log logger.Logger
	db  *mongo.Database

	collections map[string]*mongo.Collection
}

func newMongo(
	ctx context.Context,
	uri string, database string,
	log logger.Logger,
) (MongoExecutor, error) {
	mgo := &Mongo{
		collections: make(map[string]*mongo.Collection),
	}

	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := cli.Ping(ctx, nil); err != nil {
		return nil, err
	}

	mgo.cli = cli
	mgo.log = log
	mgo.db = cli.Database(database)

	return mgo, nil
}

func (m *Mongo) Close() error {
	return m.cli.Disconnect(context.Background())
}

func (m *Mongo) AddCollection(col string) {
	if _, exists := m.collections[col]; !exists {
		m.collections[col] = m.db.Collection(col)
	}
}

func (m *Mongo) Count(ctx context.Context, col string) (int64, error) {
	return m.collections[col].EstimatedDocumentCount(ctx)
}

func (m *Mongo) FindOne(ctx context.Context, filter interface{}, col string) (Result, error) {
	var res Result

	return res, m.collections[col].FindOne(ctx, filter).Decode(&res)
}

func (m *Mongo) Find(ctx context.Context, filter interface{}, col string) <-chan Result {
	resCh := make(chan Result)

	go func() {
		defer close(resCh)

		cursor, err := m.collections[col].Find(ctx, filter)
		if err != nil {
			m.log.Fatal(err.Error())
			return
		}
		defer func() {
			_ = cursor.Close(ctx)
		}()

		for cursor.Next(ctx) {
			var res Result
			if err := cursor.Decode(&res); err != nil {
				m.log.Fatal(err.Error())
				return
			}
			resCh <- res
		}
	}()

	return resCh
}

func (m *Mongo) FindLimit(ctx context.Context, limit int64, col string) (Cursor, error) {
	count, err := m.Count(ctx, col)
	if err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(count) / float64(limit)))

	return &cur{
		col:   m.collections[col],
		limit: limit,
		pages: totalPages,
		page:  0,
		total: count,
		err:   nil,
		res:   make([]*Result, 0),
	}, nil
}

func (m *Mongo) Watcher(ctx context.Context, col string) (<-chan func() (wType WatcherType, res WatchResult), error) {
	resCh := make(chan func() (wType WatcherType, res WatchResult))

	cs, err := m.collections[col].Watch(ctx, buildChangeStreamAggregationPipeline())
	if err != nil {
		return nil, err
	}

	go func() {
		for cs.Next(ctx) {
			var changeEvent mongoChangeEvent

			err := cs.Decode(&changeEvent)
			if err != nil {
				m.log.Error("failed to decode event", "err", err)
				continue
			}

			res := WatchResult{
				DocumentId: changeEvent.DocumentKey,
				Document:   changeEvent.FullDocument,
				Update: struct {
					UpdateFields Result
					RemoveFields []string
				}{UpdateFields: changeEvent.UpdateDescription.UpdatedFields,
					RemoveFields: changeEvent.UpdateDescription.RemovedFields,
				},
			}

			resCh <- func() (WatcherType, WatchResult) {
				return m.wType(changeEvent.OperationType), res
			}
		}
	}()

	return resCh, nil
}

func (m *Mongo) wType(op string) WatcherType {
	switch op {
	case OnInsert.String():
		return OnInsert
	case OnUpdate.String():
		return OnUpdate
	case OnDelete.String():
		return OnDelete
	case OnReplace.String():
		return OnReplace
	default:
		return Null
	}
}

type cur struct {
	total  int64
	pages  int64
	page   int64
	limit  int64
	col    *mongo.Collection
	cursor *mongo.Cursor
	err    error
	res    []*Result
}

func (c *cur) Next(ctx context.Context) bool {
	if c.page >= c.pages {
		return false
	}

	skip := c.page * c.limit
	opts := options.Find().SetLimit(c.limit).SetSkip(skip)

	if c.cursor != nil {
		c.err = c.cursor.Close(ctx)
	}

	c.cursor, c.err = c.col.Find(ctx, bson.D{}, opts)
	if c.err != nil {
		return false
	}

	c.res = make([]*Result, 0)

	if c.err = c.cursor.All(ctx, &c.res); c.err != nil {
		return false
	}

	c.page++
	return true
}

func (c *cur) Result() ([]*Result, error) {
	return c.res, c.err
}

func buildChangeStreamAggregationPipeline() mongo.Pipeline {
	pipeline := mongo.Pipeline{bson.D{
		{Key: "$addFields", Value: bson.D{
			{Key: "documentKey", Value: "$documentKey._id"},
		},
		},
	},
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "operationType", Value: 1},
				{Key: "documentKey", Value: 1},
				{Key: "fullDocument", Value: 1},
				{Key: "updateDescription", Value: 1},
			}}}}

	return pipeline
}
