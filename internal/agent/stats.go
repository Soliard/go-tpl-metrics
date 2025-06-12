package agent

import (
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

func (sc *StatsCollector) Collect() error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	sc.Gauges["Alloc"] = float64(m.Alloc)
	sc.Gauges["BuckHashSys"] = float64(m.BuckHashSys)
	sc.Gauges["Frees"] = float64(m.Frees)
	sc.Gauges["GCCPUFraction"] = float64(m.GCCPUFraction)
	sc.Gauges["GCSys"] = float64(m.GCSys)
	sc.Gauges["HeapAlloc"] = float64(m.HeapAlloc)
	sc.Gauges["HeapIdle"] = float64(m.HeapIdle)
	sc.Gauges["HeapInuse"] = float64(m.HeapInuse)
	sc.Gauges["HeapObjects"] = float64(m.HeapObjects)
	sc.Gauges["HeapReleased"] = float64(m.HeapReleased)
	sc.Gauges["HeapSys"] = float64(m.HeapSys)
	sc.Gauges["LastGC"] = float64(m.LastGC)
	sc.Gauges["Lookups"] = float64(m.Lookups)
	sc.Gauges["MCacheInuse"] = float64(m.MCacheInuse)
	sc.Gauges["MCacheSys"] = float64(m.MCacheSys)
	sc.Gauges["MSpanInuse"] = float64(m.MSpanInuse)
	sc.Gauges["MSpanSys"] = float64(m.MSpanSys)
	sc.Gauges["Mallocs"] = float64(m.Mallocs)
	sc.Gauges["NextGC"] = float64(m.NextGC)
	sc.Gauges["NumForcedGC"] = float64(m.NumForcedGC)
	sc.Gauges["NumGC"] = float64(m.NumGC)
	sc.Gauges["OtherSys"] = float64(m.OtherSys)
	sc.Gauges["PauseTotalNs"] = float64(m.PauseTotalNs)
	sc.Gauges["StackInuse"] = float64(m.StackInuse)
	sc.Gauges["StackSys"] = float64(m.StackSys)
	sc.Gauges["Sys"] = float64(m.Sys)
	sc.Gauges["TotalAlloc"] = float64(m.TotalAlloc)
	sc.Gauges["RandomValue"] = float64(rand.Float64())

	sc.Counters["PollCount"]++

	return nil
}
