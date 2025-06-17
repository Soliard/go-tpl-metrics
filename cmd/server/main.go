package main

import (
	"fmt"
	"net/http"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/Soliard/go-tpl-metrics/internal/store"
)

func main() {
	if err := logger.InitLogger("server"); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	logger.LogInfo("server", "Starting server...")

	storage := store.NewStorage()
	config := config.New()
	logger.LogConfig("server", config)

	service := server.NewService(storage, config)
	metricRouter := server.MetricRouter(service)

	logger.LogInfo("server", fmt.Sprintf("Server starting to listen on %s", service.ServerHost))
	err := http.ListenAndServe(service.ServerHost, metricRouter)
	if err != nil {
		logger.LogError("server", err)
		panic(err)
	}
}
