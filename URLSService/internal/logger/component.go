package logger

import (
	"context"

	"go.uber.org/zap"
)

type Component string

const (
	ComponentPostgres Component = "POSTGRES"
	ComponentRedis    Component = "REDIS"
	ComponentApp      Component = "APP"
)

// No context component loggers

func Postgres() *zap.Logger {
	return L().Named(string(ComponentPostgres))
}

func Redis() *zap.Logger {
	return L().Named(string(ComponentRedis))
}

func App() *zap.Logger {
	return L().Named(string(ComponentApp))
}

// Postgres loggers

func PgLogDebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentPostgres)).
		WithOptions(zap.AddCallerSkip(1)).
		Debug(msg, fields...)
}

func PgLogInfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentPostgres)).
		WithOptions(zap.AddCallerSkip(1)).
		Info(msg, fields...)
}

func PgLogWarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentPostgres)).
		WithOptions(zap.AddCallerSkip(1)).
		Warn(msg, fields...)
}

func PgLogErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentPostgres)).
		WithOptions(zap.AddCallerSkip(1)).
		Error(msg, fields...)
}

// Non-context versions

func PgLogDebug(msg string, fields ...zap.Field) {
	Postgres().
		WithOptions(zap.AddCallerSkip(1)).
		Debug(msg, fields...)
}

func PgLogInfo(msg string, fields ...zap.Field) {
	Postgres().
		WithOptions(zap.AddCallerSkip(1)).
		Info(msg, fields...)
}

func PgLogWarn(msg string, fields ...zap.Field) {
	Postgres().
		WithOptions(zap.AddCallerSkip(1)).
		Warn(msg, fields...)
}

func PgLogError(msg string, fields ...zap.Field) {
	Postgres().
		WithOptions(zap.AddCallerSkip(1)).
		Error(msg, fields...)
}

// Redis loggers

func RedisLogDebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentRedis)).
		WithOptions(zap.AddCallerSkip(1)).
		Debug(msg, fields...)
}

func RedisLogInfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentRedis)).
		WithOptions(zap.AddCallerSkip(1)).
		Info(msg, fields...)
}

func RedisLogWarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentRedis)).
		WithOptions(zap.AddCallerSkip(1)).
		Warn(msg, fields...)
}

func RedisLogErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentRedis)).
		WithOptions(zap.AddCallerSkip(1)).
		Error(msg, fields...)
}

// Non-context versions

func RedisLogDebug(msg string, fields ...zap.Field) {
	Redis().
		WithOptions(zap.AddCallerSkip(1)).
		Debug(msg, fields...)
}

func RedisLogInfo(msg string, fields ...zap.Field) {
	Redis().
		WithOptions(zap.AddCallerSkip(1)).
		Info(msg, fields...)
}

func RedisLogWarn(msg string, fields ...zap.Field) {
	Redis().
		WithOptions(zap.AddCallerSkip(1)).
		Warn(msg, fields...)
}

func RedisLogError(msg string, fields ...zap.Field) {
	Redis().
		WithOptions(zap.AddCallerSkip(1)).
		Error(msg, fields...)
}

// App loggers

func AppLogDebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentApp)).
		WithOptions(zap.AddCallerSkip(1)).
		Debug(msg, fields...)
}

func AppLogInfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentApp)).
		WithOptions(zap.AddCallerSkip(1)).
		Info(msg, fields...)
}

func AppLogWarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentApp)).
		WithOptions(zap.AddCallerSkip(1)).
		Warn(msg, fields...)
}

func AppLogErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).
		Named(string(ComponentApp)).
		WithOptions(zap.AddCallerSkip(1)).
		Error(msg, fields...)
}

// Non-context versions

func AppLogDebug(msg string, fields ...zap.Field) {
	App().
		WithOptions(zap.AddCallerSkip(1)).
		Debug(msg, fields...)
}

func AppLogInfo(msg string, fields ...zap.Field) {
	App().
		WithOptions(zap.AddCallerSkip(1)).
		Info(msg, fields...)
}

func AppLogWarn(msg string, fields ...zap.Field) {
	App().
		WithOptions(zap.AddCallerSkip(1)).
		Warn(msg, fields...)
}

func AppLogError(msg string, fields ...zap.Field) {
	App().
		WithOptions(zap.AddCallerSkip(1)).
		Error(msg, fields...)
}
