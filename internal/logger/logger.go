package logger

import (
	"context"

	"go.uber.org/zap"
)

const (
	ctxKeyLogger ctxKey = "logger"
)

// исключает перезапись ключей контекста сторонними сервисами/либами
type ctxKey string

func New(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	return zl, nil
}

func CtxWithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, l)
}

func LoggerFromCtx(ctx context.Context, defaultLogger *zap.Logger) *zap.Logger {
	loggerFromCtx, ok := ctx.Value(ctxKeyLogger).(*zap.Logger)
	if !ok {
		return defaultLogger
	}
	return loggerFromCtx
}
