package agent

import (
	"context"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sodiqit/metricpulse.git/internal/agent/reporter"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
	"github.com/sodiqit/metricpulse.git/pkg/signer"
)

type MetricCounters struct {
	PollCount   int64
	RandomValue float64
}

func CollectMetrics(mc *MetricCounters) map[string]interface{} {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	mc.PollCount++
	mc.RandomValue = float64(rand.Intn(100))

	return map[string]interface{}{
		"Alloc":         float64(rtm.Alloc),
		"BuckHashSys":   float64(rtm.BuckHashSys),
		"Frees":         float64(rtm.Frees),
		"GCCPUFraction": float64(rtm.GCCPUFraction),
		"GCSys":         float64(rtm.GCSys),
		"HeapAlloc":     float64(rtm.HeapAlloc),
		"HeapIdle":      float64(rtm.HeapIdle),
		"HeapInuse":     float64(rtm.HeapInuse),
		"HeapObjects":   float64(rtm.HeapObjects),
		"HeapReleased":  float64(rtm.HeapReleased),
		"HeapSys":       float64(rtm.HeapSys),
		"LastGC":        float64(rtm.LastGC),
		"Lookups":       float64(rtm.Lookups),
		"MCacheInuse":   float64(rtm.MCacheInuse),
		"MCacheSys":     float64(rtm.MCacheSys),
		"MSpanInuse":    float64(rtm.MSpanInuse),
		"MSpanSys":      float64(rtm.MSpanSys),
		"Mallocs":       float64(rtm.Mallocs),
		"NextGC":        float64(rtm.NextGC),
		"NumForcedGC":   float64(rtm.NumForcedGC),
		"NumGC":         float64(rtm.NumGC),
		"OtherSys":      float64(rtm.OtherSys),
		"PauseTotalNs":  float64(rtm.PauseTotalNs),
		"StackInuse":    float64(rtm.StackInuse),
		"StackSys":      float64(rtm.StackSys),
		"Sys":           float64(rtm.Sys),
		"TotalAlloc":    float64(rtm.TotalAlloc),
		"PollCount":     mc.PollCount,
		"RandomValue":   mc.RandomValue,
	}
}

func RunCollector(serverAddr string, pollInterval time.Duration, reportInterval time.Duration, logLevel string, signerKey string) error {
	ctx := context.Background()
	mc := MetricCounters{}
	logger, err := logger.Initialize(logLevel)

	if err != nil {
		return err
	}

	client := resty.New()

	var s signer.Signer

	if signerKey != "" {
		s = signer.NewSHA256Signer(signerKey)
	}

	reporter := reporter.NewMetricReporter(serverAddr, client, logger, s)

	go func() {
		for {
			CollectMetrics(&mc)
			time.Sleep(pollInterval)
		}
	}()

	for {
		metrics := CollectMetrics(&mc)
		reporter.SendMetrics(ctx, metrics, retry.NewBaseBackoff())
		time.Sleep(reportInterval)
	}
}
