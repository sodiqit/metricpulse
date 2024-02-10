package main

import (
	"log"

	"github.com/sodiqit/metricpulse.git/internal/server"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
)

func main() {
	cfg := config.ParseConfig()

	err := server.Run(cfg)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
