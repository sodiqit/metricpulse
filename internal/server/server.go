package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sodiqit/metricpulse.git/internal/server/controllers"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

func RegisterDeps(r chi.Router) {
	storage := storage.NewMemStorage()
	metricService := services.NewMetricService(storage)
	metricController := controllers.NewMetricController(&metricService)

	r.Mount("/", metricController.Route())
}

func NewRouter() chi.Router {
	r := chi.NewRouter()

	RegisterDeps(r)

	return r
}

func RunServer(addr string, r chi.Router) error {
	log.Printf("Server is starting on %s", addr)
	return http.ListenAndServe(addr, r)
}
