package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

type (
	HandleType  uint8
	Environment uint8
)

const (
	CONSOLE_HANDLER HandleType = iota
	TEXT_HANDLER
	JSON_HANDLER
)

const (
	DEVELOPMENT Environment = iota
	PRODUCTION
	RELEASE
)

type Log struct {
	skipCaller int
	slog       *slog.Logger
}

type Options struct {
	Development  bool // Development add development details of machine
	Debug        bool // Debug show debug devel message
	EnableCaller bool // EnableCaller show caller in line code
	// SkipCaller skip caller level of CallerFrames
	// https://github.com/golang/go/issues/59145#issuecomment-1481920720
	SkipCaller        int
	CustomSlog        *slog.Logger      // CustomSlog set custom slog.Logger
	CustomJsonHandler *slog.JSONHandler // CustomJsonHandler set custom slog.JsonHandler
}

type Logger interface {
	// Debug show debug log
	Debug(msg string, args ...any)
	// DebugContext show debug log with passing context
	DebugContext(ctx context.Context, msg string, args ...any)
	// Info show info log
	Info(msg string, args ...any)
	// InfoContext show info log with passing context
	InfoContext(ctx context.Context, msg string, args ...any)
	// Warn show warn log
	Warn(msg string, args ...any)
	// WarnContext show warn log with passing context
	WarnContext(ctx context.Context, msg string, args ...any)
	// Error show error log
	Error(msg string, args ...any)
	// ErrorContext show error log with passing context
	ErrorContext(ctx context.Context, msg string, args ...any)
	// Fatal show error log and exit application with os.Exit(1)
	Fatal(msg string, args ...any)
	// FatalContext show error log with passing context and exit application with os.Exit(1)
	FatalContext(ctx context.Context, msg string, args ...any)
}

// DefaultLogger is the default [Log] with console handler.
var DefaultLogger = New(CONSOLE_HANDLER, Options{
	Development:  false,
	Debug:        false,
	EnableCaller: false,
	SkipCaller:   3,
})

// New create new Logger
func New(
	handler HandleType,
	loggerOption Options,
) Logger {
	log := new(Log)
	logger := slog.Default()

	if loggerOption.CustomSlog != nil {
		logger = loggerOption.CustomSlog
	}

	slogHandlerOpt := new(slog.HandlerOptions)
	slogHandlerOpt.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		return a
	}

	if loggerOption.Debug {
		slogHandlerOpt.Level = slog.LevelDebug
	}

	if loggerOption.EnableCaller {
		slogHandlerOpt.AddSource = true
	}

	switch handler {
	case JSON_HANDLER:
		logger = slog.New(slog.NewJSONHandler(os.Stderr, slogHandlerOpt))
		if loggerOption.CustomJsonHandler != nil {
			logger = slog.New(loggerOption.CustomJsonHandler)
		}
	case TEXT_HANDLER:
		logger = slog.New(slog.NewTextHandler(os.Stderr, slogHandlerOpt))
	case CONSOLE_HANDLER:
		logger = slog.New(NewConsoleHandler(slogHandlerOpt))
	}

	if loggerOption.Development {
		buildInfo, _ := debug.ReadBuildInfo()
		logger = logger.With(slog.Group("debug_info",
			slog.String("go_version", buildInfo.GoVersion),
			slog.Int("pid", os.Getpid()),
			slog.String("os", runtime.GOOS),
			slog.String("os_arch", runtime.GOARCH),
		))
	}

	log.slog = logger
	log.skipCaller = loggerOption.SkipCaller

	return log
}

func (l *Log) Debug(msg string, keyValues ...any) {
	l.log(context.Background(), slog.LevelDebug, msg, keyValues...)
}

func (l *Log) DebugContext(ctx context.Context, msg string, keyValues ...any) {
	l.log(ctx, slog.LevelDebug, msg, keyValues...)
}

func (l *Log) Info(msg string, keyValues ...any) {
	l.log(context.Background(), slog.LevelInfo, msg, keyValues...)
}

func (l *Log) InfoContext(ctx context.Context, msg string, keyValues ...any) {
	l.log(ctx, slog.LevelInfo, msg, keyValues...)
}

func (l *Log) Warn(msg string, keyValues ...any) {
	l.log(context.Background(), slog.LevelWarn, msg, keyValues...)
}

func (l *Log) WarnContext(ctx context.Context, msg string, keyValues ...any) {
	l.log(ctx, slog.LevelWarn, msg, keyValues...)
}

func (l *Log) Error(msg string, keyValues ...any) {
	l.log(context.Background(), slog.LevelError, msg, keyValues...)
}

func (l *Log) ErrorContext(ctx context.Context, msg string, keyValues ...any) {
	l.log(ctx, slog.LevelError, msg, keyValues...)
}

func (l *Log) Fatal(msg string, keyValues ...any) {
	defer os.Exit(1)
	l.log(context.Background(), slog.LevelError, msg, keyValues...)
}

func (l *Log) FatalContext(ctx context.Context, msg string, keyValues ...any) {
	defer os.Exit(1)
	l.log(ctx, slog.LevelError, msg, keyValues...)
}

func (l *Log) log(ctx context.Context, level slog.Level, msg string, keyValues ...any) {
	var pcs [1]uintptr
	runtime.Callers(l.skipCaller, pcs[:])
	rec := slog.NewRecord(time.Now(), level, msg, pcs[0])
	rec.Add(keyValues...)

	_ = l.slog.Handler().Handle(ctx, rec)
}

func (e Environment) String() string {
	switch e {
	case DEVELOPMENT:
		return "development"
	case PRODUCTION:
		return "production"
	case RELEASE:
		return "release"
	}
	return ""
}
