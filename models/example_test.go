package models_test

import (
	"fmt"

	"github.com/Soliard/go-tpl-metrics/models"
)

// ExampleNewGaugeMetric демонстрирует создание метрики типа gauge
func ExampleNewGaugeMetric() {
	// Создаем gauge метрику
	metric := models.NewGaugeMetric("memory_usage", 85.3)

	fmt.Printf("Type: %s\n", metric.MType)
	fmt.Printf("ID: %s\n", metric.ID)
	fmt.Printf("Value: %.1f\n", *metric.Value)
	// Output:
	// Type: gauge
	// ID: memory_usage
	// Value: 85.3
}

// ExampleNewCounterMetric демонстрирует создание метрики типа counter
func ExampleNewCounterMetric() {
	// Создаем counter метрику
	metric := models.NewCounterMetric("request_count", 150)

	fmt.Printf("Type: %s\n", metric.MType)
	fmt.Printf("ID: %s\n", metric.ID)
	fmt.Printf("Delta: %d\n", *metric.Delta)
	// Output:
	// Type: counter
	// ID: request_count
	// Delta: 150
}

// ExampleMetrics_StringifyValue демонстрирует форматирование значения gauge метрики
func ExampleMetrics_StringifyValue() {
	metric := models.NewGaugeMetric("cpu_usage", 75.2500)

	fmt.Printf("Formatted value: '%s'\n", metric.StringifyValue())
	// Output:
	// Formatted value: '75.25'
}

// ExampleMetrics_StringifyDelta демонстрирует форматирование значения counter метрики
func ExampleMetrics_StringifyDelta() {
	metric := models.NewCounterMetric("errors", 42)

	fmt.Printf("Formatted delta: '%s'\n", metric.StringifyDelta())
	// Output:
	// Formatted delta: '42'
}

// ExampleMetrics_String демонстрирует строковое представление метрики
func ExampleMetrics_String() {
	metric := models.NewGaugeMetric("temperature", 23.5)

	fmt.Printf("String representation: %s\n", metric.String())
	// Output:
	// String representation: {ID: temperature, Type: gauge, Value: 23.5, Delta: , Hash: }
}

// ExamplePFloat демонстрирует создание указателя на float64
func ExamplePFloat() {
	value := 42.5
	ptr := models.PFloat(value)

	fmt.Printf("Original: %.1f\n", value)
	fmt.Printf("Pointer value: %.1f\n", *ptr)
	// Output:
	// Original: 42.5
	// Pointer value: 42.5
}

// ExamplePInt демонстрирует создание указателя на int64
func ExamplePInt() {
	value := int64(100)
	ptr := models.PInt(value)

	fmt.Printf("Original: %d\n", value)
	fmt.Printf("Pointer value: %d\n", *ptr)
	// Output:
	// Original: 100
	// Pointer value: 100
}
