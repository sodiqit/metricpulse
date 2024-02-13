package metricprocessor

import (
	"fmt"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricuploader"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

type MetricService interface {
	SaveMetric(metricType string, metricName string, metricValue MetricValue) (MetricValue, error)
	GetMetric(metricType string, metricName string) (MetricValue, error)
	GetAllMetrics() (entities.TotalMetrics, error)
}

type MetricProcessor struct {
	storage  storage.Storage
	uploader metricuploader.Uploader
	config   *config.Config
}

type MetricValue struct {
	Gauge   float64
	Counter int64
}

func (s *MetricProcessor) SaveMetric(metricType string, metricName string, metricValue MetricValue) (MetricValue, error) {
	var result MetricValue
	var saveErr error

	switch metricType {
	case constants.MetricTypeGauge:
		val, err := s.storage.SaveGaugeMetric(metricName, metricValue.Gauge)
		result, saveErr = MetricValue{Gauge: val}, err
	case constants.MetricTypeCounter:
		val, err := s.storage.SaveCounterMetric(metricName, metricValue.Counter)
		result, saveErr = MetricValue{Counter: val}, err
	default:
		saveErr = fmt.Errorf("unsupported metricType: %s", metricType)
	}

	if s.config.StoreInterval == 0 && saveErr == nil {
		err := s.uploader.Save()
		saveErr = err
	}

	return result, saveErr
}

func (s *MetricProcessor) GetMetric(metricType string, metricName string) (MetricValue, error) {
	switch metricType {
	case constants.MetricTypeGauge:
		val, err := s.storage.GetGaugeMetric(metricName)
		return MetricValue{Gauge: val}, err
	case constants.MetricTypeCounter:
		val, err := s.storage.GetCounterMetric(metricName)
		return MetricValue{Counter: val}, err
	}

	return MetricValue{}, fmt.Errorf("unsupported metricType: %s", metricType)
}

func (s *MetricProcessor) GetAllMetrics() (entities.TotalMetrics, error) {
	return s.storage.GetAllMetrics()
}

func New(storage storage.Storage, uploader metricuploader.Uploader, cfg *config.Config) *MetricProcessor {
	return &MetricProcessor{storage, uploader, cfg}
}
