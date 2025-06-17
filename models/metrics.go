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

func (m *Metrics) StringifyDelta() string {
	if m.Delta != nil {
		return fmt.Sprintf(`%d`, *m.Delta)
	} else {
		return ``
	}
}

func (m *Metrics) StringifyValue() string {
	if m.Value != nil {
		return strings.TrimRight(fmt.Sprintf(`%.3f`, *m.Value), `0`)
	} else {
		return ``
	}
}

func (m *Metrics) String() string {
	return fmt.Sprintf("{ID: %s, Type: %s, Value: %s, Delta: %s, Hash: %s}",
		m.ID,
		m.MType,
		m.StringifyValue(),
		m.StringifyDelta(),
		m.Hash)
}
