package config

import (
	"testing"
	"time"
)

// TestModeManager_GetLastSuccessTime tests GetLastSuccessTime
func TestModeManager_GetLastSuccessTime(t *testing.T) {
	manager := NewModeManager(nil)

	// Initially zero
	zeroTime := manager.GetLastSuccessTime()
	if !zeroTime.IsZero() {
		t.Error("Expected zero time initially")
	}

	// After success, should be updated
	manager.RecordHeartbeatSuccess()
	successTime := manager.GetLastSuccessTime()
	if successTime.IsZero() {
		t.Error("Expected non-zero time after success")
	}
	if time.Since(successTime) > 2*time.Second {
		t.Error("Expected recent success time")
	}
}

// TestModeManager_GetLastFailureTime tests GetLastFailureTime
func TestModeManager_GetLastFailureTime(t *testing.T) {
	manager := NewModeManager(nil)

	// Initially zero
	zeroTime := manager.GetLastFailureTime()
	if !zeroTime.IsZero() {
		t.Error("Expected zero time initially")
	}

	// After failure, should be updated
	manager.RecordHeartbeatFailure()
	failureTime := manager.GetLastFailureTime()
	if failureTime.IsZero() {
		t.Error("Expected non-zero time after failure")
	}
	if time.Since(failureTime) > 2*time.Second {
		t.Error("Expected recent failure time")
	}
}

// TestResumeConfig_Validate tests validation of ResumeConfig
func TestResumeConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ResumeConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: ResumeConfig{
				Enabled:             true,
				MaxCacheSizeBytes:   10 * 1024 * 1024,
				AlertReservePercent: 30,
			},
			wantErr: false,
		},
		{
			name: "zero values are valid",
			cfg:  ResumeConfig{},
			wantErr: false,
		},
		{
			name: "max cache size too large",
			cfg: ResumeConfig{
				MaxCacheSizeBytes: 101 * 1024 * 1024, // > 100MB
			},
			wantErr: true,
		},
		{
			name: "negative max cache size",
			cfg: ResumeConfig{
				MaxCacheSizeBytes: -1,
			},
			wantErr: true,
		},
		{
			name: "alert reserve percent too high",
			cfg: ResumeConfig{
				AlertReservePercent: 101,
			},
			wantErr: true,
		},
		{
			name: "alert reserve percent negative",
			cfg: ResumeConfig{
				AlertReservePercent: -1,
			},
			wantErr: true,
		},
		{
			name: "max cache size exactly 100MB",
			cfg: ResumeConfig{
				MaxCacheSizeBytes: 100 * 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "alert reserve percent 100",
			cfg: ResumeConfig{
				AlertReservePercent: 100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDefaultResumeConfig tests DefaultResumeConfig
func TestDefaultResumeConfig(t *testing.T) {
	cfg := DefaultResumeConfig()

	if !cfg.Enabled {
		t.Error("Expected Enabled=true")
	}
	if cfg.MaxCacheSizeBytes != 10*1024*1024 {
		t.Errorf("Expected MaxCacheSizeBytes=10MB, got %d", cfg.MaxCacheSizeBytes)
	}
	if cfg.CacheFilePath == "" {
		t.Error("Expected non-empty CacheFilePath")
	}
	if !cfg.AlertPriorityMode {
		t.Error("Expected AlertPriorityMode=true")
	}
	if cfg.AlertReservePercent != 30 {
		t.Errorf("Expected AlertReservePercent=30, got %d", cfg.AlertReservePercent)
	}

	// Validate default config is valid
	if err := cfg.Validate(); err != nil {
		t.Errorf("Default config should be valid, got: %v", err)
	}
}
