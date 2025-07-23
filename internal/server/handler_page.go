package server

import (
	"bytes"
	"net/http"
	"sort"
	"text/template"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/server/templates"
	"go.uber.org/zap"
)

func (s *MetricsService) MetricsPageHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := logger.LoggerFromCtx(ctx, s.Logger)
	logger.Info("recieved request for metrics page handler")
	tmpl, err := template.New("metrics").Parse(templates.MetricsTemplate)

	if err != nil {
		logger.Error("error while loading template", zap.Error(err))
		http.Error(res, "error while loading page", http.StatusInternalServerError)
		return
	}

	data, err := s.storage.GetAllMetrics(ctx)
	if err != nil {
		logger.Error("error while getting all metrics for page table", zap.Error(err))
		http.Error(res, "something went wrong", http.StatusInternalServerError)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].ID < data[j].ID
	})

	var bufTemplate bytes.Buffer
	if err := tmpl.Execute(&bufTemplate, data); err != nil {
		http.Error(res, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write(bufTemplate.Bytes())
}
