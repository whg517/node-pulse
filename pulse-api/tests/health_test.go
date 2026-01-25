package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kevin/node-pulse/pulse-api/internal/health"
)

// MockDB is a mock database for testing
type MockDB struct{}

func (m *MockDB) Check(ctx context.Context) error {
	return nil // Always healthy
}

// UnhealthyDB simulates an unhealthy database
type UnhealthyDB struct{}

func (u *UnhealthyDB) Check(ctx context.Context) error {
	return ctx.Err() // Simulate connection error
}

// TestHealthHandler_Healthy tests health check with healthy database
func TestHealthHandler_Healthy(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a mock database
	mockDB := &MockDB{}
	healthChecker := health.New(mockDB)

	// Create a test router
	router := gin.New()
	router.GET("/api/v1/health", healthChecker.Handler)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Unmarshal response to verify structure
	var healthResponse health.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &healthResponse); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify status
	if healthResponse.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", healthResponse.Status)
	}

	// Verify database check
	dbStatus, ok := healthResponse.Checks["database"]
	if !ok {
		t.Error("Expected 'database' check in response")
	}
	if dbStatus != "ok" {
		t.Errorf("Expected database status 'ok', got '%s'", dbStatus)
	}
}

// TestHealthHandler_Unhealthy tests health check with unhealthy database
func TestHealthHandler_Unhealthy(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create an unhealthy mock database
	unhealthyDB := &UnhealthyDB{}
	healthChecker := health.New(unhealthyDB)

	// Create a test router
	router := gin.New()
	router.GET("/api/v1/health", healthChecker.Handler)

	// Create a test request with a context that errors
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", "/api/v1/health", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	// Unmarshal response to verify structure
	var healthResponse health.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &healthResponse); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify status
	if healthResponse.Status != "unhealthy" {
		t.Errorf("Expected status 'unhealthy', got '%s'", healthResponse.Status)
	}

	// Verify database check contains error
	dbStatus, ok := healthResponse.Checks["database"]
	if !ok {
		t.Error("Expected 'database' check in response")
	}
	if dbStatus == "ok" {
		t.Error("Expected database status not 'ok' when unhealthy")
	}
}

// TestHealthHandler_NoDatabase tests health check with nil database
func TestHealthHandler_NoDatabase(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create health checker with nil database
	healthChecker := health.New(nil)

	// Create a test router
	router := gin.New()
	router.GET("/api/v1/health", healthChecker.Handler)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Unmarshal response to verify structure
	var healthResponse health.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &healthResponse); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify status is still healthy when database is disabled
	if healthResponse.Status != "healthy" {
		t.Errorf("Expected status 'healthy' when database is disabled, got '%s'", healthResponse.Status)
	}

	// Verify database check shows disabled
	dbStatus, ok := healthResponse.Checks["database"]
	if !ok {
		t.Error("Expected 'database' check in response")
	}
	if dbStatus != "disabled" {
		t.Errorf("Expected database status 'disabled', got '%s'", dbStatus)
	}
}
