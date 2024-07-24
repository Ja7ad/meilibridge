package database

import (
	"context"
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"sync"
)

var _pool *sync.Map

func init() {
	_pool = new(sync.Map)
}

func AddEngine(
	ctx context.Context,
	engine config.Engine,
	uri, database string,
	log logger.Logger,
) error {
	switch engine {
	case config.MONGO:
		mgoExec, err := newMongo(ctx, uri, database, log)
		if err != nil {
			return err
		}
		_pool.Store(config.MONGO, mgoExec)
		return nil
	default:
		return ErrEngineNotSupported
	}
}

func GetEngine[T GlobalExecutor](engine config.Engine) T {
	eng, _ := _pool.Load(engine)
	return eng.(T)
}
