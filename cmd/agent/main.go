package main

import (
	"time"

	"github.com/sodiqit/metricpulse.git/internal/agent"
)

func main() {
	parseFlags()

	pollIntervalDuration := time.Duration(agentFlags.pollInterval) * time.Second
	reportIntervalDuration := time.Duration(agentFlags.reportInterval) * time.Second

	agent.RunCollector(agentFlags.serverAddr, pollIntervalDuration, reportIntervalDuration)
}
