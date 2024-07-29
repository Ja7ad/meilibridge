package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/Ja7ad/meilibridge/pkg/meilisearch"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
)

func New(
	bridges []*config.Bridge,
	log logger.Logger,
) *Bridge {
	return newBridge(bridges, log)
}

func newBridge(
	bridges []*config.Bridge,
	log logger.Logger,
) *Bridge {
	b := &Bridge{
		log:     log,
		bridges: bridges,
	}

	return b
}

func (b *Bridge) Sync(ctx context.Context) error {
	var wg sync.WaitGroup

	syncer, err := b.initSyncers(ctx)
	if err != nil {
		return err
	}

	for _, s := range syncer {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.OnDemand(ctx)
			b.log.InfoContext(ctx, fmt.Sprintf("started on demand sync bridge %s", s.Name()))
		}()
	}

	b.log.InfoContext(ctx, "on demand is idle...")
	wg.Wait()

	return nil
}

func (b *Bridge) BulkSync(ctx context.Context, isContinue bool) error {
	var wg sync.WaitGroup

	syncer, err := b.initSyncers(ctx)
	if err != nil {
		return err
	}

	for _, s := range syncer {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.log.InfoContext(ctx, fmt.Sprintf("starting bulk sync bridge %s", s.Name()))
			s.Bulk(ctx, isContinue)
		}()
	}

	wg.Wait()
	b.log.InfoContext(ctx, "finished bulk sync")

	return nil
}

func (b *Bridge) initSyncers(ctx context.Context) ([]Syncer, error) {
	syncer := make([]Syncer, 0)

	for _, bridge := range b.bridges {
		switch bridge.Database.Engine {
		case config.MONGO:
			mgo := new(mongo)
			mgo.name = bridge.Name
			eng := database.GetEngine[database.MongoExecutor](config.MONGO)
			mgo.executor = eng
			mgo.indexMap = bridge.IndexMap
			mgo.log = b.log

			m, err := meilisearch.New(ctx, bridge.Meilisearch.APIURL, bridge.Meilisearch.APIKey, b.log)
			if err != nil {
				return nil, err
			}
			mgo.meili = m

			syncer = append(syncer, mgo)
		case config.MYSQL:
			sq := new(sql)
			sq.name = bridge.Name
			eng := database.GetEngine[database.SQLExecutor](config.MYSQL)
			sq.executor = eng
			sq.indexMap = bridge.IndexMap
			sq.log = b.log

			m, err := meilisearch.New(ctx, bridge.Meilisearch.APIURL, bridge.Meilisearch.APIKey, b.log)
			if err != nil {
				return nil, err
			}
			sq.meili = m

			syncer = append(syncer, sq)
		case config.POSTGRES:
			sq := new(sql)
			sq.name = bridge.Name
			eng := database.GetEngine[database.SQLExecutor](config.POSTGRES)
			sq.executor = eng
			sq.indexMap = bridge.IndexMap
			sq.log = b.log

			m, err := meilisearch.New(ctx, bridge.Meilisearch.APIURL, bridge.Meilisearch.APIKey, b.log)
			if err != nil {
				return nil, err
			}
			sq.meili = m

			syncer = append(syncer, sq)
		}
	}

	return syncer, nil
}
