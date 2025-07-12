package store

import (
	"context"
	"testing"

	"github.com/Soliard/go-tpl-metrics/models"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	t.Run("gauge operations", func(t *testing.T) {
		// Test UpdateGauge
		_, err := storage.UpdateMetric(ctx, models.NewGaugeMetric("testGauge", 123.45))
		assert.NoError(t, err)
		metric, err := storage.GetMetric(ctx, "testGauge")
		assert.NoError(t, err)
		assert.Equal(t, 123.45, *metric.Value)

		// Test overwrite
		_, err = storage.UpdateMetric(ctx, models.NewGaugeMetric("testGauge", 67.89))
		assert.NoError(t, err)
		metric, err = storage.GetMetric(ctx, "testGauge")
		assert.NoError(t, err)
		assert.Equal(t, 67.89, *metric.Value)

		// Test non-existent gauge
		_, err = storage.GetMetric(ctx, "nonExistent")
		assert.Error(t, err)
	})

	t.Run("counter operations", func(t *testing.T) {
		// Test UpdateCounter
		_, err := storage.UpdateMetric(ctx, models.NewCounterMetric("testCounter", 10))
		assert.NoError(t, err)
		metric, err := storage.GetMetric(ctx, "testCounter")
		assert.NoError(t, err)
		assert.Equal(t, int64(10), *metric.Delta)

		// Test increment
		_, err = storage.UpdateMetric(ctx, models.NewCounterMetric("testCounter", 20))
		assert.NoError(t, err)
		metric, err = storage.GetMetric(ctx, "testCounter")
		assert.NoError(t, err)
		assert.Equal(t, int64(30), *metric.Delta) // 10 + 20

		// Test non-existent counter
		_, err = storage.GetMetric(ctx, "nonExistent")
		assert.Error(t, err)
	})

	t.Run("gauge and counter interaction", func(t *testing.T) {
		_, err := storage.UpdateMetric(ctx, models.NewGaugeMetric("mixed", 100.0))
		assert.NoError(t, err)
		_, err = storage.UpdateMetric(ctx, models.NewCounterMetric("mixed", 200))
		assert.Error(t, err)
		metric, err := storage.GetMetric(ctx, "mixed")
		assert.NoError(t, err)
		assert.Equal(t, 100.0, *metric.Value)
	})
}
