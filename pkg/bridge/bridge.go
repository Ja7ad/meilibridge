package bridge

import (
	"context"
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
	b.log.InfoContext(ctx, "starting on demand sync...")
	syncer := b.initSyncers()

	for _, s := range syncer {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.OnDemand(ctx)
		}()
	}

	b.log.InfoContext(ctx, "on demand is idle...")
	wg.Wait()

	return nil
}

func (b *Bridge) BulkSync(ctx context.Context, isContinue bool) error {
	var wg sync.WaitGroup

	b.log.InfoContext(ctx, "starting bulk sync")
	syncer := b.initSyncers()

	for _, s := range syncer {
		wg.Add(1)
		go func() {
			defer wg.Done()
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
			eng := database.GetEngine[database.MongoExecutor](config.MONGO)
			mgo.executor = eng
			mgo.meili = b.meili
			mgo.indexMap = bridge.IndexMap
			mgo.log = b.log
			syncer = append(syncer, mgo)
		}
	}

	return syncer
}
