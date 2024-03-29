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
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricprocessor"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
	"github.com/sodiqit/metricpulse.git/pkg/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUpdateMetricHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := chi.NewRouter()

	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	storageMock := storage.NewMockStorage(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := metric.New(metricServiceMock, storageMock, logger, nil)

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
				metricServiceMock.EXPECT().SaveMetric(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(tc.returnValue, nil)
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

	storageMock := storage.NewMockStorage(ctrl)
	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := metric.New(metricServiceMock, storageMock, logger, nil)

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
				metricServiceMock.EXPECT().GetMetric(gomock.Any(), constants.MetricTypeGauge, "temp").Times(1).Return(metricprocessor.MetricValue{Gauge: 100.156}, nil)
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
				metricServiceMock.EXPECT().GetMetric(gomock.Any(), constants.MetricTypeCounter, "temp").Times(1).Return(metricprocessor.MetricValue{Counter: 100}, nil)
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
				metricServiceMock.EXPECT().GetMetric(gomock.Any(), constants.MetricTypeGauge, "temp").Times(1).Return(metricprocessor.MetricValue{}, storage.NewErrNotFound(errors.New(""), map[string]interface{}{}))
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
	storageMock := storage.NewMockStorage(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := metric.New(metricServiceMock, storageMock, logger, nil)

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
				metricServiceMock.EXPECT().SaveMetric(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(metricprocessor.MetricValue{}, nil)
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
	storageMock := storage.NewMockStorage(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := metric.New(metricServiceMock, storageMock, logger, nil)

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
				metricServiceMock.EXPECT().GetMetric(gomock.Any(), constants.MetricTypeGauge, "temp").Times(1).Return(metricprocessor.MetricValue{Gauge: 100.156}, nil)
			},
			expectedResult: "100.156",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "valid counter metric",
			method: http.MethodGet,
			url:    "/value/counter/temp",
			setupMock: func() {
				metricServiceMock.EXPECT().GetMetric(gomock.Any(), constants.MetricTypeCounter, "temp").Times(1).Return(metricprocessor.MetricValue{Counter: 100}, nil)
			},
			expectedResult: "100",
			expectedStatus: http.StatusOK,
		},
		{
			name:   "not found metric",
			method: http.MethodGet,
			url:    "/value/gauge/temp",
			setupMock: func() {
				metricServiceMock.EXPECT().GetMetric(gomock.Any(), constants.MetricTypeGauge, "temp").Times(1).Return(metricprocessor.MetricValue{}, storage.NewErrNotFound(errors.New(""), map[string]interface{}{}))
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
	storageMock := storage.NewMockStorage(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := metric.New(metricServiceMock, storageMock, logger, nil)

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
				metricServiceMock.EXPECT().GetAllMetrics(gomock.Any()).Times(1).Return(entities.TotalMetrics{}, nil)
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

func TestPingHandler(t *testing.T) {
	r := chi.NewRouter()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	storageMock := storage.NewMockStorage(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := metric.New(metricServiceMock, storageMock, logger, nil)

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
			name:   "valid storage connection",
			method: http.MethodGet,
			url:    "/ping",
			setupMock: func() {
				storageMock.EXPECT().Ping(gomock.Any()).Times(1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid storage connection",
			method: http.MethodGet,
			url:    "/ping",
			setupMock: func() {
				storageMock.EXPECT().Ping(gomock.Any()).Times(1).Return(errors.New("ping error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid method",
			method:         http.MethodPost,
			url:            "/ping",
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
		})
	}
}

func TestBatchUpdatesMetricHandler(t *testing.T) {
	r := chi.NewRouter()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	storageMock := storage.NewMockStorage(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	c := metric.New(metricServiceMock, storageMock, logger, nil)

	r.Mount("/", c.Route())

	ts := httptest.NewServer(r)
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL).SetHeader("Content-Type", "application/json")

	tests := []struct {
		name                string
		method              string
		url                 string
		body                string
		setupMock           func()
		expectedContentType string
		expectedStatus      int
	}{
		{
			name:   "valid batch update",
			method: http.MethodPost,
			url:    "/updates/",
			body: `[
				{
				  "id": "test",
				  "type": "counter",
				  "delta": 100
				},
				{
					"id": "test",
					"type": "gauge",
					"value": 200.123125
				}
			]`,
			setupMock: func() {
				storageMock.EXPECT().SaveMetricBatch(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid update",
			method: http.MethodPost,
			body:   `[{"id": "test", "type": "counter", "delta": 100}]`,
			url:    "/updates/",
			setupMock: func() {
				storageMock.EXPECT().SaveMetricBatch(gomock.Any(), gomock.Any()).Times(1).Return(errors.New("save error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			req := client.R()

			req.Method = tc.method
			req.URL = tc.url
			req.Body = tc.body

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode())
		})
	}
}

func TestSetupSignerInAdapter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricServiceMock := metricprocessor.NewMockMetricService(ctrl)
	signerMock := signer.NewMockSigner(ctrl)
	storageMock := storage.NewMockStorage(ctrl)
	logger, err := logger.Initialize("info")

	if err != nil {
		log.Fatalf(err.Error())
	}

	client := resty.New().SetHeader("Content-Type", "application/json")

	body := `[
				{
				  "id": "test",
				  "type": "counter",
				  "delta": 100
				},
				{
					"id": "test",
					"type": "gauge",
					"value": 200.123125
				}
			]`

	tests := []struct {
		name       string
		setupSuite func() *httptest.Server
	}{
		{
			name: "should validate sign if provide signer",
			setupSuite: func() *httptest.Server {
				r := chi.NewRouter()
				c := metric.New(metricServiceMock, storageMock, logger, signerMock)

				r.Mount("/", c.Route())

				ts := httptest.NewServer(r)

				storageMock.EXPECT().SaveMetricBatch(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				signerMock.EXPECT().Verify(gomock.Any(), "test-signature").Times(1).Return(true)
				signerMock.EXPECT().Sign(gomock.Any()).MinTimes(1).Return("signature")

				return ts
			},
		},
		{
			name: "should not validate sign if signer not provided",
			setupSuite: func() *httptest.Server {
				r := chi.NewRouter()
				c := metric.New(metricServiceMock, storageMock, logger, nil)

				r.Mount("/", c.Route())

				ts := httptest.NewServer(r)

				storageMock.EXPECT().SaveMetricBatch(gomock.Any(), gomock.Any()).Times(1).Return(nil)
				signerMock.EXPECT().Verify(gomock.Any(), gomock.Any()).Times(0)
				signerMock.EXPECT().Sign(gomock.Any()).Times(0)

				return ts
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := tc.setupSuite()
			defer ts.Close()

			resp, err := client.R().SetBody(body).SetHeader(constants.HashHeader, "test-signature").Post(ts.URL + "/updates/")

			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode())
		})
	}
}
