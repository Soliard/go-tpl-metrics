package store

import (
	"context"
	"errors"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/models"
)

type Storage interface {
	UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error)
	UpdateMetrics(ctx context.Context, metrics []*models.Metrics) error
	GetMetric(ctx context.Context, name string) (*models.Metrics, error)
	GetAllMetrics(ctx context.Context) ([]models.Metrics, error)
}

var ErrNotFound = errors.New("not found")
var ErrInvalidMetricReceived = errors.New("invalid metric recieved")

func New(ctx context.Context, config *config.Config) (Storage, error) {
	if config.FileStoragePath != "" {
		return NewFileStorage(config.FileStoragePath, config.IsRestoreFromFile)
	} else if config.DatabaseDSN != "" {
		return NewDatabaseStorage(ctx, config.DatabaseDSN)
	} else {
		return NewMemoryStorage(), nil
	}
}
