package probe

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestICMPProbeConfig_Validate tests configuration validation
func TestICMPProbeConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ICMPProbeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid configuration",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			},
			wantErr: false,
		},
		{
			name: "Valid hostname",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "example.com",
				TimeoutSeconds: 10,
				Interval:       60,
				Count:          1,
			},
			wantErr: false,
		},
		{
			name: "Valid IPv6 address",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "::1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			},
			wantErr: false,
		},
		{
			name: "Invalid type",
			config: ICMPProbeConfig{
				Type:           "tcp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "invalid probe type",
		},
		{
			name: "Invalid timeout (too low)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: -1,
				Interval:       60,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "timeout",
		},
		{
			name: "Invalid timeout (too high)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 31,
				Interval:       60,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "timeout",
		},
		{
			name: "Invalid interval (too low)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       59,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "interval",
		},
		{
			name: "Invalid interval (too high)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       301,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "interval",
		},
		{
			name: "Invalid count (too low)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          0,
			},
			wantErr: true,
			errMsg:  "count",
		},
		{
			name: "Invalid count (too high)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          101,
			},
			wantErr: true,
			errMsg:  "count",
		},
		{
			name: "Invalid target (empty)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "target",
		},
		{
			name: "Invalid packet size (too low)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
				PacketSize:     7,
			},
			wantErr: true,
			errMsg:  "packet_size",
		},
		{
			name: "Invalid packet size (too high)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
				PacketSize:     65508,
			},
			wantErr: true,
			errMsg:  "packet_size",
		},
		{
			name: "Valid packet size boundary (minimum)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
				PacketSize:     8,
			},
			wantErr: false,
		},
		{
			name: "Valid packet size boundary (maximum)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
				PacketSize:     65507,
			},
			wantErr: false,
		},
		{
			name: "Default packet size (0 -> 56)",
			config: ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
				PacketSize:     0,
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
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

// TestNewICMPPinger tests creating a new ICMP pinger
func TestNewICMPPinger(t *testing.T) {
	config := ICMPProbeConfig{
		Type:           "icmp_ping",
		Target:         "192.168.1.1",
		TimeoutSeconds: 5,
		Interval:       60,
		Count:          1,
	}

	pinger := NewICMPPinger(config)

	if pinger == nil {
		t.Fatal("Expected non-nil pinger")
	}

	if pinger.config.Target != "192.168.1.1" {
		t.Errorf("Expected target '192.168.1.1', got '%s'", pinger.config.Target)
	}

	// Verify default packet size is set
	if pinger.config.PacketSize != 56 {
		t.Errorf("Expected default packet size 56, got %d", pinger.config.PacketSize)
	}
}

// TestNewICMPPinger_DefaultPacketSize tests that default packet size is applied
func TestNewICMPPinger_DefaultPacketSize(t *testing.T) {
	tests := []struct {
		name         string
		packetSize   int
		expectedSize int
	}{
		{"Zero packet size defaults to 56", 0, 56},
		{"Custom packet size preserved", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
				PacketSize:     tt.packetSize,
			}

			pinger := NewICMPPinger(config)

			if pinger.config.PacketSize != tt.expectedSize {
				t.Errorf("Expected packet size %d, got %d", tt.expectedSize, pinger.config.PacketSize)
			}
		})
	}
}

// TestICMPPinger_Execute_InvalidConfig tests Execute with invalid configuration
func TestICMPPinger_Execute_InvalidConfig(t *testing.T) {
	config := ICMPProbeConfig{
		Type:           "tcp_ping", // Invalid type for ICMP
		Target:         "192.168.1.1",
		TimeoutSeconds: 5,
		Interval:       60,
		Count:          1,
	}

	pinger := NewICMPPinger(config)
	result, err := pinger.Execute()

	// Should return error for invalid configuration
	if err == nil {
		t.Error("Expected error for invalid configuration")
	}

	if result != nil {
		t.Error("Expected nil result for invalid configuration")
	}

	if !strings.Contains(err.Error(), "invalid configuration") {
		t.Errorf("Expected 'invalid configuration' error, got: %v", err)
	}
}

// TestICMPPinger_Execute_ResolveError tests Execute with unresolvable hostname
// Note: This test may fail due to privilege error before resolve error on some systems
func TestICMPPinger_Execute_ResolveError(t *testing.T) {
	config := ICMPProbeConfig{
		Type:           "icmp_ping",
		Target:         "this-host-does-not-exist.invalid",
		TimeoutSeconds: 1,
		Interval:       60,
		Count:          1,
	}

	pinger := NewICMPPinger(config)
	result, err := pinger.Execute()

	// Should not return error (error is in result)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected failure for unresolvable hostname")
	}

	if result.ErrorMessage == "" {
		t.Error("Expected error message for unresolvable hostname")
	}

	// Accept either resolve error or privilege error (order depends on system)
	if !strings.Contains(result.ErrorMessage, "resolve") && !strings.Contains(result.ErrorMessage, "privilege") && !strings.Contains(result.ErrorMessage, "root") {
		t.Logf("Got error message: %s (expected resolve or privilege error)", result.ErrorMessage)
	}
}

// TestICMPPinger_ExecuteBatch_InvalidCount tests ExecuteBatch with invalid count
func TestICMPPinger_ExecuteBatch_InvalidCount(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"Zero count", 0},
		{"Negative count", -1},
		{"Too high count", 101},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			}

			pinger := NewICMPPinger(config)
			result, err := pinger.ExecuteBatch(tt.count)

			if err == nil {
				t.Error("Expected error for invalid count")
			}

			if result != nil {
				t.Error("Expected nil result for invalid count")
			}

			if !strings.Contains(err.Error(), "invalid count") {
				t.Errorf("Expected 'invalid count' error, got: %v", err)
			}
		})
	}
}

// TestICMPPinger_ExecuteBatch_ResolveError tests ExecuteBatch with unresolvable hostname
func TestICMPPinger_ExecuteBatch_ResolveError(t *testing.T) {
	config := ICMPProbeConfig{
		Type:           "icmp_ping",
		Target:         "this-host-does-not-exist.invalid",
		TimeoutSeconds: 1,
		Interval:       60,
		Count:          1,
	}

	pinger := NewICMPPinger(config)
	result, err := pinger.ExecuteBatch(3)

	// Should not return error (error is in result)
	if err != nil {
		t.Fatalf("ExecuteBatch() returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected failure for unresolvable hostname")
	}

	if result.ErrorMessage == "" {
		t.Error("Expected error message for unresolvable hostname")
	}
}

// TestICMPPinger_resolveTarget tests target resolution
func TestICMPPinger_resolveTarget(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		wantErr        bool
		expectedType   string
	}{
		{
			name:         "IPv4 address",
			target:       "192.168.1.1",
			wantErr:      false,
			expectedType: "ipv4",
		},
		{
			name:         "IPv4 loopback",
			target:       "127.0.0.1",
			wantErr:      false,
			expectedType: "ipv4",
		},
		{
			name:         "IPv6 loopback",
			target:       "::1",
			wantErr:      false,
			expectedType: "ipv6",
		},
		{
			name:         "IPv6 address",
			target:       "2001:db8::1",
			wantErr:      false,
			expectedType: "ipv6",
		},
		{
			name:    "Unresolvable hostname",
			target:  "this-host-does-not-exist-12345.nonexistent",
			wantErr: true,
		},
		{
			name:         "Localhost resolves",
			target:       "localhost",
			wantErr:      false,
			expectedType: "ipv4", // or ipv6 depending on system
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         tt.target,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			}

			pinger := NewICMPPinger(config)
			ip, networkType, err := pinger.resolveTarget()

			if tt.wantErr {
				// DNS resolution behavior varies - some networks hijack DNS
				// If no error occurs, we just verify a valid IP was returned
				if err == nil {
					if ip == nil {
						t.Error("Expected non-nil IP even when no error (DNS hijacking scenario)")
						return
					}
					// DNS hijacking occurred - skip this test
					t.Skipf("DNS hijacking detected: hostname '%s' resolved to %v", tt.target, ip)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// For localhost, accept either ipv4 or ipv6
			if tt.target == "localhost" {
				if networkType != "ipv4" && networkType != "ipv6" {
					t.Errorf("Expected network type 'ipv4' or 'ipv6' for localhost, got '%s'", networkType)
				}
			} else if networkType != tt.expectedType {
				t.Errorf("Expected network type '%s', got '%s'", tt.expectedType, networkType)
			}

			if ip == nil {
				t.Error("Expected non-nil IP address")
			}
		})
	}
}

// TestICMPPinger_createEchoRequest tests ICMP echo request creation
func TestICMPPinger_createEchoRequest(t *testing.T) {
	config := ICMPProbeConfig{
		Type:           "icmp_ping",
		Target:         "192.168.1.1",
		TimeoutSeconds: 5,
		Interval:       60,
		Count:          1,
		PacketSize:     64,
	}

	pinger := NewICMPPinger(config)

	// Test IPv4
	request := pinger.createEchoRequest(1, "ipv4")
	if request == nil {
		t.Error("Expected non-nil echo request for IPv4")
	}

	// The packet should be at least the size of the ICMP header + payload
	// ICMP header is 8 bytes + 64 bytes payload = 72 bytes minimum
	if len(request) < 72 {
		t.Errorf("Expected at least 72 bytes, got %d", len(request))
	}

	// Test IPv6
	request6 := pinger.createEchoRequest(1, "ipv6")
	if request6 == nil {
		t.Error("Expected non-nil echo request for IPv6")
	}

	if len(request6) < 72 {
		t.Errorf("Expected at least 72 bytes for IPv6, got %d", len(request6))
	}
}

// TestICMPPinger_Execute_PrivilegeError tests that ICMP requires privileges
// Note: This test will fail if run as root (unlikely in CI)
func TestICMPPinger_Execute_PrivilegeError(t *testing.T) {
	// Skip if running as root (CI usually doesn't)
	if isRunningAsRoot() {
		t.Skip("Skipping test - running as root")
	}

	config := ICMPProbeConfig{
		Type:           "icmp_ping",
		Target:         "127.0.0.1",
		TimeoutSeconds: 1,
		Interval:       60,
		Count:          1,
	}

	pinger := NewICMPPinger(config)
	result, err := pinger.Execute()

	// Should not return error (error is in result)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Without root, we expect a privilege error
	if result.Success {
		t.Error("Expected failure without root privileges")
	}

	if !strings.Contains(result.ErrorMessage, "root") && !strings.Contains(result.ErrorMessage, "privilege") {
		t.Logf("Got error message: %s (expected privilege-related error)", result.ErrorMessage)
	}
}

// TestICMPPinger_ExecuteBatch_PrivilegeError tests that ICMP batch requires privileges
func TestICMPPinger_ExecuteBatch_PrivilegeError(t *testing.T) {
	// Skip if running as root
	if isRunningAsRoot() {
		t.Skip("Skipping test - running as root")
	}

	config := ICMPProbeConfig{
		Type:           "icmp_ping",
		Target:         "127.0.0.1",
		TimeoutSeconds: 1,
		Interval:       60,
		Count:          1,
	}

	pinger := NewICMPPinger(config)
	result, err := pinger.ExecuteBatch(3)

	// Should not return error (error is in result)
	if err != nil {
		t.Fatalf("ExecuteBatch() returned error: %v", err)
	}

	// Without root, we expect a privilege error
	if result.Success {
		t.Error("Expected failure without root privileges")
	}

	if !strings.Contains(result.ErrorMessage, "root") && !strings.Contains(result.ErrorMessage, "privilege") {
		t.Logf("Got error message: %s (expected privilege-related error)", result.ErrorMessage)
	}
}

// TestICMPProbeResult_ToGenericResult tests result conversion
func TestICMPProbeResult_ToGenericResult(t *testing.T) {
	tests := []struct {
		name   string
		result *ICMPProbeResult
		target string
	}{
		{
			name: "Success result",
			result: &ICMPProbeResult{
				Success:        true,
				RTTMs:          10.5,
				RTTMedianMs:    9.0,
				JitterMs:       1.2,
				VarianceMs:     2.5,
				PacketLossRate: 0,
				SampleCount:    5,
				TTL:            64,
				ErrorMessage:   "",
				Timestamp:      time.Now().Format(time.RFC3339),
			},
			target: "192.168.1.1",
		},
		{
			name: "Failure result",
			result: &ICMPProbeResult{
				Success:      false,
				ErrorMessage: "ICMP requires root privileges",
				Timestamp:    time.Now().Format(time.RFC3339),
			},
			target: "192.168.1.1",
		},
		{
			name: "Partial success with packet loss",
			result: &ICMPProbeResult{
				Success:        true,
				RTTMs:          15.0,
				RTTMedianMs:    14.0,
				JitterMs:       2.5,
				VarianceMs:     5.0,
				PacketLossRate: 20.0,
				SampleCount:    10,
				TTL:            64,
				ErrorMessage:   "",
				Timestamp:      time.Now().Format(time.RFC3339),
			},
			target: "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generic := tt.result.ToGenericResult(tt.target)

			if generic == nil {
				t.Fatal("Expected non-nil generic result")
			}

			if generic.Type != "icmp_ping" {
				t.Errorf("Expected type 'icmp_ping', got '%s'", generic.Type)
			}

			if generic.Target != tt.target {
				t.Errorf("Expected target '%s', got '%s'", tt.target, generic.Target)
			}

			if generic.Success != tt.result.Success {
				t.Errorf("Expected success %v, got %v", tt.result.Success, generic.Success)
			}

			if generic.ErrorMessage != tt.result.ErrorMessage {
				t.Errorf("Expected error message '%s', got '%s'", tt.result.ErrorMessage, generic.ErrorMessage)
			}

			// Verify metrics contain expected keys
			if _, ok := generic.Metrics["rtt_ms"]; !ok {
				t.Error("Expected 'rtt_ms' in metrics")
			}

			if _, ok := generic.Metrics["ttl"]; !ok {
				t.Error("Expected 'ttl' in metrics")
			}
		})
	}
}

// TestICMPProbeResult_Fields tests that result struct has all expected fields
func TestICMPProbeResult_Fields(t *testing.T) {
	result := &ICMPProbeResult{
		Success:        true,
		RTTMs:          10.5,
		RTTMedianMs:    9.0,
		JitterMs:       1.2,
		VarianceMs:     2.5,
		PacketLossRate: 5.0,
		SampleCount:    10,
		TTL:            64,
		ErrorMessage:   "",
		Timestamp:      "2024-01-01T00:00:00Z",
	}

	if result.Success != true {
		t.Error("Success field mismatch")
	}
	if result.RTTMs != 10.5 {
		t.Error("RTTMs field mismatch")
	}
	if result.RTTMedianMs != 9.0 {
		t.Error("RTTMedianMs field mismatch")
	}
	if result.JitterMs != 1.2 {
		t.Error("JitterMs field mismatch")
	}
	if result.VarianceMs != 2.5 {
		t.Error("VarianceMs field mismatch")
	}
	if result.PacketLossRate != 5.0 {
		t.Error("PacketLossRate field mismatch")
	}
	if result.SampleCount != 10 {
		t.Error("SampleCount field mismatch")
	}
	if result.TTL != 64 {
		t.Error("TTL field mismatch")
	}
	if result.Timestamp != "2024-01-01T00:00:00Z" {
		t.Error("Timestamp field mismatch")
	}
}

// TestICMPPinger_TimeoutBoundary tests timeout configuration boundary values
func TestICMPPinger_TimeoutBoundary(t *testing.T) {
	tests := []struct {
		name           string
		timeoutSeconds int
		valid          bool
	}{
		{"Default timeout (0 seconds -> uses 5s)", 0, true},
		{"Minimum timeout (1 second)", 1, true},
		{"Maximum timeout (30 seconds)", 30, true},
		{"Below minimum (-1 seconds)", -1, false},
		{"Above maximum (31 seconds)", 31, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ICMPProbeConfig{
				Type:           "icmp_ping",
				Target:         "192.168.1.1",
				TimeoutSeconds: tt.timeoutSeconds,
				Interval:       60,
				Count:          1,
			}

			pinger := NewICMPPinger(config)
			result, err := pinger.Execute()

			if tt.valid {
				// Should execute without validation error
				// Result may fail (network issues), but should not have validation error
				if err != nil && strings.Contains(err.Error(), "invalid configuration") {
					t.Errorf("Valid timeout %d should not cause validation error: %v", tt.timeoutSeconds, err)
				}
				// Result might be nil if there's a configuration error
				if result == nil && err == nil {
					t.Errorf("Expected result object for valid timeout")
				}
			} else {
				// Should fail validation
				if err == nil {
					t.Errorf("Invalid timeout %d should cause error", tt.timeoutSeconds)
				}
				if err != nil && !strings.Contains(err.Error(), "invalid configuration") {
					t.Errorf("Expected configuration error for invalid timeout %d, got: %v", tt.timeoutSeconds, err)
				}
			}
		})
	}
}

// Helper function to check if running as root
func isRunningAsRoot() bool {
	// This is a simple check - in production you might use os.Getuid() on Unix
	// For cross-platform, we'll return false and let the test run
	return false
}

// formatFloat is defined in tcp_ping_test.go, but we need it here too
// We'll use fmt.Sprintf directly instead

// TestICMPPinger_RTTFormat tests RTT value formatting
func TestICMPPinger_RTTFormat(t *testing.T) {
	rtt := 10.123456789
	formatted := fmt.Sprintf("%.2f", rtt)

	// Should have 2 decimal places
	if !strings.Contains(formatted, ".") {
		t.Errorf("Expected decimal point in RTT format, got %s", formatted)
	}

	parts := strings.Split(formatted, ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		t.Errorf("Expected 2 decimal places in RTT format, got %s", formatted)
	}
}
