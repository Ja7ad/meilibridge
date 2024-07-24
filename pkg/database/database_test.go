package database

import (
	"context"
	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_Engine(t *testing.T) {
	err := AddEngine(context.Background(), config.MONGO, os.Getenv("MONGO_URI"), testDBName, logger.DefaultLogger)
	require.Nil(t, err)

	mgoEngine := GetEngine[MongoExecutor](config.MONGO)
	require.NotNil(t, mgoEngine)
}
