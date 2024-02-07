package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ILogger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Infoln(args ...interface{})
	Info(args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	DPanicw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	Sync() error
}

func Initialize(level string) (ILogger, error) {
	defaultLogger := zap.NewNop().Sugar()
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return defaultLogger, err
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zl, err := cfg.Build()

	if err != nil {
		return defaultLogger, err
	}

	return zl.Sugar(), nil
}
