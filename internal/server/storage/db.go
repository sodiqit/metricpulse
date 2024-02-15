package storage

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
)

var ErrNotConnection = errors.New("not connection. Maybe not invoked Init() method")

func getUpdateMetricQuery(metricType string) string {
	baseQuery := `
		INSERT INTO metric
			(type, name, value)
		VALUES
			(@type, @name, @value)
		ON CONFLICT(type, name)
	`

	switch metricType {
	case constants.MetricTypeGauge:
		return baseQuery + " DO UPDATE SET value = EXCLUDED.value RETURNING value;"
	case constants.MetricTypeCounter:
		return baseQuery + " DO UPDATE SET value = metric.value + EXCLUDED.value RETURNING value;"
	default:
		return baseQuery
	}
}

var selectMetricQuery = `SELECT value FROM metric WHERE type = @type AND name = @name`

type rawMetric struct {
	ID    int
	MType string `db:"type"`
	Name  string
	Value float64
}

type PostgresStorage struct {
	cfg    *config.Config
	logger logger.ILogger
	pool   *pgxpool.Pool
}

func (s *PostgresStorage) SaveGaugeMetric(ctx context.Context, metricType string, value float64) (float64, error) {
	if s.pool == nil {
		return 0, ErrNotConnection
	}

	var result float64

	err := s.pool.QueryRow(ctx, getUpdateMetricQuery(constants.MetricTypeGauge), pgx.NamedArgs{"type": constants.MetricTypeGauge, "value": value, "name": metricType}).Scan(&result)

	if err != nil {
		s.logger.Errorw("error while save gauge metric", "error", err, "metricName", metricType, "value", value)
	}

	return result, err
}

func (s *PostgresStorage) SaveCounterMetric(ctx context.Context, metricType string, value int64) (int64, error) {
	var result int64

	err := s.pool.QueryRow(ctx, getUpdateMetricQuery(constants.MetricTypeCounter), pgx.NamedArgs{"type": constants.MetricTypeCounter, "value": value, "name": metricType}).Scan(&result)

	if err != nil {
		s.logger.Errorw("error while save counter metric", "error", err, "metricName", metricType, "value", value)
	}

	return result, err
}

func (s *PostgresStorage) SaveMetricBatch(ctx context.Context, metrics []entities.Metrics) error {
	batch := &pgx.Batch{}

	for _, metric := range metrics {
		var value float64

		if metric.MType == constants.MetricTypeGauge {
			value = *metric.Value
		} else {
			value = float64(*metric.Delta)
		}

		batch.Queue(getUpdateMetricQuery(metric.MType), pgx.NamedArgs{"type": metric.MType, "name": metric.ID, "value": value})
	}

	err := s.pool.SendBatch(ctx, batch).Close()

	return err
}

func (s *PostgresStorage) GetGaugeMetric(ctx context.Context, metricName string) (float64, error) {
	var result float64

	err := s.pool.QueryRow(ctx, selectMetricQuery, pgx.NamedArgs{"type": constants.MetricTypeGauge, "name": metricName}).Scan(&result)

	if err != nil {
		s.logger.Errorw("error while get gauge metric", "error", err, "metricName", metricName)
	}

	return result, err
}

func (s *PostgresStorage) GetCounterMetric(ctx context.Context, metricName string) (int64, error) {
	var result int64

	err := s.pool.QueryRow(ctx, selectMetricQuery, pgx.NamedArgs{"type": constants.MetricTypeCounter, "name": metricName}).Scan(&result)

	if err != nil {
		s.logger.Errorw("error while get gauge metric", "error", err, "metricName", metricName)
	}

	return result, err
}

func (s *PostgresStorage) GetAllMetrics(ctx context.Context) (entities.TotalMetrics, error) {
	var rawResult []rawMetric

	err := pgxscan.Select(ctx, s.pool, &rawResult, "SELECT * FROM metric")

	if err != nil {
		s.logger.Errorw("error while get all metrics", "error", err)
		return entities.TotalMetrics{}, err
	}

	result := entities.TotalMetrics{Gauge: make(map[string]float64), Counter: make(map[string]int64)}

	for _, rawMetric := range rawResult {
		if rawMetric.MType == constants.MetricTypeGauge {
			result.Gauge[rawMetric.Name] = rawMetric.Value
		} else {
			result.Counter[rawMetric.Name] = int64(rawMetric.Value)
		}
	}

	return result, nil
}

func (s *PostgresStorage) Init(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, s.cfg.DatabaseDSN)

	if err != nil {
		return err
	}

	s.pool = pool

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	tx.Exec(ctx, `
		CREATE TABLE metric (
			id serial PRIMARY KEY,
			type varchar(128) NOT NULL,
			name varchar(128) NOT NULL,
			value double precision NOT NULL
		)
	`)

	tx.Exec(ctx, "CREATE UNIQUE INDEX idx_type_name ON metric(type, name)")

	tx.Commit(ctx)

	return err
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	if s.pool == nil {
		return ErrNotConnection
	}

	return s.pool.Ping(ctx)
}

func (s *PostgresStorage) Close(ctx context.Context) error {
	if s.pool == nil {
		return ErrNotConnection
	}

	s.pool.Close()

	return nil
}

func NewPostgresStorage(cfg *config.Config, logger logger.ILogger) *PostgresStorage {
	return &PostgresStorage{cfg, logger, nil}
}
