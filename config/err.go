package config

import "errors"

var (
	ErrDecodeConfig             = errors.New("failed to decode config file")
	ErrMissingMeilisearchConfig = errors.New("meilisearch configuration is missing")
	ErrAPIUrlRequire            = errors.New("meilisearch api_url is required")
	ErrMissingSourceConfig      = errors.New("source configuration is missing")
	ErrNotSupportedEngine       = errors.New("engine is not supported")
	ErrSourceURIRequire         = errors.New("source uri is required")
	ErrSourceDatabaseRequire    = errors.New("source database is required")
	ErrMissingBridgeConfig      = errors.New("bridge configuration is missing")
	ErrCollectionRequire        = errors.New("bridge collection is required")
	ErrIndexNameRequire         = errors.New("bridge index_name is required")
	ErrMissingBridgeFields      = errors.New("at least one bridge fields is required")
)
