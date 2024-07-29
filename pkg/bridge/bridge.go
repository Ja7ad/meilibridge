package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
)

func New(
	bridges []*config.Bridge,
	meili meilisearch.Meilisearch,
	log logger.Logger,
) *Bridge {
	return newBridge(bridges, meili, log)
}

func newBridge(
	bridges []*config.Bridge,
	meili meilisearch.Meilisearch,
	log logger.Logger,
) *Bridge {
	b := &Bridge{
		meili:   meili,
		log:     log,
		bridges: bridges,
	}

	return b
}

func (b *Bridge) Sync(ctx context.Context) error {
	var wg sync.WaitGroup
	syncer := b.initSyncers()

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

	syncer := b.initSyncers()

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

func (b *Bridge) initSyncers() []Syncer {
	syncer := make([]Syncer, 0)

	for _, bridge := range b.bridges {
		switch bridge.Source.Engine {
		case config.MONGO:
			mgo := new(mongo)
			mgo.name = bridge.Name
			eng := database.GetEngine[database.MongoExecutor](config.MONGO)
			mgo.executor = eng
			mgo.meili = b.meili
			mgo.indexMap = bridge.IndexMap
			mgo.log = b.log
			syncer = append(syncer, mgo)
		case config.MYSQL:
			sq := new(sql)
			sq.name = bridge.Name
			eng := database.GetEngine[database.SQLExecutor](config.MYSQL)
			sq.executor = eng
			sq.meili = b.meili
			sq.indexMap = bridge.IndexMap
			sq.log = b.log
			syncer = append(syncer, sq)
		case config.POSTGRES:
			sq := new(sql)
			sq.name = bridge.Name
			eng := database.GetEngine[database.SQLExecutor](config.POSTGRES)
			sq.executor = eng
			sq.meili = b.meili
			sq.indexMap = bridge.IndexMap
			sq.log = b.log
			syncer = append(syncer, sq)
		}
	}

	return syncer
}
