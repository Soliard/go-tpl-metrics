// Package main содержит точку входа для HTTP сервера метрик.
// Инициализирует конфигурацию, логгер, хранилище и запускает HTTP сервер.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
    "net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Soliard/go-tpl-metrics/internal/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"go.uber.org/zap"
    "google.golang.org/grpc"
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

	fmt.Println("server starting...")
	defer os.Stdout.Sync()
	config, err := config.NewServerConfig()
	if err != nil {
		log.Fatalf("cannot create config: %v", err)
	}

	logger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	fmt.Printf("server config: %v", config)

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	storage, err := store.New(appCtx, config)
	if err != nil {
		logger.Fatal("error while creating storage", zap.Error(err))
	}
	fmt.Println("storage type: ", storage)

	service := server.NewMetricsService(storage, config, logger)
    // HTTP сервер (если адрес задан)
    var httpSrv *http.Server
    if config.ServerHost != "" {
        metricRouter := server.MetricRouter(service)
        httpSrv = &http.Server{
            Addr:    service.ServerHost,
            Handler: metricRouter,
        }
    }

    // gRPC сервер (если адрес задан)
    var grpcSrv *grpc.Server
    var grpcLis net.Listener
    if config.GRPCServerHost != "" {
        grpcSrv = server.NewGRPCServer(service)
        var err error
        grpcLis, err = net.Listen("tcp", config.GRPCServerHost)
        if err != nil {
            logger.Fatal("failed to listen grpc", zap.Error(err))
        }
    }

    if httpSrv == nil && grpcSrv == nil {
        logger.Fatal("no server address provided: set address and/or grpc_address")
    }

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

    if httpSrv != nil {
        go func() {
            if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                logger.Fatal("fatal error while http serving", zap.Error(err))
            }
        }()
    }
    if grpcSrv != nil {
        go func() {
            if err := grpcSrv.Serve(grpcLis); err != nil {
                logger.Fatal("fatal error while grpc serving", zap.Error(err))
            }
        }()
    }

	<-sigCh
	logger.Info("shutdown signal received, stopping server...")

    shCtx, shCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer shCancel()
    if httpSrv != nil {
        if err := httpSrv.Shutdown(shCtx); err != nil {
            logger.Error("HTTP server Shutdown", zap.Error(err))
        }
    }
    if grpcSrv != nil {
        grpcSrv.GracefulStop()
    }

	appCancel()

	fmt.Println("server shutdown gracefully")
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
