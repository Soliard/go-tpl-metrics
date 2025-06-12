package models

import "fmt"

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

type MetricStringDTO struct {
	MType    string
	ID       string
	Value    string
	Delta    string
	AccValue string
	Hash     string
}

type MetricsPageData struct {
	Metrics []MetricStringDTO
}

func СonvertToMetricStringDTO(metric Metrics) MetricStringDTO {
	return MetricStringDTO{
		ID:    metric.ID,
		MType: metric.MType,
		Delta: fmt.Sprintf("%d", func() int64 {
			if metric.Delta != nil {
				return *metric.Delta
			}
			return 0
		}()),
		Value: fmt.Sprintf("%.3f", func() float64 {
			if metric.Value != nil {
				return *metric.Value
			}
			return 0.0
		}()),
	}
}
