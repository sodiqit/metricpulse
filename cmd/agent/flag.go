package main

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address        string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
}

var cfg Config

func parseConfig() {
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port server")
	reportInterval := flag.Int("r", 10, "report interval in seconds")
	pollInterval := flag.Int("p", 2, "poll runtime interval in seconds")

	flag.Parse()

	reportIntervalDuration := time.Duration(*reportInterval) * time.Second
	pollIntervalDuration := time.Duration(*pollInterval) * time.Second

	cfg.ReportInterval = reportIntervalDuration
	cfg.PollInterval = pollIntervalDuration

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
}
