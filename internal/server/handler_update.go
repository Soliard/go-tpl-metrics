package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *MetricsService) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := logger.LoggerFromCtx(ctx, s.Logger)
	if req.Header.Get("Content-type") != "application/json" {
		http.Error(res, "only application/json content accepting", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()
	metric := &models.Metrics{}
	err := json.NewDecoder(req.Body).Decode(metric)
	if err != nil {
		logger.Warn("cant decode body to metric type", zap.Error(err))
		http.Error(res, "cant decode body to metric type", http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		logger.Warn("update handler recieved metric with empty id")
		http.Error(res, `metric id cannot be empty`, http.StatusNotFound)
		return
	}

	if metric.MType != models.Gauge && metric.MType != models.Counter {
		logger.Warn(`invalid metric type`, zap.Any("metric", metric))
		http.Error(res, `invalid metric type`, http.StatusBadRequest)
		return
	}

	err = s.UpdateMetric(ctx, metric)
	if err != nil {
		logger.Error("cant update metric", zap.Any("metric", metric))
		http.Error(res, "cant update metric", http.StatusBadRequest)
		return
	}

	retMetric, ok := s.GetMetric(ctx, metric.ID)
	if !ok {
		logger.Error("cant get metric that was updated right now",
			zap.Error(err),
			zap.String("metric id", metric.ID),
			zap.Any("recieved metric", metric))
		http.Error(res, "cant get metric that was updated right now", http.StatusInternalServerError)
		return
	}

	retBody, err := json.Marshal(retMetric)
	if err != nil {
		logger.Error("cant marshal metric",
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
	ctx := req.Context()
	logger := logger.LoggerFromCtx(ctx, s.Logger)
	metric := parseMetricURL(req)
	if metric.ID == "" {
		http.Error(res, `metric name cannot be empty`, http.StatusNotFound)
		return
	}
	if metric.Delta == nil && metric.Value == nil {
		http.Error(res, "empty value", http.StatusBadRequest)
		return
	}
	if metric.MType != models.Gauge && metric.MType != models.Counter {
		http.Error(res, `invalid metric type`, http.StatusBadRequest)
		return
	}

	err := s.UpdateMetric(ctx, &metric)
	if err != nil {
		logger.Error("cant update metric", zap.Any("metric", metric))
		http.Error(res, "cant update metric", http.StatusBadRequest)
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
