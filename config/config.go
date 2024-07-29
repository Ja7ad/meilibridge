package config

import (
	"os"

	"gopkg.in/yaml.v3"
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

	if c.Bridges == nil {
		return ErrMissingBridgeConfig
	}

	for _, bridge := range c.Bridges {
		if bridge.IndexMap == nil {
			return ErrIndexMapRequire
		}

		if bridge.Source == nil {
			return ErrMissingSourceConfig
		}

		if bridge.Source.Host == "" {
			return ErrDatabaseHostIsRequired
		}

		if bridge.Source.Port < 1 || bridge.Source.Port > 65535 {
			return ErrDatabasePortIsRequired
		}

		if bridge.Source.Database == "" {
			return ErrSourceDatabaseRequire
		}

		switch bridge.Source.Engine {
		case MONGO:
		case MYSQL, POSTGRES:
		default:
			return ErrNotSupportedEngine
		}

		for collection, index := range bridge.IndexMap {
			if collection == "" {
				return ErrCollectionNameRequire
			}

			if index == nil {
				return ErrBridgeDestinationRequire
			}

			if index.IndexName == "" {
				return ErrIndexNameRequire
			}

			if index.PrimaryKey == "" {
				return ErrPrimaryKeyIsRequire
			}

			if index.Fields != nil {
				pk, ok := index.Fields[index.PrimaryKey]
				if !ok {
					break
				}

				if pk != "" && index.PrimaryKey != pk {
					return ErrInvalidPrimaryKey
				}
			}
		}

	}

	return nil
}
