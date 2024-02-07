package main

import (
	"log"

	"github.com/sodiqit/metricpulse.git/internal/server"
)

func main() {
	parseConfig()

	err := server.Run(cfg.Address, cfg.LogLevel)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
