package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type DatabaseStorage struct {
	db *sqlx.DB
}

func NewDatabaseStorage(ctx context.Context, databaseDSN string) (Storage, error) {
	db, err := sqlx.Open("pgx", databaseDSN)
	if err != nil {
		return nil, err
	}
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	//exmpl migrate create -ext sql -dir migrations -seq create_metrics_table
	// миграции
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	migr, err := migrate.NewWithDatabaseInstance(
		//file://cmd/server/migrations
		"file://cmd/server/migrations",
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

func (s *DatabaseStorage) UpdateMetrics(ctx context.Context, metrics []*models.Metrics) error {
	query := `
	INSERT INTO metrics (id, type, value, delta, hash)
	VALUES (:id, :type, :value, :delta, :hash)
	ON CONFLICT (id) DO UPDATE SET
		value = EXCLUDED.value,
		delta = metrics.delta + EXCLUDED.delta,
		hash = EXCLUDED.hash
`
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareNamedContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range metrics {
		params := map[string]interface{}{
			"id":    m.ID,
			"type":  m.MType,
			"value": m.Value,
			"delta": m.Delta,
			"hash":  m.Hash,
		}
		_, err := stmt.ExecContext(ctx, params)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *DatabaseStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	existed, err := s.GetMetric(ctx, metric.ID)
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
	if existed.MType != metric.MType {
		return nil, errors.New("trying to update existed metric with same id, but new mtype")
	}
	if metric.MType == models.Counter {
		_, err = s.db.ExecContext(ctx, `
			UPDATE 
				metrics
			SET
				delta = delta + $1
			WHERE
				id = $2
		`, metric.Delta, metric.ID)
	} else {
		_, err = s.db.ExecContext(ctx, `
			UPDATE 
				metrics
			SET
				value = $1
			WHERE
				id = $2
		`, metric.Value, metric.ID)
	}
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
