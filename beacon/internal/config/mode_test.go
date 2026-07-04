package config

import (
	"testing"
	"time"
)

func TestNewModeManager(t *testing.T) {
	cfg := &ModeConfig{
		Mode:                       ModeRegistered,
		ConfigCheckIntervalSeconds: 60,
		DegradedModeThreshold:      3,
	}

	manager := NewModeManager(cfg)
	if manager == nil {
		t.Fatal("Expected non-nil mode manager")
	}

	if manager.GetMode() != ModeRegistered {
		t.Errorf("Expected mode %s, got %s", ModeRegistered, manager.GetMode())
	}
}

func TestNewModeManager_NilConfig(t *testing.T) {
	manager := NewModeManager(nil)
	if manager == nil {
		t.Fatal("Expected non-nil mode manager with nil config")
	}

	// Should use defaults
	if manager.GetMode() != ModeRegistered {
		t.Errorf("Expected default mode %s, got %s", ModeRegistered, manager.GetMode())
	}

	if manager.degradedThreshold != 3 {
		t.Errorf("Expected default threshold 3, got %d", manager.degradedThreshold)
	}
}

func TestModeManager_RecordHeartbeatSuccess(t *testing.T) {
	manager := NewModeManager(nil)

	// First set to degraded mode
	manager.currentMode = ModeDegraded

	manager.RecordHeartbeatSuccess()

	// Should exit degraded mode
	if manager.GetMode() != ModeRegistered {
		t.Errorf("Expected mode to be %s after success, got %s", ModeRegistered, manager.GetMode())
	}

	if manager.GetConsecutiveFailures() != 0 {
		t.Errorf("Expected consecutive failures to be 0, got %d", manager.GetConsecutiveFailures())
	}
}

func TestModeManager_RecordHeartbeatFailure(t *testing.T) {
	manager := NewModeManager(&ModeConfig{
		DegradedModeThreshold: 3,
	})

	// Record failures below threshold
	manager.RecordHeartbeatFailure()
	manager.RecordHeartbeatFailure()

	if manager.GetMode() != ModeRegistered {
		t.Errorf("Expected mode to still be %s, got %s", ModeRegistered, manager.GetMode())
	}

	// Third failure should trigger degraded mode
	manager.RecordHeartbeatFailure()

	if manager.GetMode() != ModeDegraded {
		t.Errorf("Expected mode to be %s after 3 failures, got %s", ModeDegraded, manager.GetMode())
	}

	// Config source should be cached
	if manager.GetConfigSource() != SourceCached {
		t.Errorf("Expected config source to be %s, got %s", SourceCached, manager.GetConfigSource())
	}
}

func TestModeManager_DegradedModeRecovery(t *testing.T) {
	manager := NewModeManager(&ModeConfig{
		DegradedModeThreshold: 2,
	})

	// Enter degraded mode
	manager.RecordHeartbeatFailure()
	manager.RecordHeartbeatFailure()

	if manager.GetMode() != ModeDegraded {
		t.Fatal("Expected to be in degraded mode")
	}

	// Successful heartbeat should exit degraded mode
	manager.RecordHeartbeatSuccess()

	if manager.GetMode() != ModeRegistered {
		t.Errorf("Expected mode to be %s after recovery, got %s", ModeRegistered, manager.GetMode())
	}
}

func TestModeManager_OnModeChange(t *testing.T) {
	manager := NewModeManager(&ModeConfig{
		DegradedModeThreshold: 2,
	})

	modeChanged := false
	var oldMode, newMode OperatingMode

	manager.OnModeChange(func(old, new OperatingMode) {
		modeChanged = true
		oldMode = old
		newMode = new
	})

	// Trigger mode change
	manager.RecordHeartbeatFailure()
	manager.RecordHeartbeatFailure()

	if !modeChanged {
		t.Error("Expected mode change callback to be called")
	}

	if oldMode != ModeRegistered {
		t.Errorf("Expected old mode %s, got %s", ModeRegistered, oldMode)
	}

	if newMode != ModeDegraded {
		t.Errorf("Expected new mode %s, got %s", ModeDegraded, newMode)
	}
}

func TestModeManager_IsConnected(t *testing.T) {
	manager := NewModeManager(nil)

	// Registered mode
	if !manager.IsConnected() {
		t.Error("Expected IsConnected to be true in registered mode")
	}

	// Degraded mode
	manager.currentMode = ModeDegraded
	if !manager.IsConnected() {
		t.Error("Expected IsConnected to be true in degraded mode")
	}

	// Standalone mode
	manager.currentMode = ModeStandalone
	if manager.IsConnected() {
		t.Error("Expected IsConnected to be false in standalone mode")
	}
}

func TestModeManager_IsStandalone(t *testing.T) {
	manager := NewModeManager(nil)

	if manager.IsStandalone() {
		t.Error("Expected IsStandalone to be false in registered mode")
	}

	manager.currentMode = ModeStandalone
	if !manager.IsStandalone() {
		t.Error("Expected IsStandalone to be true in standalone mode")
	}
}

func TestModeManager_GetStatus(t *testing.T) {
	manager := NewModeManager(&ModeConfig{
		Mode:                       ModeRegistered,
		ConfigCheckIntervalSeconds: 30,
		DegradedModeThreshold:      3,
	})

	status := manager.GetStatus()

	if status.CurrentMode != string(ModeRegistered) {
		t.Errorf("Expected current mode %s, got %s", ModeRegistered, status.CurrentMode)
	}

	if status.DegradedThreshold != 3 {
		t.Errorf("Expected degraded threshold 3, got %d", status.DegradedThreshold)
	}

	if status.ConfigCheckInterval != "30s" {
		t.Errorf("Expected config check interval 30s, got %s", status.ConfigCheckInterval)
	}
}

func TestModeManager_SetConfigSource(t *testing.T) {
	manager := NewModeManager(nil)

	manager.SetConfigSource(SourceServer)

	if manager.GetConfigSource() != SourceServer {
		t.Errorf("Expected config source %s, got %s", SourceServer, manager.GetConfigSource())
	}
}

func TestModeManager_GetConfigCheckInterval(t *testing.T) {
	manager := NewModeManager(&ModeConfig{
		ConfigCheckIntervalSeconds: 120,
	})

	interval := manager.GetConfigCheckInterval()
	if interval != 120*time.Second {
		t.Errorf("Expected interval 120s, got %s", interval)
	}
}

func TestModeConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ModeConfig
		wantErr bool
	}{
		{
			name:    "valid registered mode",
			config:  ModeConfig{Mode: ModeRegistered, ConfigCheckIntervalSeconds: 60},
			wantErr: false,
		},
		{
			name:    "valid standalone mode",
			config:  ModeConfig{Mode: ModeStandalone},
			wantErr: false,
		},
		{
			name:    "invalid mode",
			config:  ModeConfig{Mode: OperatingMode("invalid")},
			wantErr: true,
		},
		{
			name:    "interval too small",
			config:  ModeConfig{ConfigCheckIntervalSeconds: 5},
			wantErr: true,
		},
		{
			name:    "interval too large",
			config:  ModeConfig{ConfigCheckIntervalSeconds: 500},
			wantErr: true,
		},
		{
			name:    "threshold zero (uses default)",
			config:  ModeConfig{DegradedModeThreshold: 0},
			wantErr: false, // zero means use default, which is valid
		},
		{
			name:    "threshold too small (negative)",
			config:  ModeConfig{DegradedModeThreshold: -1},
			wantErr: true, // negative values are invalid
		},
		{
			name:    "threshold too large",
			config:  ModeConfig{DegradedModeThreshold: 15},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompressionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  CompressionConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  CompressionConfig{Enabled: true, Level: 6, MinSizeBytes: 1024},
			wantErr: false,
		},
		{
			name:    "level zero (uses default)",
			config:  CompressionConfig{Level: 0},
			wantErr: false, // zero means use default
		},
		{
			name:    "level too high",
			config:  CompressionConfig{Level: 10},
			wantErr: true,
		},
		{
			name:    "negative min size",
			config:  CompressionConfig{MinSizeBytes: -1},
			wantErr: true, // negative values are invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
