package main

import (
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/misc"
	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/Soliard/go-tpl-metrics/internal/store"
)

func main() {

	storage := store.NewStorage()
	service := server.NewService(storage)
	metricRouter := server.MetricRouter(service)

	err := http.ListenAndServe(misc.DefaultServerHost, metricRouter)
	if err != nil {
		panic(err)
	}
}
