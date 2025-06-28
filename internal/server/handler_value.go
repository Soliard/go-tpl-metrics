package server

import (
	"encoding/json"
	"net/http"

	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
)

func (s *MetricsService) ValueHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	if req.Header.Get("Content-type") != "application/json" {
		http.Error(res, "only application/json content accepting", http.StatusBadRequest)
		return
	}

	buf := json.NewDecoder(req.Body)
	defer req.Body.Close()
	metric := &models.Metrics{}
	err := buf.Decode(metric)
	if err != nil {
		s.Logger.Warn("cant decode body to metric type", zap.Error(err))
		http.Error(res, "cant decode body to metric type", http.StatusBadRequest)
	}

	retMetric, ok := s.GetMetric(metric.ID)
	if !ok {
		http.Error(res, `metric with this name doesnt exists`, http.StatusNotFound)
		return
	}
	retBody, err := json.Marshal(retMetric)
	if err != nil {
		s.Logger.Error("cant marshal metric",
			zap.Error(err),
			zap.Any("metric", retMetric))
		http.Error(res, "cant return metric", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write(retBody)

}

func (s *MetricsService) ValueViaURLHandler(res http.ResponseWriter, req *http.Request) {
	m := parseMetricURL(req)

	if m.MType == "" || m.ID == "" {
		http.Error(res, `type or name cannot be empty`, http.StatusBadRequest)
		return
	}
	if metric, exists := s.GetMetric(m.ID); exists {
		if metric.MType == m.MType {
			if m.MType == models.Counter {
				res.Write([]byte(metric.StringifyDelta()))
			} else if metric.MType == models.Gauge {
				res.Write([]byte(metric.StringifyValue()))
			}
		} else {
			http.Error(res, `invalid metric type`, http.StatusNotFound)
			return
		}
	} else {
		http.Error(res, `metric with this name doesnt exists`, http.StatusNotFound)
		return
	}
	res.Header().Set("Content-Type", "plain/text; charset=utf-8")
	res.WriteHeader(http.StatusOK)
}
