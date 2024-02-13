package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/adapters/http/metric"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricprocessor"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricuploader"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

func RunServer(config *config.Config) error {
	logger, err := logger.Initialize(config.LogLevel)

	if err != nil {
		return err
	}

	defer logger.Sync()

	storage := storage.NewMemStorage()
	uploadService, err := metricuploader.New(config, storage, logger)

	if err != nil {
		return err
	}

	defer uploadService.Close()

	metricService := metricprocessor.New(storage, uploadService, config)
	metricAdapter := metric.New(metricService, logger)

	r := chi.NewRouter()
	r.Mount("/", metricAdapter.Route())

	uploadService.StoreInterval()
	loadErr := uploadService.Load()

	if loadErr != nil {
		return loadErr
	}

	logger.Infow("start server", "address", config.Address, "config", config)
	return http.ListenAndServe(config.Address, r)
}
