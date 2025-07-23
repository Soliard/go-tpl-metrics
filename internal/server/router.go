package server

import (
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
	"github.com/go-chi/chi/v5"
)

func MetricRouter(s *MetricsService) chi.Router {
	r := chi.NewRouter()
	r.Use(
		logger.LoggingMiddleware(s.Logger),
		signer.VerifySignatureMiddleware(s.signKey, s.Logger),
		compressor.GzipMiddleware(s.Logger),
	)
	r.Get("/", s.MetricsPageHandler)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", s.UpdateHandler)
		r.Post("/{type}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })
		r.Post("/{type}/{name}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadRequest) })
		r.Post("/{type}/{name}/{value}", s.UpdateViaURLHandler)
	})
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", s.UpdatesHandler)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", s.ValueHandler)
		r.Get("/{type}/{name}", s.ValueViaURLHandler)
	})
	r.Get("/ping", s.PingHandler)

	return r
}
