package main

import (
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/misc"
	"github.com/Soliard/go-tpl-metrics/internal/server"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, server.UpdateClaimMetric)

	err := http.ListenAndServe(misc.GetServerURL(), mux)
	if err != nil {
		panic(err)
	}
}
