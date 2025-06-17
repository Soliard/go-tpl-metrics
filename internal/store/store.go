package store

import (
	"fmt"
	"sort"

	"github.com/Soliard/go-tpl-metrics/models"
)

type Storage interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetMetric(name string) (metric *models.Metrics, exists bool)
	GetAllMetrics() []models.Metrics
	GetAllMetricsStringDTO() []*models.MetricStringDTO
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
		fmt.Printf("[storage UpdateCounter] Updated counter metric: %v\n", models.СonvertToMetricStringDTO(metric))
	} else {
		newDelta := value
		s.metrics[name] = models.Metrics{ID: name, MType: models.Counter, Delta: &newDelta}
		fmt.Printf("[storage UpdateCounter] Created new counter metric: %v\n", models.СonvertToMetricStringDTO(s.metrics[name]))
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
		fmt.Printf("[storage UpdateGauge] Updated gauge metric: %v\n", models.СonvertToMetricStringDTO(metric))
	} else {
		newValue := value
		s.metrics[name] = models.Metrics{ID: name, MType: models.Gauge, Value: &newValue}
		fmt.Printf("[storage UpdateGauge] Created new gauge metric: %v\n", models.СonvertToMetricStringDTO(s.metrics[name]))
	}

	return nil
}

func (s *memStorage) GetMetric(name string) (metric *models.Metrics, exists bool) {
	val, ok := s.metrics[name]
	return &val, ok
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

func (s *memStorage) GetAllMetricsStringDTO() []*models.MetricStringDTO {
	metrics := make([]*models.MetricStringDTO, 0, len(s.metrics))
	for _, m := range s.metrics {
		metrics = append(metrics, models.СonvertToMetricStringDTO(m))
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})
	return metrics
}
