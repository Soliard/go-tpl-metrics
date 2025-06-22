package main

import (
	"fmt"
	"net/http"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"go.uber.org/zap"
)

func main() {

	config, err := config.New()
	if err != nil {
		panic(fmt.Errorf("cannot create config for agent %w", err))
	}

	logger, err := logger.New(logger.ComponentServer, config.LogLevel)

	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}
	logger.Log.Info("Agent config: ", zap.Any("config", config))

	storage := store.NewStorage()
	service := server.NewMetricsService(storage, config, logger)
	metricRouter := server.MetricRouter(service)

	logger.Log.Info("Server starting to listen on ", zap.String("ServerHost", service.ServerHost))
	err = http.ListenAndServe(service.ServerHost, metricRouter)
	if err != nil {
		logger.Log.Fatal("Fatal error while server serving", zap.Error(err))
	}
}
