package main

import (
	"time"

	"github.com/sodiqit/metricpulse.git/internal/agent"
)

func main() {
	parseConfig()

	agent.RunCollector(cfg.Address, time.Duration(cfg.PollInterval)*time.Second, time.Duration(cfg.ReportInterval)*time.Second, cfg.LogLevel)
}
