package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// PulseClient handles communication with Pulse API
type PulseClient struct {
	baseURL    string
	authToken  string
	httpClient *http.Client
}

// RegisterNodeRequest represents registration request body
type RegisterNodeRequest struct {
	NodeName string   `json:"node_name"`
	IP       string   `json:"ip"`
	Region   string   `json:"region"`
	Tags     []string `json:"tags"`
}

// RegisterNodeData represents node data in registration response
type RegisterNodeData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IP        string    `json:"ip"`
	Region    string    `json:"region"`
	Tags      string    `json:"tags,omitempty"` // JSONB stored as string (matches Pulse API response)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RegisterNodeResponse represents registration response from Pulse
type RegisterNodeResponse struct {
	Data      RegisterNodeData `json:"data"`
	Message   string          `json:"message"`
	Timestamp string          `json:"timestamp"`
}

// APIError represents error response from Pulse API
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// NewPulseClient creates a new Pulse API client
func NewPulseClient(baseURL string, authToken string, httpClient *http.Client) *PulseClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &PulseClient{
		baseURL:    baseURL,
		authToken:  authToken,
		httpClient: httpClient,
	}
}

// RegisterNode sends registration request to Pulse with exponential backoff retry
func (c *PulseClient) RegisterNode(ctx context.Context, req *RegisterNodeRequest) (*RegisterNodeResponse, error) {
	const maxRetries = 3

	var lastError error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := exponentialBackoff(attempt - 1)
			select {
			case <-time.After(delay):
				// Continue to retry
			case <-ctx.Done():
				return nil, fmt.Errorf("registration cancelled during retry: %w", ctx.Err())
			}
		}

		resp, err := c.doRegisterNode(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastError = err

		// Check if error is retryable
		statusCode, apiErr := c.extractStatusCode(err)
		if !c.isRetryableError(statusCode, apiErr) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("registration failed after %d attempts: %w", maxRetries, lastError)
}

// doRegisterNode performs a single registration attempt
func (c *PulseClient) doRegisterNode(ctx context.Context, req *RegisterNodeRequest) (*RegisterNodeResponse, error) {
	// Build request URL
	url := c.baseURL + "/api/v1/nodes"

	// Marshal request body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	// Send request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	var responseBody struct {
		Data      RegisterNodeData `json:"data"`
		Message   string          `json:"message"`
		Timestamp string          `json:"timestamp"`
		Code     string          `json:"code,omitempty"`
		Details   any             `json:"details,omitempty"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response (status %d): %w", httpResp.StatusCode, err)
	}

	// Check response status
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, &APIError{
			Code:    responseBody.Code,
			Message: responseBody.Message,
			Details: responseBody.Details,
		}
	}

	return &RegisterNodeResponse{
		Data:      responseBody.Data,
		Message:   responseBody.Message,
		Timestamp: responseBody.Timestamp,
	}, nil
}

// exponentialBackoff returns delay duration for retry attempts: 1s, 2s, 4s
func exponentialBackoff(attempt int) time.Duration {
	// attempt 0 -> 1s, attempt 1 -> 2s, attempt 2 -> 4s
	return time.Duration(1<<uint(attempt)) * time.Second
}

// isRetryableError determines if an error should trigger a retry
func (c *PulseClient) isRetryableError(statusCode int, err error) bool {
	// Network errors (no status code)
	if statusCode == 0 {
		// Check if it's a context cancellation
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false // Don't retry on explicit cancellation
		}
		return true // Retry on network errors (timeout, connection refused, etc.)
	}

	// Server errors (5xx) are retryable
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	// Client errors (4xx) are NOT retryable (except rate limit 429)
	if statusCode == 429 {
		return true // Rate limit is retryable
	}

	return false // Other client errors don't retry
}

// extractStatusCode extracts HTTP status code from error
func (c *PulseClient) extractStatusCode(err error) (int, error) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// For API errors, infer status code from error code
		switch apiErr.Code {
		case "ERR_INVALID_REQUEST":
			return http.StatusBadRequest, apiErr
		case "ERR_UNAUTHORIZED":
			return http.StatusUnauthorized, apiErr
		case "ERR_NODE_EXISTS":
			return http.StatusConflict, apiErr
		case "ERR_NODE_NOT_FOUND":
			return http.StatusNotFound, apiErr
		case "ERR_INTERNAL_SERVER":
			return http.StatusInternalServerError, apiErr
		default:
			return http.StatusInternalServerError, apiErr
		}
	}
	return 0, err // Network error (no status code)
}

// Error implementation for APIError
func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return e.Message
}
