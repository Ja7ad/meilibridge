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
	err := AddEngine(context.Background(), config.MONGO, os.Getenv("MONGO_URI"), testDBName, logger.DefaultLogger)
	require.Nil(t, err)

	mgoEngine := GetEngine[MongoExecutor](config.MONGO)
	require.NotNil(t, mgoEngine)
}
