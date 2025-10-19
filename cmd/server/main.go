// Package main содержит точку входа для HTTP сервера метрик.
// Инициализирует конфигурацию, логгер, хранилище и запускает HTTP сервер.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Soliard/go-tpl-metrics/internal/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"go.uber.org/zap"
)

// Глобальные переменные для информации о сборке
// Может быть использовано с -ldflags "-X main.buildVersion=v1 -X main.buildDate=05.10.2025"
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// main инициализирует и запускает HTTP сервер для сбора метрик.
// Создает конфигурацию, логгер, хранилище и HTTP роутер.
func main() {

	printBuildInfo()

	log.Print("server starting...")
	defer os.Stdout.Sync()
	config, err := config.NewServerConfig()
	if err != nil {
		log.Fatalf("FATAL: cannot create config: %v", err)
	}

	logger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	fmt.Printf("server config: %v", config)

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

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
