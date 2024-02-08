package reporter_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"

	"github.com/sodiqit/metricpulse.git/internal/agent/reporter"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/stretchr/testify/assert"
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
	client := resty.New()

	httpmock.ActivateNonDefault(client.GetClient())

	defer httpmock.DeactivateAndReset()

	mockURL := "http://localhost:8080/update/"

	tests := []struct {
		name          string
		metrics       map[string]interface{}
		expectedCalls int
	}{
		{
			name: "Valid gauge metric",
			metrics: map[string]interface{}{
				"testGauge": float64(1.23),
			},
			expectedCalls: 1,
		},
		{
			name: "Valid counter metric",
			metrics: map[string]interface{}{
				"testCounter": int64(1),
			},
			expectedCalls: 1,
		},
		{
			name: "Unsupported metric type",
			metrics: map[string]interface{}{
				"unknown": "unsupported value",
			},
			expectedCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Reset()

			httpmock.RegisterResponder("POST", mockURL, httpmock.NewStringResponder(200, `{"id": "test"}`))

			logger, err := logger.Initialize("info")
			if err != nil {
				t.Fatalf("Error initializing logger: %s", err)
			}

			r := reporter.NewMetricReporter("localhost:8080", client, logger)

			r.SendMetrics(tt.metrics)

			assert.Equal(t, tt.expectedCalls, httpmock.GetTotalCallCount(), "Unexpected number of calls")
		})
	}
}
