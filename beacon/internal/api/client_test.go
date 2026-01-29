package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test server with custom handler
func createTestServer(handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

func TestPulseClient_RegisterNode_Success(t *testing.T) {
	// Test: Successful registration returns 201 with node data
	testNodeID := "550e8400-e29b-41d4-a716-446655440000"
	expectedResponse := RegisterNodeResponse{
		Data: RegisterNodeData{
			ID:        testNodeID,
			Name:      "美国东部-节点01",
			IP:        "192.168.1.100",
			Region:    "us-east",
			Tags:      `["production","east-coast"]`, // JSONB stored as string
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Message:   "节点注册成功",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	server := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/nodes", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req RegisterNodeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "美国东部-节点01", req.NodeName)
		assert.Equal(t, "192.168.1.100", req.IP)
		assert.Equal(t, "us-east", req.Region)
		assert.Equal(t, []string{"production", "east-coast"}, req.Tags)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(expectedResponse)
	})
	defer server.Close()

	client := NewPulseClient(server.URL, "", &http.Client{
		Timeout: 30 * time.Second,
	})

	req := &RegisterNodeRequest{
		NodeName: "美国东部-节点01",
		IP:       "192.168.1.100",
		Region:   "us-east",
		Tags:     []string{"production", "east-coast"},
	}

	resp, err := client.RegisterNode(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, testNodeID, resp.Data.ID)
	assert.Equal(t, "美国东部-节点01", resp.Data.Name)
	assert.Equal(t, "192.168.1.100", resp.Data.IP)
	assert.Equal(t, "us-east", resp.Data.Region)
}

func TestPulseClient_RegisterNode_Duplicate(t *testing.T) {
	// Test: Duplicate registration (same node_id) returns 200 with existing node
	testNodeID := "550e8400-e29b-41d4-a716-446655440000"

	expectedResponse := RegisterNodeResponse{
		Data: RegisterNodeData{
			ID:        testNodeID,
			Name:      "美国东部-节点01",
			IP:        "192.168.1.100",
			Region:    "us-east",
			Tags:      `["production"]`,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Message:   "节点已存在，已更新信息",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	server := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 instead of 201 for existing node
		json.NewEncoder(w).Encode(expectedResponse)
	})
	defer server.Close()

	client := NewPulseClient(server.URL, "", &http.Client{
		Timeout: 30 * time.Second,
	})

	req := &RegisterNodeRequest{
		NodeName: "美国东部-节点01",
		IP:       "192.168.1.100",
		Region:   "us-east",
		Tags:     []string{"production"},
	}

	resp, err := client.RegisterNode(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, testNodeID, resp.Data.ID)
	assert.Equal(t, "节点已存在，已更新信息", resp.Message)
}

func TestPulseClient_RegisterNode_RetryNetworkError(t *testing.T) {
	// Test: Network errors trigger exponential backoff retry (max 3 attempts)
	attemptCount := 0
	server := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		// First two attempts fail with server error
		if attemptCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"code":    "ERR_INTERNAL_SERVER",
				"message": "Internal server error",
			})
			return
		}

		// Third attempt succeeds
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(RegisterNodeResponse{
			Data: RegisterNodeData{
				ID:        "test-node-id",
				Name:      "测试节点",
				IP:        "192.168.1.1",
				Region:    "us-east",
				Tags:      "[]",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Message:   "节点注册成功",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	})
	defer server.Close()

	client := NewPulseClient(server.URL, "", &http.Client{
		Timeout: 30 * time.Second,
	})

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "192.168.1.1",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-node-id", resp.Data.ID)
	assert.Equal(t, 3, attemptCount, "Should retry 3 times before success")
}

func TestPulseClient_RegisterNode_RetryMaxExceeded(t *testing.T) {
	// Test: Max retries (3) exhausted - return final error
	attemptCount := 0
	server := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		// Always return server error
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "ERR_INTERNAL_SERVER",
			"message": "Internal server error",
		})
	})
	defer server.Close()

	client := NewPulseClient(server.URL, "", &http.Client{
		Timeout: 30 * time.Second,
	})

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "192.168.1.1",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 3, attemptCount, "Should attempt exactly 3 times")
}

func TestPulseClient_RegisterNode_ClientErrorNoRetry(t *testing.T) {
	// Test: Client errors (4xx) should NOT trigger retry
	attemptCount := 0
	server := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		// Return 400 bad request (client error)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"code":    "ERR_INVALID_REQUEST",
			"message": "Invalid request",
		})
	})
	defer server.Close()

	client := NewPulseClient(server.URL, "", &http.Client{
		Timeout: 30 * time.Second,
	})

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "invalid-ip",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 1, attemptCount, "Should NOT retry client errors")
}

func TestPulseClient_RegisterNode_ContextCancellation(t *testing.T) {
	// Test: Context cancellation stops registration
	server := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Simulate slow server
		w.WriteHeader(http.StatusCreated)
	})
	defer server.Close()

	client := NewPulseClient(server.URL, "", &http.Client{
		Timeout: 30 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "192.168.1.1",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPulseClient_RegisterNode_InvalidURL(t *testing.T) {
	// Test: Invalid server URL should fail
	client := NewPulseClient("://invalid-url", "", &http.Client{
		Timeout: 30 * time.Second,
	})

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "192.168.1.1",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPulseClient_RegisterNode_InvalidJSONResponse(t *testing.T) {
	// Test: Server returns invalid JSON should fail
	server := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("invalid json"))
	})
	defer server.Close()

	client := NewPulseClient(server.URL, "", &http.Client{
		Timeout: 30 * time.Second,
	})

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "192.168.1.1",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPulseClient_RegisterNode_Timeout(t *testing.T) {
	// Test: Request timeout should fail
	server := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(35 * time.Second) // Longer than client timeout
		w.WriteHeader(http.StatusCreated)
	})
	defer server.Close()

	client := NewPulseClient(server.URL, "", &http.Client{
		Timeout: 1 * time.Second, // Short timeout
	})

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "192.168.1.1",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPulseClient_RegisterNode_ConnectionRefused(t *testing.T) {
	// Test: Connection refused (server not running) should fail
	client := NewPulseClient("http://localhost:9999", "", &http.Client{
		Timeout: 1 * time.Second,
	})

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "192.168.1.1",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPulseClient_TLS_Enabled(t *testing.T) {
	// Test: Client should use TLS by default (https scheme)
	client := NewPulseClient("https://pulse.example.com", "", &http.Client{
		Timeout: 30 * time.Second,
	})

	// Can't fully test TLS without actual HTTPS server, but verify URL is parsed correctly
	assert.Equal(t, "https", strings.Split(client.baseURL, ":")[0])
}

func TestPulseClient_RegisterNode_AuthToken(t *testing.T) {
	// Test: Authorization header should be sent with token
	authToken := "test-auth-token-12345"

	server := createTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer "+authToken, r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(RegisterNodeResponse{
			Data: RegisterNodeData{
				ID:        "test-node-id",
				Name:      "测试节点",
				IP:        "192.168.1.1",
				Region:    "us-east",
				Tags:      "[]",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Message:   "节点注册成功",
			Timestamp: time.Now().Format(time.RFC3339),
		})
	})
	defer server.Close()

	client := NewPulseClient(server.URL, authToken, &http.Client{
		Timeout: 30 * time.Second,
	})

	req := &RegisterNodeRequest{
		NodeName: "测试节点",
		IP:       "192.168.1.1",
		Region:   "us-east",
	}

	resp, err := client.RegisterNode(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-node-id", resp.Data.ID)
}

func TestPulseClient_exponentialBackoff(t *testing.T) {
	// Test: Verify exponential backoff intervals: 1s, 2s, 4s
	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"First retry", 0, 1 * time.Second},
		{"Second retry", 1, 2 * time.Second},
		{"Third retry", 2, 4 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := exponentialBackoff(tt.attempt)
			assert.Equal(t, tt.expected, delay, "Exponential backoff mismatch on attempt %d", tt.attempt)
		})
	}
}

func TestPulseClient_isRetryableError(t *testing.T) {
	// Test: Determine which errors should trigger retry
	tests := []struct {
		name        string
		statusCode  int
		err         error
		shouldRetry bool
	}{
		{"Network error (timeout - context cancellation)", 0, context.DeadlineExceeded, false}, // Don't retry on context cancellation
		{"Network error (connection refused)", 0, errors.New("connection refused"), true},
		{"Network error (network unreachable)", 0, errors.New("network unreachable"), true},
		{"Server error 500", http.StatusInternalServerError, nil, true},
		{"Server error 503", http.StatusServiceUnavailable, nil, true},
		{"Server error 502", http.StatusBadGateway, nil, true},
		{"Client error 400", http.StatusBadRequest, nil, false},
		{"Client error 401", http.StatusUnauthorized, nil, false},
		{"Client error 403", http.StatusForbidden, nil, false},
		{"Client error 404", http.StatusNotFound, nil, false},
		{"Client error 409", http.StatusConflict, nil, false},
		{"Success 200", http.StatusOK, nil, false},
		{"Success 201", http.StatusCreated, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &PulseClient{}
			assert.Equal(t, tt.shouldRetry, client.isRetryableError(tt.statusCode, tt.err),
				"Retry decision mismatch for %s (status: %d, err: %v)", tt.name, tt.statusCode, tt.err)
		})
	}
}

