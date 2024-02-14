package metricprocessor_test

import (
	"errors"
	"testing"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricprocessor"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricuploader"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMetricProcessor_SaveMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uploadService := metricuploader.NewMockUploader(ctrl)
	storage := storage.NewMockStorage(ctrl)

	tests := []struct {
		name        string
		config      *config.Config
		setupMock   func()
		metricType  string
		metricName  string
		metricValue metricprocessor.MetricValue
		returnValue metricprocessor.MetricValue
		err         error
	}{
		{
			name:        "invalid metric type",
			config:      &config.Config{},
			setupMock:   func() {},
			err:         errors.New("unsupported metricType: invalid"),
			metricType:  "invalid",
			metricName:  "temp",
			metricValue: metricprocessor.MetricValue{},
			returnValue: metricprocessor.MetricValue{},
		},
		{
			name:   "invalid gauge metric save",
			config: &config.Config{},
			setupMock: func() {
				storage.EXPECT().SaveGaugeMetric("temp", gomock.Any()).Times(1).Return(float64(0), errors.New("error"))
				storage.EXPECT().SaveCounterMetric(gomock.Any(), gomock.Any()).Times(0)
				uploadService.EXPECT().Save().Times(0)
			},
			err:         errors.New("error"),
			metricType:  constants.MetricTypeGauge,
			metricName:  "temp",
			metricValue: metricprocessor.MetricValue{},
			returnValue: metricprocessor.MetricValue{},
		},
		{
			name:   "invalid counter metric save",
			config: &config.Config{},
			setupMock: func() {
				storage.EXPECT().SaveGaugeMetric(gomock.Any(), gomock.Any()).Times(0)
				storage.EXPECT().SaveCounterMetric("temp", gomock.Any()).Times(1).Return(int64(0), errors.New("error"))
				uploadService.EXPECT().Save().Times(0)
			},
			err:         errors.New("error"),
			metricType:  constants.MetricTypeCounter,
			metricName:  "temp",
			metricValue: metricprocessor.MetricValue{},
			returnValue: metricprocessor.MetricValue{},
		},
		{
			name:   "success gauge metric save",
			config: &config.Config{StoreInterval: 10},
			setupMock: func() {
				storage.EXPECT().SaveGaugeMetric("temp", gomock.Any()).Times(1).Return(float64(10.5), nil)
				storage.EXPECT().SaveCounterMetric(gomock.Any(), gomock.Any()).Times(0)
				uploadService.EXPECT().Save().Times(0)
			},
			metricType:  constants.MetricTypeGauge,
			metricName:  "temp",
			metricValue: metricprocessor.MetricValue{Gauge: 10.5},
			returnValue: metricprocessor.MetricValue{Gauge: 10.5},
		},
		{
			name:   "success counter metric save",
			config: &config.Config{StoreInterval: 10},
			setupMock: func() {
				storage.EXPECT().SaveGaugeMetric(gomock.Any(), gomock.Any()).Times(0)
				storage.EXPECT().SaveCounterMetric("temp", gomock.Any()).Times(1).Return(int64(5), nil)
				uploadService.EXPECT().Save().Times(0)
			},
			metricType:  constants.MetricTypeCounter,
			metricName:  "temp",
			metricValue: metricprocessor.MetricValue{Counter: 5},
			returnValue: metricprocessor.MetricValue{Counter: 5},
		},
		{
			name:   "save on disk if store interval == 0",
			config: &config.Config{StoreInterval: 0},
			setupMock: func() {
				storage.EXPECT().SaveGaugeMetric(gomock.Any(), gomock.Any()).Times(0)
				storage.EXPECT().SaveCounterMetric("temp", gomock.Any()).Times(1).Return(int64(5), nil)
				uploadService.EXPECT().Save().Times(1).Return(nil)
			},
			metricType:  constants.MetricTypeCounter,
			metricName:  "temp",
			metricValue: metricprocessor.MetricValue{Counter: 5},
			returnValue: metricprocessor.MetricValue{Counter: 5},
		},
		{
			name:   "should return error if save failed",
			config: &config.Config{StoreInterval: 0},
			setupMock: func() {
				storage.EXPECT().SaveGaugeMetric(gomock.Any(), gomock.Any()).Times(0)
				storage.EXPECT().SaveCounterMetric("temp", gomock.Any()).Times(1).Return(int64(5), nil)
				uploadService.EXPECT().Save().Times(1).Return(errors.New("save error"))
			},
			err:         errors.New("save error"),
			metricType:  constants.MetricTypeCounter,
			metricName:  "temp",
			metricValue: metricprocessor.MetricValue{Counter: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			metricService := metricprocessor.New(storage, uploadService, tt.config)
			val, err := metricService.SaveMetric(tt.metricType, tt.metricName, tt.metricValue)

			if err != nil {
				require.NotNil(t, err)
				require.Equal(t, tt.err, err)
			} else {
				require.Equal(t, tt.returnValue, val)
			}
		})
	}
}

func TestMetricProcessor_GetMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uploadService := metricuploader.NewMockUploader(ctrl)
	storage := storage.NewMockStorage(ctrl)

	tests := []struct {
		name        string
		config      *config.Config
		setupMock   func()
		metricType  string
		metricName  string
		returnValue metricprocessor.MetricValue
		err         error
	}{
		{
			name:        "invalid metric type",
			config:      &config.Config{},
			setupMock:   func() {},
			err:         errors.New("unsupported metricType: invalid"),
			metricType:  "invalid",
			metricName:  "temp",
			returnValue: metricprocessor.MetricValue{},
		},
		{
			name:   "invalid gauge metric save",
			config: &config.Config{},
			setupMock: func() {
				storage.EXPECT().GetGaugeMetric("temp").Times(1).Return(float64(0), errors.New("error"))
				storage.EXPECT().GetCounterMetric(gomock.Any()).Times(0)
			},
			err:         errors.New("error"),
			metricType:  constants.MetricTypeGauge,
			metricName:  "temp",
			returnValue: metricprocessor.MetricValue{},
		},
		{
			name:   "invalid counter metric save",
			config: &config.Config{},
			setupMock: func() {
				storage.EXPECT().GetGaugeMetric(gomock.Any()).Times(0)
				storage.EXPECT().GetCounterMetric("temp").Times(1).Return(int64(0), errors.New("error"))
			},
			err:         errors.New("error"),
			metricType:  constants.MetricTypeCounter,
			metricName:  "temp",
			returnValue: metricprocessor.MetricValue{},
		},
		{
			name:   "success get gauge metric",
			config: &config.Config{},
			setupMock: func() {
				storage.EXPECT().GetGaugeMetric("temp").Times(1).Return(float64(10.5), nil)
				storage.EXPECT().GetCounterMetric(gomock.Any()).Times(0)
			},
			metricType:  constants.MetricTypeGauge,
			metricName:  "temp",
			returnValue: metricprocessor.MetricValue{Gauge: 10.5},
		},
		{
			name:   "success get counter metric",
			config: &config.Config{},
			setupMock: func() {
				storage.EXPECT().GetGaugeMetric(gomock.Any()).Times(0)
				storage.EXPECT().GetCounterMetric("temp").Times(1).Return(int64(5), nil)
			},
			metricType:  constants.MetricTypeCounter,
			metricName:  "temp",
			returnValue: metricprocessor.MetricValue{Counter: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			metricService := metricprocessor.New(storage, uploadService, tt.config)
			val, err := metricService.GetMetric(tt.metricType, tt.metricName)

			if err != nil {
				require.NotNil(t, err)
				require.Equal(t, tt.err, err)
			} else {
				require.Equal(t, tt.returnValue, val)
			}
		})
	}
}