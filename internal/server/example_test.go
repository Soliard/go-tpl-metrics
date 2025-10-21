package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/Soliard/go-tpl-metrics/internal/config"
	"github.com/Soliard/go-tpl-metrics/internal/server"
	"github.com/Soliard/go-tpl-metrics/internal/store"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
)

// ExampleMetricRouter_UpdateHandler демонстрирует обновление метрики через JSON API
func ExampleMetricsService_UpdateHandler() {
	// Создаем тестовый сервер
	storage := store.NewMemoryStorage()
	logger, _ := zap.NewDevelopment()
	service := server.NewMetricsService(storage, &config.ServerConfig{}, logger)
	router := server.MetricRouter(service)
	server := httptest.NewServer(router)
	defer server.Close()

	// Создаем метрику для обновления
	metric := models.Metrics{
		ID:    "test_metric",
		MType: models.Gauge,
		Value: models.PFloat(42.5),
	}

	// Кодируем в JSON
	jsonData, _ := json.Marshal(metric)

	// Отправляем POST запрос
	resp, err := http.Post(server.URL+"/update", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// ExampleMetricRouter_UpdateViaURLHandler демонстрирует обновление метрики через URL параметры
func ExampleMetricsService_UpdateViaURLHandler() {
	// Создаем тестовый сервер
	storage := store.NewMemoryStorage()
	logger, _ := zap.NewDevelopment()
	service := server.NewMetricsService(storage, &config.ServerConfig{}, logger)
	router := server.MetricRouter(service)
	server := httptest.NewServer(router)
	defer server.Close()

	// Отправляем POST запрос с метрикой в URL
	resp, err := http.Post(server.URL+"/update/gauge/memory_usage/85.3", "text/plain", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	// Output: Status: 200
}

// ExampleMetricRouter_ValueHandler демонстрирует получение метрики через JSON API
func ExampleMetricsService_ValueHandler() {
	// Создаем тестовый сервер
	storage := store.NewMemoryStorage()
	logger, _ := zap.NewDevelopment()
	service := server.NewMetricsService(storage, &config.ServerConfig{}, logger)
	router := server.MetricRouter(service)
	server := httptest.NewServer(router)
	defer server.Close()

	// Сначала создаем метрику
	metric := models.Metrics{
		ID:    "cpu_usage",
		MType: models.Gauge,
		Value: models.PFloat(75.2),
	}
	jsonData, _ := json.Marshal(metric)
	res, err := http.Post(server.URL+"/update", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer res.Body.Close()

	// Теперь получаем метрику
	requestMetric := models.Metrics{ID: "cpu_usage"}
	requestData, _ := json.Marshal(requestMetric)
	resp, err := http.Post(server.URL+"/value", "application/json", bytes.NewBuffer(requestData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result models.Metrics
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Metric: %s = %.1f\n", result.ID, *result.Value)
	// Output: Metric: cpu_usage = 75.2
}

// ExampleMetricRouter_ValueViaURLHandler демонстрирует получение метрики через URL
func ExampleMetricsService_ValueViaURLHandler() {
	// Создаем тестовый сервер
	storage := store.NewMemoryStorage()
	logger, _ := zap.NewDevelopment()
	service := server.NewMetricsService(storage, &config.ServerConfig{}, logger)
	router := server.MetricRouter(service)
	server := httptest.NewServer(router)
	defer server.Close()

	// Сначала создаем метрику
	resp, err := http.Post(server.URL+"/update/counter/request_count/5", "text/plain", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Получаем метрику через URL
	resp, err = http.Get(server.URL + "/value/counter/request_count")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	bod, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Counter value: %s\n", string(bod))
	// Output: Counter value: 5
}

// ExampleMetricRouter_UpdatesHandler демонстрирует пакетное обновление метрик
func ExampleMetricsService_UpdatesHandler() {
	// Создаем тестовый сервер
	storage := store.NewMemoryStorage()
	logger, _ := zap.NewDevelopment()
	service := server.NewMetricsService(storage, &config.ServerConfig{}, logger)
	router := server.MetricRouter(service)
	server := httptest.NewServer(router)
	defer server.Close()

	// Создаем пакет метрик
	metrics := []*models.Metrics{
		models.NewGaugeMetric("memory_usage", 85.3),
		models.NewGaugeMetric("cpu_usage", 45.7),
		models.NewCounterMetric("request_count", 100),
	}

	// Кодируем в JSON
	jsonData, _ := json.Marshal(metrics)

	// Отправляем пакетный запрос
	resp, err := http.Post(server.URL+"/updates", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Batch update status: %d\n", resp.StatusCode)
	// Output: Batch update status: 200
}

// ExampleMetricRouter_PingHandler демонстрирует проверку состояния сервера
func ExampleMetricsService_PingHandler() {
	// Создаем тестовый сервер
	storage := store.NewMemoryStorage()
	logger, _ := zap.NewDevelopment()
	service := server.NewMetricsService(storage, &config.ServerConfig{}, logger)
	router := server.MetricRouter(service)
	server := httptest.NewServer(router)
	defer server.Close()

	// Проверяем состояние сервера
	resp, err := http.Get(server.URL + "/ping")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Ping status: %d\n", resp.StatusCode)
	// Output: Ping status: 200
}

// ExampleMetricRouter_MetricsPageHandler демонстрирует получение HTML страницы с метриками
func ExampleMetricsService_MetricsPageHandler() {
	// Создаем тестовый сервер
	storage := store.NewMemoryStorage()
	logger, _ := zap.NewDevelopment()
	service := server.NewMetricsService(storage, &config.ServerConfig{}, logger)
	router := server.MetricRouter(service)
	server := httptest.NewServer(router)
	defer server.Close()

	// Добавляем несколько метрик
	resp, err := http.Post(server.URL+"/update/gauge/memory/85.3", "text/plain", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	resp, err = http.Post(server.URL+"/update/counter/requests/10", "text/plain", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Получаем HTML страницу
	resp, err = http.Get(server.URL + "/")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Page status: %d, Content-Type: %s\n", resp.StatusCode, resp.Header.Get("Content-Type"))
	// Output: Page status: 200, Content-Type: text/html; charset=utf-8
}
