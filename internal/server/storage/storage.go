package storage

import (
	"context"

	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
)

type Storage interface {
	SaveGaugeMetric(ctx context.Context, metricType string, value float64) (float64, error)
	SaveCounterMetric(ctx context.Context, metricType string, value int64) (int64, error)
	GetCounterMetric(ctx context.Context, metricType string) (int64, error)
	GetGaugeMetric(ctx context.Context, metricType string) (float64, error)
	GetAllMetrics(ctx context.Context) (entities.TotalMetrics, error)
	SaveMetricBatch(ctx context.Context, metrics []entities.Metrics) error
	Init(context.Context, retry.Backoff) error
	Ping(context.Context) error
	Close(context.Context) error
}
