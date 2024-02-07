package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/controllers"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

func registerDeps(r chi.Router, logger logger.ILogger) {
	storage := storage.NewMemStorage()
	metricService := services.NewMetricService(storage)
	metricController := controllers.NewMetricController(&metricService, logger)

	r.Mount("/", metricController.Route())
}

func Run(addr, logLevel string) error {
	logger, err := logger.Initialize(logLevel)

	if err != nil {
		return err
	}

	defer logger.Sync()

	r := chi.NewRouter()

	registerDeps(r, logger)

	logger.Infow("start server", "address", addr)
	return http.ListenAndServe(addr, r)
}
