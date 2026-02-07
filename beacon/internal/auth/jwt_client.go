package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"beacon/internal/logger"
)

const (
	// TokenExpirationBuffer is the time before token expiration to trigger refresh (2 minutes)
	TokenExpirationBuffer = 2 * time.Minute
	// DefaultTokenExpiration is the default token expiration time (15 minutes)
	DefaultTokenExpiration = 15 * time.Minute
)

// TokenResponse represents the beacon token response from Pulse
type TokenResponse struct {
	Data      TokenData `json:"data"`
	Message   string    `json:"message"`
	Timestamp string    `json:"timestamp"`
}

// TokenData contains the access token and metadata
type TokenData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // seconds
	NodeID      string `json:"node_id"`
}

// TokenRequest represents the token request
type TokenRequest struct {
	APIKey string `json:"api_key"`
}

// JWTClient manages JWT token lifecycle for beacon authentication
type JWTClient struct {
	serverURL   string
	apiKey      string
	httpClient  *http.Client

	// Token state
	mu           sync.RWMutex
	accessToken  string
	expiresAt    time.Time
	nodeID       string
	isRefreshing bool
}

// NewJWTClient creates a new JWT client
func NewJWTClient(serverURL, apiKey string, httpClient *http.Client) *JWTClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &JWTClient{
		serverURL:  serverURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// GetAccessToken returns a valid access token, refreshing if necessary
func (c *JWTClient) GetAccessToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	// Check if we have a valid token
	if c.accessToken != "" && time.Now().Add(TokenExpirationBuffer).Before(c.expiresAt) {
		token := c.accessToken
		c.mu.RUnlock()
		logger.WithField("component", "jwt_client").Debug("Using cached access token")
		return token, nil
	}
	c.mu.RUnlock()

	// Need to refresh token
	return c.refreshToken(ctx)
}

// refreshToken fetches a new access token from the server
func (c *JWTClient) refreshToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.accessToken != "" && time.Now().Add(TokenExpirationBuffer).Before(c.expiresAt) {
		return c.accessToken, nil
	}

	// Check if another goroutine is already refreshing
	if c.isRefreshing {
		// Wait for refresh to complete
		for c.isRefreshing {
			c.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			c.mu.Lock()
		}
		if c.accessToken != "" {
			return c.accessToken, nil
		}
		return "", fmt.Errorf("token refresh failed")
	}

	c.isRefreshing = true
	defer func() { c.isRefreshing = false }()

	logger.WithField("component", "jwt_client").Info("Fetching new access token")

	// Prepare request
	reqBody := TokenRequest{APIKey: c.apiKey}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	// Create HTTP request
	url := c.serverURL + "/api/v1/beacon/token"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Update token state
	c.accessToken = tokenResp.Data.AccessToken
	c.nodeID = tokenResp.Data.NodeID
	c.expiresAt = time.Now().Add(time.Duration(tokenResp.Data.ExpiresIn) * time.Second)

	logger.WithFields(map[string]interface{}{
		"component": "jwt_client",
		"node_id":   c.nodeID,
		"expires_in": tokenResp.Data.ExpiresIn,
	}).Info("Access token refreshed successfully")

	return c.accessToken, nil
}

// GetNodeID returns the node ID from the token
func (c *JWTClient) GetNodeID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nodeID
}

// InvalidateToken clears the cached token (e.g., after an auth error)
func (c *JWTClient) InvalidateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	logger.WithField("component", "jwt_client").Warn("Invalidating cached access token")
	c.accessToken = ""
	c.nodeID = ""
	c.expiresAt = time.Time{}
}
