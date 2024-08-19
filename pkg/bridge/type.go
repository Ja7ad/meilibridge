package bridge

import (
	"context"
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"net/http"
)

const (
	_bulkLimit        = int64(100)
	_triggerHeaderKey = "x-token-key"
)

type Bridge struct {
	bridges    []*config.Bridge
	mux        *http.ServeMux
	triggerCfg *config.TriggerSync
	log        logger.Logger
}

type stat struct {
	col, index     string
	indexed, total int64
	err            error
}

type task struct {
	col config.Collection
	des *config.IndexConfig
}

type Syncer interface {
	Name() string
	OnDemand(ctx context.Context)
	Bulk(ctx context.Context, isContinue bool)
	Trigger() http.HandlerFunc
}
