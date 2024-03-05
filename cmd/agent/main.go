package main

import (
	"log"

	"github.com/sodiqit/metricpulse.git/internal/agent"
)

func main() {
	cfg := agent.ParseConfig()

	a := agent.NewAgent(cfg)

	err := a.Run()
	if err != nil {
		log.Fatalf("Error while run agent: %s", err)
	}
}
