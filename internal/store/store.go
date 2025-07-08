package store

import (
	"context"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/models"
)

type Storage interface {
	UpdateMetric(ctx context.Context, metric *models.Metrics) error
	GetMetric(ctx context.Context, name string) (metric *models.Metrics, exists bool)
	GetAllMetrics(ctx context.Context) []models.Metrics
}

func New(ctx context.Context, config *config.Config) (Storage, error) {
	if config.FileStoragePath != "" {
		return NewFileStorage(config.FileStoragePath, config.IsRestoreFromFile)
	} else if config.DatabaseDSN != "" {
		return NewDatabaseStorage(ctx, config.DatabaseDSN)
	} else {
		return NewMemoryStorage(), nil
	}
}
