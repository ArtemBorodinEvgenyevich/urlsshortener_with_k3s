package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var global *zap.Logger

func Initialize() error {
	logger, err := newLogger()
	if err != nil {
		return err
	}

	global = logger
	zap.ReplaceGlobals(global)
	return nil
}

func L() *zap.Logger {
	if global == nil {
		return zap.NewNop()
	}
	return global
}

func Sync() error {
	if global != nil {
		return global.Sync()
	}
	return nil
}

func newLogger() (*zap.Logger, error) {
	env := os.Getenv("ENV")
	var config zap.Config

	if env == "" || env == "development" {
		config = developmentConfig()
	} else {
		config = productionConfig()
	}

	return config.Build(zap.AddCaller())
}

func developmentConfig() zap.Config {
	return zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: false,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "@timestamp",
			NameKey:        "logger",
			LevelKey:       "level",
			MessageKey:     "message",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout(time.DateTime),
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func productionConfig() zap.Config {
	return zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "@timestamp",
			NameKey:        "logger",
			LevelKey:       "level",
			MessageKey:     "message",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}
