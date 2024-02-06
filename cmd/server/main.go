package main

import (
	"log"

	"github.com/sodiqit/metricpulse.git/internal/server"
)

func main() {
	parseFlags()

	router := server.NewRouter()

	err := server.RunServer(runAddr, router)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
