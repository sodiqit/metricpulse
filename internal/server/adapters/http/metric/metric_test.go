package metric_test

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/adapters/http/metric"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricprocessor"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricuploader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSyncSaveMetricsInFile(t *testing.T) {
	t.Run("should not save metric in file if storage interval > 0", func(t *testing.T) {
		r := chi.NewRouter()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
		uploadServiceMock := metricuploader.NewMockUploader(ctrl)
		logger, err := logger.Initialize("info")

		if err != nil {
			log.Fatalf(err.Error())
		}

		cfg := &config.Config{StoreInterval: 100}

		c := metric.New(metricServiceMock, logger, uploadServiceMock, cfg)

		r.Mount("/", c.Route())

		ts := httptest.NewServer(r)
		defer ts.Close()

		metricServiceMock.EXPECT().SaveMetric(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(metricprocessor.MetricValue{Gauge: 23.5}, nil)
		uploadServiceMock.EXPECT().Save().Times(0)

		client := resty.New().SetBaseURL(ts.URL).SetHeader("Content-Type", "application/json")

		resp, httpErr := client.R().SetBody(`{"id": "temp", "type": "gauge", "value": 23.5}`).Post("/update/")
		require.NoError(t, httpErr)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	})

	t.Run("should save metric in file if storage interval = 0", func(t *testing.T) {
		r := chi.NewRouter()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
		uploadServiceMock := metricuploader.NewMockUploader(ctrl)
		logger, err := logger.Initialize("info")

		if err != nil {
			log.Fatalf(err.Error())
		}

		cfg := &config.Config{StoreInterval: 0}

		c := metric.New(metricServiceMock, logger, uploadServiceMock, cfg)

		r.Mount("/", c.Route())

		ts := httptest.NewServer(r)
		defer ts.Close()

		client := resty.New().SetBaseURL(ts.URL)

		metricServiceMock.EXPECT().SaveMetric(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(metricprocessor.MetricValue{Gauge: 23.5}, nil)
		uploadServiceMock.EXPECT().Save().Times(1).Return(nil)

		resp, httpErr := client.R().SetBody(`{"id": "temp", "type": "gauge", "value": 23.5}`).SetHeader("Content-Type", "application/json").Post("/update/")
		require.NoError(t, httpErr)
		require.Equal(t, http.StatusOK, resp.StatusCode())
	})
}

func TestUpdateMetricHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := chi.NewRouter()

	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	uploadServiceMock := metricuploader.NewMockUploader(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	cfg := &config.Config{StoreInterval: 100}

	c := metric.New(metricServiceMock, logger, uploadServiceMock, cfg)

	r.Mount("/", c.Route())

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		contentType    string
		returnValue    metricprocessor.MetricValue
		expectedResult string
		expectedStatus int
	}{
		{
			name:           "invalid Content-type",
			method:         http.MethodPost,
			url:            "/update/",
			body:           "test",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid update gauge metric",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "gauge", "value": 23.5}`,
			returnValue:    metricprocessor.MetricValue{Gauge: 23.5},
			contentType:    "application/json",
			expectedResult: `{"id":"temp","type":"gauge","value":23.5}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid counter update metric",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "counter", "delta": 23}`,
			returnValue:    metricprocessor.MetricValue{Counter: 23},
			contentType:    "application/json",
			expectedResult: `{"id":"temp","type":"counter","delta":23}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid metric type",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid gauge value",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "gauge", "value": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid method",
			method:         http.MethodGet,
			url:            "/update/",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid counter value",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "counter", "delta": 23.5}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedStatus == http.StatusOK {
				metricServiceMock.EXPECT().SaveMetric(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(tc.returnValue, nil)
			}

			req := client.R().SetBody(tc.body)

			req.Method = tc.method
			req.URL = tc.url

			if tc.contentType != "" {
				req.SetHeader("Content-Type", tc.contentType)
			}

			resp, err := req.Send()

			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode())
			if tc.expectedResult != "" {
				assert.JSONEq(t, tc.expectedResult, resp.String())
			}
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	r := chi.NewRouter()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	uploadServiceMock := metricuploader.NewMockUploader(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	cfg := &config.Config{}

	c := metric.New(metricServiceMock, logger, uploadServiceMock, cfg)

	r.Mount("/", c.Route())

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	tests := []struct {
		name           string
		method         string
		url            string
		setupMock      func()
		body           string
		contentType    string
		expectedResult string
		expectedStatus int
	}{
		{
			name:           "invalid Content-type",
			method:         http.MethodPost,
			url:            "/value/",
			setupMock:      func() {},
			body:           "test",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "valid gauge metric",
			method: http.MethodPost,
			url:    "/value/",
			setupMock: func() {
				metricServiceMock.EXPECT().GetMetric(constants.MetricTypeGauge, "temp").Times(1).Return(metricprocessor.MetricValue{Gauge: 100.156}, nil)
			},
			contentType:    "application/json",
			body:           `{"id": "temp", "type": "gauge"}`,
			expectedResult: `{"id": "temp", "type": "gauge", "value": 100.156}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "valid counter metric",
			method:      http.MethodPost,
			url:         "/value/",
			contentType: "application/json",
			setupMock: func() {
				metricServiceMock.EXPECT().GetMetric(constants.MetricTypeCounter, "temp").Times(1).Return(metricprocessor.MetricValue{Counter: 100}, nil)
			},
			body:           `{"id": "temp", "type": "counter"}`,
			expectedResult: `{"id": "temp", "type": "counter", "delta": 100}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "not found metric",
			method:      http.MethodPost,
			url:         "/value/",
			body:        `{"id": "temp", "type": "gauge"}`,
			contentType: "application/json",
			setupMock: func() {
				metricServiceMock.EXPECT().GetMetric(constants.MetricTypeGauge, "temp").Times(1).Return(metricprocessor.MetricValue{}, errors.New("not found metric"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid method",
			method:         http.MethodGet,
			url:            "/value/",
			setupMock:      func() {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid metric type",
			method:         http.MethodPost,
			url:            "/value/",
			contentType:    "application/json",
			setupMock:      func() {},
			body:           `{"id": "temp", "type": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			req := client.R().SetBody(tc.body)

			req.Method = tc.method
			req.URL = tc.url

			if tc.contentType != "" {
				req.SetHeader("Content-Type", tc.contentType)
			}

			resp, err := req.Send()
			require.NoError(t, err)

			require.Equal(t, tc.expectedStatus, resp.StatusCode())

			if tc.expectedStatus == http.StatusOK {
				assert.JSONEq(t, tc.expectedResult, resp.String())
			}
		})
	}
}

func TestTextUpdateMetricHandler(t *testing.T) {
	r := chi.NewRouter()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	uploadServiceMock := metricuploader.NewMockUploader(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	cfg := &config.Config{StoreInterval: 100}

	c := metric.New(metricServiceMock, logger, uploadServiceMock, cfg)

	r.Mount("/", c.Route())

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "valid gauge metric",
			method:         http.MethodPost,
			url:            "/update/gauge/temp/23.5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid counter metric",
			method:         http.MethodPost,
			url:            "/update/counter/temp/10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid method",
			method:         http.MethodGet,
			url:            "/update/gauge/temp/23.5",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid metric type",
			method:         http.MethodPost,
			url:            "/update/invalid/temp/23.5",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedStatus == http.StatusOK {
				metricServiceMock.EXPECT().SaveMetric(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(metricprocessor.MetricValue{}, nil)
			}

			req := client.R()

			req.URL = tc.url
			req.Method = tc.method

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode())
		})
	}
}

func TestTextGetMetricHandler(t *testing.T) {
	r := chi.NewRouter()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	uploadServiceMock := metricuploader.NewMockUploader(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	cfg := &config.Config{}

	c := metric.New(metricServiceMock, logger, uploadServiceMock, cfg)

	r.Mount("/", c.Route())

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	tests := []struct {
		name           string
		method         string
		url            string
		setupMock      func()
		expectedResult string
		expectedStatus int
	}{
		{
			name:   "valid gauge metric",
			method: http.MethodGet,
			url:    "/value/gauge/temp",
			setupMock: func() {
				metricServiceMock.EXPECT().GetMetric(constants.MetricTypeGauge, "temp").Times(1).Return(metricprocessor.MetricValue{Gauge: 100.156}, nil)
			},
			expectedResult: "100.156",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "valid counter metric",
			method: http.MethodGet,
			url:    "/value/counter/temp",
			setupMock: func() {
				metricServiceMock.EXPECT().GetMetric(constants.MetricTypeCounter, "temp").Times(1).Return(metricprocessor.MetricValue{Counter: 100}, nil)
			},
			expectedResult: "100",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "not found metric",
			method: http.MethodGet,
			url:    "/value/gauge/temp",
			setupMock: func() {
				metricServiceMock.EXPECT().GetMetric(constants.MetricTypeGauge, "temp").Times(1).Return(metricprocessor.MetricValue{}, errors.New("not found metric"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid method",
			method:         http.MethodPost,
			url:            "/value/gauge/temp",
			setupMock:      func() {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid metric type",
			method:         http.MethodGet,
			url:            "/value/invalid/temp",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			req := client.R()

			req.Method = tc.method
			req.URL = tc.url

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode())

			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, tc.expectedResult, resp.String())
			}
		})
	}
}

func TestGetAllMetricsHandler(t *testing.T) {
	r := chi.NewRouter()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	uploadServiceMock := metricuploader.NewMockUploader(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	cfg := &config.Config{}

	c := metric.New(metricServiceMock, logger, uploadServiceMock, cfg)

	r.Mount("/", c.Route())

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	tests := []struct {
		name                string
		method              string
		url                 string
		setupMock           func()
		expectedContentType string
		expectedStatus      int
	}{
		{
			name:   "valid result",
			method: http.MethodGet,
			url:    "/",
			setupMock: func() {
				metricServiceMock.EXPECT().GetAllMetrics().Times(1).Return(entities.TotalMetrics{}, nil)
			},
			expectedContentType: "text/html",
			expectedStatus:      http.StatusOK,
		},
		{
			name:           "Invalid method",
			method:         http.MethodPost,
			url:            "/",
			setupMock:      func() {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			req := client.R()

			req.Method = tc.method
			req.URL = tc.url

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode())

			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, tc.expectedContentType, resp.Header().Get("Content-Type"))
			}
		})
	}
}
