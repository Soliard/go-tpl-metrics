package agent

import (
	"testing"
)

func TestNewStatsCollector(t *testing.T) {
	collector := NewStatsCollector()
	if collector.Gauges == nil {
		t.Error("Gauges map should be initialized")
	}
	if collector.counters == nil {
		t.Error("Counters map should be initialized")
	}
}

func TestCollect(t *testing.T) {
	collector := NewStatsCollector()

	if len(collector.Gauges) != 0 {
		t.Error("Gauges should be empty initially")
	}
	if collector.counters["PollCount"] != 0 {
		t.Error("PollCount should be 0 initially")
	}

	err := collector.Collect()
	if err != nil {
		t.Errorf("Collect should not return error: %v", err)
	}

	if len(collector.Gauges) == 0 {
		t.Error("Gauges should not be empty after collection")
	}
	if collector.counters["PollCount"] != 1 {
		t.Error("PollCount should be 1 after first collection")
	}

	requiredMetrics := []string{"Alloc", "GCCPUFraction", "HeapAlloc", "RandomValue"}
	for _, metric := range requiredMetrics {
		if _, exists := collector.Gauges[metric]; !exists {
			t.Errorf("Required metric %s not found", metric)
		}
	}
}
