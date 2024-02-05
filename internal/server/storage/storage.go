package storage

import (
	"fmt"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string][]int64
}

func (m *MemStorage) SaveGaugeMetric(metricType string, value float64) {
	m.gauge[metricType] = value
}

func (m *MemStorage) SaveCounterMetric(metricType string, value int64) {
	val, ok := m.counter[metricType]

	if ok {
		m.counter[metricType] = append(val, value)
	} else {
		m.counter[metricType] = []int64{value}
	}
}

func (m *MemStorage) GetGaugeMetric(metricName string) (float64, error) {
	val, ok := m.gauge[metricName]

	if ok {
		return val, nil
	} else {
		return 0, fmt.Errorf("not found metric: %s", metricName)
	}
}

func (m *MemStorage) GetCounterMetric(metricName string) (int64, error) {
	val, ok := m.counter[metricName]

	if ok {
		return val[len(val)-1], nil
	} else {
		return 0, fmt.Errorf("not found metric: %s", metricName)
	}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string][]int64),
	}
}
