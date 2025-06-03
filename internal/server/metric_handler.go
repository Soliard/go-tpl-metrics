package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
)

var storage store.Storage = store.NewMemStorage()

func UpdateClaimMetric(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, `only POST method is allowed`, http.StatusMethodNotAllowed)
		return
	}

	metricName, metricType, metricStrValue, err := parseClaimMetricUrl(req.URL.Path)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if metricName == "" {
		http.Error(res, `metric name cannot be empty`, http.StatusNotFound)
		return
	}

	switch metricType {
	case models.Gauge:
		err := updateGaugeMetric(metricName, metricStrValue, res)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

	case models.Counter:
		err := updateCounterMetric(metricName, metricStrValue, res)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(res, `invalid metric type`, http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func parseClaimMetricUrl(url string) (metricType string, metricName string, metricValue string, err error) {
	urlParts := make([]string, 0)
	for _, v := range strings.Split(url, `/`) {
		if v != "" {
			urlParts = append(urlParts, v)
		}
	}

	if len(urlParts) < 4 {
		return "", "", "", fmt.Errorf(`URL must be /update/{metricType}/{metricName}/{metricValue}`)
	}

	return urlParts[1], urlParts[2], urlParts[3], nil
}

func updateCounterMetric(name string, valueStr string, res http.ResponseWriter) error {
	metricValue, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return fmt.Errorf(`invalid metric value, must be an integer`)
	}
	storage.UpdateCounter(name, metricValue)
	sum, _ := storage.GetCounter(name)
	fmt.Fprintf(res, "counter metric '%s' updated to %d", name, sum)

	return nil
}

func updateGaugeMetric(name string, valueStr string, res http.ResponseWriter) error {
	metricValue, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf(`invalid metric value, must be a float`)
	}
	storage.UpdateGauge(name, metricValue)
	fmt.Fprintf(res, "gauge metric '%s' updated to %f", name, metricValue)

	return nil
}
