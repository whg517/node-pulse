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
	Message   string           `json:"message"`
	Timestamp string           `json:"timestamp"`
}

// ProbeConfig represents a server-assigned probe configuration.
type ProbeConfig struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	Target          string `json:"target"`
	Port            int    `json:"port"`
	IntervalSeconds int    `json:"interval_seconds"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	Count           int    `json:"count"`
	MaxHops         int    `json:"max_hops,omitempty"`
	PacketSize      int    `json:"packet_size,omitempty"`
}

// BeaconConfigData represents server-managed beacon configuration.
type BeaconConfigData struct {
	Probes          []ProbeConfig `json:"probes"`
	IntervalSeconds int           `json:"interval_seconds"`
	TimeoutSeconds  int           `json:"timeout_seconds"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Version         int           `json:"version"`
}

// BeaconConfigResponse represents the config response from Pulse.
type BeaconConfigResponse struct {
	Data      BeaconConfigData `json:"data"`
	Message   string           `json:"message"`
	Timestamp string           `json:"timestamp"`
}

// BeaconConfigAckRequest reports server config application status.
type BeaconConfigAckRequest struct {
	NodeID       string `json:"node_id"`
	Version      int    `json:"version"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// MTRHop represents a single MTR hop sent to Pulse.
type MTRHop struct {
	HopNumber  int     `json:"hop_number"`
	IP         string  `json:"ip"`
	Hostname   string  `json:"hostname,omitempty"`
	ASNumber   string  `json:"as_number,omitempty"`
	Sent       int     `json:"sent"`
	Received   int     `json:"received"`
	LossRate   float64 `json:"loss_rate"`
	LastRTTMs  float64 `json:"last_rtt_ms"`
	AvgRTTMs   float64 `json:"avg_rtt_ms"`
	BestRTTMs  float64 `json:"best_rtt_ms"`
	WorstRTTMs float64 `json:"worst_rtt_ms"`
	StdDevMs   float64 `json:"std_dev_ms"`
	Location   string  `json:"location,omitempty"`
}

// MTRResultRequest sends route-hop results to Pulse.
type MTRResultRequest struct {
	NodeID       string   `json:"node_id"`
	ProbeID      string   `json:"probe_id,omitempty"`
	Target       string   `json:"target"`
	TotalHops    int      `json:"total_hops"`
	Hops         []MTRHop `json:"hops"`
	CompletedAt  string   `json:"completed_at"`
	Success      bool     `json:"success"`
	ErrorMessage string   `json:"error_message,omitempty"`
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

// GetBeaconConfig fetches the latest server-assigned probe config for a beacon.
func (c *PulseClient) GetBeaconConfig(ctx context.Context, beaconID string) (*BeaconConfigResponse, error) {
	if strings.TrimSpace(beaconID) == "" {
		return nil, fmt.Errorf("beaconID is required")
	}

	url := c.baseURL + "/api/v1/beacons/" + beaconID + "/config"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	var responseBody struct {
		Data      BeaconConfigData `json:"data"`
		Message   string           `json:"message"`
		Timestamp string           `json:"timestamp"`
		Code      string           `json:"code,omitempty"`
		Details   any              `json:"details,omitempty"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response (status %d): %w", httpResp.StatusCode, err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, &APIError{
			Code:    responseBody.Code,
			Message: responseBody.Message,
			Details: responseBody.Details,
		}
	}

	return &BeaconConfigResponse{
		Data:      responseBody.Data,
		Message:   responseBody.Message,
		Timestamp: responseBody.Timestamp,
	}, nil
}

// AcknowledgeBeaconConfig reports that a server-assigned config was applied or failed.
func (c *PulseClient) AcknowledgeBeaconConfig(ctx context.Context, req *BeaconConfigAckRequest) error {
	if req == nil {
		return fmt.Errorf("ack request is required")
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/api/v1/beacon/config/ack"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	var responseBody struct {
		Message string `json:"message"`
		Code    string `json:"code,omitempty"`
		Details any    `json:"details,omitempty"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("failed to decode response (status %d): %w", httpResp.StatusCode, err)
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return &APIError{
			Code:    responseBody.Code,
			Message: responseBody.Message,
			Details: responseBody.Details,
		}
	}

	return nil
}

// SendMTRResult reports an MTR route-hop result to Pulse.
func (c *PulseClient) SendMTRResult(ctx context.Context, req *MTRResultRequest) error {
	if req == nil {
		return fmt.Errorf("mtr result request is required")
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + "/api/v1/beacon/mtr"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(reqBody)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if c.authToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	var responseBody struct {
		Message string `json:"message"`
		Code    string `json:"code,omitempty"`
		Details any    `json:"details,omitempty"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&responseBody); err != nil {
		return fmt.Errorf("failed to decode response (status %d): %w", httpResp.StatusCode, err)
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return &APIError{
			Code:    responseBody.Code,
			Message: responseBody.Message,
			Details: responseBody.Details,
		}
	}

	return nil
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
	defer func() { _ = httpResp.Body.Close() }()

	// Read response body
	var responseBody struct {
		Data      RegisterNodeData `json:"data"`
		Message   string           `json:"message"`
		Timestamp string           `json:"timestamp"`
		Code      string           `json:"code,omitempty"`
		Details   any              `json:"details,omitempty"`
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
