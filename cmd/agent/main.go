package main

import (
	"github.com/sodiqit/metricpulse.git/internal/agent"
)

func main() {
	parseConfig()

	agent.RunCollector(cfg.Address, cfg.PollInterval, cfg.ReportInterval)
}
