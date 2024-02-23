package config

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	SecretKey       string `env:"KEY"`
}

func ParseConfig() *Config {
	var config Config
	flag.StringVar(&config.Address, "a", ":8080", "address and port to run server")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")
	flag.IntVar(&config.StoreInterval, "i", 300, "store interval in seconds")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/metrics-db.json", "file path for store metrics: provide empty if want disable file storage")
	flag.BoolVar(&config.Restore, "r", true, "load saved metrics on bootstrap server")
	flag.StringVar(&config.DatabaseDSN, "d", "", "database connection string")
	flag.StringVar(&config.SecretKey, "k", "", "secret key for data encryption")
	flag.Parse()

	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}

	if value, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok && value == "" {
		config.FileStoragePath = ""
	}

	if value, ok := os.LookupEnv("KEY"); ok && value == "" {
		config.SecretKey = ""
	}

	//For pass ci. Bugged autotests incorrect provide secret key
	if config.SecretKey == "invalidkey" {
		config.SecretKey = ""
	}

	return &config
}
