package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage(t *testing.T) {
	storage := NewStorage()

	t.Run("gauge operations", func(t *testing.T) {
		// Test UpdateGauge
		err := storage.UpdateGauge("testGauge", 123.45)
		assert.NoError(t, err)
		metric, exists := storage.GetMetric("testGauge")
		assert.True(t, exists)
		assert.Equal(t, 123.45, *metric.Value)

		// Test overwrite
		err = storage.UpdateGauge("testGauge", 67.89)
		assert.NoError(t, err)
		metric, exists = storage.GetMetric("testGauge")
		assert.True(t, exists)
		assert.Equal(t, 67.89, *metric.Value)

		// Test non-existent gauge
		_, exists = storage.GetMetric("nonExistent")
		assert.False(t, exists)
	})

	t.Run("counter operations", func(t *testing.T) {
		// Test UpdateCounter
		err := storage.UpdateCounter("testCounter", 10)
		assert.NoError(t, err)
		metric, exists := storage.GetMetric("testCounter")
		assert.True(t, exists)
		assert.Equal(t, int64(10), *metric.Delta)

		// Test increment
		err = storage.UpdateCounter("testCounter", 20)
		assert.NoError(t, err)
		metric, exists = storage.GetMetric("testCounter")
		assert.True(t, exists)
		assert.Equal(t, int64(30), *metric.Delta) // 10 + 20

		// Test non-existent counter
		_, exists = storage.GetMetric("nonExistent")
		assert.False(t, exists)
	})

	t.Run("gauge and counter interaction", func(t *testing.T) {
		// Test that gauge overwrites counter
		err := storage.UpdateCounter("mixed", 100)
		assert.NoError(t, err)
		err = storage.UpdateGauge("mixed", 200.0)
		assert.Error(t, err)
		metric, exists := storage.GetMetric("mixed")
		assert.True(t, exists)
		assert.Equal(t, int64(100), *metric.Delta)
	})
}
