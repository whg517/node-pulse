package reporter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockJWTClientWithError is a mock JWT client that can be configured to fail auth
type mockJWTClientWithError struct {
	shouldFailAuth bool
	accessToken    string
	nodeID         string
}

func (m *mockJWTClientWithError) GetAccessToken(ctx context.Context) (string, error) {
	if m.shouldFailAuth {
		return "", errors.New("auth failed: invalid credentials")
	}
	return m.accessToken, nil
}

func (m *mockJWTClientWithError) GetNodeID() string {
	return m.nodeID
}

func (m *mockJWTClientWithError) InvalidateToken() {
	m.accessToken = ""
}

// TestSendHeartbeat_GetAccessTokenFails tests when GetAccessToken fails
func TestSendHeartbeat_GetAccessTokenFails(t *testing.T) {
	jwtClient := &mockJWTClientWithError{
		shouldFailAuth: true,
		accessToken:    "",
		nodeID:         "test-node",
	}

	apiClient := &PulseAPIClient{
		serverURL:  "http://localhost:16599",
		httpClient: &http.Client{Timeout: 1 * time.Second},
		jwtClient:  jwtClient,
	}

	data := &HeartbeatData{
		NodeID:   "test-node",
	}

	err := apiClient.SendHeartbeat(context.Background(), data)
	if err == nil {
		t.Error("Expected error when GetAccessToken fails")
	}
}

// TestSendHeartbeat_Unauthorized tests 401 response
func TestSendHeartbeat_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer server.Close()

	jwtClient := &mockJWTClientWithError{
		shouldFailAuth: false,
		accessToken:    "test-token",
		nodeID:         "test-node",
	}

	apiClient := &PulseAPIClient{
		serverURL:  server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		jwtClient:  jwtClient,
	}

	data := &HeartbeatData{
		NodeID:   "test-node",
	}

	err := apiClient.SendHeartbeat(context.Background(), data)
	if err == nil {
		t.Error("Expected error for 401 unauthorized response")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("Expected authentication error, got: %v", err)
	}
}

// TestSendHeartbeat_ServerError tests non-200/non-401 response
func TestSendHeartbeat_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	jwtClient := &mockJWTClientWithError{
		shouldFailAuth: false,
		accessToken:    "test-token",
		nodeID:         "test-node",
	}

	apiClient := &PulseAPIClient{
		serverURL:  server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		jwtClient:  jwtClient,
	}

	data := &HeartbeatData{
		NodeID:   "test-node",
	}

	err := apiClient.SendHeartbeat(context.Background(), data)
	if err == nil {
		t.Error("Expected error for 500 server error response")
	}
}

// TestSendHeartbeat_InvalidURL tests SendHeartbeat with invalid server URL (creates request error)
func TestSendHeartbeat_InvalidURL(t *testing.T) {
	jwtClient := &mockJWTClientWithError{
		shouldFailAuth: false,
		accessToken:    "test-token",
		nodeID:         "test-node",
	}

	apiClient := &PulseAPIClient{
		serverURL:  "http://invalid\x00host",
		httpClient: &http.Client{Timeout: 1 * time.Second},
		jwtClient:  jwtClient,
	}

	data := &HeartbeatData{
		NodeID:   "test-node",
	}

	err := apiClient.SendHeartbeat(context.Background(), data)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}
