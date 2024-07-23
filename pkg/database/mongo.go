package database

import (
	"context"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	cli *mongo.Client
	log logger.Logger
	*MongoCollection
}

type MongoCollection struct {
	col *mongo.Collection
	db  *mongo.Database
}

type mongoChangeEvent struct {
	OperationType     string             `bson:"operationType" json:"operationType"`
	DocumentKey       primitive.ObjectID `bson:"documentKey" json:"documentKey"`
	FullDocument      Result             `bson:"fullDocument" json:"fullDocument"`
	UpdateDescription struct {
		UpdatedFields Result   `bson:"updatedFields" json:"updatedFields"`
		RemovedFields []string `bson:"removedFields" json:"removedFields"`
	} `bson:"updateDescription" json:"updateDescription"`
}

type documentKey struct {
	ID primitive.ObjectID `bson:"_id"`
}

func NewMongo(
	ctx context.Context,
	uri string, database string,
	log logger.Logger,
) (Engine, error) {
	mgo := new(Mongo)

	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := cli.Ping(ctx, nil); err != nil {
		return nil, err
	}

	mgo.cli = cli
	mgo.log = log
	mgo.MongoCollection = &MongoCollection{
		db: cli.Database(database),
	}

	return mgo, nil
}

func (m *Mongo) Close() error {
	return m.cli.Disconnect(context.Background())
}

func (m *Mongo) Collection(col string) Operation {
	m.MongoCollection.col = m.db.Collection(col)
	return m
}

func (m *Mongo) Count(ctx context.Context) (int64, error) {
	return m.col.EstimatedDocumentCount(ctx)
}

func (m *Mongo) FindOne(ctx context.Context, filter interface{}) (Result, error) {
	var res Result

	return res, m.col.FindOne(ctx, filter).Decode(&res)
}

func (m *Mongo) Find(ctx context.Context, filter interface{}) ([]Result, error) {
	res := make([]Result, 0)

	cursor, err := m.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	return res, cursor.All(ctx, &res)
}

func (m *Mongo) FindLimit(ctx context.Context, limit int64) (Cursor, error) {
	count, err := m.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalPages := count / limit

	return &cur{
		col:   m.col,
		limit: limit,
		pages: totalPages,
		page:  0,
		total: count,
		err:   nil,
		res:   make([]*Result, 0),
	}, nil
}

func (m *Mongo) Watcher(ctx context.Context) (<-chan func() (wType WatcherType, res WatchResult), error) {
	resCh := make(chan func() (wType WatcherType, res WatchResult))

	cs, err := m.col.Watch(ctx, buildChangeStreamAggregationPipeline())
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
