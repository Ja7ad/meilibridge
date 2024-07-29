package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Result      map[string]interface{}
	WatchResult struct {
		DocumentId primitive.ObjectID
		Document   Result
		Update     struct {
			UpdateFields Result
			RemoveFields []string
		}
	}
	WatcherType uint8
)

const (
	Null WatcherType = iota
	OnInsert
	OnUpdate
	OnDelete
	OnReplace
)

func (w WatcherType) String() string {
	switch w {
	case OnInsert:
		return "insert"
	case OnUpdate:
		return "update"
	case OnDelete:
		return "delete"
	case OnReplace:
		return "replace"
	default:
		return ""
	}
}

type Cursor interface {
	Next(ctx context.Context) bool
	Result() ([]*Result, error)
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

type GlobalExecutor interface {
	Close() error
}

type MongoExecutor interface {
	GlobalExecutor

	AddCollection(col string)
	Count(ctx context.Context, col string) (int64, error)
	FindOne(ctx context.Context, filter interface{}, col string) (Result, error)
	FindLimit(ctx context.Context, limit int64, col string) (Cursor, error)
	Watcher(ctx context.Context, col string) (<-chan func() (WatcherType, WatchResult), error)
}

type SQLExecutor interface {
	GlobalExecutor

	Count(ctx context.Context, table string) (int64, error)
	FindOne(ctx context.Context, table string, query map[string]interface{}) (Result, error)
	FindLimit(ctx context.Context, table string, limit int64) (Cursor, error)
	// TODO: support trigger for realtime sync
}
