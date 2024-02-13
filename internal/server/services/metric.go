package services

import (
	"errors"
	"fmt"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

type IMetricService interface {
	SaveMetric(metricType string, metricName string, metricValue MetricValue) (MetricValue, error)
	GetMetric(metricType string, metricName string) (MetricValue, error)
	GetAllMetrics() (entities.TotalMetrics, error)
}

type MetricService struct {
	storage storage.IStorage
}

type MetricValue struct {
	Gauge   float64
	Counter int64
}

func (s *MetricService) SaveMetric(metricType string, metricName string, metricValue MetricValue) (MetricValue, error) {
	switch metricType {
	case constants.MetricTypeGauge:
		val, err := s.storage.SaveGaugeMetric(metricName, metricValue.Gauge)

		if err != nil {
			return MetricValue{}, err
		}

		return MetricValue{Gauge: val}, nil
	case constants.MetricTypeCounter:
		val, err := s.storage.SaveCounterMetric(metricName, metricValue.Counter)

		if err != nil {
			return MetricValue{}, err
		}

		return MetricValue{Counter: val}, nil
	}

	return MetricValue{}, fmt.Errorf("unsupported metricType: %s", metricType)
}

func (s *MetricService) GetMetric(metricType string, metricName string) (MetricValue, error) {
	switch metricType {
	case constants.MetricTypeGauge:
		val, err := s.storage.GetGaugeMetric(metricName)
		return MetricValue{Gauge: val}, err
	case constants.MetricTypeCounter:
		val, err := s.storage.GetCounterMetric(metricName)
		return MetricValue{Counter: val}, err
	}

	return MetricValue{}, errors.New("not correct metric type")
}

func (s *MetricService) GetAllMetrics() (entities.TotalMetrics, error) {
	return s.storage.GetAllMetrics()
}

func NewMetricService(storage storage.IStorage) *MetricService {
	return &MetricService{storage}
}
