package services

import (
	"errors"
	"fmt"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

type IMetricService interface {
	SaveMetric(metricType string, metricName string, metricValue MetricValue) (MetricValue, error)
	GetMetric(metricType string, metricName string) (MetricValue, error)
	GetAllMetrics() *storage.MemStorage
}

type MetricService struct {
	Storage *storage.MemStorage
}

type MetricValue struct {
	Gauge   float64
	Counter int64
}

func (s *MetricService) SaveMetric(metricType string, metricName string, metricValue MetricValue) (MetricValue, error) {
	switch metricType {
	case constants.MetricTypeGauge:
		return MetricValue{Gauge: s.Storage.SaveGaugeMetric(metricName, metricValue.Gauge)}, nil
	case constants.MetricTypeCounter:
		return MetricValue{Counter: s.Storage.SaveCounterMetric(metricName, metricValue.Counter)}, nil
	}

	return MetricValue{}, fmt.Errorf("unsupported metricType: %s", metricType)
}

func (s *MetricService) GetMetric(metricType string, metricName string) (MetricValue, error) {
	switch metricType {
	case constants.MetricTypeGauge:
		val, err := s.Storage.GetGaugeMetric(metricName)
		return MetricValue{Gauge: val}, err
	case constants.MetricTypeCounter:
		val, err := s.Storage.GetCounterMetric(metricName)
		return MetricValue{Counter: val}, err
	}

	return MetricValue{}, errors.New("not correct metric type")
}

func (s *MetricService) GetAllMetrics() *storage.MemStorage {
	return s.Storage
}

func NewMetricService(storage *storage.MemStorage) MetricService {
	return MetricService{Storage: storage}
}
