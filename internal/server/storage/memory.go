package storage

import (
	"context"
	"errors"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func (m *MemStorage) SaveGaugeMetric(ctx context.Context, metricType string, value float64) (float64, error) {
	m.gauge[metricType] = value
	return value, nil
}

func (m *MemStorage) SaveCounterMetric(ctx context.Context, metricType string, value int64) (int64, error) {
	val, ok := m.counter[metricType]

	if ok {
		m.counter[metricType] = val + value
	} else {
		m.counter[metricType] = value
	}

	return m.counter[metricType], nil
}

func (m *MemStorage) GetGaugeMetric(ctx context.Context, metricName string) (float64, error) {
	val, ok := m.gauge[metricName]

	if ok {
		return val, nil
	} else {
		return val, NewErrNotFound(errors.New("not found metric"), map[string]interface{}{"metricName": metricName})
	}
}

func (m *MemStorage) GetCounterMetric(ctx context.Context, metricName string) (int64, error) {
	val, ok := m.counter[metricName]

	if ok {
		return val, nil
	} else {
		return val, NewErrNotFound(errors.New("not found metric"), map[string]interface{}{"metricName": metricName})
	}
}

func (m *MemStorage) GetAllMetrics(ctx context.Context) (entities.TotalMetrics, error) {
	return entities.TotalMetrics{Gauge: m.gauge, Counter: m.counter}, nil
}

func (m *MemStorage) SaveMetricBatch(ctx context.Context, metrics []entities.Metrics) error {
	for _, metric := range metrics {
		if metric.MType == constants.MetricTypeGauge {
			m.SaveGaugeMetric(ctx, metric.ID, *metric.Value)
		} else {
			m.SaveCounterMetric(ctx, metric.ID, *metric.Delta)
		}
	}

	return nil
}

func (m *MemStorage) InitMetrics(metrics entities.TotalMetrics) error {
	m.counter = metrics.Counter
	m.gauge = metrics.Gauge
	return nil
}

func (m *MemStorage) Init(context.Context, retry.Backoff) error {
	return nil
}

func (m *MemStorage) Ping(context.Context) error {
	return nil
}

func (m *MemStorage) Close(context.Context) error {
	return nil
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}
