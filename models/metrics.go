package models

import (
	"fmt"
	"strings"
)

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
	var delta, value string
	if metric.Delta != nil {
		delta = fmt.Sprintf(`%d`, *metric.Delta)
	} else {
		delta = ``
	}

	if metric.Value != nil {
		value = fmt.Sprintf(`%.3f`, *metric.Value)
	} else {
		value = ``
	}
	value = strings.TrimRight(value, `0`)

	return MetricStringDTO{
		ID:    metric.ID,
		MType: metric.MType,
		Delta: delta,
		Value: value,
	}
}
