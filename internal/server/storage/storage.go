package storage

import (
	"context"

	"github.com/sodiqit/metricpulse.git/internal/entities"
)

type Storage interface {
	SaveGaugeMetric(ctx context.Context, metricType string, value float64) (float64, error)
	SaveCounterMetric(ctx context.Context, metricType string, value int64) (int64, error)
	GetCounterMetric(ctx context.Context, metricType string) (int64, error)
	GetGaugeMetric(ctx context.Context, metricType string) (float64, error)
	GetAllMetrics(ctx context.Context) (entities.TotalMetrics, error)
	Init(context.Context) error
	Ping(context.Context) error
	Close(context.Context) error
}
