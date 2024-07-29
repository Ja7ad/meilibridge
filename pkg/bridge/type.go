package bridge

import (
	"context"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/Ja7ad/meilibridge/pkg/meilisearch"
)

const _bulkLimit = int64(100)

type Bridge struct {
	meili   meilisearch.Meilisearch
	bridges []*config.Bridge
	log     logger.Logger
}

type stat struct {
	col, index     string
	indexed, total int64
	err            error
}

type task struct {
	col config.Collection
	des *config.Destination
}

type Syncer interface {
	Name() string
	OnDemand(ctx context.Context)
	Bulk(ctx context.Context, isContinue bool)
}
