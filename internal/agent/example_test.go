package agent_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Soliard/go-tpl-metrics/cmd/agent/config"
	"github.com/Soliard/go-tpl-metrics/internal/agent"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
)

// ExampleAgent_Run демонстрирует запуск агента (с ограниченным временем)
func ExampleAgent_Run() {
	// Создаем конфигурацию
	cfg := &config.Config{
		ServerHost:            "localhost:8080",
		PollIntervalSeconds:   1,
		ReportIntervalSeconds: 2,
		LogLevel:              "warn",
		RequestsLimit:         5,
	}

	// Создаем логгер
	logger, _ := zap.NewDevelopment()

	// Создаем агента
	agent := agent.New(cfg, logger)

	// Запускаем агента в контексте с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fmt.Println("Starting agent...")
	agent.Run(ctx)
	fmt.Println("Agent stopped")
	// Output:
	// Starting agent...
	// Agent stopped
}

// ExampleAgent_Collector демонстрирует работу сборщика метрик
func ExampleAgent_Collector() {
	// Создаем конфигурацию
	cfg := &config.Config{
		PollIntervalSeconds: 1,
		LogLevel:            "warn",
	}

	// Создаем логгер
	logger, _ := zap.NewDevelopment()

	// Создаем агента
	agent := agent.New(cfg, logger)

	// Создаем канал для метрик
	metricsChan := make(chan []*models.Metrics, 1)

	// Запускаем сборщик в отдельной горутине
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go agent.Collector(1, metricsChan)

	// Получаем метрики
	select {
	case metrics := <-metricsChan:
		fmt.Printf("Collected %d metrics\n", len(metrics))
		for _, m := range metrics[:3] { // Показываем первые 3
			fmt.Printf("- %s: %s\n", m.ID, m.MType)
		}
	case <-ctx.Done():
		fmt.Println("Timeout waiting for metrics")
	}
	// Output:
	// Collected 29 metrics
	// - Alloc: gauge
	// - BuckHashSys: gauge
	// - Frees: gauge
}

// ExampleAgent_CollectorPS демонстрирует работу сборщика системных метрик
func ExampleAgent_CollectorPS() {
	// Создаем конфигурацию
	cfg := &config.Config{
		PollIntervalSeconds: 1,
		LogLevel:            "warn",
	}

	// Создаем логгер
	logger, _ := zap.NewDevelopment()

	// Создаем агента
	agent := agent.New(cfg, logger)

	// Создаем канал для метрик
	metricsChan := make(chan []*models.Metrics, 1)

	// Запускаем сборщик в отдельной горутине
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go agent.CollectorPS(1, metricsChan)

	// Получаем метрики
	select {
	case metrics := <-metricsChan:
		fmt.Printf("Collected %d system metrics\n", len(metrics))
	case <-ctx.Done():
		fmt.Println("Timeout waiting for metrics")
	}
	// Output:
	// Collected 18 system metrics
}
