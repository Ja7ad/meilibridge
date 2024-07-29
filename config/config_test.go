package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Meilisearch: &Meilisearch{
							APIURL: "http://localhost:7700",
							APIKey: "masterKey",
						},
						Database: &Database{
							Engine:       "mongo",
							Host:         "127.0.0.1",
							Port:         27017,
							User:         "root",
							Password:     "foobar",
							Database:     "mydb",
							CustomParams: make(map[string]interface{}),
						},
						IndexMap: map[Collection]*IndexConfig{
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
				Bridges: []*Bridge{
					{
						Name:        "bridge1",
						Meilisearch: nil,
						Database: &Database{
							Engine:       "mongo",
							Host:         "127.0.0.1",
							Port:         27017,
							User:         "root",
							Password:     "foobar",
							Database:     "mydb",
							CustomParams: make(map[string]interface{}),
						},
						IndexMap: map[Collection]*IndexConfig{
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
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Meilisearch: &Meilisearch{
							APIURL: "",
							APIKey: "masterKey",
						},
						Database: &Database{
							Engine:       "mongo",
							Host:         "127.0.0.1",
							Port:         27017,
							User:         "root",
							Password:     "foobar",
							Database:     "mydb",
							CustomParams: make(map[string]interface{}),
						},
						IndexMap: map[Collection]*IndexConfig{
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
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Meilisearch: &Meilisearch{
							APIURL: "http://localhost:7700",
							APIKey: "masterKey",
						},
						Database: nil,
						IndexMap: map[Collection]*IndexConfig{
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
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Meilisearch: &Meilisearch{
							APIURL: "http://localhost:7700",
							APIKey: "masterKey",
						},
						Database: &Database{
							Engine:       "unsupported",
							Host:         "127.0.0.1",
							Port:         27017,
							User:         "root",
							Password:     "foobar",
							Database:     "mydb",
							CustomParams: make(map[string]interface{}),
						},
						IndexMap: map[Collection]*IndexConfig{
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
			name: "missing database host",
			config: &Config{
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Meilisearch: &Meilisearch{
							APIURL: "http://localhost:7700",
							APIKey: "masterKey",
						},
						Database: &Database{
							Engine:   "mongo",
							Database: "mydb",
						},
						IndexMap: map[Collection]*IndexConfig{
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
			wantError: ErrDatabaseHostIsRequired,
		},
		{
			name: "missing source database",
			config: &Config{
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Meilisearch: &Meilisearch{
							APIURL: "http://localhost:7700",
							APIKey: "masterKey",
						},
						Database: &Database{
							Engine:       "mongo",
							Host:         "127.0.0.1",
							Port:         27017,
							User:         "root",
							Password:     "foobar",
							CustomParams: make(map[string]interface{}),
						},
						IndexMap: map[Collection]*IndexConfig{
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
			name: "missing source port",
			config: &Config{
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Meilisearch: &Meilisearch{
							APIURL: "http://localhost:7700",
							APIKey: "masterKey",
						},
						Database: &Database{
							Engine:       "mongo",
							Host:         "127.0.0.1",
							User:         "root",
							Password:     "foobar",
							CustomParams: make(map[string]interface{}),
						},
						IndexMap: map[Collection]*IndexConfig{
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
			wantError: ErrDatabasePortIsRequired,
		},
		{
			name: "missing bridge",
			config: &Config{
				Bridges: nil,
			},
			wantError: ErrMissingBridgeConfig,
		},
		{
			name: "missing bridge index map",
			config: &Config{
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Meilisearch: &Meilisearch{
							APIURL: "http://localhost:7700",
							APIKey: "masterKey",
						},
						Database: &Database{
							Engine:       "mongo",
							Host:         "127.0.0.1",
							User:         "root",
							Password:     "foobar",
							CustomParams: make(map[string]interface{}),
						},
						IndexMap: nil,
					},
				},
			},
			wantError: ErrIndexMapRequire,
		},
		{
			name: "missing primary key",
			config: &Config{
				Bridges: []*Bridge{
					{
						Name: "bridge1",
						Meilisearch: &Meilisearch{
							APIURL: "http://localhost:7700",
							APIKey: "masterKey",
						},
						Database: &Database{
							Engine:       "mongo",
							Host:         "127.0.0.1",
							Port:         27017,
							User:         "root",
							Password:     "foobar",
							Database:     "mydb",
							CustomParams: make(map[string]interface{}),
						},
						IndexMap: map[Collection]*IndexConfig{
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
			wantError: ErrPrimaryKeyIsRequire,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.Equal(t, tt.wantError, err)
		})
	}
}

func TestCollection_GetCollectionAndView(t *testing.T) {
	tests := []struct {
		input    Collection
		wantCol  string
		wantView string
	}{
		{"collection:view", "collection", "view"},
		{"singleitem", "", ""},
		{"anothercollection:anotherview", "anothercollection", "anotherview"},
		{"justcollection:", "justcollection", ""},
		{":justview", "", "justview"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			col, view := tt.input.GetCollectionAndView()
			if col != tt.wantCol {
				t.Errorf("GetCollectionAndView() col = %v, want %v", col, tt.wantCol)
			}
			if view != tt.wantView {
				t.Errorf("GetCollectionAndView() view = %v, want %v", view, tt.wantView)
			}
		})
	}
}

func TestCollection_HasView(t *testing.T) {
	tests := []struct {
		input Collection
		want  bool
	}{
		{"collection:view", true},
		{"singleitem", false},
		{"anothercollection:anotherview", true},
		{"justcollection:", true},
		{":justview", true},
		{"no:view:here", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			if got := tt.input.HasView(); got != tt.want {
				t.Errorf("HasView() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollection_GetView(t *testing.T) {
	tests := []struct {
		input    Collection
		wantView string
	}{
		{"collection:view", "view"},
		{"singleitem", ""},
		{"anothercollection:anotherview", "anotherview"},
		{"justcollection:", ""},
		{":justview", "justview"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			if got := tt.input.GetView(); got != tt.wantView {
				t.Errorf("GetView() = %v, want %v", got, tt.wantView)
			}
		})
	}
}
