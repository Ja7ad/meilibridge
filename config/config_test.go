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
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Source: &Source{
							Engine:   "mongo",
							URI:      "mongodb://localhost:27017",
							Database: "mydb",
						},
						IndexMap: map[Collection]*Destination{
							"col1": {
								IndexName:  "idx1",
								PrimaryKey: "id",
								Fields: map[string]string{
									"foo": "foo",
									"bar": "",
								},
							},
							"col2": {
								IndexName:  "idx1",
								PrimaryKey: "id",
							},
						},
					},
				},
			},
			wantError: nil,
		},
		{
			name: "missing meilisearch",
			config: &Config{
				Meilisearch: nil,
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Source: &Source{
							Engine:   "mongo",
							URI:      "mongodb://localhost:27017",
							Database: "mydb",
						},
						IndexMap: map[Collection]*Destination{
							"col1": {
								IndexName:  "idx1",
								PrimaryKey: "id",
								Fields: map[string]string{
									"foo": "foo",
									"bar": "",
								},
							},
							"col2": {
								IndexName: "idx1",
							},
						},
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
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Source: &Source{
							Engine:   "mongo",
							URI:      "mongodb://localhost:27017",
							Database: "mydb",
						},
						IndexMap: map[Collection]*Destination{
							"col1": {
								IndexName:  "idx1",
								PrimaryKey: "id",
								Fields: map[string]string{
									"foo": "foo",
									"bar": "",
								},
							},
							"col2": {
								IndexName: "idx1",
							},
						},
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
				Bridges: []*Bridge{
					{
						Name:   "bridge1",
						Source: nil,
						IndexMap: map[Collection]*Destination{
							"col1": {
								IndexName:  "idx1",
								PrimaryKey: "id",
								Fields: map[string]string{
									"foo": "foo",
									"bar": "",
								},
							},
							"col2": {
								IndexName: "idx1",
							},
						},
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
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Source: &Source{
							Engine:   "unsupported",
							URI:      "mongodb://localhost:27017",
							Database: "mydb",
						},
						IndexMap: map[Collection]*Destination{
							"col1": {
								IndexName:  "idx1",
								PrimaryKey: "id",
								Fields: map[string]string{
									"foo": "foo",
									"bar": "",
								},
							},
							"col2": {
								IndexName: "idx1",
							},
						},
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
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Source: &Source{
							Engine:   "mongo",
							Database: "mydb",
						},
						IndexMap: map[Collection]*Destination{
							"col1": {
								IndexName:  "idx1",
								PrimaryKey: "id",
								Fields: map[string]string{
									"foo": "foo",
									"bar": "",
								},
							},
							"col2": {
								IndexName: "idx1",
							},
						},
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
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Source: &Source{
							Engine: "mongo",
							URI:    "mongodb://localhost:27017",
						},
						IndexMap: map[Collection]*Destination{
							"col1": {
								IndexName:  "idx1",
								PrimaryKey: "id",
								Fields: map[string]string{
									"foo": "foo",
									"bar": "",
								},
							},
							"col2": {
								IndexName: "idx1",
							},
						},
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
				Bridges: nil,
			},
			wantError: ErrMissingBridgeConfig,
		},
		{
			name: "missing bridge index map",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Bridges: nil,
			},
			wantError: ErrMissingBridgeConfig,
		},
		{
			name: "missing primary key",
			config: &Config{
				Meilisearch: &Meilisearch{
					APIURL: "http://localhost:7700",
					APIKey: "masterKey",
				},
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Source: &Source{
							Engine: "mongo",
							URI:    "mongodb://localhost:27017",
						},
						IndexMap: map[Collection]*Destination{
							"col1": {
								IndexName: "idx1",
								Fields: map[string]string{
									"foo": "foo",
									"bar": "",
								},
							},
							"col2": {
								IndexName: "idx1",
							},
						},
					},
				},
			},
			wantError: ErrSourceDatabaseRequire,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.Equal(t, tt.wantError, err)
		})
	}
}
