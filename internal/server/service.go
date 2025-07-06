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

func (s *MetricsService) UpdateCounter(ctx context.Context, name string, value *int64) error {
	err := s.storage.UpdateCounter(ctx, name, value)
	return err
}

func (s *MetricsService) UpdateGauge(ctx context.Context, name string, value *float64) error {
	err := s.storage.UpdateGauge(ctx, name, value)
	return err
}

func (s *MetricsService) GetMetric(ctx context.Context, name string) (metric *models.Metrics, exists bool) {
	metric, exists = s.storage.GetMetric(ctx, name)
	return
}

func (s *MetricsService) GetAllMetrics(ctx context.Context) []models.Metrics {
	return s.storage.GetAllMetrics(ctx)
}
