package cfg

type Config struct {
	Meilisearch *Meilisearch `yaml:"meilisearch"`
	Source      *Source      `yaml:"source"`
	Bridge      *Bridge      `yaml:"bridge"`
}

type Meilisearch struct {
	APIURL string `yaml:"api_url"`
	APIKey string `yaml:"api_key"`
}

type Source struct {
	Engine   string `yaml:"engine"`
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
}

type Bridge struct {
	Collection  string            `yaml:"collection"`
	IndexName   string            `yaml:"index_name"`
	PrimaryKeys []string          `yaml:"primary_keys"`
	Fields      map[string]string `yaml:"fields"`
	Settings    *BridgeSettings   `yaml:"settings"`
}

type BridgeSettings struct {
	Dictionary           []string            `yaml:"dictionary"`
	DisplayedAttributes  []string            `yaml:"displayed_attributes"`
	Faceting             *Faceting           `yaml:"faceting"`
	FilterableAttributes []string            `yaml:"filterable_attributes"`
	SearchableAttributes []string            `yaml:"searchable_attributes"`
	SortableAttributes   []string            `yaml:"sortable_attributes"`
	Pagination           *Pagination         `yaml:"pagination"`
	ProximityPrecision   string              `yaml:"proximity_precision"`
	RankingRules         []string            `yaml:"ranking_rules"`
	SearchCutoffMs       int                 `yaml:"search_cutoff_ms"`
	SeparatorTokens      []string            `yaml:"separator_tokens"`
	NonSeparatorTokens   []string            `yaml:"non_separator_tokens"`
	StopWords            []string            `yaml:"stop_words"`
	Synonyms             map[string][]string `yaml:"synonyms"`
	TypoTolerance        *TypoTolerance      `yaml:"typo_tolerance"`
}

type Faceting struct {
	MaxValuesPerFacet int               `yaml:"max_values_perFacet"`
	SortFacetValuesBy map[string]string `yaml:"sort_facet_values_by"`
}

type Pagination struct {
	MaxTotalHits int `yaml:"max_total_hits"`
}

type TypoTolerance struct {
	Enabled             bool `yaml:"enabled"`
	MinWordSizeForTypos struct {
		OneTypo  int `yaml:"one_typo"`
		TwoTypos int `yaml:"two_typos"`
	} `yaml:"min_word_size_for_typos"`
	DisableOnWords      []string `yaml:"disable_on_words"`
	DisableOnAttributes []string `yaml:"disable_on_attributes"`
}

type Engine string

const (
	MONGO    Engine = "mongo"
	MYSQL    Engine = "mysql"
	POSTGRES Engine = "postgres"
)

func (e Engine) String() string {
	return string(e)
}
