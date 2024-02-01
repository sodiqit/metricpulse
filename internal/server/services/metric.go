package services

import (
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

type IMetricService interface {
	SaveMetric(metricType string, metricKind string, metricValue interface{})
}

type MetricService struct {
	Storage *storage.MemStorage
}

func (s *MetricService) SaveMetric(metricType string, metricKind string, metricValue interface{}) {
	switch metricType {
	case "gauge":
		if value, ok := metricValue.(float64); ok {
			s.Storage.SaveGaugeMetric(metricKind, value)
		}
	case "counter":
		if value, ok := metricValue.(int64); ok {
			s.Storage.SaveCounterMetric(metricKind, value)
		}
	}
}

func NewMetricService(storage *storage.MemStorage) MetricService {
	return MetricService{Storage: storage}
}
