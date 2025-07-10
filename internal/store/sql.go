package store

import (
	"context"
	"database/sql"

	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DatabaseStorage struct {
	db *sql.DB
}

func NewDatabaseStorage(ctx context.Context, databaseDSN string) (Storage, error) {
	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	//exmpl migrate create -ext sql -dir migrations -seq create_metrics_table
	// миграции
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	migr, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		databaseDSN,
		driver,
	)
	if err != nil {
		return nil, err
	}
	if err := migr.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return &DatabaseStorage{db: db}, nil
}

func (s *DatabaseStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	_, err := s.GetMetric(ctx, metric.ID)
	if err != nil {
		if err == ErrNotFound {
			_, err := s.db.ExecContext(ctx, `
				INSERT INTO metrics 
					(id, type, value, delta, hash) 
				VALUES 
					($1, $2, $3, $4, $5)
			`, metric.ID, metric.MType, metric.Value, metric.Delta, metric.Hash)
			if err != nil {
				return nil, err
			}
			return metric, nil
		}
		return nil, err
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE 
			metrics
		SET
			value = $1,
			delta = $2
		WHERE
			id = $3
	`, metric.Value, metric.Delta, metric.ID)
	if err != nil {
		return nil, err
	}
	return metric, nil
}

func (s *DatabaseStorage) GetMetric(ctx context.Context, name string) (*models.Metrics, error) {
	var metric models.Metrics
	err := s.db.QueryRowContext(ctx, `
		SELECT 
			id, type, delta, value, hash
		FROM
			metrics
		WHERE 
			id = $1
	`, name).Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value, &metric.Hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &metric, nil
}

func (s *DatabaseStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	var metrics []models.Metrics

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			id, type, delta, value, hash
		FROM metrics
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Metrics
		err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
