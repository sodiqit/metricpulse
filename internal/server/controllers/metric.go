package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sodiqit/metricpulse.git/internal/server/services"
)

func parseMetricValue(metricType string, value string) (interface{}, error) {
	if metricType == "gauge" {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return val, errors.New("invalid value metric: provide float64")
		}

		return val, nil
	}

	if metricType == "counter" {
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return val, errors.New("invalid value metric: provide int64")
		}

		return val, nil
	}
	return 0, fmt.Errorf("unknown metricType: %s", metricType)
}

func UpdateMetricHandler(metricService services.IMetricService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		pathSegments := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")

		if len(pathSegments) != 3 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		metricType := pathSegments[0]
		metricKind := pathSegments[1]
		metricValue := pathSegments[2]

		if metricType != "gauge" && metricType != "counter" {
			http.Error(w, "Supported metrics: gauge | counter", http.StatusBadRequest)
			return
		}

		val, err := parseMetricValue(metricType, metricValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metricService.SaveMetric(metricType, metricKind, val)

		w.Write([]byte{})
	}
}

func NewMetricController(metricService services.IMetricService) *Controller {
	controller := &Controller{}

	controller.Routes = map[string]http.HandlerFunc{
		"/update/": UpdateMetricHandler(metricService),
	}

	return controller
}
