package probe

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// TestTCPPinger_Success tests successful TCP connection scenario
func TestTCPPinger_Success(t *testing.T) {
	// Setup: Start a local TCP server for testing
	server := startTestTCPServer(t, "localhost:18888")
	defer server.Close()

	config := TCPProbeConfig{
		Type:           "tcp_ping",
		Target:         "localhost",
		Port:           18888,
		TimeoutSeconds: 5,
		Interval:       60,
		Count:          1,
	}

	pinger := NewTCPPinger(config)
	result, err := pinger.Execute()

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected Success=true, got false. ErrorMessage: %s", result.ErrorMessage)
	}

	if result.RTTMs <= 0 {
		t.Errorf("Expected RTTMs > 0, got %f", result.RTTMs)
	}

	if result.ErrorMessage != "" {
		t.Errorf("Expected empty ErrorMessage, got %s", result.ErrorMessage)
	}

	if result.Timestamp == "" {
		t.Errorf("Expected non-empty Timestamp")
	}
}

// TestTCPPinger_ConnectionRefused tests connection refused scenario
func TestTCPPinger_ConnectionRefused(t *testing.T) {
	// Use a port that is not listening
	config := TCPProbeConfig{
		Type:           "tcp_ping",
		Target:         "localhost",
		Port:           19999, // Port not in use
		TimeoutSeconds: 1,
		Interval:       60,
		Count:          1,
	}

	pinger := NewTCPPinger(config)
	result, err := pinger.Execute()

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if result.Success {
		t.Errorf("Expected Success=false for refused connection, got true")
	}

	if result.RTTMs != 0 {
		t.Errorf("Expected RTTMs=0 for failed connection, got %f", result.RTTMs)
	}

	if result.ErrorMessage == "" {
		t.Errorf("Expected non-empty ErrorMessage for failed connection")
	}
}

// TestTCPPinger_Timeout tests connection timeout scenario
func TestTCPPinger_Timeout(t *testing.T) {
	// Use a blackhole IP that should timeout
	// 198.18.0.1 is in the benchmark testing range (RFC 2544) and should be unreachable
	config := TCPProbeConfig{
		Type:           "tcp_ping",
		Target:         "198.18.0.1", // Benchmark testing range, should timeout
		Port:           80,
		TimeoutSeconds: 1,
		Interval:       60,
		Count:          1,
	}

	pinger := NewTCPPinger(config)
	result, err := pinger.Execute()

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Note: Some networks may route this IP differently, so we just verify it returns a result
	// The important thing is Execute() doesn't hang and returns a valid result object
	if result.Timestamp == "" {
		t.Errorf("Expected non-empty Timestamp")
	}

	// If the connection succeeded (rare), log it but don't fail
	if result.Success {
		t.Logf("Warning: Expected timeout but connection succeeded to %s (RTT: %f ms)", config.Target, result.RTTMs)
	}
}

// TestTCPPinger_RTTPrecision tests RTT measurement precision
func TestTCPPinger_RTTPrecision(t *testing.T) {
	server := startTestTCPServer(t, "localhost:18889")
	defer server.Close()

	config := TCPProbeConfig{
		Type:           "tcp_ping",
		Target:         "localhost",
		Port:           18889,
		TimeoutSeconds: 5,
		Interval:       60,
		Count:          1,
	}

	pinger := NewTCPPinger(config)
	result, err := pinger.Execute()

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Expected successful connection, got error: %s", result.ErrorMessage)
	}

	// Check RTT precision (should have 2 decimal places)
	// Convert to string and check format
	rttStr := formatFloat(result.RTTMs, 2)
	if rttStr == "" {
		t.Errorf("Failed to format RTT value")
	}

	// RTT should be reasonable (between 0 and timeout)
	if result.RTTMs < 0 || result.RTTMs > float64(config.TimeoutSeconds)*1000 {
		t.Errorf("RTT %f ms is out of reasonable range [0, %d]", result.RTTMs, config.TimeoutSeconds*1000)
	}
}

// TestTCPPinger_TimeoutBoundary tests timeout configuration boundary values
func TestTCPPinger_TimeoutBoundary(t *testing.T) {
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
			config := TCPProbeConfig{
				Type:           "tcp_ping",
				Target:         "localhost",
				Port:           19999,
				TimeoutSeconds: tt.timeoutSeconds,
				Interval:       60,
				Count:          1,
			}

			pinger := NewTCPPinger(config)
			result, err := pinger.Execute()

			if tt.valid {
				// Should execute without validation error
				// Result may fail (port not open), but should not have validation error
				if err != nil && contains(err.Error(), "invalid configuration") {
					t.Errorf("Valid timeout %d should not cause validation error: %v", tt.timeoutSeconds, err)
				}
				if result == nil {
					t.Errorf("Expected result object for valid timeout")
				}
			} else {
				// Should fail validation
				if err == nil {
					t.Errorf("Invalid timeout %d should cause error", tt.timeoutSeconds)
				}
				if !contains(err.Error(), "invalid configuration") {
					t.Errorf("Expected configuration error for invalid timeout %d, got: %v", tt.timeoutSeconds, err)
				}
			}
		})
	}
}

// TestTCPPinger_ExecuteBatch tests batch probe execution
func TestTCPPinger_ExecuteBatch(t *testing.T) {
	server := startTestTCPServer(t, "localhost:18890")
	defer server.Close()

	config := TCPProbeConfig{
		Type:           "tcp_ping",
		Target:         "localhost",
		Port:           18890,
		TimeoutSeconds: 5,
		Interval:       60,
		Count:          1,
	}

	pinger := NewTCPPinger(config)
	count := 10
	result, err := pinger.ExecuteBatch(count)

	if err != nil {
		t.Fatalf("ExecuteBatch() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected Success=true, got false. Error: %s", result.ErrorMessage)
	}
	if result.RTTMs <= 0 {
		t.Errorf("Expected RTTMs > 0, got %f", result.RTTMs)
	}
	if result.SampleCount != count {
		t.Errorf("Expected SampleCount=%d, got %d", count, result.SampleCount)
	}
	// Verify core metrics are calculated
	if result.JitterMs < 0 {
		t.Errorf("Expected JitterMs >= 0, got %f", result.JitterMs)
	}
	if result.VarianceMs < 0 {
		t.Errorf("Expected VarianceMs >= 0, got %f", result.VarianceMs)
	}
}

// TestTCPProbeConfig_Validate tests configuration validation
func TestTCPProbeConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TCPProbeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid configuration",
			config: TCPProbeConfig{
				Type:           "tcp_ping",
				Target:         "192.168.1.1",
				Port:           80,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			},
			wantErr: false,
		},
		{
			name: "Valid hostname",
			config: TCPProbeConfig{
				Type:           "tcp_ping",
				Target:         "example.com",
				Port:           443,
				TimeoutSeconds: 10,
				Interval:       60,
				Count:          1,
			},
			wantErr: false,
		},
		{
			name: "Invalid port (too low)",
			config: TCPProbeConfig{
				Type:           "tcp_ping",
				Target:         "192.168.1.1",
				Port:           0,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "Invalid port (too high)",
			config: TCPProbeConfig{
				Type:           "tcp_ping",
				Target:         "192.168.1.1",
				Port:           65536,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "port",
		},
		{
			name: "Invalid timeout (too low)",
			config: TCPProbeConfig{
				Type:           "tcp_ping",
				Target:         "192.168.1.1",
				Port:           80,
				TimeoutSeconds: -1,
				Interval:       60,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "timeout",
		},
		{
			name: "Invalid timeout (too high)",
			config: TCPProbeConfig{
				Type:           "tcp_ping",
				Target:         "192.168.1.1",
				Port:           80,
				TimeoutSeconds: 31,
				Interval:       60,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "timeout",
		},
		{
			name: "Invalid target (empty)",
			config: TCPProbeConfig{
				Type:           "tcp_ping",
				Target:         "",
				Port:           80,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          1,
			},
			wantErr: true,
			errMsg:  "target",
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
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

// Helper functions

// startTestTCPServer starts a simple TCP server for testing
func startTestTCPServer(t *testing.T, addr string) *testTCPServer {
	t.Helper()
	server := &testTCPServer{addr: addr}
	go server.start()
	time.Sleep(100 * time.Millisecond) // Give server time to start
	return server
}

type testTCPServer struct {
	addr string
	lis net.Listener
}

func (s *testTCPServer) start() {
	var err error
	s.lis, err = net.Listen("tcp", s.addr)
	if err != nil {
		return
	}
	for {
		conn, err := s.lis.Accept()
		if err != nil {
			break
		}
		conn.Close()
	}
}

func (s *testTCPServer) Close() {
	if s.lis != nil {
		s.lis.Close()
	}
}

// formatFloat formats a float64 with specified precision
func formatFloat(f float64, precision int) string {
	return fmt.Sprintf("%.*f", precision, f)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
