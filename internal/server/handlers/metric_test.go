package handlers_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sodiqit/metricpulse.git/internal/server/handlers"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MetricServiceMock struct {
	mock.Mock
}

func (m *MetricServiceMock) SaveMetric(metricType, metricKind string, val services.MetricValue) {
	m.Called(metricType, metricKind, val)
}

func (m *MetricServiceMock) GetMetric(metricType, metricKind string) (services.MetricValue, error) {
	args := m.Called(metricType, metricKind)
	return args.Get(0).(services.MetricValue), args.Error(1)
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestUpdateMetricHandler(t *testing.T) {
	r := chi.NewRouter()

	metricServiceMock := new(MetricServiceMock)

	handlers.RegisterMetricRouter(r, metricServiceMock)

	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "Valid gauge metric",
			method:         http.MethodPost,
			url:            "/update/gauge/temp/23.5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid counter metric",
			method:         http.MethodPost,
			url:            "/update/counter/temp/10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			url:            "/update/gauge/temp/23.5",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid metric type",
			method:         http.MethodPost,
			url:            "/update/invalid/temp/23.5",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedStatus == http.StatusOK {
				metricServiceMock.On("SaveMetric", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			}

			resp, _ := testRequest(t, ts, tc.method, tc.url)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			metricServiceMock.AssertExpectations(t)
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	r := chi.NewRouter()

	metricServiceMock := new(MetricServiceMock)

	handlers.RegisterMetricRouter(r, metricServiceMock)

	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name           string
		method         string
		url            string
		setupMock      func()
		expectedResult string
		expectedStatus int
	}{
		{
			name:   "Valid gauge metric",
			method: http.MethodGet,
			url:    "/value/gauge/temp",
			setupMock: func() {
				metricServiceMock.On("GetMetric", mock.Anything, mock.Anything).Once().Return(services.MetricValue{Gauge: 100.156}, nil)
			},
			expectedResult: "100.156",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Valid counter metric",
			method: http.MethodGet,
			url:    "/value/counter/temp",
			setupMock: func() {
				metricServiceMock.On("GetMetric", mock.Anything, mock.Anything).Once().Return(services.MetricValue{Counter: 100}, nil)
			},
			expectedResult: "100",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Not found metric",
			method: http.MethodGet,
			url:    "/value/gauge/temp",
			setupMock: func() {
				metricServiceMock.On("GetMetric", mock.Anything, mock.Anything).Once().Return(services.MetricValue{}, errors.New("not found metric"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid method",
			method:         http.MethodPost,
			url:            "/value/gauge/temp",
			setupMock:      func() {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Invalid metric type",
			method:         http.MethodGet,
			url:            "/value/invalid/temp",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			resp, body := testRequest(t, ts, tc.method, tc.url)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			metricServiceMock.AssertExpectations(t)

			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, tc.expectedResult, body)
			}
		})
	}
}
