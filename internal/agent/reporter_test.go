package agent_test

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"go.uber.org/mock/gomock"

	"github.com/sodiqit/metricpulse.git/internal/agent"
	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
	"github.com/sodiqit/metricpulse.git/pkg/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricReporter_SendMetrics(t *testing.T) {
	client := resty.New()

	httpmock.ActivateNonDefault(client.GetClient())

	defer httpmock.DeactivateAndReset()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signerMock := signer.NewMockSigner(ctrl)

	mockURL := "http://localhost:8080/updates/"

	tests := []struct {
		name               string
		signer             signer.Signer
		expectResponseBody string
		expectedCalls      int
	}{
		{
			name:               "Valid metric snapshot",
			signer:             signerMock,
			expectResponseBody: `[{"id":"TestCounter","type":"counter","delta":1},{"id":"TestGauge","type":"gauge","value":1}]`,
			expectedCalls:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.Reset()

			if tt.signer != nil {
				s, _ := tt.signer.(*signer.MockSigner)
				s.EXPECT().Sign(gomock.Any()).Times(1).Return("test-signature")
			}

			httpmock.RegisterResponder("POST", mockURL, func(req *http.Request) (*http.Response, error) {
				contentEncoding := req.Header.Get("Content-Encoding")
				signature := req.Header.Get(constants.HashHeader)
				isSendsGzip := strings.Contains(contentEncoding, "gzip")
				if !isSendsGzip {
					t.Errorf("Expected Content-Encoding header', got '%s'", contentEncoding)
				}

				if tt.signer != nil && signature == "" {
					t.Errorf("Expected %s header', got '%s'", constants.HashHeader, signature)
				}

				data, err := gzip.NewReader(req.Body)
				require.NoError(t, err)

				res, err := io.ReadAll(data)
				require.NoError(t, err)

				require.JSONEq(t, tt.expectResponseBody, string(res))
				return httpmock.NewStringResponse(200, "OK"), nil
			})

			logger, err := logger.Initialize("info")
			if err != nil {
				t.Fatalf("Error initializing logger: %s", err)
			}

			scope := agent.NewRootScope()

			testGauge := scope.Gauge("TestGauge")
			testGauge.Update(1)

			testCounter := scope.Counter("TestCounter")
			testCounter.Inc(1)

			snapshot := scope.Snapshot()

			options := agent.MetricReporterOptions{
				ServerAddr:     "localhost:8080",
				Scope:          scope,
				Client:         client,
				ReportInterval: 0,
				RateLimit:      2,
				Logger:         logger,
				Signer:         tt.signer,
			}

			r := agent.NewMetricReporter(options)

			r.SendBatchMetrics(context.Background(), snapshot, retry.EmptyBackoff)

			assert.Equal(t, tt.expectedCalls, httpmock.GetTotalCallCount(), "Unexpected number of calls")
		})
	}
}
