package server

import (
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/go-chi/chi/v5"
)

func MetricRouter(s *MetricsService) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware(s.Logger))
	r.Get("/", s.MetricsPageHandler)
	r.Route("/update", func(r chi.Router) {
		// Полный путь с тремя параметрами
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadRequest) })
		r.Post("/{type}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })
		r.Post("/{type}/{name}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadRequest) })
		r.Post("/{type}/{name}/{value}", s.UpdateHandler)
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", s.ValueHandler)
	})

	return r
}
