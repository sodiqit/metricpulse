package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/models"
	"github.com/sodiqit/metricpulse.git/internal/server/middlewares"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
)

type MetricController struct {
	metricService services.IMetricService
	logger        logger.ILogger
}

func (c *MetricController) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares.WithLogger(c.logger))

	r.Post("/update/", c.handleUpdateMetric)
	r.Post("/value/", c.handleGetMetric)
	r.Get("/", c.handleGetAllMetrics)

	return r
}

func (c *MetricController) handleUpdateMetric(w http.ResponseWriter, r *http.Request) {
	var metrics models.Metrics

	contentType := r.Header.Get("Content-Type")

	if contentType != "application/json" {
		http.Error(w, "need provide Content-Type: application/json", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok := isValidMetricType(metrics.MType)

	if !ok {
		http.Error(w, "Supported metrics: gauge | counter", http.StatusBadRequest)
		return
	}

	val, err := parseMetricValue(metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedValue, err := c.metricService.SaveMetric(metrics.MType, metrics.ID, val)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(models.Metrics{ID: metrics.ID, MType: metrics.MType, Delta: &updatedValue.Counter, Value: &updatedValue.Gauge})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	w.Write(result)
}

func (c *MetricController) handleGetMetric(w http.ResponseWriter, r *http.Request) {
	var metrics models.Metrics

	contentType := r.Header.Get("Content-Type")

	if contentType != "application/json" {
		http.Error(w, "need provide Content-Type: application/json", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok := isValidMetricType(metrics.MType)

	if !ok {
		http.Error(w, "Supported metrics: gauge | counter", http.StatusBadRequest)
		return
	}

	val, err := c.metricService.GetMetric(metrics.MType, metrics.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Not found metric: %s", metrics.ID), http.StatusNotFound)
		return
	}

	result, err := json.Marshal(models.Metrics{ID: metrics.ID, MType: metrics.MType, Delta: &val.Counter, Value: &val.Gauge})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	w.Write(result)
}

func (c *MetricController) handleGetAllMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := c.metricService.GetAllMetrics()

	var htmlBuilder strings.Builder
	htmlBuilder.WriteString("<html><head><title>Metrics</title></head><body>")
	htmlBuilder.WriteString("<h1>Gauge Metrics</h1><ul>")

	for name, value := range metrics.Gauge {
		htmlBuilder.WriteString(fmt.Sprintf("<li>%s: %f</li>", name, value))
	}
	htmlBuilder.WriteString("</ul>")

	htmlBuilder.WriteString("<h1>Counter Metrics</h1><ul>")
	for name, value := range metrics.Counter {
		htmlBuilder.WriteString(fmt.Sprintf("<li>%s: %v</li>", name, value))
	}
	htmlBuilder.WriteString("</ul></body></html>")

	w.Header().Add("Content-Type", "text/html")

	w.Write([]byte(htmlBuilder.String()))
}

func NewMetricController(metricService services.IMetricService, logger logger.ILogger) *MetricController {
	return &MetricController{
		metricService: metricService,
		logger:        logger,
	}
}

func isValidMetricType(metricType string) bool {
	if metricType != constants.MetricTypeGauge && metricType != constants.MetricTypeCounter {
		return false
	}

	return true
}

func parseMetricValue(metric models.Metrics) (services.MetricValue, error) {
	if metric.MType == constants.MetricTypeGauge {
		if metric.Value == nil {
			return services.MetricValue{}, errors.New("metric value not provided: provide float64")
		}

		return services.MetricValue{Gauge: *metric.Value}, nil
	}

	if metric.MType == constants.MetricTypeCounter {
		if metric.Delta == nil {
			return services.MetricValue{}, errors.New("metric value not provided: provide int64")
		}

		return services.MetricValue{Counter: *metric.Delta}, nil
	}
	return services.MetricValue{}, fmt.Errorf("unknown metricType: %s", metric.MType)
}
