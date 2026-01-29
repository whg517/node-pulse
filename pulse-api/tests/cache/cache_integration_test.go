package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kevin/node-pulse/pulse-api/internal/cache"
)

// testDBPool returns the test database pool or skips the test
func testDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Use environment variable or default to test database
	connString := "postgres://nodepulse:testpass@localhost:5432/nodepulse_test?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		t.Logf("Failed to connect to test database: %v", err)
		return nil
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		t.Logf("Failed to ping test database: %v", err)
		pool.Close()
		return nil
	}

	return pool
}

// getTestPool returns the test database pool or skips the test
func getTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testDBPool(t)
	if pool == nil {
		t.Skip("No test database available")
	}
	return pool
}

// setupMetricsTable creates the metrics table for testing
func setupMetricsTable(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS metrics (
			id BIGSERIAL PRIMARY KEY,
			node_id UUID NOT NULL,
			probe_id UUID NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			latency_ms DOUBLE PRECISION NOT NULL,
			packet_loss_rate DOUBLE PRECISION NOT NULL,
			jitter_ms DOUBLE PRECISION NOT NULL,
			is_aggregated BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err, "Failed to create metrics table")
}

// cleanupMetricsTable drops the metrics table
func cleanupMetricsTable(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(ctx, "DROP TABLE IF EXISTS metrics")
	// Ignore error during cleanup
	_ = err
}

// TestCacheIntegration_FullFlow tests the complete data flow: API → memory cache → PostgreSQL
func TestCacheIntegration_FullFlow(t *testing.T) {
	ctx := context.Background()
	pool := getTestPool(t)
	setupMetricsTable(t, ctx, pool)
	defer cleanupMetricsTable(t, ctx, pool)

	// Create memory cache and batch writer
	memoryCache := cache.NewMemoryCache()
	batchWriter := cache.NewBatchWriter(pool, 100, 10) // Small batch size for testing
	batchWriter.Start()
	defer batchWriter.Stop()

	// Simulate Beacon heartbeat data
	nodeID := uuid.New().String()
	probeID := uuid.New().String()
	baseTime := time.Now()

	// Write 15 records (more than batch size of 10)
	numRecords := 15
	for i := 0; i < numRecords; i++ {
		point := &cache.MetricPoint{
			Timestamp:      baseTime.Add(time.Duration(i) * time.Second),
			LatencyMs:      float64(100 + i*10),
			PacketLossRate: 0.1,
			JitterMs:       5.0,
		}

		// Write to memory cache
		err := memoryCache.Store(nodeID, point)
		require.NoError(t, err, "Failed to write to memory cache")

		// Send to batch writer
		record := &cache.MetricRecord{
			NodeID:         nodeID,
			ProbeID:        probeID,
			Timestamp:      point.Timestamp,
			LatencyMs:      point.LatencyMs,
			PacketLossRate: point.PacketLossRate,
			JitterMs:       point.JitterMs,
			IsAggregated:   false,
		}

		err = batchWriter.Write(record)
		require.NoError(t, err, "Failed to write to batch buffer")
	}

	// Wait for batch processing
	time.Sleep(2 * time.Second)

	// Verify memory cache has data
	cachedPoints := memoryCache.Get(nodeID)
	assert.NotEmpty(t, cachedPoints, "Memory cache should have data")

	// Verify PostgreSQL has data (at least one batch)
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM metrics").Scan(&count)
	require.NoError(t, err, "Failed to query metrics table")

	// Should have at least 10 records (first batch)
	assert.GreaterOrEqual(t, count, 10, "PostgreSQL should have at least one batch")
}

// TestCacheIntegration_ConcurrentNodes tests multiple nodes reporting concurrently
func TestCacheIntegration_ConcurrentNodes(t *testing.T) {
	ctx := context.Background()
	pool := getTestPool(t)
	setupMetricsTable(t, ctx, pool)
	defer cleanupMetricsTable(t, ctx, pool)

	// Create memory cache and batch writer
	memoryCache := cache.NewMemoryCache()
	batchWriter := cache.NewBatchWriter(pool, 1000, 50)
	batchWriter.Start()
	defer batchWriter.Stop()

	// Simulate 10 nodes reporting concurrently
	numNodes := 10
	recordsPerNode := 20
	done := make(chan bool, numNodes)

	for i := 0; i < numNodes; i++ {
		go func(nodeIndex int) {
			defer func() { done <- true }()

			nodeID := uuid.New().String()
			probeID := uuid.New().String()
			baseTime := time.Now()

			for j := 0; j < recordsPerNode; j++ {
				point := &cache.MetricPoint{
					Timestamp:      baseTime.Add(time.Duration(j) * time.Second),
					LatencyMs:      float64(nodeIndex*100 + j*10),
					PacketLossRate: 0.1,
					JitterMs:       5.0,
				}

				memoryCache.Store(nodeID, point)

				record := &cache.MetricRecord{
					NodeID:         nodeID,
					ProbeID:        probeID,
					Timestamp:      point.Timestamp,
					LatencyMs:      point.LatencyMs,
					PacketLossRate: point.PacketLossRate,
					JitterMs:       point.JitterMs,
					IsAggregated:   false,
				}

				batchWriter.Write(record)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numNodes; i++ {
		<-done
	}

	// Wait for batch processing
	time.Sleep(3 * time.Second)

	// Verify memory cache has all nodes
	nodeIDs := memoryCache.GetAllNodeIDs()
	assert.Equal(t, numNodes, len(nodeIDs), "Memory cache should have all nodes")

	// Verify memory cache size matches requirement (supports at least 10 nodes)
	assert.GreaterOrEqual(t, memoryCache.GetSize(), 10, "Memory cache should support at least 10 nodes")

	// Verify PostgreSQL has data
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM metrics").Scan(&count)
	require.NoError(t, err, "Failed to query metrics table")

	// Should have at least some records
	assert.Greater(t, count, 0, "PostgreSQL should have records")
}

// TestCacheIntegration_MemoryOccupancy tests memory usage (estimation)
func TestCacheIntegration_MemoryOccupancy(t *testing.T) {
	// This test verifies memory estimation logic
	// Expected: ~30KB for 10 nodes × 1 hour × 60 minutes × 50 bytes/point

	memoryCache := cache.NewMemoryCache()

	// Simulate 10 nodes with 1 hour of data each
	numNodes := 10
	pointsPerNode := 60 // 1 per minute for 1 hour

	totalPoints := 0
	for i := 0; i < numNodes; i++ {
		nodeID := uuid.New().String()
		baseTime := time.Now()

		for j := 0; j < pointsPerNode; j++ {
			point := &cache.MetricPoint{
				Timestamp:      baseTime.Add(time.Duration(j) * time.Minute),
				LatencyMs:      float64(i*100 + j),
				PacketLossRate: 0.1,
				JitterMs:       5.0,
			}

			memoryCache.Store(nodeID, point)
			totalPoints++
		}
	}

	// Verify all nodes are in cache
	nodeIDs := memoryCache.GetAllNodeIDs()
	assert.Equal(t, numNodes, len(nodeIDs), "Should have all nodes")

	// Verify total points (should be numNodes × pointsPerNode)
	expectedTotalPoints := numNodes * pointsPerNode

	// Count total points in cache
	actualTotalPoints := 0
	for _, nodeID := range nodeIDs {
		points := memoryCache.Get(nodeID)
		actualTotalPoints += len(points)
	}

	assert.Equal(t, expectedTotalPoints, actualTotalPoints, "Should have all data points")

	// Memory estimation:
	// Each MetricPoint: ~48 bytes (Timestamp=24, LatencyMs=8, PacketLossRate=8, JitterMs=8)
	// Total: 600 points × 48 bytes = ~28.8 KB
	// This is well within the ~30KB requirement
	t.Logf("Total points in cache: %d", actualTotalPoints)
	t.Logf("Estimated memory usage: ~%.2f KB", float64(actualTotalPoints*48)/1024)

	// Verify memory usage is reasonable (should be < 100KB)
	estimatedKB := float64(actualTotalPoints*48) / 1024
	assert.Less(t, estimatedKB, 100.0, "Memory usage should be reasonable")
}

// TestCacheIntegration_FIFOEviction tests FIFO eviction after 1 hour
func TestCacheIntegration_FIFOEviction(t *testing.T) {
	memoryCache := cache.NewMemoryCache()

	nodeID := uuid.New().String()
	baseTime := time.Now()

	// Write 70 data points (more than 60 - 1 hour capacity)
	for i := 0; i < 70; i++ {
		point := &cache.MetricPoint{
			Timestamp:      baseTime.Add(time.Duration(i) * time.Minute),
			LatencyMs:      float64(i),
			PacketLossRate: 0.0,
			JitterMs:       5.0,
		}
		memoryCache.Store(nodeID, point)
	}

	points := memoryCache.Get(nodeID)

	// Should have at most 60 points (1 hour capacity)
	assert.LessOrEqual(t, len(points), 60, "Should have at most 60 points after FIFO eviction")

	// First data point should not be the original first (latency 0)
	if len(points) > 0 {
		assert.NotEqual(t, 0.0, points[0].LatencyMs, "First data points should be evicted after 1 hour")
	}
}

// TestCacheIntegration_Aggregation tests 1-minute aggregation logic
func TestCacheIntegration_Aggregation(t *testing.T) {
	memoryCache := cache.NewMemoryCache()

	nodeID := uuid.New().String()
	baseTime := time.Now().Truncate(time.Minute)

	// Write multiple points within the same minute
	pointsInMinute := 5
	for i := 0; i < pointsInMinute; i++ {
		point := &cache.MetricPoint{
			Timestamp:      baseTime.Add(time.Duration(i) * time.Second),
			LatencyMs:      float64(100 + i*10), // 100, 110, 120, 130, 140
			PacketLossRate: 0.1,
			JitterMs:       5.0,
		}
		memoryCache.Store(nodeID, point)
	}

	// Get aggregated metrics
	aggregated := memoryCache.AggregateMetricsByNode(nodeID)

	require.NotEmpty(t, aggregated, "Should have aggregated metrics")

	// Check average: (100+110+120+130+140)/5 = 120
	agg := aggregated[0]
	expectedAvg := 120.0
	assert.InDelta(t, expectedAvg, agg.LatencyMs, 0.01, "Average latency should be 120")

	// Check max
	expectedMax := 140.0
	assert.Equal(t, expectedMax, agg.MaxLatencyMs, "Max latency should be 140")

	// Check min
	expectedMin := 100.0
	assert.Equal(t, expectedMin, agg.MinLatencyMs, "Min latency should be 100")
}
