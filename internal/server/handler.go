package server

import (
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/Soliard/go-tpl-metrics/internal/server/templates"
	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/go-chi/chi/v5"
)

func (s *Service) UpdateHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("[UpdateHandler] Входящий запрос: %s %s\n", req.Method, req.URL.RawPath)
	metric := parseMetricURL(req)
	fmt.Printf("[UpdateHandler] Парсинг метрики: type=%s, name=%s, value=%s\n",
		metric.MType, metric.ID, metric.Value)
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
		err := s.updateCounterMetric(metric.ID, metric.Value)
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

func (s *Service) ValueHandler(res http.ResponseWriter, req *http.Request) {
	m := parseMetricURL(req)

	if m.MType == "" || m.ID == "" {
		http.Error(res, `type or name cannot be empty`, http.StatusBadRequest)
		return
	}
	if metric, exists := s.GetMetric(m.ID); exists {
		if metric.MType == m.MType {
			dto := models.СonvertToMetricStringDTO(metric)
			if m.MType == models.Counter {
				res.Write([]byte(fmt.Sprintf(`%s`, dto.Delta)))
			} else if dto.MType == models.Gauge {
				res.Write([]byte(fmt.Sprintf(`%s`, dto.Value)))
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

func (s *Service) MetricsPageHandler(res http.ResponseWriter, req *http.Request) {
	tmpl, err := template.New("metrics").Parse(templates.MetricsTemplate)

	if err != nil {
		http.Error(res, "Error loading template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := models.MetricsPageData{
		Metrics: s.storage.GetAllMetricsStringDTO(),
	}

	res.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := tmpl.Execute(res, data); err != nil {
		http.Error(res, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func parseMetricURL(req *http.Request) models.MetricStringDTO {
	metric := models.MetricStringDTO{}
	metric.MType = chi.URLParam(req, "type")
	metric.ID = chi.URLParam(req, "name")
	metric.Value = chi.URLParam(req, "value")
	return metric
}

func (s *Service) updateCounterMetric(name string, valueStr string) error {
	metricValue, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return fmt.Errorf(`invalid metric value, must be an integer`)
	}
	err = s.UpdateCounter(name, metricValue)

	return err
}

func (s *Service) updateGaugeMetric(name string, valueStr string) error {
	metricValue, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf(`invalid metric value, must be a float`)
	}
	s.storage.UpdateGauge(name, metricValue)

	return nil
}
