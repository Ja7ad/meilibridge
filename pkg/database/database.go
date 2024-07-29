package database

import (
	"context"
	"sync"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
)

var _pool *sync.Map

func init() {
	_pool = new(sync.Map)
}

func AddEngine(
	ctx context.Context,
	source *config.Database,
	log logger.Logger,
) error {
	switch source.Engine {
	case config.MONGO:
		mgoExec, err := newMongo(ctx, source, log)
		if err != nil {
			return err
		}
		_pool.Store(config.MONGO, mgoExec)
		return nil
	case config.MYSQL:
		sqlExec, err := newSQL(source, log)
		if err != nil {
			return err
		}
		_pool.Store(config.MYSQL, sqlExec)
		return nil
	case config.POSTGRES:
		sqlExec, err := newSQL(source, log)
		if err != nil {
			return err
		}
		_pool.Store(config.POSTGRES, sqlExec)
		return nil
	default:
		return ErrEngineNotSupported
	}
}

func GetEngine[T GlobalExecutor](engine config.Engine) T {
	eng, _ := _pool.Load(engine)
	return eng.(T)
}
