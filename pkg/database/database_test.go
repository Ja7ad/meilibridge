package database

import (
	"context"
	"os"
	"testing"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/stretchr/testify/require"
)

func Test_Engine(t *testing.T) {
	err := AddEngine(context.Background(), &config.Database{
		Engine:   config.MONGO,
		Host:     os.Getenv("MONGO_HOST"),
		Database: os.Getenv("MONGO_DATABASE"),
	}, logger.DefaultLogger)
	require.Nil(t, err)

	mgoEngine := GetEngine[MongoExecutor](config.MONGO)
	require.NotNil(t, mgoEngine)
}
