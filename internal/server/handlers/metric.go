package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
)

func isValidMetricType(metricType string) bool {
	if metricType != constants.MetricTypeGauge && metricType != constants.MetricTypeCounter {
		return false
	}

	return true
}

func parseMetricValue(metricType string, value string) (services.MetricValue, error) {
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

func UpdateMetricHandler(metricService services.IMetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		ok := isValidMetricType(metricType)

		if !ok {
			http.Error(w, "Supported metrics: gauge | counter", http.StatusBadRequest)
			return
		}

		val, err := parseMetricValue(metricType, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metricService.SaveMetric(metricType, metricName, val)

		w.Write([]byte{})
	}
}

func GetMetricHandler(metricService services.IMetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		ok := isValidMetricType(metricType)

		if !ok {
			http.Error(w, "Supported metrics: gauge | counter", http.StatusBadRequest)
			return
		}

		val, err := metricService.GetMetric(metricType, metricName)

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
}

func GetAllMetricsHandler(metricService services.IMetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := metricService.GetAllMetrics()

		var htmlBuilder strings.Builder
		htmlBuilder.WriteString("<html><head><title>Metrics</title></head><body>")
		htmlBuilder.WriteString("<h1>Gauge Metrics</h1><ul>")

		for name, value := range metrics.Gauge {
			htmlBuilder.WriteString(fmt.Sprintf("<li>%s: %f</li>", name, value))
		}
		htmlBuilder.WriteString("</ul>")

		htmlBuilder.WriteString("<h1>Counter Metrics</h1><ul>")
		for name, values := range metrics.Counter {
			valuesStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(values)), ", "), "[]")
			htmlBuilder.WriteString(fmt.Sprintf("<li>%s: [%s]</li>", name, valuesStr))
		}
		htmlBuilder.WriteString("</ul></body></html>")

		w.Header().Add("Content-Type", "text/html")

		w.Write([]byte(htmlBuilder.String()))
	}
}

func RegisterMetricRouter(r chi.Router, metricService services.IMetricService) chi.Router {
	r.Post("/update/{metricType}/{metricName}/{metricValue}", UpdateMetricHandler(metricService))
	r.Get("/value/{metricType}/{metricName}", GetMetricHandler(metricService))
	r.Get("/", GetAllMetricsHandler(metricService))

	return r
}
