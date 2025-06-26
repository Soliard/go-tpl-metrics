package agent

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/Soliard/go-tpl-metrics/models"
)

type StatsCollector struct {
	Metrics map[string]*models.Metrics
}

func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		Metrics: map[string]*models.Metrics{},
	}
}

func (s *StatsCollector) UpdateGauge(id string, value float64) {
	if v, ok := s.Metrics[id]; ok {
		if v.MType != models.Gauge {
			panic(errors.New("collector trying to update counter metric with gauge value"))
		}
		v.Value = &value
	} else {
		s.Metrics[id] = models.NewGaugeMetric(id, value)
	}
}

func (s *StatsCollector) UpdateCounter(id string) {
	if v, ok := s.Metrics[id]; ok {
		if v.MType != models.Counter {
			panic(errors.New("collector trying to update gauge metric with counter value"))
		}
		*v.Delta += 1
	} else {
		s.Metrics[id] = models.NewCounterMetric(id, 1)
	}
}

func (s *StatsCollector) Collect() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	s.UpdateGauge("Alloc", float64(m.Alloc))
	s.UpdateGauge("BuckHashSys", float64(m.BuckHashSys))
	s.UpdateGauge("Frees", float64(m.Frees))
	s.UpdateGauge("GCCPUFraction", float64(m.GCCPUFraction))
	s.UpdateGauge("GCSys", float64(m.GCSys))
	s.UpdateGauge("HeapAlloc", float64(m.HeapAlloc))
	s.UpdateGauge("HeapIdle", float64(m.HeapIdle))
	s.UpdateGauge("HeapInuse", float64(m.HeapInuse))
	s.UpdateGauge("HeapObjects", float64(m.HeapObjects))
	s.UpdateGauge("HeapReleased", float64(m.HeapReleased))
	s.UpdateGauge("HeapSys", float64(m.HeapSys))
	s.UpdateGauge("LastGC", float64(m.LastGC))
	s.UpdateGauge("Lookups", float64(m.Lookups))
	s.UpdateGauge("MCacheInuse", float64(m.MCacheInuse))
	s.UpdateGauge("MCacheSys", float64(m.MCacheSys))
	s.UpdateGauge("MSpanInuse", float64(m.MSpanInuse))
	s.UpdateGauge("MSpanSys", float64(m.MSpanSys))
	s.UpdateGauge("Mallocs", float64(m.Mallocs))
	s.UpdateGauge("NextGC", float64(m.NextGC))
	s.UpdateGauge("NumForcedGC", float64(m.NumForcedGC))
	s.UpdateGauge("NumGC", float64(m.NumGC))
	s.UpdateGauge("OtherSys", float64(m.OtherSys))
	s.UpdateGauge("PauseTotalNs", float64(m.PauseTotalNs))
	s.UpdateGauge("StackInuse", float64(m.StackInuse))
	s.UpdateGauge("StackSys", float64(m.StackSys))
	s.UpdateGauge("Sys", float64(m.Sys))
	s.UpdateGauge("TotalAlloc", float64(m.TotalAlloc))
	s.UpdateGauge("RandomValue", float64(rand.Float64()))

	s.UpdateCounter("PollCount")

	fmt.Println("Stats collected")

	return nil
}
