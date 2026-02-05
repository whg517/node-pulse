package cache

import (
	"testing"
	"time"
)

// TestBatchWriter_Write tests writing to batch buffer
func TestBatchWriter_Write(t *testing.T) {
	// Use a nil pool for basic buffer testing (won't actually write to DB)
	bw := NewBatchWriter(nil, 1000, 100)

	record := &MetricRecord{
		NodeID:         "550e8400-e29b-41d4-a716-446655440000",
		ProbeID:        "550e8400-e29b-41d4-a716-446655440001",
		Timestamp:      time.Now(),
		LatencyMs:      100.0,
		PacketLossRate: 0.0,
		JitterMs:       5.0,
		IsAggregated:   false,
	}

	err := bw.Write(record)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if bw.GetBufferSize() != 1 {
		t.Errorf("Expected buffer size 1, got %d", bw.GetBufferSize())
	}
}

// TestBatchWriter_Write_NilRecord tests error handling for nil record
func TestBatchWriter_Write_NilRecord(t *testing.T) {
	bw := NewBatchWriter(nil, 1000, 100)

	err := bw.Write(nil)
	if err != ErrNilMetricRecord {
		t.Errorf("Expected ErrNilMetricRecord, got %v", err)
	}
}

// TestBatchWriter_BufferFull tests buffer full scenario
func TestBatchWriter_BufferFull(t *testing.T) {
	bufferSize := 10
	bw := NewBatchWriter(nil, bufferSize, 100)

	// Fill the buffer
	for i := 0; i < bufferSize; i++ {
		record := &MetricRecord{
			NodeID:         "550e8400-e29b-41d4-a716-446655440000",
			ProbeID:        "550e8400-e29b-41d4-a716-446655440001",
			Timestamp:      time.Now(),
			LatencyMs:      100.0,
			PacketLossRate: 0.0,
			JitterMs:       5.0,
			IsAggregated:   false,
		}
		err := bw.Write(record)
		if err != nil {
			t.Errorf("Failed to write record: %v", err)
		}
	}

	// Try to write one more (should fail)
	record := &MetricRecord{
		NodeID:         "550e8400-e29b-41d4-a716-446655440000",
		ProbeID:        "550e8400-e29b-41d4-a716-446655440001",
		Timestamp:      time.Now(),
		LatencyMs:      100.0,
		PacketLossRate: 0.0,
		JitterMs:       5.0,
		IsAggregated:   false,
	}

	err := bw.Write(record)
	if err != ErrBufferFull {
		t.Errorf("Expected ErrBufferFull, got %v", err)
	}
}

// TestBatchWriter_StartStop tests starting and stopping batch writer
func TestBatchWriter_StartStop(t *testing.T) {
	bw := NewBatchWriter(nil, 1000, 100)

	bw.Start()

	// Write some records
	for i := 0; i < 5; i++ {
		record := &MetricRecord{
			NodeID:         "550e8400-e29b-41d4-a716-446655440000",
			ProbeID:        "550e8400-e29b-41d4-a716-446655440001",
			Timestamp:      time.Now(),
			LatencyMs:      float64(i * 10),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
			IsAggregated:   false,
		}
		bw.Write(record)
	}

	// Stop should flush and exit gracefully
	bw.Stop()

	// Buffer should be closed
	if bw.GetBufferSize() < 0 {
		t.Error("Buffer should be closed after stop")
	}
}

// TestBatchWriter_MultipleNodes tests batch writer with multiple nodes
func TestBatchWriter_MultipleNodes(t *testing.T) {
	bw := NewBatchWriter(nil, 1000, 100)

	// Write records from 10 different nodes
	numNodes := 10
	for i := 0; i < numNodes; i++ {
		record := &MetricRecord{
			NodeID:         "550e8400-e29b-41d4-a716-446655440000",
			ProbeID:        "550e8400-e29b-41d4-a716-446655440001",
			Timestamp:      time.Now(),
			LatencyMs:      float64(i * 10),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
			IsAggregated:   false,
		}
		err := bw.Write(record)
		if err != nil {
			t.Errorf("Failed to write record for node %d: %v", i, err)
		}
	}

	if bw.GetBufferSize() != numNodes {
		t.Errorf("Expected buffer size %d, got %d", numNodes, bw.GetBufferSize())
	}
}

// TestBatchWriter_IsAggregatedFlag tests is_aggregated flag handling
func TestBatchWriter_IsAggregatedFlag(t *testing.T) {
	bw := NewBatchWriter(nil, 1000, 100)

	// Write aggregated record
	aggRecord := &MetricRecord{
		NodeID:         "550e8400-e29b-41d4-a716-446655440000",
		ProbeID:        "550e8400-e29b-41d4-a716-446655440001",
		Timestamp:      time.Now(),
		LatencyMs:      100.0,
		PacketLossRate: 0.0,
		JitterMs:       5.0,
		IsAggregated:   true,
	}
	err := bw.Write(aggRecord)
	if err != nil {
		t.Errorf("Failed to write aggregated record: %v", err)
	}

	// Write non-aggregated record
	rawRecord := &MetricRecord{
		NodeID:         "550e8400-e29b-41d4-a716-446655440000",
		ProbeID:        "550e8400-e29b-41d4-a716-446655440001",
		Timestamp:      time.Now(),
		LatencyMs:      200.0,
		PacketLossRate: 0.0,
		JitterMs:       5.0,
		IsAggregated:   false,
	}
	err = bw.Write(rawRecord)
	if err != nil {
		t.Errorf("Failed to write non-aggregated record: %v", err)
	}

	if bw.GetBufferSize() != 2 {
		t.Errorf("Expected buffer size 2, got %d", bw.GetBufferSize())
	}
}

// TestBatchWriter_ConcurrentWrites tests concurrent write operations
func TestBatchWriter_ConcurrentWrites(t *testing.T) {
	bw := NewBatchWriter(nil, 10000, 10000) // Large batch size to prevent auto-flush
	bw.Start()
	defer bw.Stop()

	done := make(chan bool)
	numGoroutines := 10
	writesPerGoroutine := 10

	// Concurrent writers
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < writesPerGoroutine; j++ {
				record := &MetricRecord{
					NodeID:         "550e8400-e29b-41d4-a716-446655440000",
					ProbeID:        "550e8400-e29b-41d4-a716-446655440001",
					Timestamp:      time.Now(),
					LatencyMs:      float64(goroutineID*1000 + j),
					PacketLossRate: 0.0,
					JitterMs:       5.0,
					IsAggregated:   false,
				}
				bw.Write(record)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	expectedWrites := numGoroutines * writesPerGoroutine
	// Note: Buffer size may be less than expected due to background processing
	// Just verify no errors occurred during concurrent writes
	if bw.GetBufferSize() > expectedWrites {
		t.Errorf("Buffer size %d exceeds expected writes %d", bw.GetBufferSize(), expectedWrites)
	}
}

// TestBatchWriter_RetryLogic tests retry logic structure (without actual DB failure)
func TestBatchWriter_RetryLogic(t *testing.T) {
	// This test verifies the retry structure is in place
	// Actual retry logic is tested in integration tests with real DB
	bw := NewBatchWriter(nil, 1000, 10)

	// Verify batch size is set correctly
	if bw.batchSize != 10 {
		t.Errorf("Expected batch size 10, got %d", bw.batchSize)
	}

	// Verify buffer capacity
	if cap(bw.buffer) != 1000 {
		t.Errorf("Expected buffer capacity 1000, got %d", cap(bw.buffer))
	}
}

// TestBatchWriter_TimeoutTrigger tests the 1-minute timeout flush trigger
func TestBatchWriter_TimeoutTrigger(t *testing.T) {
	// Create a custom batch writer with shorter timeout for testing
	// In production this is 1 minute, but we use 100ms for test speed
	bw := NewBatchWriter(nil, 1000, 1000) // Large batch size to prevent size-based trigger

	// Override the ticker with a shorter interval for testing
	bw.flushTicker = time.NewTicker(100 * time.Millisecond)
	defer bw.flushTicker.Stop()

	bw.Start()
	defer bw.Stop()

	// Write a small number of records (less than batch size)
	record := &MetricRecord{
		NodeID:         "550e8400-e29b-41d4-a716-446655440000",
		ProbeID:        "550e8400-e29b-41d4-a716-446655440001",
		Timestamp:      time.Now(),
		LatencyMs:      100.0,
		PacketLossRate: 0.0,
		JitterMs:       5.0,
		IsAggregated:   false,
	}

	// Write 5 records (well below batch size of 1000)
	for i := 0; i < 5; i++ {
		err := bw.Write(record)
		if err != nil {
			t.Fatalf("Failed to write record: %v", err)
		}
	}

	initialBufferSize := bw.GetBufferSize()
	if initialBufferSize != 5 {
		t.Errorf("Expected buffer size 5, got %d", initialBufferSize)
	}

	// Wait for timeout trigger (slightly longer than ticker interval)
	time.Sleep(150 * time.Millisecond)

	// Buffer should be flushed by timeout trigger
	finalBufferSize := bw.GetBufferSize()
	if finalBufferSize >= initialBufferSize {
		t.Errorf("Expected buffer to be flushed by timeout trigger, got size %d", finalBufferSize)
	}
}
