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

	metricType, metricName, metricStrValue, err := parseClaimMetricURL(req.URL.Path)
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
		err := updateGaugeMetric(metricName, metricStrValue)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

	case models.Counter:
		err := updateCounterMetric(metricName, metricStrValue)
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

func parseClaimMetricURL(url string) (metricType string, metricName string, metricValue string, err error) {
	segments := getUrlSegments(url)
	if len(segments) < 2 {
		err = fmt.Errorf(`URL must be /update/{metricType}/{metricName}/{metricValue}`)
		return
	}
	if len(segments) > 1 {
		metricType = segments[1]
	}
	if len(segments) > 2 {
		metricName = segments[2]
	}
	if len(segments) > 3 {
		metricValue = segments[3]
	}

	return
}

func getUrlSegments(url string) []string {
	segments := make([]string, 0)
	for _, v := range strings.Split(url, `/`) {
		if v != "" {
			segments = append(segments, v)
		}
	}
	return segments
}

func updateCounterMetric(name string, valueStr string) error {
	metricValue, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return fmt.Errorf(`invalid metric value, must be an integer`)
	}
	storage.UpdateCounter(name, metricValue)

	return nil
}

func updateGaugeMetric(name string, valueStr string) error {
	metricValue, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf(`invalid metric value, must be a float`)
	}
	storage.UpdateGauge(name, metricValue)

	return nil
}
