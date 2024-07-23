package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name        string
		handler     HandleType
		options     Options
		level       slog.Level
		method      func(logger Logger, msg string, args ...any)
		methodCtx   func(logger Logger, ctx context.Context, msg string, args ...any)
		expectedMsg string
	}{
		{
			name:        "debug log",
			handler:     JSON_HANDLER,
			options:     Options{Debug: true, EnableCaller: true},
			level:       slog.LevelDebug,
			method:      Logger.Debug,
			methodCtx:   Logger.DebugContext,
			expectedMsg: "debug message",
		},
		{
			name:        "info log",
			handler:     JSON_HANDLER,
			options:     Options{EnableCaller: true},
			level:       slog.LevelInfo,
			method:      Logger.Info,
			methodCtx:   Logger.InfoContext,
			expectedMsg: "info message",
		},
		{
			name:        "warn log",
			handler:     JSON_HANDLER,
			options:     Options{EnableCaller: true},
			level:       slog.LevelWarn,
			method:      Logger.Warn,
			methodCtx:   Logger.WarnContext,
			expectedMsg: "warn message",
		},
		{
			name:        "error log",
			handler:     JSON_HANDLER,
			options:     Options{EnableCaller: true},
			level:       slog.LevelError,
			method:      Logger.Error,
			methodCtx:   Logger.ErrorContext,
			expectedMsg: "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.options.CustomJsonHandler = slog.NewJSONHandler(&buf, nil)

			logger := New(tt.handler, tt.options)
			assert.NotNil(t, logger)

			tt.method(logger, tt.expectedMsg)
			assert.Contains(t, buf.String(), tt.expectedMsg)

			buf.Reset()

			ctx := context.Background()
			tt.methodCtx(logger, ctx, tt.expectedMsg)
			assert.Contains(t, buf.String(), tt.expectedMsg)
		})
	}
}
