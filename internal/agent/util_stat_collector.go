package agent

import (
	"context"
	"math/rand"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sodiqit/metricpulse.git/internal/logger"
)

type UtilStatsCollector struct {
	logger       logger.ILogger
	pollInterval time.Duration
	scope        Scope
}

func (c *UtilStatsCollector) PollLoop(ctx context.Context) error {
	scope := c.scope

	rPollCount := scope.Counter("PollCount")
	rRandomValue := scope.Gauge("RandomValue")

	rTotalMemory := scope.Gauge("TotalMemory")
	rFreeMemory := scope.Gauge("FreeMemory")
	rCPUutilization1 := scope.Gauge("CPUutilization1")

	ticker := time.NewTicker(c.pollInterval)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			c.logger.Infow("vm monitor: terminate goroutine", "reason", ctx.Err())
			return ctx.Err()
		}

		c.logger.Infow("vm monitor: update vm and cpu metrics", "interval", c.pollInterval)
		rPollCount.Inc(1)
		rRandomValue.Update(rand.Float64() * 100)

		vm, err := mem.VirtualMemory()

		if err != nil {
			c.logger.Errorw("vm monitor: error while getting stats", "error", err)
			return err
		}

		cpu, err := cpu.Percent(time.Duration(0), true)

		if err != nil {
			c.logger.Errorw("vm monitor: error while getting cpu stats", "error", err)
			return err
		}

		rTotalMemory.Update(float64(vm.Total))
		rFreeMemory.Update(float64(vm.Free))
		rCPUutilization1.Update(float64(cpu[0]))
	}
}

func NewUtilStatsCollector(logger logger.ILogger, pollInterval time.Duration, scope Scope) *UtilStatsCollector {
	return &UtilStatsCollector{
		logger:       logger,
		pollInterval: pollInterval,
		scope:        scope,
	}
}
