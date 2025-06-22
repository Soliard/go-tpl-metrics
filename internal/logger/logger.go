package logger

import (
	"go.uber.org/zap"
)

type Component string

const (
	ComponentServer Component = "server"
	ComponentAgent  Component = "agent"
)

type Logger struct {
	Log *zap.Logger
}

func New(component Component, level string) (*Logger, error) {
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
	// устанавливаем синглтон
	return &Logger{Log: zl}, nil
}
