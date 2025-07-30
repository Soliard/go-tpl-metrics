package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *MetricsService) UpdatesHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := logger.LoggerFromCtx(ctx, s.Logger)
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "only application/json content accepting", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	metrics := []*models.Metrics{}
	err := json.NewDecoder(req.Body).Decode(&metrics)
	if err != nil {
		logger.Warn("cant decode body to metric slice", zap.Error(err))
		http.Error(res, "cant decode body to metrics", http.StatusBadRequest)
		return
	}
	err = s.UpdateMetrics(ctx, metrics)
	if err != nil {
		if errors.Is(err, store.ErrInvalidMetricReceived) {
			logger.Warn("recieved one or more invalid metric", zap.Error(err))
			http.Error(res, "recieved one or more invalid metric", http.StatusBadRequest)
			return
		}
		logger.Error("error while batch metrics update", zap.Error(err))
		http.Error(res, "error while batch metrics update", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)

}

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
	retMetric, err := s.UpdateMetric(ctx, metric)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(res, `metric is not found or id is empty`, http.StatusNotFound)
			return
		}
		if errors.Is(err, store.ErrInvalidMetricReceived) {
			http.Error(res, "invalid metric recieved", http.StatusBadRequest)
			return
		}
		logger.Error("cant update metric",
			zap.Any("metric", metric),
			zap.Error(err))
		http.Error(res, "cant update metric", http.StatusInternalServerError)
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

	_, err := s.UpdateMetric(ctx, &metric)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(res, "metric is not found or id is empty", http.StatusNotFound)
			return
		}
		if errors.Is(err, store.ErrInvalidMetricReceived) {
			http.Error(res, "invalid metric recieved", http.StatusBadRequest)
			return
		}
		logger.Error("cant update metric",
			zap.Any("metric", metric),
			zap.Error(err))
		http.Error(res, "cant update metric", http.StatusInternalServerError)
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
