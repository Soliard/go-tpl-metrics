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
	defer os.Stdout.Sync()
	config, err := config.New()
	if err != nil {
		log.Fatalf("FATAL: cannot create config: %v", err)
	}

	logger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	logger.Warn("server config: ", zap.Any("config", config))

	storage, err := store.New(context.TODO(), config)
	if err != nil {
		logger.Fatal("error while creating storage", zap.Error(err))
	}
	logger.Sugar().Warnf("storage type: %T", storage)

	service := server.NewMetricsService(storage, config, logger)
	metricRouter := server.MetricRouter(service)

	err = http.ListenAndServe(service.ServerHost, metricRouter)
	if err != nil {
		logger.Fatal("fatal error while server serving", zap.Error(err))
	}
}
