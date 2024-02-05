package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

func RegisterMetricRouter(r chi.Router, metricService services.IMetricService) chi.Router {
	r.Post("/update/{metricType}/{metricName}/{metricValue}", UpdateMetricHandler(metricService))
	r.Get("/value/{metricType}/{metricName}", GetMetricHandler(metricService))

	return r
}