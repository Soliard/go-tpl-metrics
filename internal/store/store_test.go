package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()
	t.Run("gauge operations", func(t *testing.T) {
		// Test UpdateGauge
		var delta1 = 123.45
		err := storage.UpdateGauge(ctx, "testGauge", &delta1)
		assert.NoError(t, err)
		metric, exists := storage.GetMetric(ctx, "testGauge")
		assert.True(t, exists)
		assert.Equal(t, 123.45, *metric.Value)

		// Test overwrite
		var delta2 = 67.89
		err = storage.UpdateGauge(ctx, "testGauge", &delta2)
		assert.NoError(t, err)
		metric, exists = storage.GetMetric(ctx, "testGauge")
		assert.True(t, exists)
		assert.Equal(t, 67.89, *metric.Value)

		// Test non-existent gauge
		_, exists = storage.GetMetric(ctx, "nonExistent")
		assert.False(t, exists)
	})

	t.Run("counter operations", func(t *testing.T) {
		// Test UpdateCounter
		var value1 int64 = 10
		err := storage.UpdateCounter(ctx, "testCounter", &value1)
		assert.NoError(t, err)
		metric, exists := storage.GetMetric(ctx, "testCounter")
		assert.True(t, exists)
		assert.Equal(t, int64(10), *metric.Delta)

		// Test increment
		var value2 int64 = 20
		err = storage.UpdateCounter(ctx, "testCounter", &value2)
		assert.NoError(t, err)
		metric, exists = storage.GetMetric(ctx, "testCounter")
		assert.True(t, exists)
		assert.Equal(t, int64(30), *metric.Delta) // 10 + 20

		// Test non-existent counter
		_, exists = storage.GetMetric(ctx, "nonExistent")
		assert.False(t, exists)
	})

	t.Run("gauge and counter interaction", func(t *testing.T) {
		// Test that gauge overwrites counter
		var value1 int64 = 100
		err := storage.UpdateCounter(ctx, "mixed", &value1)
		assert.NoError(t, err)
		var delta1 = 200.0
		err = storage.UpdateGauge(ctx, "mixed", &delta1)
		assert.Error(t, err)
		metric, exists := storage.GetMetric(ctx, "mixed")
		assert.True(t, exists)
		assert.Equal(t, int64(100), *metric.Delta)
	})
}
