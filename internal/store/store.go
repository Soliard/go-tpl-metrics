// Package store для хранения данных - один интерфейс Storage с тремя реализациями:
// в памяти, в файле и в базе данных.
// Основная цель - обеспечить гибкость выбора хранилища в зависимости от конфигурации.
package store

import (
	"context"
	"errors"

	"github.com/Soliard/go-tpl-metrics/internal/config"
	"github.com/Soliard/go-tpl-metrics/models"
)

// Storage определяет интерфейс для работы с хранилищем метрик.
// Поддерживает различные типы хранилищ: память, файл, база данных.
type Storage interface {
	// UpdateMetric обновляет или создает одну метрику
	UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error)
	// UpdateMetrics обновляет или создает несколько метрик за одну операцию
	UpdateMetrics(ctx context.Context, metrics []*models.Metrics) error
	// GetMetric получает метрику по имени
	GetMetric(ctx context.Context, name string) (*models.Metrics, error)
	// GetAllMetrics получает все метрики из хранилища
	GetAllMetrics(ctx context.Context) ([]*models.Metrics, error)
}

// ErrNotFound возвращается когда метрика не найдена в хранилище
var ErrNotFound = errors.New("not found")

// ErrInvalidMetricReceived возвращается когда получена некорректная метрика
var ErrInvalidMetricReceived = errors.New("invalid metric recieved")

// New создает новое хранилище на основе конфигурации.
// Выбирает тип хранилища в зависимости от настроек:
// - FileStoragePath указан -> файловое хранилище
// - DatabaseDSN указан -> хранилище в базе данных
// - иначе -> хранилище в памяти
func New(ctx context.Context, config *config.ServerConfig) (Storage, error) {
	if config.FileStoragePath != "" {
		return NewFileStorage(config.FileStoragePath, config.IsRestoreFromFile)
	} else if config.DatabaseDSN != "" {
		return NewDatabaseStorage(ctx, config.DatabaseDSN)
	} else {
		return NewMemoryStorage(), nil
	}
}
