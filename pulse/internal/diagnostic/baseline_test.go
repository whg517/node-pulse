package diagnostic

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLogger captures log messages for testing
type mockLogger struct {
	mu       sync.Mutex
	messages []string
}

func (l *mockLogger) Info(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, msg)
}

func (l *mockLogger) Error(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, msg)
}

func (l *mockLogger) Warn(msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = append(l.messages, msg)
}

func TestBaselineCache_GetSet(t *testing.T) {
	cache := &BaselineCache{}

	// Test empty cache returns nil
	assert.Nil(t, cache.Get())

	// Test setting and getting value
	baseline := &Baseline{
		LatencyMs:      45.5,
		PacketLossRate: 0.008,
		JitterMs:       1.8,
		CalculatedAt:   time.Now(),
		NodeCount:      10,
		DataPointCount: 1000,
	}

	cache.Set(baseline, 5*time.Minute)
	result := cache.Get()
	require.NotNil(t, result)
	assert.Equal(t, 45.5, result.LatencyMs)
	assert.Equal(t, 0.008, result.PacketLossRate)
	assert.Equal(t, 1.8, result.JitterMs)
	assert.Equal(t, 10, result.NodeCount)
	assert.Equal(t, 1000, result.DataPointCount)
}

func TestBaselineCache_Expiration(t *testing.T) {
	cache := &BaselineCache{}

	baseline := &Baseline{
		LatencyMs:    50.0,
		CalculatedAt: time.Now(),
	}

	// Set with very short TTL
	cache.Set(baseline, 1*time.Millisecond)

	// Should be available immediately
	assert.NotNil(t, cache.Get())

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Should be nil after expiration
	assert.Nil(t, cache.Get())
}

func TestBaselineCache_Clear(t *testing.T) {
	cache := &BaselineCache{}

	baseline := &Baseline{LatencyMs: 50.0}
	cache.Set(baseline, 5*time.Minute)

	assert.NotNil(t, cache.Get())

	cache.Clear()

	assert.Nil(t, cache.Get())
}

func TestBaselineCalculator_GetBaselines_CacheHit(t *testing.T) {
	// Create calculator with nil DB (won't be used for cache hit)
	calc := NewBaselineCalculator(nil)

	// Pre-populate cache
	expectedBaseline := &Baseline{
		LatencyMs:      42.5,
		PacketLossRate: 0.015,
		JitterMs:       2.3,
		CalculatedAt:   time.Now(),
		NodeCount:      5,
		DataPointCount: 500,
	}
	calc.cache.Set(expectedBaseline, 5*time.Minute)

	// GetBaselines should return cached value without hitting DB
	result, err := calc.GetBaselines(context.Background())
	require.NoError(t, err)
	assert.Equal(t, expectedBaseline.LatencyMs, result.LatencyMs)
	assert.Equal(t, expectedBaseline.PacketLossRate, result.PacketLossRate)
	assert.Equal(t, expectedBaseline.JitterMs, result.JitterMs)
}

func TestBaselineCalculator_Options(t *testing.T) {
	logger := &mockLogger{}

	calc := NewBaselineCalculator(
		nil,
		WithUpdateInterval(5*time.Minute),
		WithCacheTTL(10*time.Minute),
		WithLogger(logger),
	)

	assert.Equal(t, 5*time.Minute, calc.updateInterval)
	assert.Equal(t, 10*time.Minute, calc.cacheTTL)
	assert.Equal(t, logger, calc.logger)
}

func TestBaselineCalculator_StartStop(t *testing.T) {
	// Test with nil DB - background worker should skip calculations gracefully
	calc := NewBaselineCalculator(
		nil,
		WithUpdateInterval(100*time.Millisecond),
	)

	// Start the worker
	err := calc.Start(context.Background())
	require.NoError(t, err)

	// Starting again should fail
	err = calc.Start(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Stop the worker
	err = calc.Stop()
	require.NoError(t, err)

	// Stop again should be no-op (no error)
	err = calc.Stop()
	require.NoError(t, err)
}

func TestBaselineCalculator_StopWhileRunning(t *testing.T) {
	calc := NewBaselineCalculator(
		nil,
		WithUpdateInterval(10*time.Second), // Long interval so we stop before next tick
	)

	err := calc.Start(context.Background())
	require.NoError(t, err)

	// Give worker time to start
	time.Sleep(50 * time.Millisecond)

	// Stop immediately
	err = calc.Stop()
	require.NoError(t, err)

	assert.False(t, calc.running)
}

func TestBaselineCalculator_IntegrationWithDiagnosticEngine(t *testing.T) {
	// Create a baseline calculator with pre-cached values
	calc := NewBaselineCalculator(nil)
	customBaseline := &Baseline{
		LatencyMs:      75.0,
		PacketLossRate: 0.02,
		JitterMs:       3.5,
		CalculatedAt:   time.Now(),
		NodeCount:      10,
		DataPointCount: 1000,
	}
	calc.cache.Set(customBaseline, 5*time.Minute)

	// Get baselines from calculator
	baseline, err := calc.GetBaselines(context.Background())
	require.NoError(t, err)

	// Create diagnostic engine with calculated baselines
	engine := NewDiagnosticEngineWithBaselines(
		baseline.LatencyMs,
		baseline.PacketLossRate,
		baseline.JitterMs,
	)

	// Verify baselines are set correctly
	assert.Equal(t, 75.0, engine.baselineLatency)
	assert.Equal(t, 0.02, engine.baselinePacketLoss)
	assert.Equal(t, 3.5, engine.baselineJitter)
}

func TestBaselineCalculator_DefaultLogger(t *testing.T) {
	calc := NewBaselineCalculator(nil)

	// Should have no-op logger by default
	_, ok := calc.logger.(*noopLogger)
	assert.True(t, ok)

	// These should not panic
	calc.logger.Info("test")
	calc.logger.Error("test")
	calc.logger.Warn("test")
}

// TestBaselineCache_ConcurrentAccess tests thread safety
func TestBaselineCache_ConcurrentAccess(t *testing.T) {
	cache := &BaselineCache{}
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(val float64) {
			defer wg.Done()
			cache.Set(&Baseline{LatencyMs: val}, 5*time.Minute)
		}(float64(i))
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cache.Get()
		}()
	}

	wg.Wait()

	// Cache should have a value (one of the writers set it)
	assert.NotNil(t, cache.Get())
}

// Integration test with real database requires Docker
// This test is skipped if no database is available
func TestBaselineCalculator_CalculateBaselines_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires a real database connection
	// In CI, this would be set up via Docker
	dbURL := getTestDBURL()
	if dbURL == "" {
		t.Skip("No test database URL available")
	}

	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Set up test data
	setupBaselineTestData(t, db)
	defer cleanupBaselineTestData(t, db)

	calc := NewBaselineCalculator(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	baseline, err := calc.CalculateBaselines(ctx)
	require.NoError(t, err)
	require.NotNil(t, baseline)

	// Verify calculated values are reasonable
	assert.Positive(t, baseline.LatencyMs)
	assert.Positive(t, baseline.PacketLossRate)
	assert.Positive(t, baseline.JitterMs)
	assert.Positive(t, baseline.NodeCount)
	assert.Positive(t, baseline.DataPointCount)
	assert.False(t, baseline.CalculatedAt.IsZero())
}

func getTestDBURL() string {
	// Would be set via environment variable in CI
	return ""
}

func setupBaselineTestData(t *testing.T, db *sql.DB) {
	// Insert test metrics data for baseline calculation
	t.Helper()
}

func cleanupBaselineTestData(t *testing.T, db *sql.DB) {
	// Clean up test data
	t.Helper()
}
