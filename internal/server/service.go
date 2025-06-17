package server

import (
	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
)

type Service struct {
	ServerHost string
	storage    store.Storage
	Logger     *logger.Logger
}

func NewService(storage store.Storage, config *config.Config, logger *logger.Logger) *Service {
	return &Service{
		storage:    storage,
		ServerHost: config.ServerHost,
		Logger:     logger,
	}
}

func (s *Service) UpdateCounter(name string, value int64) error {
	err := s.storage.UpdateCounter(name, value)
	return err
}

func (s *Service) UpdateGauge(name string, value float64) error {
	err := s.storage.UpdateGauge(name, value)
	return err
}

func (s *Service) GetMetric(name string) (metric *models.Metrics, exists bool) {
	metric, exists = s.storage.GetMetric(name)
	return
}

func (s *Service) GetAllMetrics() []models.Metrics {
	return s.storage.GetAllMetrics()
}
