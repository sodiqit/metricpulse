package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
	"github.com/sodiqit/metricpulse.git/pkg/signer"
	"golang.org/x/sync/errgroup"
)

type WorkerPool struct {
	jobs chan func() error
	size int
}

func (wp *WorkerPool) Close() error {
	close(wp.jobs)

	return nil
}

func (wp *WorkerPool) Run(ctx context.Context, job func() error) {
	select {
	case wp.jobs <- job:
	case <-ctx.Done():
		return
	}
}

func (wp *WorkerPool) Start(ctx context.Context, g *errgroup.Group) {
	for i := 1; i <= wp.size; i++ {
		g.Go(func() error {
			for {
				select {
				case job, ok := <-wp.jobs:
					if !ok {
						return nil
					}

					err := job()

					if err != nil {
						return err
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}
}

func newWorkerPool(size int) *WorkerPool {
	return &WorkerPool{
		make(chan func() error, size),
		size,
	}
}

type MetricReporterOptions struct {
	ServerAddr     string
	Scope          Scope
	Client         *resty.Client
	ReportInterval time.Duration
	RateLimit      int
	Logger         logger.ILogger
	Signer         signer.Signer
}

type MetricReporter struct {
	client         *resty.Client
	scope          Scope
	reportInterval time.Duration
	rateLimit      int
	logger         logger.ILogger
	serverAddr     string
	signer         signer.Signer
}

func (r *MetricReporter) ReportLoop(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	wp := newWorkerPool(r.rateLimit)
	defer wp.Close()
	wp.Start(ctx, g)

	ticker := time.NewTicker(r.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			r.logger.Infow("reporter: terminate goroutine", "reason", ctx.Err())
			return ctx.Err()
		}

		wp.Run(ctx, func() error {
			snapshot := r.scope.Snapshot()

			return r.SendBatchMetrics(ctx, snapshot, retry.NewBaseBackoff())
		})
	}
}

func (r *MetricReporter) SendBatchMetrics(ctx context.Context, snapshot MetricSnapshot, backoff retry.Backoff) error {
	var metricsList []entities.Metrics

	for metricName, metricValue := range snapshot.Counters {
		val := metricValue.Value()
		metric := entities.Metrics{ID: metricName, MType: constants.MetricTypeCounter, Delta: &val}
		metricsList = append(metricsList, metric)
	}

	for metricName, metricValue := range snapshot.Gauges {
		val := metricValue.Value()
		metric := entities.Metrics{ID: metricName, MType: constants.MetricTypeGauge, Value: &val}
		metricsList = append(metricsList, metric)
	}

	if len(metricsList) == 0 {
		return nil
	}

	buf, err := wrapBodyInGzip(metricsList)
	if err != nil {
		r.logger.Errorw("error while wrapping body in gzip", "error", err)
		return err
	}

	// Отправка сжатого списка метрик
	url := fmt.Sprintf("http://%s/updates/", r.serverAddr)

	resp, err := retry.DoWithData(ctx, backoff, func(ctx context.Context) (*resty.Response, error) {
		r.logger.Infow("try send metric on server")

		req := r.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip")

		resp, err := signRequest(buf.Bytes(), req, r.signer).Post(url)

		if err != nil || resp.StatusCode() >= 500 {
			return resp, retry.RetryableError(err)
		}

		return resp, err
	})

	if err != nil {
		r.logger.Errorw("error while sending metrics batch", "error", err)
		return err
	}

	if resp.StatusCode() >= http.StatusBadRequest {
		r.logger.Errorw("error while sending metrics batch", "error", resp.String())
		return nil
	}

	r.logger.Infow("success sending metrics batch", "result", resp.String())

	return nil
}

func NewMetricReporter(options MetricReporterOptions) *MetricReporter {
	return &MetricReporter{
		client:         options.Client,
		scope:          options.Scope,
		serverAddr:     options.ServerAddr,
		reportInterval: options.ReportInterval,
		rateLimit:      options.RateLimit,
		logger:         options.Logger,
		signer:         options.Signer,
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

func signRequest(body []byte, r *resty.Request, signer signer.Signer) *resty.Request {
	if signer == nil {
		return r.SetBody(body)
	}

	signature := signer.Sign(body)

	return r.SetHeader(constants.HashHeader, signature).SetBody(body)
}
