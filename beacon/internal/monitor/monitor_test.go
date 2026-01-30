package monitor

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// mockProbeManager is a mock implementation of ProbeManager for testing
type mockProbeManager struct {
	mu              sync.Mutex
	lastMultiplier int
	intervalUpdates []int
}

func (m *mockProbeManager) UpdateProbeInterval(multiplier int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastMultiplier = multiplier
	m.intervalUpdates = append(m.intervalUpdates, multiplier)
	return nil
}

// mockLogger is a mock implementation of Logger for testing
type mockLogger struct {
	mu       sync.Mutex
	messages []string
}

func (m *mockLogger) Info(args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "INFO: "+fmt.Sprint(args...))
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "INFO: "+fmt.Sprintf(format, args...))
}

func (m *mockLogger) Warn(args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "WARN: "+fmt.Sprint(args...))
}

func (m *mockLogger) Warnf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "WARN: "+fmt.Sprintf(format, args...))
}

func (m *mockLogger) Error(args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "ERROR: "+fmt.Sprint(args...))
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "ERROR: "+fmt.Sprintf(format, args...))
}

func (m *mockLogger) Debug(args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "DEBUG: "+fmt.Sprint(args...))
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, "DEBUG: "+fmt.Sprintf(format, args...))
}

func TestNewMonitor(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *ResourceMonitorConfig
		probeMgr    ProbeManager
		logger      Logger
		expectError bool
	}{
		{
			name:        "nil config",
			cfg:         nil,
			probeMgr:    &mockProbeManager{},
			logger:      &mockLogger{},
			expectError: true,
		},
		{
			name: "monitor disabled",
			cfg: &ResourceMonitorConfig{
				Enabled: false,
			},
			probeMgr:    &mockProbeManager{},
			logger:      &mockLogger{},
			expectError: true,
		},
		{
			name: "valid config",
			cfg: &ResourceMonitorConfig{
				Enabled:              true,
				CheckIntervalSeconds: 1,
				Thresholds: ThresholdsConfig{
					CPUMicrocores: 100,
					MemoryMB:      100,
				},
				Degradation: DegradationConfig{
					DegradedLevel: DegradationLevelConfig{
						CPUMicrocores:      200,
						MemoryMB:           150,
						IntervalMultiplier: 2,
					},
					CriticalLevel: DegradationLevelConfig{
						CPUMicrocores:      300,
						MemoryMB:           200,
						IntervalMultiplier: 3,
					},
					Recovery: RecoveryConfig{
						ConsecutiveNormalChecks: 3,
					},
				},
				Alerting: AlertingConfig{
					SuppressionWindowSeconds: 5,
				},
			},
			probeMgr:    &mockProbeManager{},
			logger:      &mockLogger{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor, err := NewMonitor(tt.cfg, tt.probeMgr, tt.logger)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if monitor != nil {
					t.Error("Expected nil monitor but got non-nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if monitor == nil {
					t.Error("Expected monitor but got nil")
				}
			}
		})
	}
}

func TestMonitor_StartStop(t *testing.T) {
	cfg := &ResourceMonitorConfig{
		Enabled:              true,
		CheckIntervalSeconds: 1,
		Thresholds: ThresholdsConfig{
			CPUMicrocores: 100,
			MemoryMB:      100,
		},
		Degradation: DegradationConfig{
			DegradedLevel: DegradationLevelConfig{
				CPUMicrocores:      200,
				MemoryMB:           150,
				IntervalMultiplier: 2,
			},
			CriticalLevel: DegradationLevelConfig{
				CPUMicrocores:      300,
				MemoryMB:           200,
				IntervalMultiplier: 3,
			},
			Recovery: RecoveryConfig{
				ConsecutiveNormalChecks: 3,
			},
		},
		Alerting: AlertingConfig{
			SuppressionWindowSeconds: 5,
		},
	}

	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	// Test start
	if err := mon.Start(); err != nil {
		t.Errorf("Failed to start monitor: %v", err)
	}

	if !mon.IsRunning() {
		t.Error("Monitor should be running after Start()")
	}

	// Test start while already running
	if err := mon.Start(); err == nil {
		t.Error("Expected error when starting already running monitor")
	}

	// Let it run for a short time
	time.Sleep(100 * time.Millisecond)

	// Test stop
	mon.Stop()

	if mon.IsRunning() {
		t.Error("Monitor should not be running after Stop()")
	}

	// Test stop is idempotent
	mon.Stop() // Should not panic
}

func TestMonitor_DegradationLevel(t *testing.T) {
	cfg := &ResourceMonitorConfig{
		Enabled:              true,
		CheckIntervalSeconds: 1,
		Thresholds: ThresholdsConfig{
			CPUMicrocores: 100,
			MemoryMB:      100,
		},
		Degradation: DegradationConfig{
			DegradedLevel: DegradationLevelConfig{
				CPUMicrocores:      200,
				MemoryMB:           150,
				IntervalMultiplier: 2,
			},
			CriticalLevel: DegradationLevelConfig{
				CPUMicrocores:      300,
				MemoryMB:           200,
				IntervalMultiplier: 3,
			},
			Recovery: RecoveryConfig{
				ConsecutiveNormalChecks: 3,
			},
		},
		Alerting: AlertingConfig{
			SuppressionWindowSeconds: 5,
		},
	}

	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	// Initial level should be Normal
	level := mon.GetDegradationLevel()
	if level != DegradationLevelNormal {
		t.Errorf("Expected DegradationLevelNormal, got %s", level.String())
	}
}

func TestMonitor_GetResourceUsage(t *testing.T) {
	cfg := &ResourceMonitorConfig{
		Enabled:              true,
		CheckIntervalSeconds: 1,
		Thresholds: ThresholdsConfig{
			CPUMicrocores: 100,
			MemoryMB:      100,
		},
	}

	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	// Start monitor to begin resource collection
	if err := mon.Start(); err != nil {
		t.Fatalf("Failed to start monitor: %v", err)
	}
	defer mon.Stop()

	// Wait for at least one collection cycle
	time.Sleep(1100 * time.Millisecond)

	usage := mon.GetResourceUsage()
	if usage == nil {
		t.Error("Expected resource usage but got nil")
	} else {
		if usage.CPUMicrocores < 0 {
			t.Error("CPU microcores should be non-negative")
		}
		if usage.MemoryMB < 0 {
			t.Error("Memory MB should be non-negative")
		}
		if usage.Timestamp == 0 {
			t.Error("Timestamp should be set")
		}
	}
}

func TestMonitor_GetAlerts(t *testing.T) {
	cfg := &ResourceMonitorConfig{
		Enabled:              true,
		CheckIntervalSeconds: 1,
		Thresholds: ThresholdsConfig{
			CPUMicrocores: 100,
			MemoryMB:      100,
		},
	}

	probeMgr := &mockProbeManager{}
	logger := &mockLogger{}

	mon, err := NewMonitor(cfg, probeMgr, logger)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}

	// Initially no alerts
	alerts := mon.GetAlerts()
	if len(alerts) != 0 {
		t.Errorf("Expected no alerts initially, got %d", len(alerts))
	}
}

func TestDegradationLevel_String(t *testing.T) {
	tests := []struct {
		level    DegradationLevel
		expected string
	}{
		{DegradationLevelNormal, "normal"},
		{DegradationLevelDegraded, "degraded"},
		{DegradationLevelCritical, "critical"},
		{DegradationLevel(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}
