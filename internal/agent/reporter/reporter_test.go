package reporter_test

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"

	"github.com/sodiqit/metricpulse.git/internal/agent/reporter"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

	mockURL := "http://localhost:8080/updates/"

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

			httpmock.RegisterResponder("POST", mockURL, func(req *http.Request) (*http.Response, error) {
				contentEncoding := req.Header.Get("Content-Encoding")
				isSendsGzip := strings.Contains(contentEncoding, "gzip")
				if !isSendsGzip {
					t.Errorf("Expected Content-Encoding header', got '%s'", contentEncoding)
				}

				_, err := gzip.NewReader(req.Body)
				require.NoError(t, err)
				return httpmock.NewStringResponse(200, "OK"), nil
			})

			logger, err := logger.Initialize("info")
			if err != nil {
				t.Fatalf("Error initializing logger: %s", err)
			}

			r := reporter.NewMetricReporter("localhost:8080", client, logger)

			r.SendMetrics(context.Background(), tt.metrics, retry.EmptyBackoff)

			assert.Equal(t, tt.expectedCalls, httpmock.GetTotalCallCount(), "Unexpected number of calls")
		})
	}
}
