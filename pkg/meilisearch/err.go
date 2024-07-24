package meilisearch

import "errors"

var (
	ErrMeilisearchIsUnhealthy = errors.New("meilisearch is unhealthy")
	ErrIndexNotFound          = errors.New("index not found")
	ErrTaskCanceled           = errors.New("task canceled")
	ErrTaskFailed             = errors.New("task failed")
	ErrTaskUnknown            = errors.New("task unknown")
	ErrUpdateSettings         = errors.New("update settings failed")
)
