package bridge

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Ja7ad/meilibridge/pkg/meilisearch"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
)

func New(
	bridges []*config.Bridge,
	triggerCfg *config.TriggerSync,
	log logger.Logger,
) *Bridge {
	return newBridge(bridges, triggerCfg, log)
}

func newBridge(
	bridges []*config.Bridge,
	triggerCfg *config.TriggerSync,
	log logger.Logger,
) *Bridge {
	b := &Bridge{
		log:        log,
		bridges:    bridges,
		triggerCfg: triggerCfg,
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

func (b *Bridge) TriggerSync(ctx context.Context) error {
	b.mux = http.NewServeMux()

	_, err := b.initSyncers(ctx)
	if err != nil {
		return err
	}

	b.mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	sv := &http.Server{
		Handler: b.mux,
		Addr:    b.triggerCfg.Listen,
	}

	go func() {
		<-ctx.Done()
		sv.Shutdown(ctx)
	}()

	b.log.Info("started trigger sync webhook", "addr", b.triggerCfg.Listen)

	return sv.ListenAndServe()
}

func (b *Bridge) initSyncers(ctx context.Context) ([]Syncer, error) {
	syncer := make([]Syncer, 0)

	for _, bridge := range b.bridges {
		switch bridge.Database.Engine {
		case config.MONGO:
			mgo := new(mongo)
			mgo.name = bridge.Name
			mgo.executor = database.GetEngine[database.MongoExecutor](config.MONGO)
			mgo.indexMap = bridge.IndexMap
			mgo.log = b.log

			m, err := meilisearch.New(ctx, bridge.Meilisearch.APIURL, bridge.Meilisearch.APIKey, b.log)
			if err != nil {
				return nil, err
			}
			mgo.meili = m

			if b.mux != nil {
				mgo.queue = newQueue(b.log)
				mgo.triggerToken = b.triggerCfg.Token

				go func() {
					mgo.queue.Process(ctx, mgo.processTrigger)
				}()

				for col, idx := range bridge.IndexMap {
					mgo.executor.AddCollection(col.GetView())
					pattern := fmt.Sprintf("/%s/%s", bridge.Name, idx.IndexName)
					b.mux.HandleFunc(pattern, mgo.Trigger())
					b.log.Info(fmt.Sprintf("add trigger webhook for %s", idx.IndexName), "pattern", pattern)

				}
			}

			syncer = append(syncer, mgo)
		case config.POSTGRES, config.MYSQL:
			sq := new(sql)
			sq.name = bridge.Name
			sq.executor = database.GetEngine[database.SQLExecutor](bridge.Database.Engine)
			sq.indexMap = bridge.IndexMap
			sq.log = b.log

			m, err := meilisearch.New(ctx, bridge.Meilisearch.APIURL, bridge.Meilisearch.APIKey, b.log)
			if err != nil {
				return nil, err
			}
			sq.meili = m

			if b.mux != nil {
				sq.queue = newQueue(b.log)
				sq.triggerToken = b.triggerCfg.Token

				go func() {
					sq.queue.Process(ctx, sq.processTrigger)
				}()

				for _, idx := range bridge.IndexMap {
					pattern := fmt.Sprintf("/%s/%s", bridge.Name, idx.IndexName)
					b.mux.HandleFunc(pattern, sq.Trigger())
					b.log.Info(fmt.Sprintf("add trigger webhook for %s", idx.IndexName), "pattern", pattern)
				}
			}

			syncer = append(syncer, sq)
		}
	}

	return syncer, nil
}
