package reporter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// MockPulseServer is a mock Pulse API server for testing
type MockPulseServer struct {
	server      *httptest.Server
	heartbeatCount int
	mu          sync.Mutex
	responseStatusCode int
	delay       time.Duration
}

// NewMockPulseServer creates a new mock Pulse API server
func NewMockPulseServer() *MockPulseServer {
	mock := &MockPulseServer{
		responseStatusCode: http.StatusOK,
		delay:             0,
	}

	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.handleRequest(w, r)
	}))

	return mock
}

// handleRequest handles incoming HTTP requests
func (m *MockPulseServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Apply delay if configured
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	// Only handle POST to /api/v1/beacon/heartbeat
	if r.URL.Path != "/api/v1/beacon/heartbeat" || r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Decode request body
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate heartbeat data
	nodeID, ok := data["node_id"].(string)
	if !ok || nodeID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid node_id"})
		return
	}

	// Validate metric ranges
	latencyMs, _ := data["latency_ms"].(float64)
	if latencyMs < 0 || latencyMs > 60000 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid metrics"})
		return
	}

	// Increment heartbeat counter
	m.mu.Lock()
	m.heartbeatCount++
	m.mu.Unlock()

	// Return configured response status
	w.WriteHeader(m.responseStatusCode)
	if m.responseStatusCode == http.StatusOK {
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Heartbeat received",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
	}
}

// GetURL returns the mock server URL
func (m *MockPulseServer) GetURL() string {
	return m.server.URL
}

// GetHeartbeatCount returns the number of heartbeats received
func (m *MockPulseServer) GetHeartbeatCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.heartbeatCount
}

// SetResponseStatusCode sets the response status code for heartbeats
func (m *MockPulseServer) SetResponseStatusCode(code int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseStatusCode = code
}

// SetDelay sets the response delay (for testing timeout scenarios)
func (m *MockPulseServer) SetDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delay = delay
}

// ResetHeartbeatCount resets the heartbeat counter
func (m *MockPulseServer) ResetHeartbeatCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.heartbeatCount = 0
}

// Close closes the mock server
func (m *MockPulseServer) Close() {
	m.server.Close()
}
