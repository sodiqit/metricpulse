package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/adapters/http/metric"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

func storeMetricsInterval(uploadService services.IUploadService, cfg *config.Config) {

}

func RunServer(config *config.Config) error {
	logger, err := logger.Initialize(config.LogLevel)

	if err != nil {
		return err
	}

	defer logger.Sync()

	storage := storage.NewMemStorage()
	uploadService, err := services.NewUploadService(config, storage, logger)

	if err != nil {
		return err
	}

	defer uploadService.Close()

	metricService := services.NewMetricService(storage)
	metricAdapter := metric.New(metricService, logger, uploadService, config)

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
