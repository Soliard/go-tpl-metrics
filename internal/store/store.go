package store

import (
	"context"
	"errors"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
)

type Storage interface {
	UpdateMetric(ctx context.Context, metric *models.Metrics) error
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

func (s *memStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) error {
	logger := logger.LoggerFromCtx(ctx, zap.Must(zap.NewProduction()))
	if metric == nil {
		logger.Warn("recieved nil metric to update")
		return errors.New("metric cannot be empty")
	}
	if metric.ID == "" {
		logger.Warn("recieved mitric with empty id to update")
		return errors.New("metric id cannot be empty")
	}

	if existed, ok := s.GetMetric(ctx, metric.ID); ok {
		if existed.MType != metric.MType {
			logger.Error("trying to update existed metric with same id, but new mtyper",
				zap.Any("existed", existed),
				zap.Any("new", metric))
			return errors.New("trying to update existed metric with same id, but new mtyper")
		}

		switch metric.MType {
		case models.Gauge:
			{
				*existed.Value = *metric.Value
				logger.Info("updated gauge metric",
					zap.Any("metric before", existed),
					zap.Any("metric after", metric))
				return nil
			}
		case models.Counter:
			{
				*existed.Delta += *metric.Delta
				logger.Info("updated counter metric",
					zap.Any("metric before", existed),
					zap.Any("metric after", metric))
				return nil
			}

		}
	}

	s.metrics[metric.ID] = metric
	logger.Info("created new metric", zap.Any("metric", metric))
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
