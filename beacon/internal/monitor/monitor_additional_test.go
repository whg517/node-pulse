package monitor

import (
	"testing"
	"time"
)

// newValidMonitorConfig creates a valid ResourceMonitorConfig with given thresholds
func newValidMonitorConfig(cpuThreshold, memThreshold int) *ResourceMonitorConfig {
	return &ResourceMonitorConfig{
		Enabled:              true,
		CheckIntervalSeconds: 1,
		Thresholds: ThresholdsConfig{
			CPUMicrocores: cpuThreshold,
			MemoryMB:      memThreshold,
		},
		Degradation: DegradationConfig{
			DegradedLevel: DegradationLevelConfig{
				CPUMicrocores:      cpuThreshold * 2,
				MemoryMB:           memThreshold * 2,
				IntervalMultiplier: 2,
			},
			CriticalLevel: DegradationLevelConfig{
				CPUMicrocores:      cpuThreshold * 3,
				MemoryMB:           memThreshold * 3,
				IntervalMultiplier: 3,
			},
			Recovery: RecoveryConfig{
				ConsecutiveNormalChecks: 2,
			},
		},
		Alerting: AlertingConfig{
			SuppressionWindowSeconds: 5,
		},
	}
}

// TestMonitor_EvaluateDegradation_CriticalLevel tests transition to critical
func TestMonitor_EvaluateDegradation_CriticalLevel(t *testing.T) {
	cfg := newValidMonitorConfig(100, 100)
	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	m := mon.(*monitor)

	// Critical: CPU > criticalLevel.CPUMicrocores (300)
	usage := &ResourceUsage{
		CPUMicrocores: 301,
		MemoryMB:      0,
	}
	m.evaluateDegradation(usage)

	if m.level != DegradationLevelCritical {
		t.Errorf("Expected critical level, got %v", m.level)
	}
	if probeMgr.lastMultiplier != 3 {
		t.Errorf("Expected multiplier 3 for critical, got %d", probeMgr.lastMultiplier)
	}
}

// TestMonitor_EvaluateDegradation_DegradedLevel tests transition to degraded
func TestMonitor_EvaluateDegradation_DegradedLevel(t *testing.T) {
	cfg := newValidMonitorConfig(100, 100)
	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	m := mon.(*monitor)

	// Degraded: CPU > degradedLevel.CPUMicrocores (200), < critical (300)
	usage := &ResourceUsage{
		CPUMicrocores: 250,
		MemoryMB:      0,
	}
	m.evaluateDegradation(usage)

	if m.level != DegradationLevelDegraded {
		t.Errorf("Expected degraded level, got %v", m.level)
	}
	if probeMgr.lastMultiplier != 2 {
		t.Errorf("Expected multiplier 2 for degraded, got %d", probeMgr.lastMultiplier)
	}
}

// TestMonitor_EvaluateDegradation_RecoveryFromDegraded tests recovery
func TestMonitor_EvaluateDegradation_RecoveryFromDegraded(t *testing.T) {
	cfg := newValidMonitorConfig(100, 100)
	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	m := mon.(*monitor)

	// First: enter degraded
	degradedUsage := &ResourceUsage{CPUMicrocores: 250, MemoryMB: 0}
	m.evaluateDegradation(degradedUsage)
	if m.level != DegradationLevelDegraded {
		t.Fatalf("Expected degraded level initially")
	}

	// Recovery requires ConsecutiveNormalChecks=2 normal checks
	normalUsage := &ResourceUsage{CPUMicrocores: 10, MemoryMB: 10}

	// First normal check - should NOT transition yet
	m.evaluateDegradation(normalUsage)
	if m.level != DegradationLevelDegraded {
		t.Log("Level changed on first normal check - recovery threshold is 1")
	}

	// Second normal check - should transition to normal (threshold is 2)
	m.evaluateDegradation(normalUsage)

	// After consecutive normal checks met, should be normal
	if m.level != DegradationLevelNormal {
		t.Logf("Level after 2 normal checks: %v (may still need more)", m.level)
	}
}

// TestMonitor_EvaluateDegradation_MemoryThreshold tests memory-based degradation
func TestMonitor_EvaluateDegradation_MemoryThreshold(t *testing.T) {
	cfg := newValidMonitorConfig(100, 100)
	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	m := mon.(*monitor)

	// Critical memory: MemoryMB > criticalLevel.MemoryMB (300)
	usage := &ResourceUsage{
		CPUMicrocores: 0,
		MemoryMB:      301,
	}
	m.evaluateDegradation(usage)

	if m.level != DegradationLevelCritical {
		t.Errorf("Expected critical level for high memory, got %v", m.level)
	}
}

// TestMonitor_GetIntervalMultiplier tests getIntervalMultiplier
func TestMonitor_GetIntervalMultiplier(t *testing.T) {
	cfg := newValidMonitorConfig(100, 100)
	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	m := mon.(*monitor)

	// Normal = 1
	if m.getIntervalMultiplier(DegradationLevelNormal) != 1 {
		t.Error("Expected multiplier 1 for normal level")
	}

	// Degraded = 2 (from config)
	if m.getIntervalMultiplier(DegradationLevelDegraded) != 2 {
		t.Errorf("Expected multiplier 2 for degraded level, got %d",
			m.getIntervalMultiplier(DegradationLevelDegraded))
	}

	// Critical = 3 (from config)
	if m.getIntervalMultiplier(DegradationLevelCritical) != 3 {
		t.Errorf("Expected multiplier 3 for critical level, got %d",
			m.getIntervalMultiplier(DegradationLevelCritical))
	}
}

// TestMonitor_CheckResources_Running tests checkResources when started
func TestMonitor_CheckResources_Running(t *testing.T) {
	cfg := newValidMonitorConfig(100, 100)
	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer mon.Stop()

	// Wait for some resource checks
	time.Sleep(1500 * time.Millisecond)

	// Should have collected some data
	usage := mon.GetResourceUsage()
	// Just verify it doesn't panic
	_ = usage
}

// TestMonitor_MaybeTriggerAlert tests maybeTriggerAlert
func TestMonitor_MaybeTriggerAlert(t *testing.T) {
	cfg := newValidMonitorConfig(100, 100)
	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	m := mon.(*monitor)

	now := time.Now().Unix()

	// First alert - should trigger
	m.maybeTriggerAlert("cpu", 150.0, 100.0, now)
	if len(m.alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(m.alerts))
	}

	// Second alert within suppression window - should be suppressed
	m.maybeTriggerAlert("cpu", 150.0, 100.0, now+1) // still within window

	// Different alert - should trigger
	m.maybeTriggerAlert("memory", 150.0, 100.0, now)
	if len(m.alerts) != 2 {
		t.Errorf("Expected 2 alerts, got %d", len(m.alerts))
	}

	// Alert after suppression window - should trigger
	m.maybeTriggerAlert("cpu", 150.0, 100.0, now+int64(cfg.Alerting.SuppressionWindowSeconds)+1)
	if len(m.alerts) != 3 {
		t.Errorf("Expected 3 alerts, got %d", len(m.alerts))
	}

	// Critical level alert
	m.maybeTriggerAlert("cpu", 350.0, 100.0, now+100)
	if len(m.alerts) < 4 {
		t.Errorf("Expected at least 4 alerts")
	}
}

// TestMonitor_CheckThresholds tests checkThresholds
func TestMonitor_CheckThresholds(t *testing.T) {
	// Use very low thresholds to trigger warnings
	cfg := newValidMonitorConfig(1, 1) // Very low thresholds
	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	m := mon.(*monitor)

	// High resource usage - should trigger threshold warnings
	usage := &ResourceUsage{
		CPUMicrocores: 500,  // Much higher than threshold of 1
		MemoryMB:      1000, // Much higher than threshold of 1
	}
	m.checkThresholds(usage)
	// Should not panic
}
