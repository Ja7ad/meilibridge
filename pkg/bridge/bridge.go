package bridge

import (
	"context"
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
	meili "github.com/meilisearch/meilisearch-go"
)

type Bridge struct {
	meili  meilisearch.Meilisearch
	db     database.Engine
	col    database.Executor
	index  *meili.Index
	fields map[string]string
}

func New(
	ctx context.Context,
	cfg *config.Bridge,
	db database.Engine,
	meili meilisearch.Meilisearch,
) (*Bridge, error) {
	return newBridge(ctx, cfg, db, meili)
}

func newBridge(
	ctx context.Context,
	cfg *config.Bridge,
	db database.Engine,
	meili meilisearch.Meilisearch,
) (*Bridge, error) {
	idx := meili.Index()
}

func (s *Bridge) Sync() {}

func (s *Bridge) BulkSync() {

}
