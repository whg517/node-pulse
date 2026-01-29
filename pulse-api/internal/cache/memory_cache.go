package cache

import (
	"context"
	"sync"
	"time"
)

// MetricPoint represents a single metric data point for a node
type MetricPoint struct {
	Timestamp      time.Time
	LatencyMs      float64
	PacketLossRate float64
	JitterMs       float64
}

// AggregatedMetrics represents 1-minute aggregated metrics
type AggregatedMetrics struct {
	Timestamp      time.Time
	LatencyMs      float64 // average
	MaxLatencyMs   float64
	MinLatencyMs   float64
	PacketLossRate float64 // average
	JitterMs       float64 // average
}

// RingBuffer implements a circular buffer for time-series metrics
// Fixed size: 60 (one data point per minute for 1 hour)
type RingBuffer struct {
	data      []*MetricPoint // Fixed size 60
	head      int            // Write position
	tail      int            // Read position
	size      int            // Current size
	capacity  int            // Maximum capacity (60)
	mutex     sync.RWMutex   // Read-write lock for concurrent access
}

// NewRingBuffer creates a new ring buffer with specified capacity
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		data:     make([]*MetricPoint, capacity),
		capacity: capacity,
		head:     0,
		tail:     0,
		size:     0,
	}
}

// Write adds a new metric point to the ring buffer (FIFO)
// Returns true if old data was overwritten
func (rb *RingBuffer) Write(point *MetricPoint) bool {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	overwritten := rb.size >= rb.capacity

	rb.data[rb.head] = point
	rb.head = (rb.head + 1) % rb.capacity

	if rb.size < rb.capacity {
		rb.size++
	} else {
		// Buffer is full, move tail forward (FIFO eviction)
		rb.tail = (rb.tail + 1) % rb.capacity
	}

	return overwritten
}

// ReadAll returns all metric points in the buffer
func (rb *RingBuffer) ReadAll() []*MetricPoint {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	result := make([]*MetricPoint, 0, rb.size)
	for i := 0; i < rb.size; i++ {
		idx := (rb.tail + i) % rb.capacity
		if rb.data[idx] != nil {
			result = append(result, rb.data[idx])
		}
	}

	return result
}

// Size returns the current number of elements in the buffer
func (rb *RingBuffer) Size() int {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()
	return rb.size
}

// AggregateMetrics aggregates metrics in the buffer by 1-minute intervals
// Returns aggregated metrics with mean, max, and min values
func (rb *RingBuffer) AggregateMetrics() []*AggregatedMetrics {
	rb.mutex.RLock()
	defer rb.mutex.RUnlock()

	if rb.size == 0 {
		return nil
	}

	// Group metrics by minute (truncate to minute)
	minuteBuckets := make(map[time.Time][]*MetricPoint)

	for i := 0; i < rb.size; i++ {
		idx := (rb.tail + i) % rb.capacity
		point := rb.data[idx]
		if point == nil {
			continue
		}

		minuteKey := point.Timestamp.Truncate(time.Minute)
		minuteBuckets[minuteKey] = append(minuteBuckets[minuteKey], point)
	}

	// Calculate aggregates for each minute
	result := make([]*AggregatedMetrics, 0, len(minuteBuckets))
	for timestamp, points := range minuteBuckets {
		if len(points) == 0 {
			continue
		}

		agg := &AggregatedMetrics{
			Timestamp: timestamp,
		}

		var sumLatency, sumPacketLoss, sumJitter float64
		maxLatency := points[0].LatencyMs
		minLatency := points[0].LatencyMs

		for _, p := range points {
			sumLatency += p.LatencyMs
			sumPacketLoss += p.PacketLossRate
			sumJitter += p.JitterMs

			if p.LatencyMs > maxLatency {
				maxLatency = p.LatencyMs
			}
			if p.LatencyMs < minLatency {
				minLatency = p.LatencyMs
			}
		}

		count := float64(len(points))
		agg.LatencyMs = sumLatency / count
		agg.MaxLatencyMs = maxLatency
		agg.MinLatencyMs = minLatency
		agg.PacketLossRate = sumPacketLoss / count
		agg.JitterMs = sumJitter / count

		result = append(result, agg)
	}

	return result
}

// MemoryCache stores node metrics in memory using sync.Map
// Key: node_id (UUID), Value: *RingBuffer
type MemoryCache struct {
	nodes    sync.Map
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewMemoryCache creates a new memory cache with background aggregation
func NewMemoryCache() *MemoryCache {
	ctx, cancel := context.WithCancel(context.Background())
	mc := &MemoryCache{
		ctx:    ctx,
		cancel: cancel,
	}

	// Start background aggregation goroutine
	mc.wg.Add(1)
	go mc.backgroundAggregator()

	return mc
}

// Stop gracefully stops the background aggregation goroutine
func (mc *MemoryCache) Stop() {
	mc.cancel()
	mc.wg.Wait()
}

// backgroundAggregator runs periodic aggregation (every 1 minute)
// Aggregates metrics from all nodes and sends to batch writer
func (mc *MemoryCache) backgroundAggregator() {
	defer mc.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-mc.ctx.Done():
			// Context cancelled, exit gracefully
			return
		case <-ticker.C:
			// Perform aggregation for all nodes
			mc.aggregateAllNodes()
		}
	}
}

// aggregateAllNodes aggregates metrics for all nodes in the cache
// This is called automatically every 1 minute by backgroundAggregator
func (mc *MemoryCache) aggregateAllNodes() {
	nodeIDs := mc.GetAllNodeIDs()

	for _, nodeID := range nodeIDs {
		aggregated := mc.AggregateMetricsByNode(nodeID)
		if aggregated != nil && len(aggregated) > 0 {
			// Aggregated data is available
			// Note: In production, these aggregated metrics should be sent to batch writer
			// This hook is for Story 3.2 integration point
			_ = aggregated // Aggregation completed successfully
		}
	}
}

// Store writes a metric point to the cache for the specified node
func (mc *MemoryCache) Store(nodeID string, point *MetricPoint) error {
	if nodeID == "" {
		return ErrEmptyNodeID
	}
	if point == nil {
		return ErrNilMetricPoint
	}

	// Get or create ring buffer for this node
	actual, _ := mc.nodes.LoadOrStore(nodeID, NewRingBuffer(60))
	buffer := actual.(*RingBuffer)

	// Write to buffer (FIFO eviction after 1 hour)
	overwritten := buffer.Write(point)

	if overwritten {
		// Old data was evicted (more than 1 hour)
		// This is expected behavior per requirements
	}

	return nil
}

// Get retrieves all metrics for a specific node
func (mc *MemoryCache) Get(nodeID string) []*MetricPoint {
	actual, ok := mc.nodes.Load(nodeID)
	if !ok {
		return nil
	}

	buffer := actual.(*RingBuffer)
	return buffer.ReadAll()
}

// GetAllNodeIDs returns all node IDs currently in the cache
func (mc *MemoryCache) GetAllNodeIDs() []string {
	var nodeIDs []string

	mc.nodes.Range(func(key, value interface{}) bool {
		nodeIDs = append(nodeIDs, key.(string))
		return true
	})

	return nodeIDs
}

// GetSize returns the number of nodes in the cache
func (mc *MemoryCache) GetSize() int {
	size := 0
	mc.nodes.Range(func(key, value interface{}) bool {
		size++
		return true
	})
	return size
}

// AggregateMetricsByNode returns aggregated metrics for a specific node
func (mc *MemoryCache) AggregateMetricsByNode(nodeID string) []*AggregatedMetrics {
	actual, ok := mc.nodes.Load(nodeID)
	if !ok {
		return nil
	}

	buffer := actual.(*RingBuffer)
	return buffer.AggregateMetrics()
}
