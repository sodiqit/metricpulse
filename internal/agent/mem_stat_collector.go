package agent

import (
	"context"
	"math/rand"
	"runtime"
	"time"

	"github.com/sodiqit/metricpulse.git/internal/logger"
)

type MemStatsCollector struct {
	logger       logger.ILogger
	pollInterval time.Duration
	scope        Scope
}

func (c *MemStatsCollector) PollLoop(ctx context.Context) error {
	scope := c.scope

	rPollCount := scope.Counter("PollCount")
	rRandomValue := scope.Gauge("RandomValue")

	rAlloc := scope.Gauge("Alloc")
	rTotalAlloc := scope.Gauge("TotalAlloc")
	rSys := scope.Gauge("Sys")
	rLookups := scope.Gauge("Lookups")
	rMallocs := scope.Gauge("Mallocs")
	rFrees := scope.Gauge("Frees")
	rHeapAlloc := scope.Gauge("HeapAlloc")
	rHeapSys := scope.Gauge("HeapSys")
	rHeapIdle := scope.Gauge("HeapIdle")
	rHeapInuse := scope.Gauge("HeapInuse")
	rHeapReleased := scope.Gauge("HeapReleased")
	rHeapObjects := scope.Gauge("HeapObjects")
	rStackInuse := scope.Gauge("StackInuse")
	rStackSys := scope.Gauge("StackSys")
	rMSpanInuse := scope.Gauge("MSpanInuse")
	rMSpanSys := scope.Gauge("MSpanSys")
	rMCacheInuse := scope.Gauge("MCacheInuse")
	rMCacheSys := scope.Gauge("MCacheSys")
	rBuckHashSys := scope.Gauge("BuckHashSys")
	rGCSys := scope.Gauge("GCSys")
	rOtherSys := scope.Gauge("OtherSys")
	rNextGC := scope.Gauge("NextGC")
	rLastGC := scope.Gauge("LastGC")
	rPauseTotalNs := scope.Gauge("PauseTotalNs")
	rNumGC := scope.Gauge("NumGC")
	rNumForcedGC := scope.Gauge("NumForcedGC")
	rGCCPUFraction := scope.Gauge("GCCPUFraction")

	ticker := time.NewTicker(c.pollInterval)

	defer ticker.Stop()
	var rtm runtime.MemStats

	// Обновляем счетчики с заданной периодичностью.
	// Значения (и описание) получаем из пакета runtime.
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			c.logger.Infow("monitor: teminate goroutine", "reason", ctx.Err())
			return ctx.Err()
		}

		c.logger.Infow("monitor: update mem metrics", "interval", c.pollInterval)
		rPollCount.Inc(1)
		rRandomValue.Update(rand.Float64() * 100)

		// Read full mem stats
		runtime.ReadMemStats(&rtm)

		rAlloc.Update(float64(rtm.Alloc))
		rTotalAlloc.Update(float64(rtm.TotalAlloc))
		rSys.Update(float64(rtm.Sys))
		rLookups.Update(float64(rtm.Lookups))
		rMallocs.Update(float64(rtm.Mallocs))
		rFrees.Update(float64(rtm.Frees))

		rHeapAlloc.Update(float64(rtm.HeapAlloc))
		rHeapSys.Update(float64(rtm.HeapSys))
		rHeapIdle.Update(float64(rtm.HeapIdle))
		rHeapInuse.Update(float64(rtm.HeapInuse))
		rHeapReleased.Update(float64(rtm.HeapReleased))
		rHeapObjects.Update(float64(rtm.HeapObjects))

		rStackInuse.Update(float64(rtm.StackInuse))

		rStackSys.Update(float64(rtm.StackSys))

		rMSpanInuse.Update(float64(rtm.MSpanInuse))

		rMSpanSys.Update(float64(rtm.MSpanSys))

		rMCacheInuse.Update(float64(rtm.MCacheInuse))
		rMCacheSys.Update(float64(rtm.MCacheSys))
		rBuckHashSys.Update(float64(rtm.BuckHashSys))
		rGCSys.Update(float64(rtm.GCSys))
		rOtherSys.Update(float64(rtm.OtherSys))
		rNextGC.Update(float64(rtm.NextGC))
		rLastGC.Update(float64(rtm.LastGC))
		rPauseTotalNs.Update(float64(rtm.PauseTotalNs))
		rNumGC.Update(float64(rtm.NumGC))
		rNumForcedGC.Update(float64(rtm.NumForcedGC))
		rGCCPUFraction.Update(float64(rtm.GCCPUFraction))
	}
}

func NewMemStatsCollector(logger logger.ILogger, pollInterval time.Duration, scope Scope) *MemStatsCollector {
	return &MemStatsCollector{
		logger:       logger,
		pollInterval: pollInterval,
		scope:        scope,
	}
}
