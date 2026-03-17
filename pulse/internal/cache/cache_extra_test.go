package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAggregateAllNodes tests the aggregateAllNodes method
func TestAggregateAllNodes_NoNodes(t *testing.T) {
	mc := NewMemoryCache()
	// No Start method needed
	defer mc.Stop()

	// Should not panic with no nodes
	assert.NotPanics(t, func() {
		mc.aggregateAllNodes()
	})
}

func TestAggregateAllNodes_WithNodes(t *testing.T) {
	mc := NewMemoryCache()
	// No Start method needed
	defer mc.Stop()

	// Add some data to the cache
	now := time.Now()
	for i := 0; i < 5; i++ {
		point := &MetricPoint{
			Timestamp:      now.Add(time.Duration(i) * time.Second),
			LatencyMs:      float64(10 + i),
			PacketLossRate: 0.01,
			JitterMs:       float64(2 + i),
		}
		require.NoError(t, mc.Store("node-1", point))
		require.NoError(t, mc.Store("node-2", point))
	}

	// Should not panic
	assert.NotPanics(t, func() {
		mc.aggregateAllNodes()
	})
}

// TestWriteBatch tests the writeBatch method with nil db
func TestWriteBatch_NilDB(t *testing.T) {
	bw := NewBatchWriter(nil, 100, 10)

	records := []*MetricRecord{
		{NodeID: "node-1", ProbeID: "probe-1", Timestamp: time.Now(), LatencyMs: 10},
	}

	// With nil DB, writeBatch should succeed (skip writing)
	err := bw.writeBatch(records)
	assert.NoError(t, err)
}

func TestWriteBatch_EmptyBatch(t *testing.T) {
	bw := NewBatchWriter(nil, 100, 10)

	// Empty batch should return nil
	err := bw.writeBatch([]*MetricRecord{})
	assert.NoError(t, err)
}

func TestWriteBatchWithRetry_NilDB(t *testing.T) {
	bw := NewBatchWriter(nil, 100, 10)

	records := []*MetricRecord{
		{NodeID: "node-1", ProbeID: "probe-1", Timestamp: time.Now(), LatencyMs: 10},
	}

	// writeBatchWithRetry with nil DB should not panic
	assert.NotPanics(t, func() {
		bw.writeBatchWithRetry(records)
	})
}

// TestBatchWriter_Flush tests the flush method
func TestBatchWriter_Flush(t *testing.T) {
	bw := NewBatchWriter(nil, 100, 10)

	// Add some records to the buffer
	for i := 0; i < 5; i++ {
		record := &MetricRecord{
			NodeID:    "node-1",
			ProbeID:   "probe-1",
			Timestamp: time.Now(),
			LatencyMs: float64(i * 10),
		}
		err := bw.Write(record)
		require.NoError(t, err)
	}

	// flush should drain the buffer
	assert.NotPanics(t, func() {
		bw.flush()
	})
}

// TestBatchWriter_ProcessBatches tests the processBatches goroutine
func TestBatchWriter_ProcessBatches_ContextCancel(t *testing.T) {
	bw := NewBatchWriter(nil, 100, 10)
	bw.Start()

	// Send some records
	for i := 0; i < 5; i++ {
		record := &MetricRecord{
			NodeID:    "node-1",
			ProbeID:   "probe-1",
			Timestamp: time.Now(),
			LatencyMs: float64(i * 10),
		}
		_ = bw.Write(record)
	}

	// Stop should gracefully cancel and drain
	assert.NotPanics(t, func() {
		bw.Stop()
	})
}

// TestBatchWriter_ProcessBatches_BatchSizeTriggered tests batch-size-triggered writes
func TestBatchWriter_ProcessBatches_BatchSizeTrigger(t *testing.T) {
	batchSize := 3
	bw := NewBatchWriter(nil, 100, batchSize)
	bw.Start()

	// Send enough records to trigger a batch write
	for i := 0; i < batchSize+1; i++ {
		record := &MetricRecord{
			NodeID:    "node-1",
			ProbeID:   "probe-1",
			Timestamp: time.Now(),
			LatencyMs: float64(i * 10),
		}
		_ = bw.Write(record)
	}

	// Give the goroutine time to process
	time.Sleep(50 * time.Millisecond)

	bw.Stop()
}
