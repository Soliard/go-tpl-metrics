// Package logger предоставляет утилиты для работы с логгером.
// Включает создание логгера и работу с контекстом для передачи логгера между слоями.
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

// New создает новый логгер с указанным уровнем логирования.
// Использует production конфигурацию с настраиваемым уровнем.
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

// CtxWithLogger добавляет логгер в контекст для передачи между слоями приложения.
func CtxWithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, l)
}

// LoggerFromCtx извлекает логгер из контекста или возвращает fallback логгер.
func LoggerFromCtx(ctx context.Context, defaultLogger *zap.Logger) *zap.Logger {
	loggerFromCtx, ok := ctx.Value(ctxKeyLogger).(*zap.Logger)
	if !ok {
		return defaultLogger
	}
	return loggerFromCtx
}
