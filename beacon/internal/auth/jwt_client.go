package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"beacon/internal/logger"
)

// Token configuration constants (matching Pulse server defaults)
const (
	// TokenExpirationBuffer is the time before token expiration to trigger refresh (2 minutes)
	TokenExpirationBuffer = 2 * time.Minute
	// DefaultTokenExpiration is the default token expiration time (15 minutes)
	DefaultTokenExpiration = 15 * time.Minute
	// DefaultRefreshTokenExpiration is the default refresh token expiration (7 days)
	DefaultRefreshTokenExpiration = 7 * 24 * time.Hour
	// DefaultHTTPTimeout is the default HTTP client timeout
	DefaultHTTPTimeout = 30 * time.Second
)

// Error definitions
var (
	// ErrInvalidCredentials is returned when API key or refresh token is rejected
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrNetworkError is returned when network communication fails
	ErrNetworkError = errors.New("network error")
	// ErrServerUnavailable is returned when server returns 5xx error
	ErrServerUnavailable = errors.New("server unavailable")
	// ErrRateLimited is returned when rate limit is exceeded
	ErrRateLimited = errors.New("rate limited")
	// ErrContextCanceled is returned when context is canceled during operation
	ErrContextCanceled = errors.New("context canceled")
	// ErrInvalidConfig is returned when client configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")
)

// TokenResponse represents the beacon token response from Pulse
// Matches Pulse's models.TokenResponse structure
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`         // seconds
	RefreshExpiresIn int    `json:"refresh_expires_in"` // seconds
	NodeID           string `json:"node_id,omitempty"`
}

// TokenRequest represents the token exchange request
type TokenRequest struct {
	APIKey       string `json:"api_key"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TokenState holds the current token state
type TokenState struct {
	AccessToken      string
	RefreshToken     string
	ExpiresAt        time.Time
	RefreshExpiresAt time.Time
	NodeID           string
}

// JWTClient manages JWT token lifecycle for beacon authentication
// Implements the TokenProvider interface used by reporter package
type JWTClient struct {
	serverURL  string
	apiKey     string
	httpClient *http.Client

	// Token state protected by mutex
	mu    sync.RWMutex
	state TokenState

	// Refresh state
	isRefreshing bool
	refreshCond  *sync.Cond
}

// NewJWTClient creates a new JWT client
// Returns error if serverURL or apiKey is empty
func NewJWTClient(serverURL, apiKey string, httpClient *http.Client) (*JWTClient, error) {
	// F3: Input validation
	if serverURL == "" {
		return nil, fmt.Errorf("%w: serverURL is required", ErrInvalidConfig)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("%w: apiKey is required", ErrInvalidConfig)
	}

	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: DefaultHTTPTimeout,
		}
	}

	client := &JWTClient{
		serverURL:  serverURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
	client.refreshCond = sync.NewCond(&client.mu)

	return client, nil
}

// GetAccessToken returns a valid access token, refreshing if necessary
// Implements TokenProvider interface
func (c *JWTClient) GetAccessToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	// Check if we have a valid access token
	if c.state.AccessToken != "" && time.Now().Add(TokenExpirationBuffer).Before(c.state.ExpiresAt) {
		token := c.state.AccessToken
		c.mu.RUnlock()
		logger.WithField("component", "jwt_client").Debug("Using cached access token")
		return token, nil
	}
	c.mu.RUnlock()

	// Need to refresh token
	return c.refreshToken(ctx)
}

// GetNodeID returns the node ID from the token
// Implements TokenProvider interface
func (c *JWTClient) GetNodeID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state.NodeID
}

// InvalidateToken clears the cached token (e.g., after an auth error)
// Implements TokenProvider interface
func (c *JWTClient) InvalidateToken() {
	c.mu.Lock()
	defer c.mu.Unlock()
	logger.WithField("component", "jwt_client").Warn("Invalidating cached tokens")
	c.state = TokenState{}
}

// RefreshToken forces a token refresh regardless of expiration
func (c *JWTClient) RefreshToken(ctx context.Context) error {
	_, err := c.refreshToken(ctx)
	return err
}

// refreshToken fetches a new access token from the server
func (c *JWTClient) refreshToken(ctx context.Context) (string, error) {
	c.mu.Lock()

	// Double-check after acquiring write lock
	if c.state.AccessToken != "" && time.Now().Add(TokenExpirationBuffer).Before(c.state.ExpiresAt) {
		token := c.state.AccessToken
		c.mu.Unlock()
		return token, nil
	}

	// Wait if another goroutine is already refreshing
	// F2: Add context cancellation check in wait loop
	for c.isRefreshing {
		// Check context before waiting
		if err := ctx.Err(); err != nil {
			c.mu.Unlock()
			return "", ErrContextCanceled
		}
		c.refreshCond.Wait()
		// After waking, check if token is now valid
		if c.state.AccessToken != "" && time.Now().Add(TokenExpirationBuffer).Before(c.state.ExpiresAt) {
			token := c.state.AccessToken
			c.mu.Unlock()
			return token, nil
		}
	}

	// Check context before proceeding
	if err := ctx.Err(); err != nil {
		c.mu.Unlock()
		return "", ErrContextCanceled
	}

	c.isRefreshing = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.isRefreshing = false
		c.refreshCond.Broadcast()
		c.mu.Unlock()
	}()

	logger.WithField("component", "jwt_client").Info("Fetching new access token")

	// Determine if we should use refresh token or API key
	var reqBody TokenRequest
	var endpoint string
	var usingRefreshToken bool

	c.mu.RLock()
	hasRefreshToken := c.state.RefreshToken != "" && time.Now().Before(c.state.RefreshExpiresAt)
	if hasRefreshToken {
		reqBody = TokenRequest{RefreshToken: c.state.RefreshToken}
		endpoint = "/api/v1/auth/refresh"
		usingRefreshToken = true
	} else {
		reqBody = TokenRequest{APIKey: c.apiKey}
		endpoint = "/api/v1/beacon/token"
		usingRefreshToken = false
	}
	c.mu.RUnlock()

	// F6: Log when using refresh token vs API key
	if usingRefreshToken {
		logger.WithField("component", "jwt_client").Debug("Using refresh token for authentication")
	} else {
		logger.WithField("component", "jwt_client").Debug("Using API key for authentication")
	}

	// Make the token request
	resp, err := c.makeTokenRequest(ctx, endpoint, reqBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Handle response
	if err := c.handleTokenResponse(resp, usingRefreshToken); err != nil {
		// F6: Log fallback when refresh token fails
		if usingRefreshToken && errors.Is(err, ErrInvalidCredentials) {
			logger.WithField("component", "jwt_client").Warn("Refresh token failed, will retry with API key on next attempt")
		}
		return "", err
	}

	c.mu.RLock()
	token := c.state.AccessToken
	c.mu.RUnlock()

	return token, nil
}

// makeTokenRequest creates and sends a token request
func (c *JWTClient) makeTokenRequest(ctx context.Context, endpoint string, reqBody TokenRequest) (*http.Response, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	url := c.serverURL + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.WithFields(map[string]any{
			"component": "jwt_client",
			"error":     err.Error(),
		}).Error("Token request failed")
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}

	return resp, nil
}

// handleTokenResponse processes the token response from the server
func (c *JWTClient) handleTokenResponse(resp *http.Response, isRefresh bool) error {
	// Check for rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("%w: too many requests", ErrRateLimited)
	}

	// Handle authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
		// If refresh failed, clear tokens and require re-auth with API key
		if isRefresh {
			c.InvalidateToken()
		}
		return fmt.Errorf("%w: authentication rejected", ErrInvalidCredentials)
	}

	// Handle server errors
	if resp.StatusCode >= 500 {
		body, _ := io.ReadAll(resp.Body)
		logger.WithFields(map[string]any{
			"component": "jwt_client",
			"status":    resp.StatusCode,
			"response":  string(body),
		}).Error("Server error during token request")
		return fmt.Errorf("%w: status %d", ErrServerUnavailable, resp.StatusCode)
	}

	// Check for other non-OK responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
			return fmt.Errorf("token request failed: %s", errResp.Message)
		}
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	// Update token state
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.state.AccessToken = tokenResp.AccessToken
	c.state.ExpiresAt = now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Update refresh token if provided (rotation)
	if tokenResp.RefreshToken != "" {
		c.state.RefreshToken = tokenResp.RefreshToken
		c.state.RefreshExpiresAt = now.Add(time.Duration(tokenResp.RefreshExpiresIn) * time.Second)
	}

	// Update node ID if provided
	if tokenResp.NodeID != "" {
		c.state.NodeID = tokenResp.NodeID
	}

	logger.WithFields(map[string]any{
		"component":          "jwt_client",
		"node_id":            c.state.NodeID,
		"expires_in":         tokenResp.ExpiresIn,
		"refresh_expires_in": tokenResp.RefreshExpiresIn,
	}).Info("Access token refreshed successfully")

	return nil
}

// GetTokenState returns the current token state (for testing/debugging)
func (c *JWTClient) GetTokenState() TokenState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// IsTokenValid checks if the current access token is valid
func (c *JWTClient) IsTokenValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state.AccessToken != "" && time.Now().Add(TokenExpirationBuffer).Before(c.state.ExpiresAt)
}

// HasRefreshToken checks if a valid refresh token exists
func (c *JWTClient) HasRefreshToken() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state.RefreshToken != "" && time.Now().Before(c.state.RefreshExpiresAt)
}
