# meilibridge

`meilibridge` is a robust package designed to seamlessly sync data from both SQL and NoSQL databases to Meilisearch, 
providing an efficient and unified search solution.

## Features

- Supports multiple data sources
- Compatible with various databases such as MongoDB (currently supported), MySQL, and PostgreSQL
- Index configuration options
- Real-time synchronization
- Bulk sync support with options to continue or reindex
- Concurrent data bridging to Meilisearch
- Customizable fields for indexing
- Set primary key for index

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

```yaml
version: "3"
services:
  meilibridge:
    image: ja7adr/meilibridge:latest
    volumes:
      - ./config.yml:/etc/meilibridge/config.yml
    restart: always
    command:
      - sync start
```

## Usage

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

## TODO

- [x] MongoDB engine
- [x] Bulk sync
- [x] Bulk sync resync with continue
- [x] Real-time sync
- [ ] MySQL engine
- [ ] PostgreSQL engine
