package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage(t *testing.T) {
	storage := NewStorage()

	t.Run("gauge operations", func(t *testing.T) {
		// Test UpdateGauge
		storage.UpdateGauge("testGauge", 123.45)
		value, exists := storage.GetGauge("testGauge")
		assert.True(t, exists)
		assert.Equal(t, 123.45, value)

		// Test overwrite
		storage.UpdateGauge("testGauge", 67.89)
		value, exists = storage.GetGauge("testGauge")
		assert.True(t, exists)
		assert.Equal(t, 67.89, value)

		// Test non-existent gauge
		_, exists = storage.GetGauge("nonExistent")
		assert.False(t, exists)
	})

	t.Run("counter operations", func(t *testing.T) {
		// Test UpdateCounter
		storage.UpdateCounter("testCounter", 10)
		value, exists := storage.GetCounter("testCounter")
		assert.True(t, exists)
		assert.Equal(t, int64(10), value)

		// Test increment
		storage.UpdateCounter("testCounter", 20)
		value, exists = storage.GetCounter("testCounter")
		assert.True(t, exists)
		assert.Equal(t, int64(30), value) // 10 + 20

		// Test non-existent counter
		_, exists = storage.GetCounter("nonExistent")
		assert.False(t, exists)
	})

	t.Run("gauge and counter interaction", func(t *testing.T) {
		// Test that gauge overwrites counter
		storage.UpdateCounter("mixed", 100)
		storage.UpdateGauge("mixed", 200.0)
		gValue, exists := storage.GetGauge("mixed")
		assert.True(t, exists)
		assert.Equal(t, 200.0, gValue)
		_, exists = storage.GetCounter("mixed")
		assert.False(t, exists)

		// Test that counter overwrites gauge
		storage.UpdateGauge("mixed", 300.0)
		storage.UpdateCounter("mixed", 400)
		cValue, exists := storage.GetCounter("mixed")
		assert.True(t, exists)
		assert.Equal(t, int64(400), cValue)
		_, exists = storage.GetGauge("mixed")
		assert.False(t, exists)
	})
}
