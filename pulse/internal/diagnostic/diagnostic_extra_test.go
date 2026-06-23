package diagnostic

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiagnosticEngine_SetBaselineCalculator tests SetBaselineCalculator
func TestDiagnosticEngine_SetBaselineCalculator(t *testing.T) {
	engine := NewDiagnosticEngine()
	calc := &BaselineCalculator{}

	engine.SetBaselineCalculator(calc)
	assert.Equal(t, calc, engine.GetBaselineCalculator())
}

// TestDiagnosticEngine_GetBaselineCalculator_Nil tests when nil
func TestDiagnosticEngine_GetBaselineCalculator_Nil(t *testing.T) {
	engine := NewDiagnosticEngine()
	assert.Nil(t, engine.GetBaselineCalculator())
}

// TestDiagnosticEngine_UpdateBaselinesFromCalculator_NilCalc tests nil calculator
func TestDiagnosticEngine_UpdateBaselinesFromCalculator_NilCalc(t *testing.T) {
	engine := NewDiagnosticEngine()
	updated := engine.UpdateBaselinesFromCalculator(context.Background())
	assert.False(t, updated)
}

// TestBaselineCalculator_GetBaselines_NilDB tests GetBaselines with nil DB panics
func TestBaselineCalculator_GetBaselines_NilDB(t *testing.T) {
	calc := NewBaselineCalculator(nil)

	// With nil DB, GetBaselines panics
	assert.Panics(t, func() {
		_, _ = calc.GetBaselines(context.Background())
	})
}

// TestBaselineCalculator_Info_Error_Warn tests logger methods
func TestBaselineCalculator_LoggerMethods(t *testing.T) {
	calc := NewBaselineCalculator(nil)

	// The noopLogger should not panic
	assert.NotPanics(t, func() {
		calc.logger.Info("test message")
		calc.logger.Error("test error")
		calc.logger.Warn("test warning")
	})
}

// TestBaselineCalculator_New_WithOptions tests options
func TestBaselineCalculator_New_WithOptions(t *testing.T) {
	customInterval := 5 * time.Minute
	customTTL := 10 * time.Minute

	calc := NewBaselineCalculator(nil,
		WithUpdateInterval(customInterval),
		WithCacheTTL(customTTL),
	)

	assert.Equal(t, customInterval, calc.updateInterval)
	assert.Equal(t, customTTL, calc.cacheTTL)
}

// TestBaselineCalculator_Start_Stop tests goroutine lifecycle
func TestBaselineCalculator_Start_Stop(t *testing.T) {
	calc := NewBaselineCalculator(nil)

	// Start without DB - background worker runs but can't update
	assert.NotPanics(t, func() {
		require.NoError(t, calc.Start(context.Background()))
	})

	// Short delay
	time.Sleep(10 * time.Millisecond)

	// Stop should work
	assert.NotPanics(t, func() {
		require.NoError(t, calc.Stop())
	})
}

// TestDiagnosticEngine_CalculateConfidence tests confidence calculation
func TestDiagnosticEngine_Diagnose_HighConfidence(t *testing.T) {
	engine := NewDiagnosticEngine()

	// All nodes show similar degradation → high confidence
	data := make([]MetricData, 10)
	for i := range data {
		data[i] = MetricData{
			NodeID:         fmt.Sprintf("node-%d", i),
			Region:         "us-east",
			Latency:      300,
			PacketLossRate: 0.5,
			Jitter:       50,
		}
	}

	result, err := engine.Diagnose(data)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Confidence)
}

// TestDiagnosticEngine_Diagnose_AllHealthy tests with healthy nodes
func TestDiagnosticEngine_Diagnose_AllHealthy(t *testing.T) {
	engine := NewDiagnosticEngine()

	data := []MetricData{
		{NodeID: "node-1", Region: "us-east", Latency: 20, PacketLossRate: 0.001, Jitter: 1},
		{NodeID: "node-2", Region: "us-west", Latency: 25, PacketLossRate: 0.002, Jitter: 1},
		{NodeID: "node-3", Region: "eu-west", Latency: 22, PacketLossRate: 0.001, Jitter: 2},
	}

	result, err := engine.Diagnose(data)
	require.NoError(t, err)
	assert.NotNil(t, result)
}
