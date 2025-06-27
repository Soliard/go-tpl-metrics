package server

import (
	"net/http"
	"sort"
	"text/template"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/server/templates"
)

func (s *MetricsService) MetricsPageHandler(res http.ResponseWriter, req *http.Request) {
	logger := logger.LoggerFromCtx(req.Context(), s.Logger)
	logger.Info("recieved request for metrics page handler")
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
