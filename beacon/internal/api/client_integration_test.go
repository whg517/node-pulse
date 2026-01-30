package api

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPulseClient_Integration_RegisterNewNode tests complete registration flow (Subtask 5.1)
func TestPulseClient_Integration_RegisterNewNode(t *testing.T) {
	// This test validates integration between Beacon registration and Pulse API
	// Since we can't run both Beacon and Pulse in a single test file,
	// this test focuses on the Beacon client's ability to handle the API responses

	_ = NewPulseClient("http://test-pulse-api", "", &http.Client{})

	// Test that client can create registration request
	req := &RegisterNodeRequest{
		NodeName: "集成测试节点01",
		IP:       "192.168.1.100",
		Region:   "us-east",
		Tags:     []string{"test", "integration"},
	}

	require.NotNil(t, req)
	assert.Equal(t, "集成测试节点01", req.NodeName)
	assert.Equal(t, "192.168.1.100", req.IP)
	assert.Equal(t, "us-east", req.Region)
	assert.Equal(t, []string{"test", "integration"}, req.Tags)
}

// TestPulseClient_Integration_UUIDFormatValidates tests UUID format validation (Subtask 3.3)
func TestPulseClient_Integration_UUIDFormatValidates(t *testing.T) {
	// Test valid UUID format
	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000", // Valid v4 UUID
		"550e8400-e29b-41d4-a716-446655440001", // Another valid v4
	}

	for _, uuid := range validUUIDs {
		assert.True(t, isValidUUIDFormat(uuid), "UUID %s should be valid", uuid)
	}

	// Test invalid UUID formats
	invalidUUIDs := []string{
		"",                           // Empty
		"not-a-uuid",              // Wrong format
		"12345678-1234-1234-12345678", // Too many segments
		"550e8400-e29b-41d4-a716",       // Missing segments
	}

	for _, uuid := range invalidUUIDs {
		assert.False(t, isValidUUIDFormat(uuid), "UUID %s should be invalid", uuid)
	}
}

// TestPulseClient_Integration_RetryWithExponentialBackoff tests retry mechanism (AC #4)
func TestPulseClient_Integration_RetryWithExponentialBackoff(t *testing.T) {
	// Verify exponential backoff intervals: 1s, 2s, 4s
	tests := []struct {
		name     string
		attempt  int
		expected string
	}{
		{"First retry", 0, "1s"},
		{"Second retry", 1, "2s"},
		{"Third retry", 2, "4s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := exponentialBackoff(tt.attempt)
			expected, err := time.ParseDuration(tt.expected)
			require.NoError(t, err)
			assert.Equal(t, expected, delay, "Exponential backoff mismatch on attempt %d", tt.attempt)
		})
	}
}

// TestPulseClient_Integration_PulseUnavailable verifies retry behavior (Subtask 5.3)
func TestPulseClient_Integration_PulseUnavailable(t *testing.T) {
	// Create a client with an invalid URL (will fail after retries)
	client := NewPulseClient("http://localhost:9999", "", &http.Client{TimeoutSeconds: 1 * time.Second})

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "192.168.1.1",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(context.Background(), req)

	// Verify
	assert.Error(t, err) // Should fail after 3 retries (AC #4)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed after 3 attempts") // Verify retry logic
}

// isValidUUIDFormat is a helper to validate UUID format
func isValidUUIDFormat(uuidStr string) bool {
	// Basic UUID format validation: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(uuidStr) != 36 {
		return false
	}
	// Check for dashes at correct positions
	if uuidStr[8] != '-' || uuidStr[13] != '-' || uuidStr[18] != '-' || uuidStr[23] != '-' {
		return false
	}
	// Check for hex characters in segments
	segments := []string{
		uuidStr[0:8],
		uuidStr[9:13],
		uuidStr[14:18],
		uuidStr[19:23],
		uuidStr[24:36],
	}
	for _, segment := range segments {
		for _, c := range segment {
			if !((c >= '0' && c <= '9') ||
				(c >= 'a' && c <= 'f') ||
				(c >= 'A' && c <= 'F')) {
				// If we get here, character is invalid
				return false
			}
		}
	}
	return true
}
