package server

import (
	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
)

type MetricsService struct {
	ServerHost string
	storage    store.Storage
	Logger     *logger.Logger
}

func NewMetricsService(storage store.Storage, config *config.Config, logger *logger.Logger) *MetricsService {
	return &MetricsService{
		storage:    storage,
		ServerHost: config.ServerHost,
		Logger:     logger,
	}
}

func (s *MetricsService) UpdateCounter(name string, value int64) error {
	err := s.storage.UpdateCounter(name, value)
	return err
}

func (s *MetricsService) UpdateGauge(name string, value float64) error {
	err := s.storage.UpdateGauge(name, value)
	return err
}

func (s *MetricsService) GetMetric(name string) (metric *models.Metrics, exists bool) {
	metric, exists = s.storage.GetMetric(name)
	return
}

func (s *MetricsService) GetAllMetrics() []models.Metrics {
	return s.storage.GetAllMetrics()
}
