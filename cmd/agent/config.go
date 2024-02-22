package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	LogLevel       string `env:"LOG_LEVEL"`
	SecretKey      string `env:"KEY"`
}

var cfg Config

func parseConfig() {
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&cfg.PollInterval, "p", 2, "poll runtime interval in seconds")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.SecretKey, "k", "", "key for data encryption")

	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
}
