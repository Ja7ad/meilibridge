package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_New(t *testing.T) {
	cfg, err := New("../config.example.yml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError error
	}{
		{
			name: "valid config",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Source: &Source{
					Engine:   "mongo",
					URI:      "mongodb://localhost:27017",
					Database: "mydb",
				},
				Bridge: &Bridge{
					Collection: "mycollection",
					IndexName:  "myindex",
					PrimaryKeys: []string{
						"id",
					},
					Fields: map[string]string{
						"foo": "foo",
						"bar": "",
					},
				},
			},
			wantError: nil,
		},
		{
			name: "missing meilisearch",
			config: &Config{
				Meilisearch: nil,
				Source: &Source{
					Engine:   "mongo",
					URI:      "mongodb://localhost:27017",
					Database: "mydb",
				},
				Bridge: &Bridge{
					Collection: "mycollection",
					IndexName:  "myindex",
					PrimaryKeys: []string{
						"id",
					},
				},
			},
			wantError: ErrMissingMeilisearchConfig,
		},
		{
			name: "missing meilisearch APIURL",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "",
					APIKey: "masterKey",
				},
				Source: &Source{
					Engine:   "mongo",
					URI:      "mongodb://localhost:27017",
					Database: "mydb",
				},
				Bridge: &Bridge{
					Collection: "mycollection",
					IndexName:  "myindex",
					PrimaryKeys: []string{
						"id",
					},
				},
			},
			wantError: ErrAPIUrlRequire,
		},
		{
			name: "missing source",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Source: nil,
				Bridge: &Bridge{
					Collection: "mycollection",
					IndexName:  "myindex",
					PrimaryKeys: []string{
						"id",
					},
				},
			},
			wantError: ErrMissingSourceConfig,
		},
		{
			name: "unsupported source engine",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Source: &Source{
					Engine:   "unsupported",
					URI:      "mongodb://localhost:27017",
					Database: "mydb",
				},
				Bridge: &Bridge{
					Collection: "mycollection",
					IndexName:  "myindex",
					PrimaryKeys: []string{
						"id",
					},
				},
			},
			wantError: ErrNotSupportedEngine,
		},
		{
			name: "missing source URI",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Source: &Source{
					Engine:   "mongo",
					URI:      "",
					Database: "mydb",
				},
				Bridge: &Bridge{
					Collection: "mycollection",
					IndexName:  "myindex",
					PrimaryKeys: []string{
						"id",
					},
				},
			},
			wantError: ErrSourceURIRequire,
		},
		{
			name: "missing source database",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Source: &Source{
					Engine:   "mongo",
					URI:      "mongodb://localhost:27017",
					Database: "",
				},
				Bridge: &Bridge{
					Collection: "mycollection",
					IndexName:  "myindex",
					PrimaryKeys: []string{
						"id",
					},
				},
			},
			wantError: ErrSourceDatabaseRequire,
		},
		{
			name: "missing bridge",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Source: &Source{
					Engine:   "mongo",
					URI:      "mongodb://localhost:27017",
					Database: "mydb",
				},
				Bridge: nil,
			},
			wantError: ErrMissingBridgeConfig,
		},
		{
			name: "missing bridge collection",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Source: &Source{
					Engine:   "mongo",
					URI:      "mongodb://localhost:27017",
					Database: "mydb",
				},
				Bridge: &Bridge{
					Collection: "",
					IndexName:  "myindex",
					PrimaryKeys: []string{
						"id",
					},
				},
			},
			wantError: ErrCollectionRequire,
		},
		{
			name: "missing bridge index name",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Source: &Source{
					Engine:   "mongo",
					URI:      "mongodb://localhost:27017",
					Database: "mydb",
				},
				Bridge: &Bridge{
					Collection: "mycollection",
					IndexName:  "",
					PrimaryKeys: []string{
						"id",
					},
				},
			},
			wantError: ErrIndexNameRequire,
		},
		{
			name: "missing bridge primary keys",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Source: &Source{
					Engine:   "mongo",
					URI:      "mongodb://localhost:27017",
					Database: "mydb",
				},
				Bridge: &Bridge{
					Collection:  "mycollection",
					IndexName:   "myindex",
					PrimaryKeys: []string{},
				},
			},
			wantError: ErrMissingBridgeFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.Equal(t, tt.wantError, err)
		})
	}
}
