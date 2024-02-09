package reporter

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/models"
)

type IMetricReporter interface {
	SendMetrics(metrics map[string]interface{})
}

type MetricReporter struct {
	client     *resty.Client
	logger     logger.ILogger
	serverAddr string
}

type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

func (r *MetricReporter) SendMetrics(metrics map[string]interface{}) {

	for metricName, metricValue := range metrics {
		body := models.Metrics{ID: metricName}

		switch val := metricValue.(type) {
		case float64:
			body.MType = constants.MetricTypeGauge
			body.Value = &val
		case int64:
			body.MType = constants.MetricTypeCounter
			body.Delta = &val
		default:
			r.logger.Errorw("unsupported metric type", "metricName", metricName)
			continue
		}

		url := fmt.Sprintf("http://%s/update/", r.serverAddr)

		buf, err := wrapBodyInGzip(body)

		if err != nil {
			r.logger.Errorw("error while wrap body in gzip", "metricName", metricName, "error", err.Error())
			continue
		}

		resp, err := r.client.R().SetHeader("Content-Type", "application/json").SetHeader("Content-Encoding", "gzip").SetBody(buf).Post(url)

		if err != nil {
			r.logger.Errorw("error while sending metric", "metricName", metricName, "error", err.Error())
			continue
		}

		r.logger.Infow("success sending metric", "metricName", metricName, "result", resp.String())
	}
}

func NewMetricReporter(serverAddr string, client *resty.Client, logger logger.ILogger) *MetricReporter {
	return &MetricReporter{
		client:     client,
		serverAddr: serverAddr,
		logger:     logger,
	}
}

func wrapBodyInGzip(body interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	stringBody, err := json.Marshal(body)

	if err != nil {
		return buf, fmt.Errorf("cannot marshal body: %v", body)
	}

	_, writerErr := zb.Write([]byte(stringBody))

	if writerErr != nil {
		return buf, fmt.Errorf("cannot write in gzip body: %s", stringBody)
	}

	err = zb.Close()

	if err != nil {
		return buf, fmt.Errorf("cannot close gzip writer native err: %s", err.Error())
	}

	return buf, nil
}
