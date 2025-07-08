package store

import (
	"context"
	"errors"

	"github.com/Soliard/go-tpl-metrics/models"
)

type memStorage struct {
	metrics map[string]*models.Metrics
}

func NewMemoryStorage() Storage {
	return &memStorage{
		metrics: map[string]*models.Metrics{},
	}
}

func (s *memStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) error {
	if metric == nil {
		return errors.New("metric cannot be empty")
	}
	if metric.ID == "" {
		return errors.New("metric id cannot be empty")
	}

	if existed, ok := s.GetMetric(ctx, metric.ID); ok {
		if existed.MType != metric.MType {
			return errors.New("trying to update existed metric with same id, but new mtype")
		}

		switch metric.MType {
		case models.Gauge:
			{
				*existed.Value = *metric.Value
				return nil
			}
		case models.Counter:
			{
				*existed.Delta += *metric.Delta
				return nil
			}

		}
	}

	s.metrics[metric.ID] = metric
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
