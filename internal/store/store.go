package store

import (
	"context"
	"fmt"

	"github.com/Soliard/go-tpl-metrics/models"
)

type Storage interface {
	UpdateGauge(ctx context.Context, name string, value *float64) error
	UpdateCounter(ctx context.Context, name string, value *int64) error
	GetMetric(ctx context.Context, name string) (metric *models.Metrics, exists bool)
	GetAllMetrics(ctx context.Context) []models.Metrics
}

type memStorage struct {
	metrics map[string]*models.Metrics
}

func NewMemoryStorage() Storage {
	return &memStorage{
		metrics: map[string]*models.Metrics{},
	}
}

func (s *memStorage) UpdateCounter(ctx context.Context, name string, value *int64) error {
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
		*metric.Delta += *value
		fmt.Printf("[storage UpdateCounter] Updated counter metric: %v\n", metric.String())
	} else {
		newMetric := &models.Metrics{ID: name, MType: models.Counter, Delta: value}
		s.metrics[name] = newMetric
		fmt.Printf("[storage UpdateCounter] Created new counter metric: %v\n", newMetric.String())
	}

	return nil
}

func (s *memStorage) UpdateGauge(ctx context.Context, name string, value *float64) error {
	if name == "" {
		return fmt.Errorf("metric name cannot be empty")
	}
	if metric, exists := s.metrics[name]; exists {
		if metric.MType != models.Gauge {
			return fmt.Errorf("metric is not gauge type")
		}
		metric.Value = value
		fmt.Printf("[storage UpdateGauge] Updated gauge metric: %v\n", metric.String())
	} else {
		newMetric := &models.Metrics{ID: name, MType: models.Gauge, Value: value}
		s.metrics[name] = newMetric
		fmt.Printf("[storage UpdateGauge] Created new gauge metric: %v\n", newMetric.String())
	}

	return nil
}

func (s *memStorage) GetMetric(ctx context.Context, name string) (metric *models.Metrics, exists bool) {
	metric, ok := s.metrics[name]
	return metric, ok
}

func (s *memStorage) GetAllMetrics(ctx context.Context) []models.Metrics {
	metrics := make([]models.Metrics, len(s.metrics))
	idx := 0
	for _, m := range s.metrics {
		metrics[idx] = *m
		idx++
	}
	return metrics
}
