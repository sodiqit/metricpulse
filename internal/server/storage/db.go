package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
)

var ErrNotConnection = errors.New("not connection. Maybe not invoked Init() method")

type PostgresStorage struct {
	cfg  *config.Config
	conn *pgx.Conn
}

func (s *PostgresStorage) SaveGaugeMetric(metricType string, value float64) (float64, error) {
	return 0, nil
}

func (s *PostgresStorage) SaveCounterMetric(metricType string, value int64) (int64, error) {
	return 0, nil
}

func (s *PostgresStorage) GetGaugeMetric(metricName string) (float64, error) {
	return 0, nil
}

func (s *PostgresStorage) GetCounterMetric(metricName string) (int64, error) {
	return 0, nil
}

func (s *PostgresStorage) GetAllMetrics() (entities.TotalMetrics, error) {
	return entities.TotalMetrics{}, nil
}

func (s *PostgresStorage) Init(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, s.cfg.DatabaseDSN)
	s.conn = conn

	return err
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	if s.conn != nil {
		return ErrNotConnection
	}

	return s.conn.Ping(ctx)
}

func (s *PostgresStorage) Close(ctx context.Context) error {
	if s.conn != nil {
		return ErrNotConnection
	}

	return s.conn.Close(ctx)
}

func NewPostgresStorage(cfg *config.Config) *PostgresStorage {
	return &PostgresStorage{cfg, nil}
}
