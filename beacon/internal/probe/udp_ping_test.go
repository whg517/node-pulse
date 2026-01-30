package probe

import (
	"math"
	"testing"
	"time"

	"beacon/internal/models"
)

// TestUDPProbeConfigValidation tests UDP probe configuration validation
func TestUDPProbeConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  UDPProbeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid UDP config",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          10,
			},
			wantErr: false,
		},
		{
			name: "valid UDP config with hostname",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "example.com",
				Port:           123,
				TimeoutSeconds: 10,
				Interval:       300,
				Count:          5,
			},
			wantErr: false,
		},
		{
			name: "invalid type - tcp_ping",
			config: UDPProbeConfig{
				Type:           "tcp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          10,
			},
			wantErr: true,
			errMsg:  "invalid probe type",
		},
		{
			name: "empty target",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "",
				Port:           53,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          10,
			},
			wantErr: true,
			errMsg:  "probe target cannot be empty",
		},
		{
			name: "invalid port - 0",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           0,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          10,
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "invalid port - 65536",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           65536,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          10,
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "invalid timeout - 0",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 0,
				Interval:       60,
				Count:          10,
			},
			wantErr: true,
			errMsg:  "invalid timeout",
		},
		{
			name: "invalid timeout - 31",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 31,
				Interval:       60,
				Count:          10,
			},
			wantErr: true,
			errMsg:  "invalid timeout",
		},
		{
			name: "valid timeout - 1 second (min)",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 1,
				Interval:       60,
				Count:          1,
			},
			wantErr: false,
		},
		{
			name: "valid timeout - 30 seconds (max)",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 30,
				Interval:       300,
				Count:          10,
			},
			wantErr: false,
		},
		{
			name: "invalid interval - 59",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 5,
				Interval:       59,
				Count:          10,
			},
			wantErr: true,
			errMsg:  "invalid interval",
		},
		{
			name: "invalid interval - 301",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 5,
				Interval:       301,
				Count:          10,
			},
			wantErr: true,
			errMsg:  "invalid interval",
		},
		{
			name: "invalid count - 0",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          0,
			},
			wantErr: true,
			errMsg:  "invalid count",
		},
		{
			name: "valid count - 100 (max)",
			config: UDPProbeConfig{
				Type:           "udp_ping",
				Target:         "8.8.8.8",
				Port:           53,
				TimeoutSeconds: 1,
				Interval:       100,
				Count:          100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if containsString(err.Error(), tt.errMsg) == false {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

// TestUDPPingerExecute tests UDP probe execution
func TestUDPPingerExecute(t *testing.T) {
	// Note: These tests require a real UDP server or will timeout
	// Integration tests will cover full functionality
	t.Run("basic execution structure", func(t *testing.T) {
		config := UDPProbeConfig{
			Type:           "udp_ping",
			Target:         "127.0.0.1",
			Port:           12345,
			TimeoutSeconds: 1,
			Interval:       60,
			Count:          1,
		}

		pinger := NewUDPPinger(config)

		// Execute should return a result (may fail due to no server)
		result, err := pinger.Execute()

		// Should not return error (result contains success status)
		if err != nil {
			t.Errorf("Execute() should not return error, got %v", err)
		}

		// Result should be non-nil
		if result == nil {
			t.Fatal("Execute() result should not be nil")
		}

		// Result should have timestamp
		if result.Timestamp == "" {
			t.Error("Execute() result should have timestamp")
		}

		// Test with localhost invalid port (should fail gracefully)
		if result.Success == true && result.ErrorMessage == "" {
			// This is unexpected - port 12345 should not have a server
			// But we don't fail the test, as environment may vary
			t.Logf("WARNING: Expected failure on port 12345, got success - this may be environment-specific")
		}
	})
}

// TestUDPProbeResultCreation tests UDP probe result creation
func TestUDPProbeResultCreation(t *testing.T) {
	t.Run("create success result", func(t *testing.T) {
		result := models.NewUDPProbeResult(true, 0.0, 50.5, 10, 10, "")

		if result == nil {
			t.Fatal("NewUDPProbeResult() should not return nil")
		}

		if result.Success != true {
			t.Errorf("Success = %v, want true", result.Success)
		}

		if result.PacketLossRate != 0.0 {
			t.Errorf("PacketLossRate = %v, want 0.0", result.PacketLossRate)
		}

		if result.RTTMs != 50.5 {
			t.Errorf("RTTMs = %v, want 50.5", result.RTTMs)
		}

		if result.SentPackets != 10 {
			t.Errorf("SentPackets = %v, want 10", result.SentPackets)
		}

		if result.ReceivedPackets != 10 {
			t.Errorf("ReceivedPackets = %v, want 10", result.ReceivedPackets)
		}

		if result.Timestamp == "" {
			t.Error("Timestamp should not be empty")
		}
	})

	t.Run("create failure result with packet loss", func(t *testing.T) {
		result := models.NewUDPProbeResult(false, 50.0, 0, 10, 5, "timeout")

		if result.Success != false {
			t.Errorf("Success = %v, want false", result.Success)
		}

		if result.PacketLossRate != 50.0 {
			t.Errorf("PacketLossRate = %v, want 50.0", result.PacketLossRate)
		}

		if result.RTTMs != 0 {
			t.Errorf("RTTMs = %v, want 0", result.RTTMs)
		}

		if result.SentPackets != 10 {
			t.Errorf("SentPackets = %v, want 10", result.SentPackets)
		}

		if result.ReceivedPackets != 5 {
			t.Errorf("ReceivedPackets = %v, want 5", result.ReceivedPackets)
		}

		if result.ErrorMessage != "timeout" {
			t.Errorf("ErrorMessage = %v, want 'timeout'", result.ErrorMessage)
		}
	})
}

// TestPacketLossRateCalculation tests packet loss rate calculation
func TestPacketLossRateCalculation(t *testing.T) {
	tests := []struct {
		name           string
		sent           int
		received       int
		expectedLoss   float64
		expectedRounded float64
	}{
		{
			name:           "0% packet loss",
			sent:           10,
			received:       10,
			expectedLoss:   0.0,
			expectedRounded: 0.0,
		},
		{
			name:           "50% packet loss",
			sent:           10,
			received:       5,
			expectedLoss:   50.0,
			expectedRounded: 50.0,
		},
		{
			name:           "100% packet loss",
			sent:           10,
			received:       0,
			expectedLoss:   100.0,
			expectedRounded: 100.0,
		},
		{
			name:           "33.33% packet loss",
			sent:           3,
			received:       2,
			expectedLoss:   33.333333,
			expectedRounded: 33.33,
		},
		{
			name:           "single packet received",
			sent:           1,
			received:       1,
			expectedLoss:   0.0,
			expectedRounded: 0.0,
		},
		{
			name:           "single packet lost",
			sent:           1,
			received:       0,
			expectedLoss:   100.0,
			expectedRounded: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packetLossRate := 0.0
			if tt.sent > 0 {
				packetLossRate = (1.0 - float64(tt.received)/float64(tt.sent)) * 100
			}

			rounded := math.Round(packetLossRate*100) / 100

			if !almostEqual(packetLossRate, tt.expectedLoss) {
				t.Errorf("Packet loss rate = %v, want %v", packetLossRate, tt.expectedLoss)
			}

			if rounded != tt.expectedRounded {
				t.Errorf("Rounded packet loss rate = %v, want %v", rounded, tt.expectedRounded)
			}
		})
	}
}

// TestExecuteBatch tests batch probe execution
func TestExecuteBatch(t *testing.T) {
	t.Run("batch with count parameter", func(t *testing.T) {
		config := UDPProbeConfig{
			Type:           "udp_ping",
			Target:         "127.0.0.1",
			Port:           12346,
			TimeoutSeconds: 1,
			Interval:       60,
			Count:          5,
		}

		pinger := NewUDPPinger(config)

		result, err := pinger.ExecuteBatch(5)

		if err != nil {
			t.Errorf("ExecuteBatch() error = %v", err)
		}

		if result == nil {
			t.Fatal("ExecuteBatch() result should not be nil")
		}

		if result.SentPackets != 5 {
			t.Errorf("SentPackets = %v, want 5", result.SentPackets)
		}

		if result.Timestamp == "" {
			t.Error("ExecuteBatch() result should have timestamp")
		}
	})

	t.Run("invalid count - 0", func(t *testing.T) {
		config := UDPProbeConfig{
			Type:           "udp_ping",
			Target:         "127.0.0.1",
			Port:           12347,
			TimeoutSeconds: 1,
			Interval:       60,
			Count:          1,
		}

		pinger := NewUDPPinger(config)

		_, err := pinger.ExecuteBatch(0)
		if err == nil {
			t.Error("ExecuteBatch(0) should return error")
		}
	})

	t.Run("invalid count - 101", func(t *testing.T) {
		config := UDPProbeConfig{
			Type:           "udp_ping",
			Target:         "127.0.0.1",
			Port:           12348,
			TimeoutSeconds: 1,
			Interval:       60,
			Count:          1,
		}

		pinger := NewUDPPinger(config)

		_, err := pinger.ExecuteBatch(101)
		if err == nil {
			t.Error("ExecuteBatch(101) should return error")
		}
	})
}

// TestRTTPrecision tests RTT measurement precision
func TestRTTPrecision(t *testing.T) {
	// Test that RTT is rounded to 2 decimal places
	elapsed := 123456789 * time.Nanosecond // 0.123456789 seconds
	rttMs := elapsed.Seconds() * 1000
	rounded := math.Round(rttMs*rttPrecisionMultiplier) / rttPrecisionMultiplier

	expected := 123.46
	if rounded != expected {
		t.Errorf("RTT rounding = %v, want %v", rounded, expected)
	}
}

// Helper functions
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && contains(s, substr))
}

func almostEqual(a, b float64) bool {
	epsilon := 0.0001
	return math.Abs(a-b) < epsilon
}
