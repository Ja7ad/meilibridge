package cfg

import (
	"gopkg.in/yaml.v3"
	"os"
)

func New(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := new(Config)

	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, ErrDecodeConfig
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Meilisearch == nil {
		return ErrMissingMeilisearchConfig
	}
	if c.Meilisearch.APIURL == "" {
		return ErrAPIUrlRequire
	}

	if c.Source == nil {
		return ErrMissingSourceConfig
	}

	switch c.Source.Engine {
	case MONGO.String():
	case MYSQL.String(), POSTGRES.String():
		// TODO: currently not supported mysql and postgres
		return ErrNotSupportedEngine
	default:
		return ErrNotSupportedEngine
	}

	if c.Source.URI == "" {
		return ErrSourceURIRequire
	}
	if c.Source.Database == "" {
		return ErrSourceDatabaseRequire
	}

	if c.Bridge == nil {
		return ErrMissingBridgeConfig
	}
	if c.Bridge.Collection == "" {
		return ErrCollectionRequire
	}
	if c.Bridge.IndexName == "" {
		return ErrIndexNameRequire
	}
	if len(c.Bridge.Fields) == 0 {
		return ErrMissingBridgeFields
	}

	return nil
}
