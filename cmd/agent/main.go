package main

import (
	"time"

	"github.com/sodiqit/metricpulse.git/internal/agent"
)

func main() {
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	agent.RunCollector(pollInterval, reportInterval)
}
