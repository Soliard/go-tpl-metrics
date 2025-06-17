package agent

import (
	"fmt"
	"math/rand"
	"runtime"
)

type StatsCollector struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		Gauges:   make(map[string]float64),
		Counters: map[string]int64{},
	}
}

func (s *StatsCollector) Collect() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	s.Gauges["Alloc"] = float64(m.Alloc)
	s.Gauges["BuckHashSys"] = float64(m.BuckHashSys)
	s.Gauges["Frees"] = float64(m.Frees)
	s.Gauges["GCCPUFraction"] = float64(m.GCCPUFraction)
	s.Gauges["GCSys"] = float64(m.GCSys)
	s.Gauges["HeapAlloc"] = float64(m.HeapAlloc)
	s.Gauges["HeapIdle"] = float64(m.HeapIdle)
	s.Gauges["HeapInuse"] = float64(m.HeapInuse)
	s.Gauges["HeapObjects"] = float64(m.HeapObjects)
	s.Gauges["HeapReleased"] = float64(m.HeapReleased)
	s.Gauges["HeapSys"] = float64(m.HeapSys)
	s.Gauges["LastGC"] = float64(m.LastGC)
	s.Gauges["Lookups"] = float64(m.Lookups)
	s.Gauges["MCacheInuse"] = float64(m.MCacheInuse)
	s.Gauges["MCacheSys"] = float64(m.MCacheSys)
	s.Gauges["MSpanInuse"] = float64(m.MSpanInuse)
	s.Gauges["MSpanSys"] = float64(m.MSpanSys)
	s.Gauges["Mallocs"] = float64(m.Mallocs)
	s.Gauges["NextGC"] = float64(m.NextGC)
	s.Gauges["NumForcedGC"] = float64(m.NumForcedGC)
	s.Gauges["NumGC"] = float64(m.NumGC)
	s.Gauges["OtherSys"] = float64(m.OtherSys)
	s.Gauges["PauseTotalNs"] = float64(m.PauseTotalNs)
	s.Gauges["StackInuse"] = float64(m.StackInuse)
	s.Gauges["StackSys"] = float64(m.StackSys)
	s.Gauges["Sys"] = float64(m.Sys)
	s.Gauges["TotalAlloc"] = float64(m.TotalAlloc)
	s.Gauges["RandomValue"] = float64(rand.Float64())

	s.Counters["PollCount"]++

	fmt.Println("Stats collected")

	return nil
}
