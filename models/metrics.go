// Этот пакет содержит основные типы данных для работы с метриками.
// Основная цель - предоставить структуру для представления метрик
// и удобные функции для их создания и форматирования.
package models

import (
	"fmt"
	"strings"
)

// Counter и Gauge - константы для типов метрик
const (
	// Counter представляет тип метрики-счетчика
	// Счетчики накапливают значения (например, количество запросов)
	Counter = "counter"
	// Gauge представляет тип метрики-измерителя
	// Измерители показывают текущее значение (например, использование памяти)
	Gauge = "gauge"
)

// Metrics представляет метрику в системе мониторинга.
// Используется для передачи данных между агентом и сервером.
// Delta и Value объявлены через указатели для различения значения "0" от не заданного значения.
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // подпись метрики для проверки целостности
}

// NewGaugeMetric создает новую метрику типа gauge с указанным именем и значением.
func NewGaugeMetric(id string, value float64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: Gauge,
		Value: &value,
	}
}

// NewCounterMetric создает новую метрику типа counter с указанным именем и дельтой.
func NewCounterMetric(id string, delta int64) *Metrics {
	return &Metrics{
		ID:    id,
		MType: Counter,
		Delta: &delta,
	}
}

// StringifyDelta возвращает строковое представление поля Delta.
// Возвращает пустую строку, если Delta равно nil.
func (m *Metrics) StringifyDelta() string {
	if m.Delta != nil {
		return fmt.Sprintf(`%d`, *m.Delta)
	} else {
		return ``
	}
}

// StringifyValue возвращает строковое представление поля Value.
// Возвращает пустую строку, если Value равно nil.
// Значение форматируется с точностью до 3 знаков после запятой.
func (m *Metrics) StringifyValue() string {
	if m.Value != nil {
		return strings.TrimRight(fmt.Sprintf(`%.3f`, *m.Value), `0`)
	} else {
		return ``
	}
}

// String возвращает строковое представление метрики в формате:
// "{ID: имя, Type: тип, Value: значение, Delta: дельта, Hash: хеш}"
func (m *Metrics) String() string {
	return fmt.Sprintf("{ID: %s, Type: %s, Value: %s, Delta: %s, Hash: %s}",
		m.ID,
		m.MType,
		m.StringifyValue(),
		m.StringifyDelta(),
		m.Hash)
}

// PFloat создает указатель на float64 значение.
// Удобная функция для создания указателей на float64.
func PFloat(float float64) *float64 {
	return &float
}

// PInt создает указатель на int64 значение.
// Удобная функция для создания указателей на int64.
func PInt(int int64) *int64 {
	return &int
}
