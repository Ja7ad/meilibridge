package logger

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestHandler_Handle(t *testing.T) {
	logger := slog.New(NewConsoleHandler(&slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
	}))

	assert.NotNil(t, logger)
}
