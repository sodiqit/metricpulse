package storage

import (
	"context"
	"fmt"

	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
)

type ErrNotFound struct {
	args map[string]interface{}
	err  error
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("not found rows with args: %#v", e.args)
}

func (e *ErrNotFound) Unwrap() error {
	return e.err
}

func IsErrNotFound(err error) bool {
	_, ok := err.(*ErrNotFound)
	return ok
}

func NewErrNotFound(err error, args map[string]interface{}) error {
	return &ErrNotFound{
		args,
		err,
	}
}

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
