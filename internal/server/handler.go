package server

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"text/template"

	"github.com/Soliard/go-tpl-metrics/internal/server/templates"
	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/go-chi/chi/v5"
)

func (s *MetricsService) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	metric := parseMetricURL(req)

	if metric.ID == "" {
		http.Error(res, `metric name cannot be empty`, http.StatusNotFound)
		return
	}

	switch metric.MType {
	case models.Gauge:
		err := s.updateGaugeMetric(metric.ID, metric.Value)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case models.Counter:
		err := s.updateCounterMetric(metric.ID, metric.Delta)
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

func (s *MetricsService) ValueHandler(res http.ResponseWriter, req *http.Request) {
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

func (s *MetricsService) MetricsPageHandler(res http.ResponseWriter, req *http.Request) {
	tmpl, err := template.New("metrics").Parse(templates.MetricsTemplate)

	if err != nil {
		http.Error(res, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := s.storage.GetAllMetrics()

	sort.Slice(data, func(i, j int) bool {
		return data[i].ID < data[j].ID
	})

	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := tmpl.Execute(res, data); err != nil {
		http.Error(res, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
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

func (s *MetricsService) updateCounterMetric(name string, value *int64) error {
	if value == nil {
		return fmt.Errorf(`invalid metric value`)
	}
	return s.UpdateCounter(name, value)
}

func (s *MetricsService) updateGaugeMetric(name string, value *float64) error {
	if value == nil {
		return fmt.Errorf(`invalid metric value`)
	}
	return s.storage.UpdateGauge(name, value)
}
