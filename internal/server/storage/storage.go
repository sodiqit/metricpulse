package storage

import (
	"context"

	"github.com/sodiqit/metricpulse.git/internal/entities"
)

type Storage interface {
	SaveGaugeMetric(metricType string, value float64) (float64, error)
	SaveCounterMetric(metricType string, value int64) (int64, error)
	GetCounterMetric(metricType string) (int64, error)
	GetGaugeMetric(metricType string) (float64, error)
	GetAllMetrics() (entities.TotalMetrics, error)
	Init(context.Context) error
	Ping(context.Context) error
	Close(context.Context) error
}
