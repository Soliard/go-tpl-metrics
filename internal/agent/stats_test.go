package agent

import (
	"testing"

	"github.com/Soliard/go-tpl-metrics/internal/config"
	"github.com/Soliard/go-tpl-metrics/models"
	"go.uber.org/zap"
)

func TestCollector(t *testing.T) {
	cfg := config.AgentConfig{}
	agent := New(&cfg, zap.NewNop())
	jobs := make(chan []*models.Metrics)
	go agent.Collector(1, jobs)

	metrics := <-jobs

	if len(metrics) == 0 {
		t.Error("map should not be empty after collection")
	}

	metricMap := make(map[string]*models.Metrics)
	for _, metric := range metrics {
		metricMap[metric.ID] = metric
	}

	if pollCount, exists := metricMap["PollCount"]; !exists {
		t.Error("PollCount metric not found")
	} else if *pollCount.Delta != 1 {
		t.Error("PollCount should be 1 after first collection")
	}

	requiredMetrics := []string{"Alloc", "GCCPUFraction", "HeapAlloc", "RandomValue"}
	for _, metric := range requiredMetrics {
		if _, exists := metricMap[metric]; !exists {
			t.Errorf("Required metric %s not found", metric)
		}
	}
}
