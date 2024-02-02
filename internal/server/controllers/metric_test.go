package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sodiqit/metricpulse.git/internal/server/controllers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MetricServiceMock struct {
	mock.Mock
}

func (m *MetricServiceMock) SaveMetric(metricType, metricKind string, val interface{}) {
	m.Called(metricType, metricKind, val)
}

func TestUpdateMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           string
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
			metricServiceMock := new(MetricServiceMock)
			handler := controllers.UpdateMetricHandler(metricServiceMock)

			if tc.expectedStatus == http.StatusOK {
				metricServiceMock.On("SaveMetric", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			}

			req, err := http.NewRequest(tc.method, tc.url, bytes.NewBufferString(tc.body))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			metricServiceMock.AssertExpectations(t)
		})
	}
}
