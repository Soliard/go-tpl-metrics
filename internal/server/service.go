package server

import (
	"context"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
)

type MetricsService struct {
	ServerHost string
	storage    store.Storage
	Logger     *zap.Logger
}

func NewMetricsService(storage store.Storage, config *config.Config, logger *zap.Logger) *MetricsService {
	return &MetricsService{
		storage:    storage,
		ServerHost: config.ServerHost,
		Logger:     logger,
	}
}

func (s *MetricsService) UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	return s.storage.UpdateMetric(ctx, metric)
}

func (s *MetricsService) GetMetric(ctx context.Context, name string) (*models.Metrics, error) {
	return s.storage.GetMetric(ctx, name)
}

func (s *MetricsService) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	return s.storage.GetAllMetrics(ctx)
}
