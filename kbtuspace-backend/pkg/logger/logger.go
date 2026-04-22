package logger

import (
	"context"
	"log/slog"
	"os"
)

var DefaultLogger *slog.Logger

func Init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	DefaultLogger = slog.New(handler)
	slog.SetDefault(DefaultLogger)
}

func Info(ctx context.Context, msg string, attrs ...any) {
	DefaultLogger.InfoContext(ctx, msg, attrs...)
}

func Error(ctx context.Context, msg string, err error, attrs ...any) {
	attrs = append(attrs, slog.Any("error", err))
	DefaultLogger.ErrorContext(ctx, msg, attrs...)
}

func Warn(ctx context.Context, msg string, attrs ...any) {
	DefaultLogger.WarnContext(ctx, msg, attrs...)
}

func Debug(ctx context.Context, msg string, attrs ...any) {
	DefaultLogger.DebugContext(ctx, msg, attrs...)
}
