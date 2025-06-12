package server

import (
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
)

type Service struct {
	storage store.Storage
}

func NewService(storage store.Storage) *Service {
	return &Service{
		storage: storage,
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

func (s *Service) GetMetric(name string) (metric models.Metrics, exists bool) {
	metric, exists = s.storage.GetMetric(name)
	return
}

func (s *Service) GetAllMetrics() []models.Metrics {
	return s.storage.GetAllMetrics()
}
