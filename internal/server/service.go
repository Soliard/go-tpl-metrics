// Package server предоставляет HTTP сервер для сбора и хранения метрик.
// Включает в себя обработчики для обновления, получения и отображения метрик,
// а также сервисный слой с логикой повторных попыток и валидации.
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

// MetricsService представляет основной сервис для работы с метриками.
// Обеспечивает HTTP API для обновления и получения метрик с поддержкой повторных попыток.
type MetricsService struct {
	ServerHost string
	storage    store.Storage
	Logger     *zap.Logger
	signKey    []byte
}

var maxRetries = 3

// NewMetricsService создает новый экземпляр сервиса метрик.
// Инициализирует хранилище, логгер и ключ для подписи данных.
func NewMetricsService(storage store.Storage, config *config.Config, logger *zap.Logger) *MetricsService {
	return &MetricsService{
		storage:    storage,
		ServerHost: config.ServerHost,
		Logger:     logger,
		signKey:    []byte(config.SignKey),
	}
}

// UpdateMetrics обновляет несколько метрик с поддержкой повторных попыток.
// Валидирует все метрики перед обновлением.
func (s *MetricsService) UpdateMetrics(ctx context.Context, metrics []*models.Metrics) error {
	var err error

	for _, m := range metrics {
		mErr := validateMetric(m)
		if mErr != nil {
			err = errors.Join(err, mErr)
		}
	}
	if err != nil {
		return err
	}

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

// UpdateMetric обновляет одну метрику с поддержкой повторных попыток.
// Возвращает обновленную метрику.
func (s *MetricsService) UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	var retMetric *models.Metrics
	var err error

	err = validateMetric(metric)
	if err != nil {
		return nil, err
	}

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

// GetMetric получает метрику по имени с поддержкой повторных попыток.
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

// GetAllMetrics получает все метрики с поддержкой повторных попыток.
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

// validateMetric проверяет корректность метрики перед сохранением
func validateMetric(m *models.Metrics) error {
	if (m.Delta == nil && m.Value == nil) ||
		(m.MType != models.Gauge && m.MType != models.Counter) {
		return store.ErrInvalidMetricReceived
	}
	if m.ID == "" {
		return store.ErrNotFound
	}
	return nil
}

// waitForRetry реализует экспоненциальную задержку для повторных попыток
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

// isRetriableError определяет, можно ли повторить операцию при данной ошибке
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
