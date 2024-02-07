package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
}

var cfg Config

func parseConfig() {
	flag.StringVar(&cfg.Address, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}
}
