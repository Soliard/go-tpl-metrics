package server

import (
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/store"
)

func (s *MetricsService) PingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if db, ok := s.storage.(*store.DatabaseStorage); ok {
		if err := db.Ping(ctx); err != nil {
			http.Error(w, "database connection error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
