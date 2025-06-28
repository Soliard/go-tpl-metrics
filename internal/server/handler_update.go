package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *MetricsService) UpdateHandler(res http.ResponseWriter, req *http.Request) {
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
		return
	}

	if metric.ID == "" {
		s.Logger.Warn("update handler recieved metric with empty id")
		http.Error(res, `metric id cannot be empty`, http.StatusNotFound)
		return
	}

	switch metric.MType {
	case models.Gauge:
		err := s.UpdateGauge(metric.ID, metric.Value)
		if err != nil {
			s.Logger.Error("error while update gauge metric", zap.Error(err), zap.Any("recieved metric", metric))
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case models.Counter:
		err := s.UpdateCounter(metric.ID, metric.Delta)
		if err != nil {
			s.Logger.Error("error while update counter metric", zap.Error(err), zap.Any("recieved metric", metric))
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		s.Logger.Warn("recieved unknown metric type to update", zap.Any("recieved metric", metric))
		http.Error(res, `invalid metric type`, http.StatusBadRequest)
		return
	}

	retMetric, ok := s.GetMetric(metric.ID)
	if !ok {
		s.Logger.Error("cant get metric that was updated right now",
			zap.Error(err),
			zap.String("metric id", metric.ID),
			zap.Any("recieved metric", metric))
		http.Error(res, "cant get metric that was updated right now", http.StatusInternalServerError)
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

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(retBody)
}

func (s *MetricsService) UpdateViaURLHandler(res http.ResponseWriter, req *http.Request) {
	metric := parseMetricURL(req)

	if metric.ID == "" {
		http.Error(res, `metric name cannot be empty`, http.StatusNotFound)
		return
	}

	if metric.Delta == nil && metric.Value == nil {
		http.Error(res, "empty value", http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case models.Gauge:
		err := s.UpdateGauge(metric.ID, metric.Value)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case models.Counter:
		err := s.UpdateCounter(metric.ID, metric.Delta)
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

func parseMetricURL(req *http.Request) models.Metrics {
	metric := models.Metrics{
		MType: chi.URLParam(req, "type"),
		ID:    chi.URLParam(req, "name"),
	}

	// Парсим значение в зависимости от типа метрики
	valueStr := chi.URLParam(req, "value")
	if metric.MType == models.Gauge {
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			metric.Value = &value
		}
	} else if metric.MType == models.Counter {
		if delta, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
			metric.Delta = &delta
		}
	}

	fmt.Printf("[parseMetricURL] Parsed metric: %s\n", metric.String())
	return metric
}
