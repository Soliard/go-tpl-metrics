package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Soliard/go-tpl-metrics/cmd/server/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"go.uber.org/zap"
)

func main() {

	log.Print("server starting...")
	config, err := config.New()
	if err != nil {
		log.Printf("cannot create config for server")
		os.Stdout.Sync()
		os.Exit(1)
	}

	logger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Printf("failed to initialize logger: %v", err)
		os.Stdout.Sync()
		os.Exit(1)
	}
	logger.Info("Server config: ", zap.Any("config", config))
	storage, err := store.New(context.TODO(), config)
	if err != nil {
		logger.Error("error while creating storage", zap.Error(err))
		os.Stdout.Sync()
		os.Exit(1)
	}
	logger.Sugar().Infof("storage type: %T", storage)

	service := server.NewMetricsService(storage, config, logger)
	metricRouter := server.MetricRouter(service)

	logger.Info("Server starting to listen on ", zap.String("ServerHost", service.ServerHost))
	err = http.ListenAndServe(service.ServerHost, metricRouter)
	if err != nil {
		logger.Error("Fatal error while server serving", zap.Error(err))
		os.Stdout.Sync()
		os.Exit(1)
	}
}
