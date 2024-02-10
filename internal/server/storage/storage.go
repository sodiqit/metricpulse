package storage

import (
	"fmt"
)

type MemStorage struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
}

func (m *MemStorage) SaveGaugeMetric(metricType string, value float64) float64 {
	m.Gauge[metricType] = value
	return value
}

func (m *MemStorage) SaveCounterMetric(metricType string, value int64) int64 {
	val, ok := m.Counter[metricType]

	if ok {
		m.Counter[metricType] = val + value
	} else {
		m.Counter[metricType] = value
	}

	return m.Counter[metricType]
}

func (m *MemStorage) GetGaugeMetric(metricName string) (float64, error) {
	val, ok := m.Gauge[metricName]

	if ok {
		return val, nil
	} else {
		return 0, fmt.Errorf("not found metric: %s", metricName)
	}
}

func (m *MemStorage) GetCounterMetric(metricName string) (int64, error) {
	val, ok := m.Counter[metricName]

	if ok {
		return val, nil
	} else {
		return 0, fmt.Errorf("not found metric: %s", metricName)
	}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}
