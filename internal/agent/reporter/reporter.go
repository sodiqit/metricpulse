package reporter

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type IMetricReporter interface {
	SendMetrics(metrics map[string]interface{})
}

type MetricReporter struct {
	Client     HTTPClient
	serverAddr string
}

type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

func (r *MetricReporter) SendMetrics(metrics map[string]interface{}) {
	for metricName, metricValue := range metrics {
		var metricType string
		switch metricValue.(type) {
		case float64:
			metricType = "gauge"
		case int64:
			metricType = "counter"
		default:
			fmt.Printf("Unsupported metric type for %s\n", metricName)
			continue
		}

		url := fmt.Sprintf("http://%s/update/%s/%s/%v", r.serverAddr, metricType, metricName, metricValue)
		resp, err := r.Client.Post(url, "text/plain", nil)

		if err != nil {
			log.Error().Msg(fmt.Sprintf("Error sending metric %s: %v\n", metricName, err))
			continue
		}

		log.Log().Msg(fmt.Sprintf("Success sending metric %s", metricName))

		defer resp.Body.Close()
	}
}

func NewMetricReporter(serverAddr string, client HTTPClient) *MetricReporter {
	return &MetricReporter{
		Client:     client,
		serverAddr: serverAddr,
	}
}
