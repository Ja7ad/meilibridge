meilibridge is a robust package designed to seamlessly sync data from both SQL and NoSQL databases to Meilisearch, 
providing an efficient and unified search solution.

## Features

- Multiple bridge with different data sources
- Multiple database (mongo, mysql, postgres)
- Index configuration
- Real-time sync 
- Bulk sync support with continue or reindex
- concurrent bridge data to meilisearch
- Customize fields for indexing
- Set primary key for index

## Install

you can install or use built.

### Release

Download from [release](https://github.com/Ja7ad/meilibridge/releases)

### Go installation

```shell
go install github.com/ja7ad/meilibridge/cmd/meilibridge@latest
```

### Docker

Run meilibridge with realtime sync

```yaml
version: "3"
services:
  meilibridge:
    image: ja7adr/meilibridge
    volumes:
      - ./config.yml:/etc/meilibridge/config.yml
    restart: always
    command:
      - sync
```

## TODO

- [x] Mongodb engine
- [x] Bulk sync
- [x] Bulk sync resync with continue
- [ ] Real-time sync
- [ ] MySQL engine
- [ ] PostgresSQL engine
