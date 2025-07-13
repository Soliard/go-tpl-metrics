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

func (s *memStorage) UpdateMetrics(ctx context.Context, metrics []*models.Metrics) error {
	for _, m := range metrics {
		_, err := s.UpdateMetric(ctx, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *memStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	existed, err := s.GetMetric(ctx, metric.ID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// creating new metric
			s.metrics[metric.ID] = metric
			return metric, nil
		}
		return nil, err
	}

	if existed.MType != metric.MType {
		return nil, errors.New("trying to update existed metric with same id, but new mtype")
	}

	// updating existing metric
	switch metric.MType {
	case models.Gauge:
		{
			*existed.Value = *metric.Value
		}
	case models.Counter:
		{
			*existed.Delta += *metric.Delta
		}
	default:
		{
			return nil, errors.New("provided not supported metric type")
		}
	}
	return existed, nil

}

func (s *memStorage) GetMetric(ctx context.Context, name string) (*models.Metrics, error) {
	if metric, ok := s.metrics[name]; ok {
		return metric, nil
	}
	return nil, ErrNotFound
}

func (s *memStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, len(s.metrics))
	idx := 0
	for _, m := range s.metrics {
		metrics[idx] = *m
		idx++
	}
	return metrics, nil
}
