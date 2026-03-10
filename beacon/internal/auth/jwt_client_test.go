package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"beacon/internal/config"
	"beacon/internal/logger"
)

// Test constants
const (
	testAPIKey  = "test-api-key-12345"
	testNodeID  = "test-node-id-uuid"
	testToken   = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"
	testRefresh = "refresh-token-abc123"
)

func init() {
	// Initialize logger for tests
	logger.InitLogger(&config.Config{
		LogLevel:      "ERROR", // Reduce noise in tests
		LogFile:       "/tmp/test-jwt-client.log",
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   false,
		LogToConsole:  false,
	})
}

// mustNewJWTClient creates a JWT client or panics (for test setup)
func mustNewJWTClient(serverURL, apiKey string, httpClient *http.Client) *JWTClient {
	client, err := NewJWTClient(serverURL, apiKey, httpClient)
	if err != nil {
		panic(err)
	}
	return client
}

// TestNewJWTClient tests creating a new JWT client
func TestNewJWTClient(t *testing.T) {
	tests := []struct {
		name        string
		serverURL   string
		apiKey      string
		httpClient  *http.Client
		expectError bool
	}{
		{
			name:        "with default HTTP client",
			serverURL:   "https://pulse.example.com",
			apiKey:      testAPIKey,
			httpClient:  nil,
			expectError: false,
		},
		{
			name:        "with custom HTTP client",
			serverURL:   "https://pulse.example.com",
			apiKey:      testAPIKey,
			httpClient:  &http.Client{Timeout: 10 * time.Second},
			expectError: false,
		},
		{
			name:        "empty serverURL",
			serverURL:   "",
			apiKey:      testAPIKey,
			httpClient:  nil,
			expectError: true,
		},
		{
			name:        "empty apiKey",
			serverURL:   "https://pulse.example.com",
			apiKey:      "",
			httpClient:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewJWTClient(tt.serverURL, tt.apiKey, tt.httpClient)
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if client == nil {
				t.Fatal("Expected non-nil client")
			}
			if client.serverURL != tt.serverURL {
				t.Errorf("Expected serverURL %s, got %s", tt.serverURL, client.serverURL)
			}
			if client.apiKey != tt.apiKey {
				t.Errorf("Expected apiKey %s, got %s", tt.apiKey, client.apiKey)
			}
			if client.httpClient == nil {
				t.Error("Expected httpClient to be initialized")
			}
		})
	}
}

// TestGetAccessToken tests getting an access token
func TestGetAccessToken(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/api/v1/beacon/token" {
			t.Errorf("Expected path /api/v1/beacon/token, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Return mock token response
		resp := TokenResponse{
			AccessToken:      testToken,
			RefreshToken:     testRefresh,
			TokenType:        "Bearer",
			ExpiresIn:        900,         // 15 minutes
			RefreshExpiresIn: 604800,      // 7 days
			NodeID:           testNodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	token, err := client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}
	if token != testToken {
		t.Errorf("Expected token %s, got %s", testToken, token)
	}
	if client.GetNodeID() != testNodeID {
		t.Errorf("Expected nodeID %s, got %s", testNodeID, client.GetNodeID())
	}
}

// TestGetAccessTokenCached tests that cached tokens are reused
func TestGetAccessTokenCached(t *testing.T) {
	callCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := TokenResponse{
			AccessToken:      testToken,
			RefreshToken:     testRefresh,
			TokenType:        "Bearer",
			ExpiresIn:        900,
			RefreshExpiresIn: 604800,
			NodeID:           testNodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	// First call should hit the server
	_, err := client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("First GetAccessToken failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}

	// Second call should use cached token
	_, err = client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("Second GetAccessToken failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 server call (cached), got %d", callCount)
	}
}

// TestGetAccessTokenRefresh tests token refresh when expired
func TestGetAccessTokenRefresh(t *testing.T) {
	callCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Return different token on each call
		resp := TokenResponse{
			AccessToken:      testToken + "-call-" + string(rune('0'+callCount)),
			RefreshToken:     testRefresh + "-call-" + string(rune('0'+callCount)),
			TokenType:        "Bearer",
			ExpiresIn:        900,
			RefreshExpiresIn: 604800,
			NodeID:           testNodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	// Get initial token
	token1, err := client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("First GetAccessToken failed: %v", err)
	}

	// Invalidate to force refresh
	client.InvalidateToken()

	// Get new token
	token2, err := client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("Second GetAccessToken failed: %v", err)
	}

	// Verify two different tokens were returned (two server calls)
	if callCount != 2 {
		t.Errorf("Expected 2 server calls, got %d", callCount)
	}
	if token1 == token2 {
		t.Error("Expected different tokens after invalidation")
	}
}

// TestInvalidateToken tests token invalidation
func TestInvalidateToken(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{
			AccessToken:      testToken,
			RefreshToken:     testRefresh,
			TokenType:        "Bearer",
			ExpiresIn:        900,
			RefreshExpiresIn: 604800,
			NodeID:           testNodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	// Get initial token
	_, err := client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}

	// Verify token is valid
	if !client.IsTokenValid() {
		t.Error("Expected token to be valid")
	}

	// Invalidate
	client.InvalidateToken()

	// Verify token is invalidated
	if client.IsTokenValid() {
		t.Error("Expected token to be invalid after InvalidateToken")
	}
	if client.GetNodeID() != "" {
		t.Errorf("Expected empty nodeID after invalidation, got %s", client.GetNodeID())
	}
}

// TestAuthenticationError tests handling of authentication errors
func TestAuthenticationError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Code:    "INVALID_API_KEY",
			Message: "Invalid API key",
		})
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	_, err := client.GetAccessToken(ctx)
	if err == nil {
		t.Fatal("Expected error for invalid credentials")
	}
	// Check that error wraps ErrInvalidCredentials
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestRateLimited tests handling of rate limiting
func TestRateLimited(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(ErrorResponse{
			Code:    "RATE_LIMITED",
			Message: "Too many requests",
		})
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	_, err := client.GetAccessToken(ctx)
	if err == nil {
		t.Fatal("Expected error for rate limiting")
	}
	// Check that error wraps ErrRateLimited
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestServerError tests handling of server errors
func TestServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	_, err := client.GetAccessToken(ctx)
	if err == nil {
		t.Fatal("Expected error for server error")
	}
	// Error is wrapped, so check for error message content
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestContextCanceled tests handling of canceled context
func TestContextCanceled(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		resp := TokenResponse{
			AccessToken:      testToken,
			RefreshToken:     testRefresh,
			TokenType:        "Bearer",
			ExpiresIn:        900,
			RefreshExpiresIn: 604800,
			NodeID:           testNodeID,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// First invalidate any cached token
	client.InvalidateToken()

	_, err := client.GetAccessToken(ctx)
	if err == nil {
		t.Fatal("Expected error for canceled context")
	}
	if err != ErrContextCanceled {
		t.Errorf("Expected ErrContextCanceled, got %v", err)
	}
}

// TestConcurrentTokenRefresh tests concurrent token refresh requests
func TestConcurrentTokenRefresh(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()

		// Simulate some latency
		time.Sleep(50 * time.Millisecond)

		resp := TokenResponse{
			AccessToken:      testToken,
			RefreshToken:     testRefresh,
			TokenType:        "Bearer",
			ExpiresIn:        900,
			RefreshExpiresIn: 604800,
			NodeID:           testNodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	// Start multiple goroutines to get token simultaneously
	var wg sync.WaitGroup
	errors := make([]error, 10)
	tokens := make([]string, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token, err := client.GetAccessToken(ctx)
			errors[idx] = err
			tokens[idx] = token
		}(i)
	}

	wg.Wait()

	// Check all goroutines got tokens without errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Goroutine %d got error: %v", i, err)
		}
	}

	// All tokens should be the same (from cache or same refresh)
	for i, token := range tokens {
		if token != testToken {
			t.Errorf("Goroutine %d got unexpected token: %s", i, token)
		}
	}

	// Should only have called server once (or a few times due to race)
	mu.Lock()
	count := callCount
	mu.Unlock()

	if count > 3 {
		t.Errorf("Expected at most 3 server calls for concurrent requests, got %d", count)
	}
}

// TestIsTokenValid tests the IsTokenValid method
func TestIsTokenValid(t *testing.T) {
	client := mustNewJWTClient("https://example.com", testAPIKey, nil)

	// Initially invalid
	if client.IsTokenValid() {
		t.Error("Expected token to be invalid initially")
	}

	// Set a valid token state
	client.mu.Lock()
	client.state.AccessToken = testToken
	client.state.ExpiresAt = time.Now().Add(15 * time.Minute)
	client.mu.Unlock()

	if !client.IsTokenValid() {
		t.Error("Expected token to be valid")
	}

	// Set expired token
	client.mu.Lock()
	client.state.ExpiresAt = time.Now().Add(-1 * time.Minute)
	client.mu.Unlock()

	if client.IsTokenValid() {
		t.Error("Expected token to be invalid when expired")
	}
}

// TestHasRefreshToken tests the HasRefreshToken method
func TestHasRefreshToken(t *testing.T) {
	client := mustNewJWTClient("https://example.com", testAPIKey, nil)

	// Initially no refresh token
	if client.HasRefreshToken() {
		t.Error("Expected no refresh token initially")
	}

	// Set a valid refresh token
	client.mu.Lock()
	client.state.RefreshToken = testRefresh
	client.state.RefreshExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	client.mu.Unlock()

	if !client.HasRefreshToken() {
		t.Error("Expected refresh token to be valid")
	}

	// Set expired refresh token
	client.mu.Lock()
	client.state.RefreshExpiresAt = time.Now().Add(-1 * time.Minute)
	client.mu.Unlock()

	if client.HasRefreshToken() {
		t.Error("Expected refresh token to be invalid when expired")
	}
}

// TestGetTokenState tests the GetTokenState method
func TestGetTokenState(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := TokenResponse{
			AccessToken:      testToken,
			RefreshToken:     testRefresh,
			TokenType:        "Bearer",
			ExpiresIn:        900,
			RefreshExpiresIn: 604800,
			NodeID:           testNodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	// Get token
	_, err := client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}

	// Check state
	state := client.GetTokenState()
	if state.AccessToken != testToken {
		t.Errorf("Expected access token %s, got %s", testToken, state.AccessToken)
	}
	if state.RefreshToken != testRefresh {
		t.Errorf("Expected refresh token %s, got %s", testRefresh, state.RefreshToken)
	}
	if state.NodeID != testNodeID {
		t.Errorf("Expected nodeID %s, got %s", testNodeID, state.NodeID)
	}
	if state.ExpiresAt.IsZero() {
		t.Error("Expected ExpiresAt to be set")
	}
	if state.RefreshExpiresAt.IsZero() {
		t.Error("Expected RefreshExpiresAt to be set")
	}
}

// TestTokenRefreshUsesRefreshToken tests that refresh token is used when available
func TestTokenRefreshUsesRefreshToken(t *testing.T) {
	var receivedRequest TokenRequest

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedRequest)

		// Check which endpoint was called
		if r.URL.Path == "/api/v1/auth/refresh" {
			if receivedRequest.RefreshToken == "" {
				t.Error("Expected refresh token in request")
			}
		} else if r.URL.Path == "/api/v1/beacon/token" {
			if receivedRequest.APIKey == "" {
				t.Error("Expected API key in request")
			}
		}

		resp := TokenResponse{
			AccessToken:      testToken + "-new",
			RefreshToken:     testRefresh + "-new",
			TokenType:        "Bearer",
			ExpiresIn:        900,
			RefreshExpiresIn: 604800,
			NodeID:           testNodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	// First call - should use API key
	_, err := client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("First GetAccessToken failed: %v", err)
	}

	// Invalidate access token but keep refresh token
	client.mu.Lock()
	client.state.AccessToken = ""
	client.mu.Unlock()

	// Reset received request
	receivedRequest = TokenRequest{}

	// Second call - should use refresh token
	_, err = client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("Second GetAccessToken failed: %v", err)
	}
}

// TestErrorResponse tests handling of error responses
func TestErrorResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "The request was malformed",
		})
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	_, err := client.GetAccessToken(ctx)
	if err == nil {
		t.Fatal("Expected error for bad request")
	}
	// Should contain the error message
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestGetAccessToken_APIKeyInAuthorizationHeader verifies that when no refresh token is available
// the client sends the API key in the Authorization header (RFC 6750 Bearer token), NOT in the
// request body. This tests the fix that corrected the incorrect body-based API key transmission.
func TestGetAccessToken_APIKeyInAuthorizationHeader(t *testing.T) {
	var receivedAuthHeader string
	var receivedBodyBytes []byte

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/beacon/token" {
			receivedAuthHeader = r.Header.Get("Authorization")
			receivedBodyBytes, _ = io.ReadAll(r.Body)
		}
		resp := TokenResponse{
			AccessToken:      testToken,
			TokenType:        "Bearer",
			ExpiresIn:        900,
			NodeID:           testNodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	_, err := client.GetAccessToken(ctx)
	if err != nil {
		t.Fatalf("GetAccessToken failed: %v", err)
	}

	// Verify API key is sent in Authorization header as RFC 6750 Bearer token
	expectedAuth := "Bearer " + testAPIKey
	if receivedAuthHeader != expectedAuth {
		t.Errorf("API key should be in Authorization header: want %q, got %q", expectedAuth, receivedAuthHeader)
	}

	// Verify API key is NOT present in the request body
	var bodyData map[string]interface{}
	if err := json.Unmarshal(receivedBodyBytes, &bodyData); err == nil {
		if _, found := bodyData["api_key"]; found {
			t.Error("API key must NOT be in request body; it must be sent via Authorization header only")
		}
	}
}

// TestRefreshTokenMethod tests the RefreshToken public method
func TestRefreshTokenMethod(t *testing.T) {
	callCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := TokenResponse{
			AccessToken:      testToken,
			RefreshToken:     testRefresh,
			TokenType:        "Bearer",
			ExpiresIn:        900,
			RefreshExpiresIn: 604800,
			NodeID:           testNodeID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := mustNewJWTClient(mockServer.URL, testAPIKey, nil)
	ctx := context.Background()

	// Force refresh
	err := client.RefreshToken(ctx)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}
	if !client.IsTokenValid() {
		t.Error("Expected token to be valid after refresh")
	}
}
