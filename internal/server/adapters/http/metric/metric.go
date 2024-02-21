package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/adapters/http/middlewares"
	"github.com/sodiqit/metricpulse.git/internal/server/services/metricprocessor"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

type Adapter struct {
	metricService metricprocessor.MetricService
	logger        logger.ILogger
	storage       storage.Storage
}

func (a *Adapter) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares.WithLogger(a.logger))
	r.Use(middlewares.Gzip)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", a.handleTextUpdateMetric)
	r.Post("/update/", a.handleUpdateMetric)

	r.Get("/value/{metricType}/{metricName}", a.handleTextGetMetric)
	r.Post("/value/", a.handleGetMetric)

	r.Get("/ping", a.handlePing)
	r.Post("/updates/", a.handleUpdatesMetric)
	r.Get("/", a.handleGetAllMetrics)

	return r
}

func (a *Adapter) handleUpdatesMetric(w http.ResponseWriter, r *http.Request) {
	var metrics []entities.Metrics

	contentType := r.Header.Get("Content-Type")

	if contentType != "application/json" {
		http.Error(w, "need provide Content-Type: application/json", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := a.storage.SaveMetricBatch(r.Context(), metrics)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(""))
}

func (a *Adapter) handlePing(w http.ResponseWriter, r *http.Request) {
	err := a.storage.Ping(r.Context())

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte{})
}

func (a *Adapter) handleTextUpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

	ok := isValidMetricType(metricType)

	if !ok {
		http.Error(w, "Supported metrics: gauge | counter", http.StatusBadRequest)
		return
	}

	val, err := parseString2MetricValue(metricType, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.metricService.SaveMetric(r.Context(), metricType, metricName, val)

	w.Write([]byte{})
}

func (a *Adapter) handleTextGetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	ok := isValidMetricType(metricType)

	if !ok {
		http.Error(w, "Supported metrics: gauge | counter", http.StatusBadRequest)
		return
	}

	val, err := a.metricService.GetMetric(r.Context(), metricType, metricName)

	if storage.IsErrNotFound(err) {
		http.Error(w, fmt.Sprintf("Not found metric: %s", metricName), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
		return
	}

	switch metricType {
	case constants.MetricTypeGauge:
		w.Write([]byte(strconv.FormatFloat(val.Gauge, 'f', -1, 64)))
	case constants.MetricTypeCounter:
		w.Write([]byte(strconv.Itoa(int(val.Counter))))
	}
}

func (a *Adapter) handleUpdateMetric(w http.ResponseWriter, r *http.Request) {
	var metrics entities.Metrics

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

	updatedValue, err := a.metricService.SaveMetric(r.Context(), metrics.MType, metrics.ID, val)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := marshalMetrics(metrics.MType, metrics.ID, updatedValue)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	w.Write(result)
}

func (a *Adapter) handleGetMetric(w http.ResponseWriter, r *http.Request) {
	var metrics entities.Metrics

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

	val, err := a.metricService.GetMetric(r.Context(), metrics.MType, metrics.ID)

	if storage.IsErrNotFound(err) {
		http.Error(w, fmt.Sprintf("Not found metric: %s", metrics.ID), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error: %s", err), http.StatusInternalServerError)
		return
	}

	result, err := marshalMetrics(metrics.MType, metrics.ID, val)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	w.Write(result)
}

func (a *Adapter) handleGetAllMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := a.metricService.GetAllMetrics(r.Context())

	if err != nil {
		http.Error(w, "Cannot find metrics", http.StatusInternalServerError)
		return
	}

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

func New(metricService metricprocessor.MetricService, storage storage.Storage, logger logger.ILogger) *Adapter {
	return &Adapter{
		metricService,
		logger,
		storage,
	}
}

func isValidMetricType(metricType string) bool {
	if metricType != constants.MetricTypeGauge && metricType != constants.MetricTypeCounter {
		return false
	}

	return true
}

func marshalMetrics(metricType, metricName string, val metricprocessor.MetricValue) ([]byte, error) {
	var body entities.Metrics

	if metricType == constants.MetricTypeGauge {
		body = entities.Metrics{ID: metricName, MType: metricType, Value: &val.Gauge}
	}

	if metricType == constants.MetricTypeCounter {
		body = entities.Metrics{ID: metricName, MType: metricType, Delta: &val.Counter}
	}

	return json.Marshal(body)
}

func parseMetricValue(metric entities.Metrics) (metricprocessor.MetricValue, error) {
	if metric.MType == constants.MetricTypeGauge {
		if metric.Value == nil {
			return metricprocessor.MetricValue{}, errors.New("metric value not provided: provide float64")
		}

		return metricprocessor.MetricValue{Gauge: *metric.Value}, nil
	}

	if metric.MType == constants.MetricTypeCounter {
		if metric.Delta == nil {
			return metricprocessor.MetricValue{}, errors.New("metric value not provided: provide int64")
		}

		return metricprocessor.MetricValue{Counter: *metric.Delta}, nil
	}
	return metricprocessor.MetricValue{}, fmt.Errorf("unknown metricType: %s", metric.MType)
}

func parseString2MetricValue(metricType string, value string) (metricprocessor.MetricValue, error) {
	if metricType == constants.MetricTypeGauge {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return metricprocessor.MetricValue{}, errors.New("invalid value metric: provide float64")
		}

		return metricprocessor.MetricValue{Gauge: val}, nil
	}

	if metricType == constants.MetricTypeCounter {
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return metricprocessor.MetricValue{}, errors.New("invalid value metric: provide int64")
		}

		return metricprocessor.MetricValue{Counter: val}, nil
	}
	return metricprocessor.MetricValue{}, fmt.Errorf("unknown metricType: %s", metricType)
}
