package models

import (
	"math"
	"testing"
	"time"
)

func TestNewMTRResult(t *testing.T) {
	hops := []MTRHop{
		{HopNumber: 1, IP: "192.168.1.1", Sent: 3, Received: 3, LossRate: 0},
		{HopNumber: 2, IP: "10.0.0.1", Sent: 3, Received: 2, LossRate: 33.33},
	}

	tests := []struct {
		name         string
		target       string
		hops         []MTRHop
		success      bool
		errorMessage string
		wantHops     int
	}{
		{
			name:         "successful result",
			target:       "8.8.8.8",
			hops:         hops,
			success:      true,
			errorMessage: "",
			wantHops:     2,
		},
		{
			name:         "failed result with error",
			target:       "invalid.example.com",
			hops:         nil,
			success:      false,
			errorMessage: "DNS resolution failed",
			wantHops:     0,
		},
		{
			name:         "empty hops",
			target:       "192.168.1.1",
			hops:         []MTRHop{},
			success:      true,
			errorMessage: "",
			wantHops:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewMTRResult(tt.target, tt.hops, tt.success, tt.errorMessage)

			if result.Target != tt.target {
				t.Errorf("Target = %v, want %v", result.Target, tt.target)
			}
			if result.TotalHops != tt.wantHops {
				t.Errorf("TotalHops = %v, want %v", result.TotalHops, tt.wantHops)
			}
			if result.Success != tt.success {
				t.Errorf("Success = %v, want %v", result.Success, tt.success)
			}
			if result.ErrorMessage != tt.errorMessage {
				t.Errorf("ErrorMessage = %v, want %v", result.ErrorMessage, tt.errorMessage)
			}
			if result.CompletedAt.IsZero() {
				t.Error("CompletedAt should not be zero")
			}
		})
	}
}

func TestNewMTRHop(t *testing.T) {
	tests := []struct {
		name           string
		hopNumber      int
		ip             string
		rtts           []float64
		sent           int
		wantReceived   int
		wantLossRate   float64
		wantAvgRTT     float64
		wantBestRTT    float64
		wantWorstRTT   float64
		wantStdDev     float64
	}{
		{
			name:         "all packets received",
			hopNumber:    1,
			ip:           "192.168.1.1",
			rtts:         []float64{10.5, 12.3, 11.1},
			sent:         3,
			wantReceived: 3,
			wantLossRate: 0,
			wantAvgRTT:   11.3,
			wantBestRTT:  10.5,
			wantWorstRTT: 12.3,
			wantStdDev:   0.74, // sqrt(variance) where variance = ((10.5-11.3)^2 + (12.3-11.3)^2 + (11.1-11.3)^2) / 3
		},
		{
			name:         "partial packet loss",
			hopNumber:    2,
			ip:           "10.0.0.1",
			rtts:         []float64{20.0, 25.0},
			sent:         3,
			wantReceived: 2,
			wantLossRate: 33.33,
			wantAvgRTT:   22.5,
			wantBestRTT:  20.0,
			wantWorstRTT: 25.0,
			wantStdDev:   2.5,
		},
		{
			name:         "no responses",
			hopNumber:    3,
			ip:           "192.168.100.1",
			rtts:         []float64{},
			sent:         3,
			wantReceived: 0,
			wantLossRate: 100.0,
			wantAvgRTT:   0,
			wantBestRTT:  0,
			wantWorstRTT: 0,
			wantStdDev:   0,
		},
		{
			name:         "single response",
			hopNumber:    4,
			ip:           "172.16.0.1",
			rtts:         []float64{15.5},
			sent:         3,
			wantReceived: 1,
			wantLossRate: 66.67,
			wantAvgRTT:   15.5,
			wantBestRTT:  15.5,
			wantWorstRTT: 15.5,
			wantStdDev:   0, // std dev requires at least 2 values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hop := NewMTRHop(tt.hopNumber, tt.ip, tt.rtts, tt.sent)

			if hop.HopNumber != tt.hopNumber {
				t.Errorf("HopNumber = %v, want %v", hop.HopNumber, tt.hopNumber)
			}
			if hop.IP != tt.ip {
				t.Errorf("IP = %v, want %v", hop.IP, tt.ip)
			}
			if hop.Sent != tt.sent {
				t.Errorf("Sent = %v, want %v", hop.Sent, tt.sent)
			}
			if hop.Received != tt.wantReceived {
				t.Errorf("Received = %v, want %v", hop.Received, tt.wantReceived)
			}
			if math.Abs(hop.LossRate-tt.wantLossRate) > 0.01 {
				t.Errorf("LossRate = %v, want %v", hop.LossRate, tt.wantLossRate)
			}
			if tt.wantReceived > 0 {
				if math.Abs(hop.AvgRTTMs-tt.wantAvgRTT) > 0.1 {
					t.Errorf("AvgRTTMs = %v, want %v", hop.AvgRTTMs, tt.wantAvgRTT)
				}
				if math.Abs(hop.BestRTTMs-tt.wantBestRTT) > 0.01 {
					t.Errorf("BestRTTMs = %v, want %v", hop.BestRTTMs, tt.wantBestRTT)
				}
				if math.Abs(hop.WorstRTTMs-tt.wantWorstRTT) > 0.01 {
					t.Errorf("WorstRTTMs = %v, want %v", hop.WorstRTTMs, tt.wantWorstRTT)
				}
				if math.Abs(hop.StdDevMs-tt.wantStdDev) > 0.1 {
					t.Errorf("StdDevMs = %v, want %v", hop.StdDevMs, tt.wantStdDev)
				}
			}
		})
	}
}

func TestCalculateLossRate(t *testing.T) {
	tests := []struct {
		name     string
		sent     int
		received int
		want     float64
	}{
		{"no loss", 10, 10, 0},
		{"50% loss", 10, 5, 50.0},
		{"33.33% loss", 3, 2, 33.33},
		{"100% loss", 10, 0, 100.0},
		{"zero sent", 0, 0, 0},
		{"single packet", 1, 1, 0},
		{"single packet lost", 1, 0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateLossRate(tt.sent, tt.received)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("calculateLossRate(%v, %v) = %v, want %v", tt.sent, tt.received, got, tt.want)
			}
		})
	}
}

func TestCalculateStdDev(t *testing.T) {
	tests := []struct {
		name  string
		values []float64
		mean  float64
		want  float64
	}{
		{"empty values", []float64{}, 0, 0},
		{"single value", []float64{10.0}, 10.0, 0},
		{"two values", []float64{10.0, 20.0}, 15.0, 5.0},
		{"three values", []float64{10.0, 15.0, 20.0}, 15.0, 4.08},
		{"identical values", []float64{10.0, 10.0, 10.0}, 10.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateStdDev(tt.values, tt.mean)
			if math.Abs(got-tt.want) > 0.1 {
				t.Errorf("calculateStdDev(%v, %v) = %v, want %v", tt.values, tt.mean, got, tt.want)
			}
		})
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		name string
		x    float64
		want float64
	}{
		{"zero", 0, 0},
		{"negative", -1, 0},
		{"one", 1, 1},
		{"four", 4, 2},
		{"nine", 9, 3},
		{"two", 2, 1.414},
		{"large", 10000, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sqrt(tt.x)
			tolerance := 0.001
			if tt.x > 0 && math.Abs(got-tt.want) > tolerance {
				t.Errorf("sqrt(%v) = %v, want %v", tt.x, got, tt.want)
			}
		})
	}
}

func TestRoundToTwoDecimals(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		want float64
	}{
		{"zero", 0, 0},
		{"exact", 10.5, 10.5},
		{"round up", 10.555, 10.56},
		{"round down", 10.554, 10.55},
		{"negative", -10.555, -10.55},
		{"large", 1234.567, 1234.57},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := roundToTwoDecimals(tt.f)
			if got != tt.want {
				t.Errorf("roundToTwoDecimals(%v) = %v, want %v", tt.f, got, tt.want)
			}
		})
	}
}

func TestMTRResult_JSONSerialization(t *testing.T) {
	// Test that MTRResult can be properly serialized/deserialized
	hops := []MTRHop{
		{
			HopNumber:  1,
			IP:         "192.168.1.1",
			Hostname:   "router.local",
			ASNumber:   "AS12345",
			Sent:       3,
			Received:   3,
			LossRate:   0,
			LastRTTMs:  10.5,
			AvgRTTMs:   11.3,
			BestRTTMs:  10.5,
			WorstRTTMs: 12.3,
			StdDevMs:   0.74,
			Location:   "New York, US",
		},
	}

	result := NewMTRResult("8.8.8.8", hops, true, "")

	// Verify JSON tags are correct
	if result.Hops[0].HopNumber != 1 {
		t.Errorf("Hop number not preserved")
	}
	if result.Hops[0].Location != "New York, US" {
		t.Errorf("Location not preserved")
	}
}

func TestMTRHop_OptionalFields(t *testing.T) {
	// Test that optional fields work correctly
	hop := NewMTRHop(1, "192.168.1.1", []float64{10.0}, 1)

	// Optional fields should be empty/zero
	if hop.Hostname != "" {
		t.Errorf("Hostname should be empty by default")
	}
	if hop.ASNumber != "" {
		t.Errorf("ASNumber should be empty by default")
	}
	if hop.Location != "" {
		t.Errorf("Location should be empty by default")
	}

	// Test setting optional fields
	hop.Hostname = "router.example.com"
	hop.ASNumber = "AS15169"
	hop.Location = "Mountain View, US"

	if hop.Hostname != "router.example.com" {
		t.Errorf("Hostname not set correctly")
	}
	if hop.ASNumber != "AS15169" {
		t.Errorf("ASNumber not set correctly")
	}
	if hop.Location != "Mountain View, US" {
		t.Errorf("Location not set correctly")
	}
}

func TestMTRResult_Timestamp(t *testing.T) {
	before := time.Now()
	result := NewMTRResult("test", []MTRHop{}, true, "")
	after := time.Now()

	if result.CompletedAt.Before(before) {
		t.Error("CompletedAt should not be before creation time")
	}
	if result.CompletedAt.After(after) {
		t.Error("CompletedAt should not be after creation time")
	}
}
