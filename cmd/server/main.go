package main

import (
	"log"

	"github.com/sodiqit/metricpulse.git/internal/server"
)

func main() {
	mux := server.NewServeMux()

	err := server.RunServer(8080, mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
