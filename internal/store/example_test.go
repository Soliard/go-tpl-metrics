package store_test

import (
	"context"
	"fmt"

	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
)

// ExampleNewMemoryStorage демонстрирует создание хранилища в памяти
func ExampleNewMemoryStorage() {
	// Создаем хранилище в памяти
	storage := store.NewMemoryStorage()

	fmt.Printf("Storage type: %T\n", storage)
	// Output:
	// Storage type: *store.memStorage
}

// ExampleStorage_UpdateMetric демонстрирует обновление метрики в хранилище
func ExampleStorage_UpdateMetric() {
	// Создаем хранилище
	storage := store.NewMemoryStorage()
	ctx := context.Background()

	// Создаем gauge метрику
	metric := models.NewGaugeMetric("temperature", 25.5)

	// Обновляем метрику
	updated, err := storage.UpdateMetric(ctx, metric)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Updated metric: %s = %.1f\n", updated.ID, *updated.Value)
	// Output:
	// Updated metric: temperature = 25.5
}

// ExampleStorage_GetMetric демонстрирует получение метрики из хранилища
func ExampleStorage_GetMetric() {
	// Создаем хранилище
	storage := store.NewMemoryStorage()
	ctx := context.Background()

	// Добавляем метрику
	metric := models.NewCounterMetric("requests", 100)
	storage.UpdateMetric(ctx, metric)

	// Получаем метрику
	retrieved, err := storage.GetMetric(ctx, "requests")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Retrieved: %s = %d\n", retrieved.ID, *retrieved.Delta)
	// Output:
	// Retrieved: requests = 100
}

// ExampleStorage_GetAllMetrics демонстрирует получение всех метрик
func ExampleStorage_GetAllMetrics() {
	// Создаем хранилище
	storage := store.NewMemoryStorage()
	ctx := context.Background()

	// Добавляем несколько метрик
	storage.UpdateMetric(ctx, models.NewGaugeMetric("cpu", 75.0))
	storage.UpdateMetric(ctx, models.NewCounterMetric("requests", 50))
	storage.UpdateMetric(ctx, models.NewGaugeMetric("memory", 60.5))

	// Получаем все метрики
	metrics, err := storage.GetAllMetrics(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Total metrics: %d\n", len(metrics))
	for _, m := range metrics {
		fmt.Printf("- %s (%s)\n", m.ID, m.MType)
	}
	// Output:
	// Total metrics: 3
	// - cpu (gauge)
	// - requests (counter)
	// - memory (gauge)
}

// ExampleStorage_UpdateMetrics демонстрирует пакетное обновление метрик
func ExampleStorage_UpdateMetrics() {
	// Создаем хранилище
	storage := store.NewMemoryStorage()
	ctx := context.Background()

	// Создаем пакет метрик
	metrics := []*models.Metrics{
		models.NewGaugeMetric("cpu_usage", 85.3),
		models.NewGaugeMetric("memory_usage", 70.1),
		models.NewCounterMetric("error_count", 5),
	}

	// Обновляем пакетом
	err := storage.UpdateMetrics(ctx, metrics)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Проверяем результат
	all, _ := storage.GetAllMetrics(ctx)
	fmt.Printf("Updated %d metrics\n", len(all))
	// Output:
	// Updated 3 metrics
}
