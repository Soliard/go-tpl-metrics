package server

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type MetricsService struct {
	ServerHost string
	storage    store.Storage
	Logger     *zap.Logger
}

var maxRetries = 3

func NewMetricsService(storage store.Storage, config *config.Config, logger *zap.Logger) *MetricsService {
	return &MetricsService{
		storage:    storage,
		ServerHost: config.ServerHost,
		Logger:     logger,
	}
}

func (s *MetricsService) UpdateMetrics(ctx context.Context, metrics []*models.Metrics) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = s.storage.UpdateMetrics(ctx, metrics)
		if isRetriableError(err) {
			waitForRetry(i)
			continue
		}
		return err
	}
	return err
}

func (s *MetricsService) UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	if (metric.Delta == nil && metric.Value == nil) ||
		(metric.MType != models.Gauge && metric.MType != models.Counter) {
		return nil, store.ErrInvalidMetricReceived
	}
	if metric.ID == "" {
		return nil, store.ErrNotFound
	}

	var retMetric *models.Metrics
	var err error
	for i := 0; i < maxRetries; i++ {
		retMetric, err = s.storage.UpdateMetric(ctx, metric)
		if isRetriableError(err) {
			waitForRetry(i)
			continue
		}
		return retMetric, err
	}
	return retMetric, err
}

func (s *MetricsService) GetMetric(ctx context.Context, name string) (*models.Metrics, error) {
	var metric *models.Metrics
	var err error
	for i := 0; i < maxRetries; i++ {
		metric, err = s.storage.GetMetric(ctx, name)
		if isRetriableError(err) {
			waitForRetry(i)
			continue
		}
		return metric, err
	}
	return metric, err
}

func (s *MetricsService) GetAllMetrics(ctx context.Context) ([]*models.Metrics, error) {
	var metrics []*models.Metrics
	var err error
	for i := 0; i < maxRetries; i++ {
		metrics, err = s.storage.GetAllMetrics(ctx)
		if isRetriableError(err) {
			waitForRetry(i)
			continue
		}
		return metrics, err
	}
	return metrics, err
}

func waitForRetry(iter int) {
	switch iter {
	case 0:
		time.Sleep(time.Second * 1)
	case 1:
		time.Sleep(time.Second * 3)
	case 2:
		time.Sleep(time.Second * 5)
	default:
		time.Sleep(time.Second * 1)
	}
}

func isRetriableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.ConnectionException
	}
	if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, os.ErrPermission) {
		return true
	}
	return false
}
