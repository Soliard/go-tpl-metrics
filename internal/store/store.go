package store

import (
	"fmt"

	"github.com/Soliard/go-tpl-metrics/models"
)

type Storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetMetric(name string) (metric models.Metrics, exists bool)
	GetAllMetrics() []models.Metrics
	GetAllMetricsStringDTO() []models.MetricStringDTO
}

type memStorage struct {
	metrics map[string]models.Metrics
}

func NewStorage() Storage {
	return &memStorage{
		metrics: map[string]models.Metrics{},
	}
}

func (s *memStorage) UpdateCounter(name string, value int64) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if metric, exists := s.metrics[name]; exists {
		if metric.MType != models.Counter {
			return fmt.Errorf("metric is not counter type")
		}
		if metric.Delta == nil {
			metric.Delta = new(int64)
		}
		*metric.Delta += value
	} else {
		newDelta := value
		s.metrics[name] = models.Metrics{ID: name, MType: models.Counter, Delta: &newDelta}
	}

	return nil
}

func (s *memStorage) UpdateGauge(name string, value float64) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if metric, exists := s.metrics[name]; exists {
		if metric.MType != models.Gauge {
			return fmt.Errorf("metric is not gauge type")
		}
		*metric.Value = value
	} else {
		newValue := value
		s.metrics[name] = models.Metrics{ID: name, MType: models.Gauge, Value: &newValue}
	}

	return nil
}

func (s *memStorage) GetMetric(name string) (metric models.Metrics, exists bool) {
	val, ok := s.metrics[name]
	return val, ok
}

func (s *memStorage) GetAllMetrics() []models.Metrics {
	metrics := make([]models.Metrics, len(s.metrics))
	idx := 0
	for _, m := range s.metrics {
		metrics[idx] = m
		idx++
	}

	return metrics
}

func (s *memStorage) GetAllMetricsStringDTO() []models.MetricStringDTO {
	metrics := make([]models.MetricStringDTO, len(s.metrics))
	idx := 0
	for _, m := range s.metrics {
		metrics[idx] = models.СonvertToMetricStringDTO(m)
		idx++
	}
	return metrics
}
