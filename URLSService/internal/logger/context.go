package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

var loggerKey = ctxKey{}

func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return L()
}

func InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

func ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

func DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

func WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}
