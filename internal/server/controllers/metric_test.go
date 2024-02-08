package controllers_test

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/controllers"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MetricServiceMock struct {
	mock.Mock
}

func (m *MetricServiceMock) SaveMetric(metricType, metricKind string, val services.MetricValue) (services.MetricValue, error) {
	args := m.Called(metricType, metricKind, val)
	return args.Get(0).(services.MetricValue), args.Error(1)
}

func (m *MetricServiceMock) GetMetric(metricType, metricKind string) (services.MetricValue, error) {
	args := m.Called(metricType, metricKind)
	return args.Get(0).(services.MetricValue), args.Error(1)
}

func (m *MetricServiceMock) GetAllMetrics() *storage.MemStorage {
	args := m.Called()
	return args.Get(0).(*storage.MemStorage)
}

func TestUpdateMetricHandler(t *testing.T) {
	r := chi.NewRouter()

	metricServiceMock := new(MetricServiceMock)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := controllers.NewMetricController(metricServiceMock, logger)

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
		returnValue    services.MetricValue
		expectedResult string
		expectedStatus int
	}{
		{
			name:           "Invalid Content-type",
			method:         http.MethodPost,
			url:            "/update/",
			body:           "test",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Valid gauge metric",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "gauge", "value": 23.5}`,
			returnValue:    services.MetricValue{Gauge: 23.5},
			contentType:    "application/json",
			expectedResult: `{"id":"temp","type":"gauge","value":23.5,"delta":0}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid counter metric",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "counter", "delta": 23}`,
			returnValue:    services.MetricValue{Counter: 23},
			contentType:    "application/json",
			expectedResult: `{"id":"temp","type":"counter","delta":23,"value":0}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid metric type",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid gauge value",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "gauge", "value": "invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			url:            "/update/",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid counter value",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id": "temp", "type": "counter", "delta": 23.5}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedStatus == http.StatusOK {
				metricServiceMock.On("SaveMetric", mock.Anything, mock.Anything, mock.Anything).Once().Return(tc.returnValue, nil)
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
			metricServiceMock.AssertExpectations(t)
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	r := chi.NewRouter()

	metricServiceMock := new(MetricServiceMock)

	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := controllers.NewMetricController(metricServiceMock, logger)

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
			name:           "Invalid Content-type",
			method:         http.MethodPost,
			url:            "/value/",
			setupMock:      func() {},
			body:           "test",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Valid gauge metric",
			method: http.MethodPost,
			url:    "/value/",
			setupMock: func() {
				metricServiceMock.On("GetMetric", mock.Anything, mock.Anything).Once().Return(services.MetricValue{Gauge: 100.156}, nil)
			},
			contentType:    "application/json",
			body:           `{"id": "temp", "type": "gauge"}`,
			expectedResult: `{"id": "temp", "type": "gauge", "value": 100.156, "delta": 0}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Valid counter metric",
			method:      http.MethodPost,
			url:         "/value/",
			contentType: "application/json",
			setupMock: func() {
				metricServiceMock.On("GetMetric", mock.Anything, mock.Anything).Once().Return(services.MetricValue{Counter: 100}, nil)
			},
			body:           `{"id": "temp", "type": "counter"}`,
			expectedResult: `{"id": "temp", "type": "counter", "value": 0, "delta": 100}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Not found metric",
			method:      http.MethodPost,
			url:         "/value/",
			body:        `{"id": "temp", "type": "gauge"}`,
			contentType: "application/json",
			setupMock: func() {
				metricServiceMock.On("GetMetric", mock.Anything, mock.Anything).Once().Return(services.MetricValue{}, errors.New("not found metric"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			url:            "/value/",
			setupMock:      func() {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid metric type",
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
			metricServiceMock.AssertExpectations(t)

			if tc.expectedStatus == http.StatusOK {
				assert.JSONEq(t, tc.expectedResult, resp.String())
			}
		})
	}
}

func TestGetAllMetricsHandler(t *testing.T) {
	r := chi.NewRouter()

	metricServiceMock := new(MetricServiceMock)

	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := controllers.NewMetricController(metricServiceMock, logger)

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
			name:   "Valid result",
			method: http.MethodGet,
			url:    "/",
			setupMock: func() {
				metricServiceMock.On("GetAllMetrics").Once().Return(&storage.MemStorage{})
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
			metricServiceMock.AssertExpectations(t)

			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, tc.expectedContentType, resp.Header().Get("Content-Type"))
			}
		})
	}
}
