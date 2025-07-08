package store

import (
	"context"

	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/jackc/pgx/v5"
)

type DatabaseStorage struct {
	conn *pgx.Conn
}

func NewDatabaseStorage(ctx context.Context, databaseDSN string) (Storage, error) {
	con, err := pgx.Connect(ctx, databaseDSN)
	if err != nil {
		return nil, err
	}
	return &DatabaseStorage{conn: con}, nil
}

func (s *DatabaseStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) error {

	return nil
}

func (s *DatabaseStorage) GetMetric(ctx context.Context, name string) (metric *models.Metrics, exists bool) {

	return nil, false
}

func (s *DatabaseStorage) GetAllMetrics(ctx context.Context) []models.Metrics {
	return []models.Metrics{}
}

func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.conn.Ping(ctx)
}
