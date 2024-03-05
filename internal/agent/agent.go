package agent

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/pkg/signer"
	"golang.org/x/sync/errgroup"
)

type Reporter interface {
	ReportLoop(ctx context.Context) error
}

type Collector interface {
	PollLoop(ctx context.Context) error
}

type Scope interface {
	Counter(name string) Counter
	Gauge(name string) Gauge
	Snapshot() MetricSnapshot
}

type Agent struct {
	config *Config
}

func (a *Agent) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.Initialize(a.config.LogLevel)

	if err != nil {
		return err
	}

	client := resty.New()

	var s signer.Signer

	if a.config.SecretKey != "" {
		s = signer.NewSHA256Signer(a.config.SecretKey)
	}

	scope := NewRootScope()

	reporterOptions := MetricReporterOptions{
		ServerAddr:     a.config.Address,
		Logger:         logger,
		Scope:          scope,
		Client:         client,
		Signer:         s,
		ReportInterval: time.Duration(a.config.ReportInterval) * time.Second,
		RateLimit:      a.config.RateLimit,
	}

	reporter := NewMetricReporter(reporterOptions)

	memStatsCollector := NewMemStatsCollector(logger, time.Duration(a.config.PollInterval)*time.Second, scope)
	utilStatsCollector := NewUtilStatsCollector(logger, time.Duration(a.config.PollInterval)*time.Second, scope)

	g.Go(func() error {
		return memStatsCollector.PollLoop(ctx)
	})

	g.Go(func() error {
		return utilStatsCollector.PollLoop(ctx)
	})

	g.Go(func() error {
		return reporter.ReportLoop(ctx)
	})

	return g.Wait()
}

func NewAgent(config *Config) *Agent {
	return &Agent{
		config,
	}
}
