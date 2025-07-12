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

func (s *MetricsService) UpdateMetrics(ctx context.Context, metrics []*models.Metrics) error {
	return s.storage.UpdateMetrics(ctx, metrics)
}

func (s *MetricsService) UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	if (metric.Delta == nil && metric.Value == nil) ||
		(metric.MType != models.Gauge && metric.MType != models.Counter) {
		return nil, store.ErrInvalidMetricReceived
	}
	if metric.ID == "" {
		return nil, store.ErrNotFound
	}
	return s.storage.UpdateMetric(ctx, metric)
}

func (s *MetricsService) GetMetric(ctx context.Context, name string) (*models.Metrics, error) {
	return s.storage.GetMetric(ctx, name)
}

func (s *MetricsService) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	return s.storage.GetAllMetrics(ctx)
}
