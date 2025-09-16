package server

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/Soliard/go-tpl-metrics/internal/compressor"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/signer"
	"github.com/go-chi/chi/v5"
)

func MetricRouter(s *MetricsService) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.LoggingMiddleware(s.Logger))

	// эндпоинты без подписи
	r.Group(func(r chi.Router) {
		r.Use(compressor.GzipMiddleware(s.Logger))
		r.Get("/", s.MetricsPageHandler)
		r.Get("/ping", s.PingHandler)
		r.Route("/value", func(r chi.Router) {
			r.Post("/", s.ValueHandler)
			r.Get("/{type}/{name}", s.ValueViaURLHandler)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", s.UpdateHandler)
			r.Post("/{type}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) })
			r.Post("/{type}/{name}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusBadRequest) })
			r.Post("/{type}/{name}/{value}", s.UpdateViaURLHandler)
		})
	})

	// эндпоинты с подписью
	r.Group(func(r chi.Router) {
		r.Use(
			signer.VerifySignatureMiddleware(s.signKey, s.Logger),
			signer.SignResponseMiddleware(s.signKey, s.Logger),
			compressor.GzipMiddleware(s.Logger),
		)
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", s.UpdatesHandler)
		})
	})

	r.Mount("/debug/pprof", http.DefaultServeMux)

	return r
}
