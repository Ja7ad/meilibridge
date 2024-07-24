package config

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
	case MONGO:
	case MYSQL, POSTGRES:
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

	if len(c.Bridges) == 0 {
		return ErrMissingBridgeConfig
	}

	for _, bridge := range c.Bridges {
		if bridge == nil {
			return ErrMissingBridgeConfig
		}
		if bridge.Collection == "" {
			return ErrCollectionRequire
		}
		if bridge.IndexName == "" {
			return ErrIndexNameRequire
		}
		if bridge.Settings != nil && bridge.Fields != nil {
			pk, ok := bridge.Fields[bridge.Settings.PrimaryKey]
			if !ok {
				break
			}

			if pk != "" && bridge.Settings.PrimaryKey != pk {
				return ErrInvalidPrimaryKey
			}
		}
	}

	return nil
}
