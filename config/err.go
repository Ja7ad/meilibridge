package config

import "errors"

var (
	ErrDecodeConfig             = errors.New("failed to decode config file")
	ErrMissingMeilisearchConfig = errors.New("meilisearch configuration is missing")
	ErrAPIUrlRequire            = errors.New("meilisearch api_url is required")
	ErrMissingSourceConfig      = errors.New("source configuration is missing")
	ErrNotSupportedEngine       = errors.New("engine is not supported")
	ErrSourceDatabaseRequire    = errors.New("source database is required")
	ErrMissingBridgeConfig      = errors.New("bridge configuration is missing")
	ErrIndexMapRequire          = errors.New("bridge index map is required")
	ErrIndexNameRequire         = errors.New("bridge index is required")
	ErrCollectionNameRequire    = errors.New("bridge collection is required")
	ErrBridgeDestinationRequire = errors.New("bridge destination is required")
	ErrInvalidPrimaryKey        = errors.New("don't match primary key with field value map")
	ErrPrimaryKeyIsRequire      = errors.New("primary key is required")
	ErrDatabaseHostIsRequired   = errors.New("database host is required")
	ErrDatabasePortIsRequired   = errors.New("database port is required")
)
