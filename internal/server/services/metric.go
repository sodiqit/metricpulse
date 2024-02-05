package services

import (
	"errors"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

type IMetricService interface {
	SaveMetric(metricType string, metricName string, metricValue MetricValue)
	GetMetric(metricType string, metricName string) (MetricValue, error)
}

type MetricService struct {
	Storage *storage.MemStorage
}

type MetricValue struct {
	Gauge   float64
	Counter int64
}

func (s *MetricService) SaveMetric(metricType string, metricName string, metricValue MetricValue) {
	switch metricType {
	case constants.MetricTypeGauge:
		s.Storage.SaveGaugeMetric(metricName, metricValue.Gauge)
	case constants.MetricTypeCounter:
		s.Storage.SaveCounterMetric(metricName, metricValue.Counter)
	}
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

func NewMetricService(storage *storage.MemStorage) MetricService {
	return MetricService{Storage: storage}
}
