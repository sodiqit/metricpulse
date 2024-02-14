package metricprocessor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricprocessor"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMetricProcessor_SaveMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := storage.NewMockStorage(ctrl)

	ctx := context.Background()

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
				storage.EXPECT().SaveGaugeMetric(gomock.Any(), "temp", gomock.Any()).Times(1).Return(float64(0), errors.New("error"))
				storage.EXPECT().SaveCounterMetric(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
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
				storage.EXPECT().SaveGaugeMetric(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				storage.EXPECT().SaveCounterMetric(gomock.Any(), "temp", gomock.Any()).Times(1).Return(int64(0), errors.New("error"))
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
				storage.EXPECT().SaveGaugeMetric(gomock.Any(), "temp", gomock.Any()).Times(1).Return(float64(10.5), nil)
				storage.EXPECT().SaveCounterMetric(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
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
				storage.EXPECT().SaveGaugeMetric(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				storage.EXPECT().SaveCounterMetric(gomock.Any(), "temp", gomock.Any()).Times(1).Return(int64(5), nil)
			},
			metricType:  constants.MetricTypeCounter,
			metricName:  "temp",
			metricValue: metricprocessor.MetricValue{Counter: 5},
			returnValue: metricprocessor.MetricValue{Counter: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			metricService := metricprocessor.New(storage, tt.config)
			val, err := metricService.SaveMetric(ctx, tt.metricType, tt.metricName, tt.metricValue)

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

	storage := storage.NewMockStorage(ctrl)

	ctx := context.Background()

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
				storage.EXPECT().GetGaugeMetric(gomock.Any(), "temp").Times(1).Return(float64(0), errors.New("error"))
				storage.EXPECT().GetCounterMetric(gomock.Any(), gomock.Any()).Times(0)
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
				storage.EXPECT().GetGaugeMetric(gomock.Any(), gomock.Any()).Times(0)
				storage.EXPECT().GetCounterMetric(gomock.Any(), "temp").Times(1).Return(int64(0), errors.New("error"))
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
				storage.EXPECT().GetGaugeMetric(gomock.Any(), "temp").Times(1).Return(float64(10.5), nil)
				storage.EXPECT().GetCounterMetric(gomock.Any(), gomock.Any()).Times(0)
			},
			metricType:  constants.MetricTypeGauge,
			metricName:  "temp",
			returnValue: metricprocessor.MetricValue{Gauge: 10.5},
		},
		{
			name:   "success get counter metric",
			config: &config.Config{},
			setupMock: func() {
				storage.EXPECT().GetGaugeMetric(gomock.Any(), gomock.Any()).Times(0)
				storage.EXPECT().GetCounterMetric(gomock.Any(), "temp").Times(1).Return(int64(5), nil)
			},
			metricType:  constants.MetricTypeCounter,
			metricName:  "temp",
			returnValue: metricprocessor.MetricValue{Counter: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			metricService := metricprocessor.New(storage, tt.config)
			val, err := metricService.GetMetric(ctx, tt.metricType, tt.metricName)

			if err != nil {
				require.NotNil(t, err)
				require.Equal(t, tt.err, err)
			} else {
				require.Equal(t, tt.returnValue, val)
			}
		})
	}
}
