package agent

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
)

// Collector собирает метрики Go runtime (память, GC, горутины и т.д.).
// Запускается в отдельной горутине с заданным интервалом.
func (a *Agent) Collector(id int, result chan<- []*models.Metrics) {
	var m runtime.MemStats
	polCount := 0
	for {
		time.Sleep(a.pollInterval)
		runtime.ReadMemStats(&m)
		batch := make([]*models.Metrics, 0, 28)
		polCount++
		batch = append(batch,
			models.NewGaugeMetric("Alloc", float64(m.Alloc)),
			models.NewGaugeMetric("BuckHashSys", float64(m.BuckHashSys)),
			models.NewGaugeMetric("Frees", float64(m.Frees)),
			models.NewGaugeMetric("GCCPUFraction", float64(m.GCCPUFraction)),
			models.NewGaugeMetric("GCSys", float64(m.GCSys)),
			models.NewGaugeMetric("HeapAlloc", float64(m.HeapAlloc)),
			models.NewGaugeMetric("HeapIdle", float64(m.HeapIdle)),
			models.NewGaugeMetric("HeapInuse", float64(m.HeapInuse)),
			models.NewGaugeMetric("HeapObjects", float64(m.HeapObjects)),
			models.NewGaugeMetric("HeapReleased", float64(m.HeapReleased)),
			models.NewGaugeMetric("HeapSys", float64(m.HeapSys)),
			models.NewGaugeMetric("LastGC", float64(m.LastGC)),
			models.NewGaugeMetric("Lookups", float64(m.Lookups)),
			models.NewGaugeMetric("MCacheInuse", float64(m.MCacheInuse)),
			models.NewGaugeMetric("MCacheSys", float64(m.MCacheSys)),
			models.NewGaugeMetric("MSpanInuse", float64(m.MSpanInuse)),
			models.NewGaugeMetric("MSpanSys", float64(m.MSpanSys)),
			models.NewGaugeMetric("Mallocs", float64(m.Mallocs)),
			models.NewGaugeMetric("NextGC", float64(m.NextGC)),
			models.NewGaugeMetric("NumForcedGC", float64(m.NumForcedGC)),
			models.NewGaugeMetric("NumGC", float64(m.NumGC)),
			models.NewGaugeMetric("OtherSys", float64(m.OtherSys)),
			models.NewGaugeMetric("PauseTotalNs", float64(m.PauseTotalNs)),
			models.NewGaugeMetric("StackInuse", float64(m.StackInuse)),
			models.NewGaugeMetric("StackSys", float64(m.StackSys)),
			models.NewGaugeMetric("Sys", float64(m.Sys)),
			models.NewGaugeMetric("TotalAlloc", float64(m.TotalAlloc)),
			models.NewGaugeMetric("RandomValue", float64(rand.Float64())),
			models.NewCounterMetric("PollCount", int64(polCount)),
		)

		result <- batch

		a.Logger.Info("memory stats colected by collector", zap.Int("collector id", id))
	}
}

// CollectorPS собирает системные метрики (память и CPU) через gopsutil.
// Запускается в отдельной горутине с заданным интервалом.
func (a *Agent) CollectorPS(id int, result chan<- []*models.Metrics) {
	for {
		time.Sleep(a.pollInterval)
		memory, err := mem.VirtualMemory()
		if err != nil {
			a.Logger.Error("failed to get memory stats", zap.Error(err))
			time.Sleep(a.pollInterval)
			continue
		}

		cpuPercents, err := cpu.Percent(time.Second, true)
		if err != nil {
			a.Logger.Error("failed to get CPU stats", zap.Error(err))
			time.Sleep(a.pollInterval)
			continue
		}

		batch := make([]*models.Metrics, 0, len(cpuPercents)+2)
		batch = append(batch,
			models.NewGaugeMetric("TotalMemory", float64(memory.Total)),
			models.NewGaugeMetric("FreeMemory", float64(memory.Free)),
		)
		for i, c := range cpuPercents {
			batch = append(batch, models.NewGaugeMetric(fmt.Sprint("CPUutilization", i), c))
		}
		result <- batch

		a.Logger.Info("ps stats colected by collectorps", zap.Int("collectorps id", id))
	}
}
