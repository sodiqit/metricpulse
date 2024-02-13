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
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
)

type Adapter struct {
	metricService services.IMetricService
	logger        logger.ILogger
	uploadService services.IUploadService
	cfg           *config.Config
}

func (a *Adapter) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middlewares.WithLogger(a.logger))
	r.Use(middlewares.Gzip)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", a.handleTextUpdateMetric)
	r.Post("/update/", a.handleUpdateMetric)

	r.Get("/value/{metricType}/{metricName}", a.handleTextGetMetric)
	r.Post("/value/", a.handleGetMetric)

	r.Get("/", a.handleGetAllMetrics)

	return r
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

	a.metricService.SaveMetric(metricType, metricName, val)

	if a.cfg.StoreInterval == 0 {
		a.uploadService.Save()
	}

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

	val, err := a.metricService.GetMetric(metricType, metricName)

	if err != nil {
		http.Error(w, fmt.Sprintf("Not found metric: %s", metricName), http.StatusNotFound)
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

	updatedValue, err := a.metricService.SaveMetric(metrics.MType, metrics.ID, val)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := marshalMetrics(metrics.MType, metrics.ID, updatedValue)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if a.cfg.StoreInterval == 0 {
		a.uploadService.Save()
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

	val, err := a.metricService.GetMetric(metrics.MType, metrics.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Not found metric: %s", metrics.ID), http.StatusNotFound)
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
	metrics := a.metricService.GetAllMetrics()

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

func New(metricService services.IMetricService, logger logger.ILogger, uploadService services.IUploadService, cfg *config.Config) *Adapter {
	return &Adapter{
		metricService,
		logger,
		uploadService,
		cfg,
	}
}

func isValidMetricType(metricType string) bool {
	if metricType != constants.MetricTypeGauge && metricType != constants.MetricTypeCounter {
		return false
	}

	return true
}

func marshalMetrics(metricType, metricName string, val services.MetricValue) ([]byte, error) {
	var body entities.Metrics

	if metricType == constants.MetricTypeGauge {
		body = entities.Metrics{ID: metricName, MType: metricType, Value: &val.Gauge}
	}

	if metricType == constants.MetricTypeCounter {
		body = entities.Metrics{ID: metricName, MType: metricType, Delta: &val.Counter}
	}

	return json.Marshal(body)
}

func parseMetricValue(metric entities.Metrics) (services.MetricValue, error) {
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

func parseString2MetricValue(metricType string, value string) (services.MetricValue, error) {
	if metricType == constants.MetricTypeGauge {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return services.MetricValue{}, errors.New("invalid value metric: provide float64")
		}

		return services.MetricValue{Gauge: val}, nil
	}

	if metricType == constants.MetricTypeCounter {
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return services.MetricValue{}, errors.New("invalid value metric: provide int64")
		}

		return services.MetricValue{Counter: val}, nil
	}
	return services.MetricValue{}, fmt.Errorf("unknown metricType: %s", metricType)
}