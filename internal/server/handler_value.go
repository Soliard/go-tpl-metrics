package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
)

func (s *MetricsService) ValueHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	res.Header().Set("Content-Type", "application/json")
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "only application/json content accepting", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()
	metric := &models.Metrics{}
	err := json.NewDecoder(req.Body).Decode(metric)
	if err != nil {
		s.Logger.Warn("cant decode body to metric type", zap.Error(err))
		http.Error(res, "cant decode body to metric type", http.StatusBadRequest)
	}

	retMetric, err := s.GetMetric(ctx, metric.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(res, "metric with this name doesnt exists", http.StatusNotFound)
			return
		}
		http.Error(res, "error while getting metric", http.StatusInternalServerError)
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
	ctx := req.Context()
	m := parseMetricURL(req)

	if m.MType == "" || m.ID == "" {
		http.Error(res, `type or name cannot be empty`, http.StatusBadRequest)
		return
	}
	metric, err := s.GetMetric(ctx, m.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(res, `metric with this name doesnt exists`, http.StatusNotFound)
			return
		}
		http.Error(res, `error while getting metric`, http.StatusInternalServerError)
		return
	}

	if metric.MType != m.MType {
		http.Error(res, `invalid metric type`, http.StatusNotFound)
		return
	}

	res.Header().Set("Content-Type", "plain/text; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	if m.MType == models.Counter {
		res.Write([]byte(metric.StringifyDelta()))
	}
	if metric.MType == models.Gauge {
		res.Write([]byte(metric.StringifyValue()))
	}
}
