// Package main содержит точку входа для агента сбора метрик.
// Инициализирует конфигурацию, логгер и запускает агент для сбора системных метрик.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/Soliard/go-tpl-metrics/internal/config"
	"github.com/Soliard/go-tpl-metrics/internal/logger"
)

// Глобальные переменные для информации о сборке
// Может быть использовано с -ldflags "-X main.buildVersion=v1 -X main.buildDate=05.10.2025"
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// main инициализирует и запускает агент для сбора метрик.
// Создает конфигурацию, логгер и запускает сбор метрик в фоновом режиме.
func main() {
	printBuildInfo()

	defer os.Stdout.Sync()
	config, err := config.NewAgentConfig()
	if err != nil {
		log.Fatalf("cannot create config for agent %v", err)
	}

	logger, err := logger.New(config.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()
	fmt.Printf("agent config: %v", config)

	a := agent.New(config, logger)
	fmt.Printf("agent works with service on %s", config.ServerHost)

	// Контекст и обработка сигналов
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigCh
		fmt.Print("shutdown signal received, stopping agent...")
		cancel()
	}()

	a.Run(ctx)
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
