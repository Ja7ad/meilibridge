package sql

import (
	"context"
	"testing"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/database"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dbName           = "sampledb"
	tableUser        = "user"
	tableBook        = "book"
	viewUserBooks    = "user_books"
	triggerUserTable = "user_log"
)

func setup(t *testing.T, source *config.Source) database.SQLExecutor {
	ctx := context.Background()

	err := database.AddEngine(ctx, source, logger.DefaultLogger)
	require.NoError(t, err)

	sqldb := database.GetEngine[database.SQLExecutor](config.MYSQL)
	require.NotNil(t, sqldb)

	return sqldb
}

func cleanup(closeFunc func() error) func() {
	return func() {
		_ = closeFunc()
	}
}

func Test_Count(t *testing.T) {
	src := &config.Source{
		Engine:   config.MYSQL,
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "foo",
		Password: "bar",
		Database: dbName,
	}

	sq := setup(t, src)
	t.Cleanup(cleanup(sq.Close))

	count, err := sq.Count(context.Background(), viewUserBooks)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, count)
}

func Test_FindOne(t *testing.T) {
	src := &config.Source{
		Engine:   config.MYSQL,
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "foo",
		Password: "bar",
		Database: dbName,
	}

	sq := setup(t, src)
	t.Cleanup(cleanup(sq.Close))

	id := int64(1)

	res, err := sq.FindOne(context.Background(), tableUser, map[string]interface{}{
		"id": id,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Greater(t, len(res), 0)
	assert.Equal(t, id, res["id"])
}

func Test_FindLimit(t *testing.T) {
	src := &config.Source{
		Engine:   config.MYSQL,
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "foo",
		Password: "bar",
		Database: dbName,
	}

	sq := setup(t, src)
	t.Cleanup(cleanup(sq.Close))

	ctx := context.Background()

	total, err := sq.Count(ctx, viewUserBooks)
	require.NoError(t, err)

	limit := 2

	cur, err := sq.FindLimit(ctx, viewUserBooks, int64(limit))
	require.NoError(t, err)
	require.NotNil(t, cur)

	results := make([]*database.Result, 0)

	for cur.Next(ctx) {
		res, err := cur.Result()
		require.NoError(t, err)

		results = append(results, res...)
	}

	require.Equal(t, len(results), int(total))
}
