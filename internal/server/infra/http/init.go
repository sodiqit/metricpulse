package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/adapters/http/metric"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricprocessor"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
)

func RunServer(config *config.Config) error {
	logger, err := logger.Initialize(config.LogLevel)

	if err != nil {
		return err
	}

	defer logger.Sync()

	ctx := context.Background()

	storage := setupStorage(config, logger)
	defer storage.Close(ctx)

	err = storage.Init(ctx, retry.NewBaseBackoff())

	if err != nil {
		return err
	}

	metricService := metricprocessor.New(storage, config)
	metricAdapter := metric.New(metricService, storage, logger)

	r := chi.NewRouter()
	r.Mount("/", metricAdapter.Route())

	logger.Infow("start server", "address", config.Address, "config", config)
	return http.ListenAndServe(config.Address, r)
}

func setupStorage(cfg *config.Config, logger logger.ILogger) storage.Storage {
	memoryStorage := storage.NewMemStorage()

	if cfg.DatabaseDSN != "" {
		return storage.NewPostgresStorage(cfg, logger)
	}

	if cfg.FileStoragePath != "" {
		return storage.NewFileStorage(cfg, memoryStorage, logger)
	}

	return memoryStorage
}
