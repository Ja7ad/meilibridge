package config

import (
	"fmt"
	"os"
	"strings"

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
	if c.General == nil {
		c.General = new(General)
	}

	if c.General.AutoBulkInterval < 1 {
		c.General.AutoBulkInterval = 1
	}

	if c.Bridges == nil {
		return ErrMissingBridgeConfig
	}

	for _, bridge := range c.Bridges {
		if bridge.Name == "" {
			return ErrBridgeNameIsRequired
		}

		if strings.Contains(bridge.Name, " ") {
			bridge.Name = strings.Trim(bridge.Name, " ")
			bridge.Name = fmt.Sprintf("%s", strings.Join(strings.Split(bridge.Name, " "), "-"))
		}

		if bridge.Meilisearch == nil {
			return ErrMissingMeilisearchConfig
		}
		if bridge.Meilisearch.APIURL == "" {
			return ErrAPIUrlRequire
		}

		if bridge.IndexMap == nil {
			return ErrIndexMapRequire
		}

		if bridge.Database == nil {
			return ErrMissingSourceConfig
		}

		if bridge.Database.Host == "" {
			return ErrDatabaseHostIsRequired
		}

		if bridge.Database.Port < 1 || bridge.Database.Port > 65535 {
			return ErrDatabasePortIsRequired
		}

		if bridge.Database.Database == "" {
			return ErrSourceDatabaseRequire
		}

		switch bridge.Database.Engine {
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
