package main

import (
	"log"

	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/infra/http"
)

func main() {
	cfg := config.ParseConfig()

	err := http.RunServer(cfg)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
