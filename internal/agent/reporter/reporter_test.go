package reporter_test

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/sodiqit/metricpulse.git/internal/agent/reporter"
	"github.com/stretchr/testify/mock"
)

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	args := m.Called(url, contentType, body)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestMetricReporter_SendMetrics(t *testing.T) {
	tests := []struct {
		name           string
		metrics        map[string]interface{}
		expectedCalls  int
		mockResponse   *http.Response
		mockError      error
		expectingError bool
	}{
		{
			name: "Valid gauge metric",
			metrics: map[string]interface{}{
				"testGauge": float64(1.23),
			},
			expectedCalls: 1,
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			mockError:      nil,
			expectingError: false,
		},
		{
			name: "Valid counter metric",
			metrics: map[string]interface{}{
				"testCounter": int64(1),
			},
			expectedCalls: 1,
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			mockError:      nil,
			expectingError: false,
		},
		{
			name: "Unsupported metric type",
			metrics: map[string]interface{}{
				"unknown": "unsupported value",
			},
			expectedCalls:  0,
			mockResponse:   nil,
			mockError:      nil,
			expectingError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockHTTPClient)
			r := reporter.NewMetricReporter(mockClient)

			if tt.expectedCalls > 0 {
				mockClient.On("Post", mock.Anything, "text/plain", mock.Anything).Return(tt.mockResponse, tt.mockError).Times(tt.expectedCalls)
			}

			r.SendMetrics(tt.metrics)

			mockClient.AssertExpectations(t)
		})
	}
}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	os.Exit(m.Run())
}
