package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sodiqit/metricpulse.git/internal/server/handlers"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

func RegisterDeps(r chi.Router) {
	storage := storage.NewMemStorage()
	metricService := services.NewMetricService(storage)
	handlers.RegisterMetricRouter(r, &metricService)
}

func NewRouter() chi.Router {
	r := chi.NewRouter()

	RegisterDeps(r)

	return r
}

func RunServer(port int, r chi.Router) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server is starting on %s", addr)
	return http.ListenAndServe(addr, r)
}
