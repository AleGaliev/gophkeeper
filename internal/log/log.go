package log

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	logger *slog.Logger
}

func New(logLevel string) Logger {
	logLevelSlog := convert(logLevel)

	opts := &slog.HandlerOptions{
		Level: logLevelSlog,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	return Logger{logger: logger}
}

func (l Logger) Info(ctx context.Context, msg string, args ...any) {
	l.logger.Log(ctx, slog.LevelInfo, msg, args...)
}

func (l Logger) Error(ctx context.Context, msg string, args ...any) {
	l.logger.Log(ctx, slog.LevelError, msg, args...)
}

func (l Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.logger.Log(ctx, slog.LevelWarn, msg, args...)
}

func (l Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.logger.Log(ctx, slog.LevelDebug, msg, args...)
}

func convert(level string) slog.Level {
	switch level {
	case "":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}
	return slog.LevelInfo
}
