package cache

import (
	"sync"
	"testing"
	"time"
)

// TestRingBuffer_Write tests writing data to ring buffer
func TestRingBuffer_Write(t *testing.T) {
	rb := NewRingBuffer(60)

	// Test single write
	point := &MetricPoint{
		Timestamp:      time.Now(),
		LatencyMs:      100.0,
		PacketLossRate: 0.0,
		JitterMs:       5.0,
	}

	overwritten := rb.Write(point)
	if overwritten {
		t.Error("Expected no overwrite on first write")
	}

	if rb.Size() != 1 {
		t.Errorf("Expected size 1, got %d", rb.Size())
	}
}

// TestRingBuffer_FIFOEviction tests FIFO eviction when buffer is full
func TestRingBuffer_FIFOEviction(t *testing.T) {
	capacity := 5 // Small capacity for testing
	rb := NewRingBuffer(capacity)

	// Fill the buffer
	for i := 0; i < capacity; i++ {
		point := &MetricPoint{
			Timestamp:      time.Now().Add(time.Duration(i) * time.Minute),
			LatencyMs:      float64(i * 10),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
		}
		rb.Write(point)
	}

	if rb.Size() != capacity {
		t.Errorf("Expected size %d, got %d", capacity, rb.Size())
	}

	// Write one more to trigger FIFO eviction
	newPoint := &MetricPoint{
		Timestamp:      time.Now().Add(time.Duration(capacity) * time.Minute),
		LatencyMs:      999.0,
		PacketLossRate: 0.0,
		JitterMs:       5.0,
	}

	overwritten := rb.Write(newPoint)
	if !overwritten {
		t.Error("Expected overwrite when buffer is full")
	}

	if rb.Size() != capacity {
		t.Errorf("Expected size %d after overwrite, got %d", capacity, rb.Size())
	}

	// Verify first data point was evicted
	points := rb.ReadAll()
	if len(points) != capacity {
		t.Errorf("Expected %d points, got %d", capacity, len(points))
	}

	// First point should not be the original first point (0*10=0)
	if points[0].LatencyMs == 0 {
		t.Error("Expected first point to be evicted, but found latency 0")
	}
}

// TestRingBuffer_ReadAll tests reading all data from buffer
func TestRingBuffer_ReadAll(t *testing.T) {
	rb := NewRingBuffer(60)

	// Write multiple points
	expectedCount := 10
	for i := 0; i < expectedCount; i++ {
		point := &MetricPoint{
			Timestamp:      time.Now().Add(time.Duration(i) * time.Minute),
			LatencyMs:      float64(i * 10),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
		}
		rb.Write(point)
	}

	points := rb.ReadAll()
	if len(points) != expectedCount {
		t.Errorf("Expected %d points, got %d", expectedCount, len(points))
	}
}

// TestRingBuffer_AggregateMetrics tests 1-minute aggregation
func TestRingBuffer_AggregateMetrics(t *testing.T) {
	rb := NewRingBuffer(60)

	baseTime := time.Now().Truncate(time.Minute)

	// Write multiple points within the same minute
	pointsInMinute := 5
	for i := 0; i < pointsInMinute; i++ {
		point := &MetricPoint{
			Timestamp:      baseTime.Add(time.Duration(i) * time.Second),
			LatencyMs:      float64(100 + i*10), // 100, 110, 120, 130, 140
			PacketLossRate: 0.1,
			JitterMs:       5.0,
		}
		rb.Write(point)
	}

	aggregated := rb.AggregateMetrics()
	if len(aggregated) != 1 {
		t.Errorf("Expected 1 aggregated bucket, got %d", len(aggregated))
	}

	agg := aggregated[0]

	// Check average: (100+110+120+130+140)/5 = 120
	expectedAvg := 120.0
	if agg.LatencyMs != expectedAvg {
		t.Errorf("Expected average latency %.2f, got %.2f", expectedAvg, agg.LatencyMs)
	}

	// Check max
	expectedMax := 140.0
	if agg.MaxLatencyMs != expectedMax {
		t.Errorf("Expected max latency %.2f, got %.2f", expectedMax, agg.MaxLatencyMs)
	}

	// Check min
	expectedMin := 100.0
	if agg.MinLatencyMs != expectedMin {
		t.Errorf("Expected min latency %.2f, got %.2f", expectedMin, agg.MinLatencyMs)
	}
}

// TestRingBuffer_ConcurrentAccess tests concurrent read/write safety
func TestRingBuffer_ConcurrentAccess(t *testing.T) {
	rb := NewRingBuffer(60)

	var wg sync.WaitGroup
	numWriters := 10
	numWritesPerWriter := 100

	// Concurrent writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < numWritesPerWriter; j++ {
				point := &MetricPoint{
					Timestamp:      time.Now(),
					LatencyMs:      float64(writerID*1000 + j),
					PacketLossRate: 0.0,
					JitterMs:       5.0,
				}
				rb.Write(point)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				rb.ReadAll()
				rb.Size()
				rb.AggregateMetrics()
			}
		}()
	}

	wg.Wait()

	// Verify buffer state is consistent
	if rb.Size() > rb.capacity {
		t.Errorf("Buffer size %d exceeds capacity %d", rb.Size(), rb.capacity)
	}
}

// TestMemoryCache_Store tests storing metrics in memory cache
func TestMemoryCache_Store(t *testing.T) {
	mc := NewMemoryCache()

	nodeID := "test-node-123"
	point := &MetricPoint{
		Timestamp:      time.Now(),
		LatencyMs:      100.0,
		PacketLossRate: 0.0,
		JitterMs:       5.0,
	}

	err := mc.Store(nodeID, point)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if mc.GetSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", mc.GetSize())
	}
}

// TestMemoryCache_Store_EmptyNodeID tests error handling for empty node ID
func TestMemoryCache_Store_EmptyNodeID(t *testing.T) {
	mc := NewMemoryCache()

	point := &MetricPoint{
		Timestamp:      time.Now(),
		LatencyMs:      100.0,
		PacketLossRate: 0.0,
		JitterMs:       5.0,
	}

	err := mc.Store("", point)
	if err != ErrEmptyNodeID {
		t.Errorf("Expected ErrEmptyNodeID, got %v", err)
	}
}

// TestMemoryCache_Store_NilPoint tests error handling for nil metric point
func TestMemoryCache_Store_NilPoint(t *testing.T) {
	mc := NewMemoryCache()

	err := mc.Store("test-node", nil)
	if err != ErrNilMetricPoint {
		t.Errorf("Expected ErrNilMetricPoint, got %v", err)
	}
}

// TestMemoryCache_Get tests retrieving metrics for a node
func TestMemoryCache_Get(t *testing.T) {
	mc := NewMemoryCache()

	nodeID := "test-node-123"
	expectedPoints := 10

	// Store multiple points
	for i := 0; i < expectedPoints; i++ {
		point := &MetricPoint{
			Timestamp:      time.Now().Add(time.Duration(i) * time.Minute),
			LatencyMs:      float64(i * 10),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
		}
		mc.Store(nodeID, point)
	}

	points := mc.Get(nodeID)
	if len(points) != expectedPoints {
		t.Errorf("Expected %d points, got %d", expectedPoints, len(points))
	}
}

// TestMemoryCache_Get_NonExistentNode tests getting non-existent node
func TestMemoryCache_Get_NonExistentNode(t *testing.T) {
	mc := NewMemoryCache()

	points := mc.Get("non-existent-node")
	if points != nil {
		t.Error("Expected nil for non-existent node, got points")
	}
}

// TestMemoryCache_GetAllNodeIDs tests retrieving all node IDs
func TestMemoryCache_GetAllNodeIDs(t *testing.T) {
	mc := NewMemoryCache()

	// Store data for 3 nodes
	nodeIDs := []string{"node-1", "node-2", "node-3"}
	for _, nodeID := range nodeIDs {
		point := &MetricPoint{
			Timestamp:      time.Now(),
			LatencyMs:      100.0,
			PacketLossRate: 0.0,
			JitterMs:       5.0,
		}
		mc.Store(nodeID, point)
	}

	retrievedNodeIDs := mc.GetAllNodeIDs()
	if len(retrievedNodeIDs) != len(nodeIDs) {
		t.Errorf("Expected %d node IDs, got %d", len(nodeIDs), len(retrievedNodeIDs))
	}

	// Verify all node IDs are present
	nodeIDMap := make(map[string]bool)
	for _, id := range retrievedNodeIDs {
		nodeIDMap[id] = true
	}

	for _, expectedID := range nodeIDs {
		if !nodeIDMap[expectedID] {
			t.Errorf("Expected node ID %s not found", expectedID)
		}
	}
}

// TestMemoryCache_AggregateMetricsByNode tests aggregation for specific node
func TestMemoryCache_AggregateMetricsByNode(t *testing.T) {
	mc := NewMemoryCache()

	nodeID := "test-node-123"
	baseTime := time.Now().Truncate(time.Minute)

	// Store points in 2 different minutes
	for i := 0; i < 10; i++ {
		point := &MetricPoint{
			Timestamp:      baseTime.Add(time.Duration(i) * 30 * time.Second),
			LatencyMs:      float64(100 + i*10),
			PacketLossRate: 0.1,
			JitterMs:       5.0,
		}
		mc.Store(nodeID, point)
	}

	aggregated := mc.AggregateMetricsByNode(nodeID)
	if len(aggregated) == 0 {
		t.Error("Expected aggregated metrics, got none")
	}
}

// TestMemoryCache_MultipleNodes tests cache with multiple nodes
func TestMemoryCache_MultipleNodes(t *testing.T) {
	mc := NewMemoryCache()

	// Store data for 10 nodes (requirement: support at least 10 nodes)
	numNodes := 10
	for i := 0; i < numNodes; i++ {
		nodeID := "node-" + string(rune('0'+i))
		point := &MetricPoint{
			Timestamp:      time.Now(),
			LatencyMs:      float64(i * 10),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
		}
		mc.Store(nodeID, point)
	}

	if mc.GetSize() != numNodes {
		t.Errorf("Expected %d nodes, got %d", numNodes, mc.GetSize())
	}

	nodeIDs := mc.GetAllNodeIDs()
	if len(nodeIDs) != numNodes {
		t.Errorf("Expected %d node IDs, got %d", numNodes, len(nodeIDs))
	}
}

// TestMemoryCache_ConcurrentAccess tests concurrent access to memory cache
func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	mc := NewMemoryCache()

	var wg sync.WaitGroup
	numGoroutines := 20
	operationsPerGoroutine := 100

	// Concurrent stores and gets
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			nodeID := "node-" + string(rune('0'+goroutineID%10))

			for j := 0; j < operationsPerGoroutine; j++ {
				// Store
				point := &MetricPoint{
					Timestamp:      time.Now(),
					LatencyMs:      float64(goroutineID*1000 + j),
					PacketLossRate: 0.0,
					JitterMs:       5.0,
				}
				mc.Store(nodeID, point)

				// Get
				mc.Get(nodeID)
				mc.AggregateMetricsByNode(nodeID)
				mc.GetAllNodeIDs()
			}
		}(i)
	}

	wg.Wait()

	// Verify cache is in consistent state
	// We expect 10 unique nodes (goroutineID % 10)
	if mc.GetSize() != 10 {
		t.Errorf("Expected cache size 10, got %d", mc.GetSize())
	}
}

// TestMemoryCache_OneHourDataEviction tests FIFO eviction after 1 hour
func TestMemoryCache_OneHourDataEviction(t *testing.T) {
	mc := NewMemoryCache()

	nodeID := "test-node-123"
	baseTime := time.Now()

	// Store 70 data points (more than 60 - 1 hour capacity)
	for i := 0; i < 70; i++ {
		point := &MetricPoint{
			Timestamp:      baseTime.Add(time.Duration(i) * time.Minute),
			LatencyMs:      float64(i),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
		}
		mc.Store(nodeID, point)
	}

	points := mc.Get(nodeID)
	// Should have at most 60 points (1 hour capacity)
	if len(points) > 60 {
		t.Errorf("Expected at most 60 points after FIFO eviction, got %d", len(points))
	}

	// First point should not be the original first point (latency 0)
	if len(points) > 0 && points[0].LatencyMs == 0 {
		t.Error("Expected first data points to be evicted after 1 hour")
	}
}

// TestMemoryCache_BackgroundAggregation tests that background aggregation goroutine starts and stops cleanly
func TestMemoryCache_BackgroundAggregation(t *testing.T) {
	mc := NewMemoryCache()

	// Add some test data
	nodeID := "test-node-aggregation"
	baseTime := time.Now()

	for i := 0; i < 10; i++ {
		point := &MetricPoint{
			Timestamp:      baseTime.Add(time.Duration(i) * time.Minute),
			LatencyMs:      float64(i * 10),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
		}
		err := mc.Store(nodeID, point)
		if err != nil {
			t.Fatalf("Failed to store metric point: %v", err)
		}
	}

	// Verify data is stored
	points := mc.Get(nodeID)
	if len(points) != 10 {
		t.Errorf("Expected 10 points, got %d", len(points))
	}

	// Stop the background goroutine (should not hang or panic)
	mc.Stop()

	// Verify cache is still functional after stop
	// Note: After Stop(), the background aggregation goroutine is terminated,
	// but basic Store/Get operations should still work
	for i := 10; i < 15; i++ {
		point := &MetricPoint{
			Timestamp:      baseTime.Add(time.Duration(i) * time.Minute),
			LatencyMs:      float64(i * 10),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
		}
		err := mc.Store(nodeID, point)
		if err != nil {
			t.Errorf("Store failed after Stop: %v", err)
		}
	}

	points = mc.Get(nodeID)
	if len(points) != 15 {
		t.Errorf("Expected 15 points after Stop, got %d", len(points))
	}
}
