package config

import "strings"

type Config struct {
	General *General  `yaml:"general"`
	Bridges []*Bridge `yaml:"bridges"`
}

type General struct {
	AutoBulkInterval int64  `yaml:"auto_bulk_interval"`
	PProf            *PProf `yaml:"pprof"`
}

type PProf struct {
	Enable bool   `yaml:"enable"`
	Listen string `yaml:"listen"`
}

type Bridge struct {
	Name        string                      `yaml:"name"`
	Meilisearch *Meilisearch                `yaml:"meilisearch"`
	Database    *Database                   `yaml:"database"`
	IndexMap    map[Collection]*IndexConfig `yaml:"index_map"`
}

type Meilisearch struct {
	APIURL string `yaml:"api_url"`
	APIKey string `yaml:"api_key"`
}

type Database struct {
	Engine       Engine                 `yaml:"engine"`
	Host         string                 `yaml:"host"`
	Port         uint16                 `yaml:"port"`
	User         string                 `yaml:"user"`
	Password     string                 `yaml:"password"`
	Database     string                 `yaml:"database"`
	CustomParams map[string]interface{} `yaml:"custom_params"`
}

type IndexConfig struct {
	IndexName  string            `yaml:"index_name"`
	PrimaryKey string            `yaml:"primary_key"`
	Fields     map[string]string `yaml:"fields"`
	Settings   *Settings         `yaml:"settings"`
}

type Settings struct {
	RankingRules         []string            `json:"rankingRules,omitempty" yaml:"ranking_rules"`
	DistinctAttribute    *string             `json:"distinctAttribute,omitempty" yaml:"distinct_attribute"`
	SearchableAttributes []string            `json:"searchableAttributes,omitempty" yaml:"searchable_attributes"`
	DisplayedAttributes  []string            `json:"displayedAttributes,omitempty" yaml:"displayed_attributes"`
	StopWords            []string            `json:"stopWords,omitempty" yaml:"stop_words"`
	Synonyms             map[string][]string `json:"synonyms,omitempty" yaml:"synonyms"`
	FilterableAttributes []string            `json:"filterableAttributes,omitempty" yaml:"filterable_attributes"`
	SortableAttributes   []string            `json:"sortableAttributes,omitempty" yaml:"sortable_attributes"`
	TypoTolerance        *TypoTolerance      `json:"typoTolerance,omitempty" yaml:"typo_tolerance"`
	Pagination           *Pagination         `json:"pagination,omitempty" yaml:"pagination"`
	Faceting             *Faceting           `json:"faceting,omitempty" yaml:"faceting"`
	Embedders            map[string]Embedder `json:"embedders,omitempty" yaml:"embedders"`
}

type TypoTolerance struct {
	Enabled             bool                `json:"enabled,omitempty" yaml:"enabled"`
	MinWordSizeForTypos MinWordSizeForTypos `json:"minWordSizeForTypos,omitempty" yaml:"min_word_size_for_typos"`
	DisableOnWords      []string            `json:"disableOnWords,omitempty" yaml:"disable_on_words"`
	DisableOnAttributes []string            `json:"disableOnAttributes,omitempty" yaml:"disable_on_attributes"`
}

type MinWordSizeForTypos struct {
	OneTypo  int64 `json:"oneTypo,omitempty" yaml:"one_typo"`
	TwoTypos int64 `json:"twoTypos,omitempty" yaml:"two_typos"`
}

type Pagination struct {
	MaxTotalHits int64 `json:"maxTotalHits" yaml:"max_total_hits"`
}

type Faceting struct {
	MaxValuesPerFacet int64 `json:"maxValuesPerFacet" yaml:"max_values_per_facet"`
}

type Embedder struct {
	Source           string `json:"source" yaml:"source"`
	ApiKey           string `json:"apiKey,omitempty" yaml:"api_key"`
	Model            string `json:"model,omitempty" yaml:"model"`
	Dimensions       int    `json:"dimensions,omitempty" yaml:"dimensions"`
	DocumentTemplate string `json:"documentTemplate,omitempty" yaml:"document_template"`
}

type (
	Engine     string
	Collection string
	Index      string
)

const (
	MONGO    Engine = "mongo"
	MYSQL    Engine = "mysql"
	POSTGRES Engine = "postgres"
	PLUGIN   Engine = "plugin"
)

func (e Engine) String() string { return string(e) }

func (c Collection) String() string { return string(c) }

func (c Collection) GetCollectionAndView() (col string, view string) {
	items := strings.Split(c.String(), ":")
	if len(items) > 1 {
		return items[0], items[1]
	}
	return "", ""
}

func (c Collection) HasView() bool {
	return len(strings.Split(c.String(), ":")) == 2
}

func (c Collection) GetView() string {
	items := strings.Split(c.String(), ":")
	if len(items) == 2 {
		return items[1]
	}
	return ""
}

func (i Index) String() string { return string(i) }
