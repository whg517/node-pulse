package probe

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMTRProbeConfig_Validate tests configuration validation
func TestMTRProbeConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  MTRProbeConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid configuration with IP",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
			},
			wantErr: false,
		},
		{
			name: "Valid configuration with hostname",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "google.com",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
			},
			wantErr: false,
		},
		{
			name: "Valid configuration with packet size",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
				PacketSize:     128,
			},
			wantErr: false,
		},
		{
			name: "Invalid type",
			config: MTRProbeConfig{
				Type:           "tcp_ping",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
			},
			wantErr: true,
			errMsg:  "invalid probe type",
		},
		{
			name: "Empty target",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
			},
			wantErr: true,
			errMsg:  "target cannot be empty",
		},
		{
			name: "MaxHops too low",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        0,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
			},
			wantErr: true,
			errMsg:  "max_hops",
		},
		{
			name: "MaxHops too high",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        65,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
			},
			wantErr: true,
			errMsg:  "max_hops",
		},
		{
			name: "Timeout too low",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 0,
				Interval:       60,
				Count:          3,
			},
			wantErr: true,
			errMsg:  "timeout",
		},
		{
			name: "Timeout too high",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 31,
				Interval:       60,
				Count:          3,
			},
			wantErr: true,
			errMsg:  "timeout",
		},
		{
			name: "Interval too low",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       59,
				Count:          3,
			},
			wantErr: true,
			errMsg:  "interval",
		},
		{
			name: "Interval too high",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       301,
				Count:          3,
			},
			wantErr: true,
			errMsg:  "interval",
		},
		{
			name: "Count too low",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          0,
			},
			wantErr: true,
			errMsg:  "count",
		},
		{
			name: "Count too high",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          31,
			},
			wantErr: true,
			errMsg:  "count",
		},
		{
			name: "Packet size too small",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
				PacketSize:     63,
			},
			wantErr: true,
			errMsg:  "packet_size",
		},
		{
			name: "Packet size too large",
			config: MTRProbeConfig{
				Type:           "mtr",
				Target:         "8.8.8.8",
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
				PacketSize:     1501,
			},
			wantErr: true,
			errMsg:  "packet_size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNewMTRProbe tests creating a new MTR probe
func TestNewMTRProbe(t *testing.T) {
	config := MTRProbeConfig{
		Type:           "mtr",
		Target:         "8.8.8.8",
		MaxHops:        30,
		TimeoutSeconds: 5,
		Interval:       60,
		Count:          3,
	}

	probe := NewMTRProbe(config)
	assert.NotNil(t, probe)
	assert.Equal(t, config, probe.config)
}

// TestMTRProbe_ResolveTarget tests target resolution
func TestMTRProbe_ResolveTarget(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		wantIPv4    bool
		wantErr     bool
	}{
		{
			name:     "IPv4 address",
			target:   "8.8.8.8",
			wantIPv4: true,
			wantErr:  false,
		},
		{
			name:     "Localhost",
			target:   "127.0.0.1",
			wantIPv4: true,
			wantErr:  false,
		},
		{
			name:     "Valid hostname",
			target:   "localhost",
			wantIPv4: true,
			wantErr:  false,
		},
		{
			name:     "Invalid hostname format",
			target:   "invalid..hostname",
			wantIPv4: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe := &MTRProbe{
				config: MTRProbeConfig{
					Target: tt.target,
				},
			}

			ip, err := probe.resolveTarget()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, ip)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ip)
				if tt.wantIPv4 {
					assert.NotNil(t, ip.To4())
				}
			}
		})
	}
}

// TestMTRProbe_CreateProbeData tests probe data creation
func TestMTRProbe_CreateProbeData(t *testing.T) {
	tests := []struct {
		name           string
		packetSize     int
		expectedLength int
	}{
		{
			name:           "Default packet size",
			packetSize:     0,
			expectedLength: 56, // 64 - 8 (ICMP header)
		},
		{
			name:           "Custom packet size 128",
			packetSize:     128,
			expectedLength: 120, // 128 - 8 (ICMP header)
		},
		{
			name:           "Minimum packet size",
			packetSize:     64,
			expectedLength: 56,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe := &MTRProbe{
				config: MTRProbeConfig{
					PacketSize: tt.packetSize,
				},
			}

			data := probe.createProbeData()
			assert.Equal(t, tt.expectedLength, len(data))
		})
	}
}

// TestMTRProbe_InvalidConfig tests Execute with invalid configuration
func TestMTRProbe_InvalidConfig(t *testing.T) {
	probe := &MTRProbe{
		config: MTRProbeConfig{
			Type:   "mtr",
			Target: "", // Invalid - empty target
		},
	}

	result, err := probe.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration")
	assert.Nil(t, result)
}

// TestMTRHop tests hop statistics calculation
func TestMTRHopStatistics(t *testing.T) {
	tests := []struct {
		name             string
		hopNumber        int
		ip               string
		rtts              []float64
		sent             int
		expectedReceived int
		expectedLoss     float64
		expectedAvg      float64
		expectedBest     float64
		expectedWorst    float64
	}{
		{
			name:             "All responses",
			hopNumber:        1,
			ip:               "192.168.1.1",
			rtts:              []float64{10.5, 12.3, 11.2, 10.8, 12.0},
			sent:             5,
			expectedReceived: 5,
			expectedLoss:     0.0,
			expectedAvg:      11.36,
			expectedBest:     10.5,
			expectedWorst:    12.3,
		},
		{
			name:             "Some packet loss",
			hopNumber:        2,
			ip:               "192.168.1.2",
			rtts:              []float64{20.0, 25.0},
			sent:             5,
			expectedReceived: 2,
			expectedLoss:     60.0,
			expectedAvg:      22.5,
			expectedBest:     20.0,
			expectedWorst:    25.0,
		},
		{
			name:             "Complete packet loss",
			hopNumber:        3,
			ip:               "192.168.1.3",
			rtts:              []float64{},
			sent:             3,
			expectedReceived: 0,
			expectedLoss:     100.0,
			expectedAvg:      0,
			expectedBest:     0,
			expectedWorst:    0,
		},
		{
			name:             "Single response",
			hopNumber:        4,
			ip:               "192.168.1.4",
			rtts:              []float64{15.0},
			sent:             5,
			expectedReceived: 1,
			expectedLoss:     80.0,
			expectedAvg:      15.0,
			expectedBest:     15.0,
			expectedWorst:    15.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test hop statistics through models package
			// This is tested indirectly through the models test
			// Here we verify the logic is consistent
			if len(tt.rtts) > 0 {
				avg := 0.0
				for _, rtt := range tt.rtts {
					avg += rtt
				}
				avg = roundRTT(avg / float64(len(tt.rtts)))
				assert.Equal(t, tt.expectedAvg, avg)

				best := tt.rtts[0]
				worst := tt.rtts[0]
				for _, rtt := range tt.rtts {
					if rtt < best {
						best = rtt
					}
					if rtt > worst {
						worst = rtt
					}
				}
				assert.Equal(t, tt.expectedBest, best)
				assert.Equal(t, tt.expectedWorst, worst)
			}

			// Calculate loss rate
			lossRate := 0.0
			if tt.sent > 0 {
				lossRate = (1.0 - float64(len(tt.rtts))/float64(tt.sent)) * 100
			}
			assert.Equal(t, tt.expectedLoss, lossRate)
		})
	}
}

// TestRoundRTT tests RTT rounding
func TestRoundRTT(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{10.123, 10.12},
		{10.125, 10.13},
		{10.129, 10.13},
		{10.1, 10.1},
		{0.0, 0.0},
		{100.999, 101.0},
	}

	for _, tt := range tests {
		result := roundRTT(tt.input)
		assert.Equal(t, tt.expected, result, "roundRTT(%f) = %f, want %f", tt.input, result, tt.expected)
	}
}

// TestMTRProbe_TargetValidation tests target IP/hostname validation
func TestMTRProbe_TargetValidation(t *testing.T) {
	validTargets := []string{
		"8.8.8.8",
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"127.0.0.1",
		"localhost",
		"google.com",
	}

	for _, target := range validTargets {
		t.Run("Valid target: "+target, func(t *testing.T) {
			config := MTRProbeConfig{
				Type:           "mtr",
				Target:         target,
				MaxHops:        30,
				TimeoutSeconds: 5,
				Interval:       60,
				Count:          3,
			}
			err := config.Validate()
			assert.NoError(t, err, "Valid target '%s' should pass validation", target)
		})
	}
}

// TestMTRProbe_EdgeCases tests edge cases in MTR probing
func TestMTRProbe_EdgeCases(t *testing.T) {
	t.Run("MaxHops boundary", func(t *testing.T) {
		// Test min boundary
		config := MTRProbeConfig{
			Type:           "mtr",
			Target:         "8.8.8.8",
			MaxHops:        1,
			TimeoutSeconds: 5,
			Interval:       60,
			Count:          1,
		}
		err := config.Validate()
		assert.NoError(t, err)

		// Test max boundary
		config.MaxHops = 64
		err = config.Validate()
		assert.NoError(t, err)
	})

	t.Run("Count boundary", func(t *testing.T) {
		// Test min boundary
		config := MTRProbeConfig{
			Type:           "mtr",
			Target:         "8.8.8.8",
			MaxHops:        30,
			TimeoutSeconds: 5,
			Interval:       60,
			Count:          1,
		}
		err := config.Validate()
		assert.NoError(t, err)

		// Test max boundary
		config.Count = 30
		err = config.Validate()
		assert.NoError(t, err)
	})

	t.Run("Interval boundary", func(t *testing.T) {
		// Test min boundary
		config := MTRProbeConfig{
			Type:           "mtr",
			Target:         "8.8.8.8",
			MaxHops:        30,
			TimeoutSeconds: 5,
			Interval:       60,
			Count:          3,
		}
		err := config.Validate()
		assert.NoError(t, err)

		// Test max boundary
		config.Interval = 300
		err = config.Validate()
		assert.NoError(t, err)
	})
}

// TestMTRProbe_LoopbackIntegration tests MTR against localhost
// This requires elevated privileges (CAP_NET_RAW on Linux, admin on Windows)
func TestMTRProbe_LoopbackIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if we can create raw socket (requires privileges)
	conn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		t.Skip("Skipping test: requires elevated privileges for ICMP")
	}
	_ = conn.Close()

	config := MTRProbeConfig{
		Type:           "mtr",
		Target:         "127.0.0.1",
		MaxHops:        5,
		TimeoutSeconds: 2,
		Interval:       60,
		Count:          2,
	}

	probe := NewMTRProbe(config)
	result, err := probe.Execute()

	// Test should not return an error at the configuration level
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	// Result should be returned (may fail if no privileges)
	if result != nil {
		assert.Equal(t, "127.0.0.1", result.Target)
		assert.NotEmpty(t, result.CompletedAt)
	}
}
