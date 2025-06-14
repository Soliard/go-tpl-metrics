package main

import (
	"fmt"
	"net/http"

	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/Soliard/go-tpl-metrics/internal/store"
)

func main() {

	storage := store.NewStorage()
	config := ParseFlags()
	service := server.NewService(storage, config)
	metricRouter := server.MetricRouter(service)

	fmt.Println("service Listen And Serve on ", service.ServerHost)
	err := http.ListenAndServe(service.ServerHost, metricRouter)
	if err != nil {
		panic(err)
	}
}
