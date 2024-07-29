# meilibridge

`meilibridge` is a robust package designed to seamlessly sync data from both SQL and NoSQL databases to Meilisearch,
providing an efficient and unified search solution.

### Table of content

- [Features](#features)
- [Installation](#installation)
  - [Release](#release)
  - [Go Installation](#go-installation)
  - [Docker](#docker)
- [Configuration](#example-configuration)
- [Usage](#how-to-run)
- [ToDo](#todo)

## Features

- Supports multiple data sources
- Compatible with various databases such as MongoDB (currently supported), MySQL, and PostgreSQL
- Index configuration options
- Real-time synchronization
- Bulk sync support with options to continue or reindex
- Concurrent data bridging to Meilisearch
- Customizable fields for indexing
- Set primary key for index
- Many meilisearch for specific bridge

## Installation

You can install `meilibridge` using different methods.

### Release

Download the latest release from [here](https://github.com/Ja7ad/meilibridge/releases).

### Go Installation

Install the package using Go:

```shell
go install github.com/Ja7ad/meilibridge/cmd/meilibridge@latest
```

### Docker

Run `meilibridge` with real-time sync using Docker:

- Docker

```shell
docker run --rm -it -v ./config.yml:/etc/meilibridge/config.yml --name meilibridge ja7adr/meilibridge
```

- Docker compose

```yaml
version: "3"
services:
  meilibridge:
    image: ja7adr/meilibridge:latest
    volumes:
      - ./config.yml:/etc/meilibridge/config.yml
    restart: always
```

## Example Configuration

example configuration for run meilibridge

```yaml
general:
  pprof:
    enable: false
    listen: 127.0.0.1:9900

bridges:
  - name: bridge 1 # name is required

    meilisearch:
      # API address of meilisearch
      api_url: http://127.0.0.1:7700
      # master key https://www.meilisearch.com/docs/learn/security/differences_master_api_keys#master-key
      # optional
      api_key: foobar

    database:
      # database engine mongo, mysql, postgres
      engine: mongo
      host: "localhost"
      port: 27017
      user: "foo"
      password: "bar"
      database: "foobar"
      # custom parameter for database engine key:val
      custom_params:
        directConnection: true
        replicaSet: test

    # index map is collection or table of data source to meilisearch index
    # source collection or table -> index
    index_map:
      # if you want sync view table should original_name_table:view_name
      col1:col1_view_name:
        index_name: idx1
        # set pk for fields in meilisearch, note if set value for fields please enter value not database key.
        # it's require.
        # for mongodb use field _id for primary key.
        # https://www.meilisearch.com/docs/learn/core_concepts/primary_key#primary-field
        primary_key: id
        fields:
          _id: id
          first_name:
          last_name:
          age:
          created_at:

        settings:
          # list of strings Meilisearch should parse as a single term, default is empty
          # https://www.meilisearch.com/docs/reference/api/settings#dictionary
          dictionary:
            - foo
            - bar

          # he distinct attribute is a special, user-designated field. It is most commonly used to prevent Meilisearch
          # from returning a set of several similar documents, instead forcing it to return only one, default is empty
          # https://www.meilisearch.com/docs/learn/relevancy/distinct_attribute#setting-a-distinct-attribute-during-configuration
          distinct_attribute: foo

          # fields displayed in the returned documents, default is all attributes
          # https://www.meilisearch.com/docs/reference/api/settings#displayed-attributes
          displayed_attributes:
            - foo
            - bar

          # faceting settings
          # https://www.meilisearch.com/docs/reference/api/settings#faceting-object
          faceting:
            # maximum number of facet values returned for each facet. Values are sorted in ascending lexicographical order
            # default is 100
            max_values_per_facet: 100

          # attributes to use as filters and facets, default is empty
          # https://www.meilisearch.com/docs/reference/api/settings#filterable-attributes
          filterable_attributes:
            - first_name
            - last_name

          # fields in which to search for matching query words sorted by order of importance, default is all attributes ["*"]
          # https://www.meilisearch.com/docs/reference/api/settings#searchable-attributes
          searchable_attributes:
            - first_name
            - last_name
            - age

          # attributes to use when sorting search results, default is empty
          # https://www.meilisearch.com/docs/reference/api/settings#sortable-attributes
          sortable_attributes:
            - age

          # pagination settings
          # https://www.meilisearch.com/docs/reference/api/settings#pagination
          pagination:
            # the maximum number of search results Meilisearch can return, default is 1000
            # note: setting maxTotalHits to a value higher than the default will negatively impact search performance.
            # setting maxTotalHits to values over 20000 may result in queries taking seconds to complete.
            max_total_hits: 500

          # precision level when calculating the proximity ranking rule, default is "byWord"
          # https://www.meilisearch.com/docs/reference/api/settings#proximity-precision
          proximity_precision: "byWord"

          # list of ranking rules in order of importance,
          # default is ["words", "typo", "proximity", "attribute", "sort", "exactness"]
          # https://www.meilisearch.com/docs/reference/api/settings#ranking-rules
          ranking_rules:
            - "words"
            - "typo"

          # maximum duration of a search query for null set 0, default is 1500
          # https://www.meilisearch.com/docs/reference/api/settings#search-cutoff
          search_cutoff_ms: 500

          # list of characters delimiting where one term begins and ends, default is empty
          # https://www.meilisearch.com/docs/reference/api/settings#separator-tokens
          separator_tokens:
            - foo
            - bar

          # list of characters not delimiting where one term begins and ends, default is empty
          # https://www.meilisearch.com/docs/reference/api/settings#non-separator-tokens
          non_separator_tokens:
            - foo
            - bar

          # list of words ignored by Meilisearch when present in search queries, default is empty
          # https://www.meilisearch.com/docs/reference/api/settings#stop-words
          stop_words:
            - foo
            - bar

          # list of associated words treated similarly, default is empty
          # https://www.meilisearch.com/docs/reference/api/settings#synonyms
          synonyms:
            wolverine:
              - foo
              - bar
            logan:
              - x
              - y
              - z

          # typo tolerance settings
          # https://www.meilisearch.com/docs/reference/api/settings#typo-tolerance
          typo_tolerance:
            # whether typo tolerance is enabled or not, default is true
            enabled: true

            # the minimum word size for accepting 2 typos; must be between oneTypo and 255, default is 9
            min_word_size_for_typos:
              one_typo: 5
              two_typos: 9

            # an array of words for which the typo tolerance feature is disabled, default is empty
            disable_on_words:
              - foo
              - bar

            # an array of attributes for which the typo tolerance feature is disabled, default is empty
            disable_on_attributes:
              - foo
              - bar

          # embedders translate documents and queries into vector embeddings. You must configure at
          # least one embedder to use AI-powered search, this is experimental.
          # https://www.meilisearch.com/docs/reference/api/settings#embedders-experimental
          embedders:
            embedder1:
              source: source1
              api_key: apikey1
              model: model1
              dimensions: 128
              document_template: template1

            embedder2:
              source: source2
              api_key: apikey2
              model: model2
              dimensions: 128
              document_template: template2

      col2:
        index_name: idx2
        primary_key: id
        fields:
        settings:

  - name: bridge 2

    meilisearch:
      api_url: http://127.0.0.1:7700
      api_key: foobar

    database:
      engine: mysql
      host: "localhost"
      port: 6315
      user: "foo"
      password: "bar"
      database: "foobar"

    index_map:
      col1:
        index_name: idx1
        primary_key: id
        fields:
        settings:
```

## How to run?

```shell
$ meilibridge -h
Meilibridge is a robust package designed to seamlessly sync data from both SQL and NoSQL databases to Meilisearch,
providing an efficient and unified search solution.

Usage:
  meilibridge [command]

Available Commands:
  help        Help about any command
  sync        Bulk or real-time sync
  version     Print the version number

Flags:
  -h, --help   Help for meilibridge

Use "meilibridge [command] --help" for more information about a command.
```

### Bulk Sync

Bulk sync recreates the index and syncs all data to Meilisearch.

```shell
$ meilibridge sync bulk -h
Start bulk sync operation.

Usage:
  meilibridge sync bulk [flags]

Flags:
  -c, --config string   Path to config file (default "/etc/meilibridge/config.yml")
      --continue        Sync new data on existing index
  -h, --help            Help for bulk
```

Example:

```shell
$ meilibridge sync bulk -c ./config.yml
```

### Bulk Sync with Continue

Bulk sync continues to sync new data to Meilisearch on an existing index.

```shell
$ meilibridge sync bulk -h
Start bulk sync operation.

Usage:
  meilibridge sync bulk [flags]

Flags:
  -c, --config string   Path to config file (default "/etc/meilibridge/config.yml")
      --continue        Sync new data on existing index
  -h, --help            Help for bulk
```

Example:

```shell
$ meilibridge sync bulk -c ./config.yml --continue
```

### Real-time Sync

`meilibridge` supports real-time data synchronization on write operations of the database by watching or triggering events.

```shell
$ meilibridge sync start -h
Start real-time sync operation.

Usage:
  meilibridge sync start [flags]

Flags:
  -c, --config string   Path to config file (default "/etc/meilibridge/config.yml")
  -h, --help            Help for start
```

Example:

```shell
$ meilibridge sync start -c ./config.yml
```
