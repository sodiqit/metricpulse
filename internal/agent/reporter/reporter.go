package reporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
)

type IMetricReporter interface {
	SendMetrics(ctx context.Context, metrics map[string]interface{}, backoff retry.Backoff)
}

type MetricReporter struct {
	client     *resty.Client
	logger     logger.ILogger
	serverAddr string
}

func (r *MetricReporter) SendMetrics(ctx context.Context, metrics map[string]interface{}, backoff retry.Backoff) {
	var metricsList []entities.Metrics

	for metricName, metricValue := range metrics {
		metric := entities.Metrics{ID: metricName}

		switch val := metricValue.(type) {
		case float64:
			metric.MType = constants.MetricTypeGauge
			metric.Value = &val
		case int64:
			metric.MType = constants.MetricTypeCounter
			metric.Delta = &val
		default:
			r.logger.Errorw("unsupported metric type", "metricName", metricName)
			continue
		}

		metricsList = append(metricsList, metric)
	}

	if len(metricsList) == 0 {
		return
	}

	buf, err := wrapBodyInGzip(metricsList)
	if err != nil {
		r.logger.Errorw("error while wrapping body in gzip", "error", err)
		return
	}

	// Отправка сжатого списка метрик
	url := fmt.Sprintf("http://%s/updates/", r.serverAddr)

	resp, err := retry.DoWithData(ctx, backoff, func(ctx context.Context) (*resty.Response, error) {
		r.logger.Infow("try send metric on server")

		resp, err := r.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(buf.Bytes()).
			Post(url)

		if err != nil || resp.StatusCode() >= 500 {
			return resp, retry.RetryableError(err)
		}

		return resp, err
	})

	if err != nil {
		r.logger.Errorw("error while sending metrics batch", "error", err)
		return
	}

	r.logger.Infow("success sending metrics batch", "result", resp.String())
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
