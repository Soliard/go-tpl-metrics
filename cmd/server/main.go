package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"go.uber.org/zap"
)

func main() {

	fmt.Println("server starting...")
	config, err := config.New()
	if err != nil {
		log.Fatal("cannot create config for server")
	}

	logger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	logger.Info("Server config: ", zap.Any("config", config))

	storage, err := store.NewFileStorage(config.FileStoragePath, config.IsRestoreFromFile)
	if err != nil {
		logger.Fatal("error while creating storage", zap.Error(err))
	}
	service := server.NewMetricsService(storage, config, logger)
	metricRouter := server.MetricRouter(service)

	logger.Info("Server starting to listen on ", zap.String("ServerHost", service.ServerHost))
	err = http.ListenAndServe(service.ServerHost, metricRouter)
	if err != nil {
		logger.Fatal("Fatal error while server serving", zap.Error(err))
	}
}
