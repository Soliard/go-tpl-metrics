package main

import (
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/server"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, server.UpdateClaimMetric)

	err := http.ListenAndServe(`localhost:8080`, mux)
	if err != nil {
		panic(err)
	}
}
