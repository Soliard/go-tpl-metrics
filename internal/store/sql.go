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

// DatabaseStorage реализует Storage интерфейс для хранения метрик в PostgreSQL.
// Автоматически выполняет миграции при создании.
type DatabaseStorage struct {
	db *sqlx.DB
}

// NewDatabaseStorage создает новое хранилище в базе данных PostgreSQL.
// Выполняет миграции из папки cmd/server/migrations.
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

// UpdateMetrics обновляет несколько метрик в базе данных за одну транзакцию.
// Для counter метрик значения суммируются, для gauge - перезаписываются.
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

// UpdateMetric обновляет или создает одну метрику в базе данных
func (s *DatabaseStorage) UpdateMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	existed, err := s.GetMetric(ctx, metric.ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if err == nil {
		if existed.MType != metric.MType {
			return nil, ErrInvalidMetricReceived
		}
	}

	var query string
	var args []interface{}

	if metric.MType == models.Counter {
		query = `
			INSERT INTO metrics (id, type, value, delta, hash)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET
				delta = metrics.delta + EXCLUDED.delta,
				hash = EXCLUDED.hash
		`
		args = []interface{}{metric.ID, metric.MType, metric.Value, metric.Delta, metric.Hash}
	} else if metric.MType == models.Gauge {
		query = `
			INSERT INTO metrics (id, type, value, delta, hash)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET
				value = EXCLUDED.value,
				hash = EXCLUDED.hash
		`
		args = []interface{}{metric.ID, metric.MType, metric.Value, metric.Delta, metric.Hash}
	} else {
		return nil, ErrInvalidMetricReceived
	}

	_, err = s.db.ExecContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}
	return metric, nil
}

// GetMetric получает метрику по имени из базы данных
func (s *DatabaseStorage) GetMetric(ctx context.Context, name string) (*models.Metrics, error) {
	query := `
		SELECT 
			id, type, delta, value, hash
		FROM
			metrics
		WHERE 
			id = $1
		`

	var metric models.Metrics
	err := s.db.QueryRowContext(ctx, query, name).
		Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value, &metric.Hash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &metric, nil
}

// GetAllMetrics возвращает все метрики из базы данных
func (s *DatabaseStorage) GetAllMetrics(ctx context.Context) ([]*models.Metrics, error) {
	var metrics []*models.Metrics
	query := `
		SELECT
			id, type, delta, value, hash
		FROM metrics
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.Metrics
		err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, &m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// Ping проверяет соединение с базой данных
func (s *DatabaseStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
