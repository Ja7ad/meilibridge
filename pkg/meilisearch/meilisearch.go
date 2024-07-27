package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	meili "github.com/meilisearch/meilisearch-go"
)

const _defaultWaitInterval = 5 * time.Second

type meilisearch struct {
	apiURL, apiKey string
	cli            *meili.Client
	isHealthy      bool
	log            logger.Logger
}

type Meilisearch interface {
	CreateIndex(ctx context.Context, uid, primaryKey string) error
	Index(uid string) *meili.Index
	GetIndex(uid string) (*meili.Index, error)
	IsExistsIndex(uid string) bool
	DeleteIndex(ctx context.Context, uid string) error
	UpdateIndexSettings(ctx context.Context, uid string, settings *config.Settings) error
	WaitForTask(ctx context.Context, task *meili.TaskInfo) error
	Stats() *meili.Stats
	Version() string
}

func New(ctx context.Context, apiURL, apiKey string, log logger.Logger) (Meilisearch, error) {
	cli := meili.NewClient(meili.ClientConfig{
		Host:   apiURL,
		APIKey: apiKey,
	})

	m := &meilisearch{
		log:       log,
		apiURL:    apiURL,
		apiKey:    apiKey,
		isHealthy: true,
	}

	if !cli.IsHealthy() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			t := time.NewTicker(_defaultWaitInterval)
			defer t.Stop()
			<-t.C
			log.Error("meilisearch is unhealthy, trying to reconnection...")
			return New(ctx, apiURL, apiKey, log)
		}
	}

	m.cli = cli

	go m.healthyCheck(ctx)

	return m, nil
}

func (m *meilisearch) Index(uid string) *meili.Index {
	return m.cli.Index(uid)
}

func (m *meilisearch) GetIndex(uid string) (*meili.Index, error) {
	if !m.isHealthy {
		return nil, ErrMeilisearchIsUnhealthy
	}

	idx, err := m.cli.GetIndex(uid)
	if err != nil {
		return nil, ErrIndexNotFound
	}

	return idx, nil
}

func (m *meilisearch) CreateIndex(ctx context.Context, uid, primaryKey string) error {
	if !m.isHealthy {
		return ErrMeilisearchIsUnhealthy
	}

	idxCfg := &meili.IndexConfig{
		Uid: uid,
	}

	if primaryKey != "" {
		idxCfg.PrimaryKey = primaryKey
	}

	t, err := m.cli.CreateIndex(idxCfg)
	if err != nil {
		return err
	}

	return m.WaitForTask(ctx, t)
}

func (m *meilisearch) IsExistsIndex(uid string) bool {
	_, err := m.cli.GetIndex(uid)
	return err == nil
}

func (m *meilisearch) DeleteIndex(ctx context.Context, uid string) error {
	if !m.isHealthy {
		return ErrMeilisearchIsUnhealthy
	}

	t, err := m.cli.DeleteIndex(uid)
	if err != nil {
		return err
	}

	return m.WaitForTask(ctx, t)
}

func (m *meilisearch) UpdateIndexSettings(ctx context.Context, uid string, settings *config.Settings) error {
	idx, err := m.cli.GetIndex(uid)
	if err != nil {
		return ErrIndexNotFound
	}

	if settings == nil {
		return nil
	}

	resT, err := idx.ResetSettings()
	if err != nil {
		return err
	}

	if err := m.WaitForTask(ctx, resT); err != nil {
		return err
	}

	var s meili.Settings

	b, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	t, err := idx.UpdateSettings(&s)
	if err != nil {
		return ErrUpdateSettings
	}

	return m.WaitForTask(ctx, t)
}

func (m *meilisearch) Stats() *meili.Stats {
	if !m.isHealthy {
		m.log.Warn("meilisearch is unhealthy")
		return nil
	}

	s, err := m.cli.GetStats()
	if err != nil {
		m.log.Error("failed to get meilisearch stats", "err", err)
		return nil
	}

	return s
}

func (m *meilisearch) Version() string {
	if !m.isHealthy {
		m.log.Warn("meilisearch is unhealthy")
		return ""
	}

	ver, err := m.cli.Version()
	if err == nil && ver != nil {
		return ver.PkgVersion
	}

	return ""
}

func (m *meilisearch) WaitForTask(
	ctx context.Context,
	task *meili.TaskInfo,
) error {
	if task.Status == meili.TaskStatusSucceeded {
		return nil
	}

	t, err := m.cli.WaitForTask(task.TaskUID, meili.WaitParams{
		Context:  ctx,
		Interval: _defaultWaitInterval,
	})
	if err != nil {
		return err
	}

	switch t.Status {
	case meili.TaskStatusSucceeded:
		return nil
	case meili.TaskStatusEnqueued, meili.TaskStatusProcessing:
		return m.WaitForTask(ctx, task)
	case meili.TaskStatusCanceled:
		return ErrTaskCanceled
	case meili.TaskStatusFailed:
		return fmt.Errorf("task %d index %s failed, error %s", t.TaskUID, t.IndexUID, t.Error.Message)
	case meili.TaskStatusUnknown:
		return ErrTaskUnknown
	}

	return nil
}

func (m *meilisearch) healthyCheck(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			if !m.cli.IsHealthy() {
				m.isHealthy = false
			}
			m.isHealthy = true
		}
	}
}
