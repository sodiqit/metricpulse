package metricprocessor

import (
	"context"
	"fmt"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

type MetricService interface {
	SaveMetric(ctx context.Context, metricType string, metricName string, metricValue MetricValue) (MetricValue, error)
	GetMetric(ctx context.Context, metricType string, metricName string) (MetricValue, error)
	GetAllMetrics(ctx context.Context) (entities.TotalMetrics, error)
}

type MetricProcessor struct {
	storage storage.Storage
	config  *config.Config
}

type MetricValue struct {
	Gauge   float64
	Counter int64
}

func (s *MetricProcessor) SaveMetric(ctx context.Context, metricType string, metricName string, metricValue MetricValue) (MetricValue, error) {
	var result MetricValue
	var saveErr error

	switch metricType {
	case constants.MetricTypeGauge:
		val, err := s.storage.SaveGaugeMetric(ctx, metricName, metricValue.Gauge)
		result, saveErr = MetricValue{Gauge: val}, err
	case constants.MetricTypeCounter:
		val, err := s.storage.SaveCounterMetric(ctx, metricName, metricValue.Counter)
		result, saveErr = MetricValue{Counter: val}, err
	default:
		saveErr = fmt.Errorf("unsupported metricType: %s", metricType)
	}

	return result, saveErr
}

func (s *MetricProcessor) GetMetric(ctx context.Context, metricType string, metricName string) (MetricValue, error) {
	switch metricType {
	case constants.MetricTypeGauge:
		val, err := s.storage.GetGaugeMetric(ctx, metricName)
		return MetricValue{Gauge: val}, err
	case constants.MetricTypeCounter:
		val, err := s.storage.GetCounterMetric(ctx, metricName)
		return MetricValue{Counter: val}, err
	}

	return MetricValue{}, fmt.Errorf("unsupported metricType: %s", metricType)
}

func (s *MetricProcessor) GetAllMetrics(ctx context.Context) (entities.TotalMetrics, error) {
	return s.storage.GetAllMetrics(ctx)
}

func New(storage storage.Storage, cfg *config.Config) *MetricProcessor {
	return &MetricProcessor{storage, cfg}
}
