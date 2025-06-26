package agent

import (
	"testing"
)

func TestNewStatsCollector(t *testing.T) {
	collector := NewStatsCollector()
	if collector.Metrics == nil {
		t.Error("map should be initialized")
	}
}

func TestCollect(t *testing.T) {
	collector := NewStatsCollector()

	if len(collector.Metrics) != 0 {
		t.Error("map should be empty initially")
	}

	err := collector.Collect()
	if err != nil {
		t.Errorf("Collect should not return error: %v", err)
	}

	if len(collector.Metrics) == 0 {
		t.Error("map should not be empty after collection")
	}
	if *collector.Metrics["PollCount"].Delta != 1 {
		t.Error("PollCount should be 1 after first collection")
	}

	requiredMetrics := []string{"Alloc", "GCCPUFraction", "HeapAlloc", "RandomValue"}
	for _, metric := range requiredMetrics {
		if _, exists := collector.Metrics[metric]; !exists {
			t.Errorf("Required metric %s not found", metric)
		}
	}
}
