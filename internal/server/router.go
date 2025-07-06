package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func MetricRouter(s *MetricsService) chi.Router {
	r := chi.NewRouter()
	r.Use(s.LoggingMiddleware, s.GzipMiddleware)
	r.Get("/", s.MetricsPageHandler)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", s.UpdateHandler)
		r.Post("/{type}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })
		r.Post("/{type}/{name}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadRequest) })
		r.Post("/{type}/{name}/{value}", s.UpdateViaURLHandler)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", s.ValueHandler)
		r.Get("/{type}/{name}", s.ValueViaURLHandler)
	})

	return r
}
