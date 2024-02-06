package main

import (
	"log"

	"github.com/sodiqit/metricpulse.git/internal/server"
)

func main() {
	parseConfig()

	router := server.NewRouter()

	err := server.RunServer(cfg.Address, router)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
