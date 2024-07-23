package database

import (
	"context"
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
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

type Engine interface {
	Close() error
	Collection(col string) Operation
}

type Operation interface {
	Count(ctx context.Context) (int64, error)
	FindOne(ctx context.Context, filter interface{}) (Result, error)
	Find(ctx context.Context, filter interface{}) ([]Result, error)
	FindLimit(ctx context.Context, limit int64) (Cursor, error)
	Watcher(ctx context.Context) (<-chan func() (WatcherType, WatchResult), error)
}

type Cursor interface {
	Next(ctx context.Context) bool
	Result() ([]*Result, error)
}

func New(
	ctx context.Context,
	engine config.Engine,
	uri, database string,
	log logger.Logger,
) (Engine, error) {
	switch engine {
	case config.MONGO:
		return NewMongo(ctx, uri, database, log)
	default:
		return nil, ErrEngineNotSupported
	}
}

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
