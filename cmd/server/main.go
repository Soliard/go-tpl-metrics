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
	logger, err := logger.New(logger.ComponentServer)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Close()
	logger.Info("Starting")

	storage := store.NewStorage()
	config := config.New(logger)
	logger.Info("Configure server ", config)

	service := server.NewService(storage, config, logger)
	metricRouter := server.MetricRouter(service)

	logger.Info("Server starting to listen on ", service.ServerHost)
	err = http.ListenAndServe(service.ServerHost, metricRouter)
	if err != nil {
		logger.Error(err)
		panic(err)
	}
}
