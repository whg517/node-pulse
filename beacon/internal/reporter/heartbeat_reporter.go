// Package reporter provides heartbeat reporting functionality for Beacon.
// It aggregates probe metrics and reports them to the Pulse server via HTTP/HTTPS.
package reporter

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"beacon/internal/logger"
	"beacon/internal/models"
)

const (
	// ReportInterval is the interval between heartbeat reports (60 seconds)
	ReportInterval = 60 * time.Second
	// MaxRetries is the maximum number of retry attempts for failed reports
	MaxRetries = 3
	// MaxUploadLatency is the maximum acceptable upload latency (NFR-PERF-001)
	MaxUploadLatency = 5 * time.Second
)

// HeartbeatData represents the heartbeat data structure for reporting to Pulse
type HeartbeatData struct {
	NodeID         string  `json:"node_id"`           // UUID from Pulse registration
	LatencyMs      float64 `json:"latency_ms"`        // RTT mean in milliseconds
	PacketLossRate float64 `json:"packet_loss_rate"`  // Packet loss rate (0-100%)
	JitterMs       float64 `json:"jitter_ms"`         // Delay jitter in milliseconds
	Timestamp      string  `json:"timestamp"`         // ISO 8601 timestamp
}

// PulseAPIClient handles HTTP/HTTPS communication with Pulse server
type PulseAPIClient struct {
	serverURL  string
	httpClient *http.Client
	timeout    time.Duration
}

// ProbeScheduler interface for accessing probe results
type ProbeScheduler interface {
	GetLatestResults() ([]*models.TCPProbeResult, []*models.UDPProbeResult)
}

// HeartbeatReporter manages scheduled heartbeat reporting to Pulse
type HeartbeatReporter struct {
	apiClient *PulseAPIClient
	nodeID    string
	scheduler ProbeScheduler
	ticker    *time.Ticker
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.Mutex
	reporting bool
}

// NewHeartbeatData creates a new HeartbeatData with current timestamp
func NewHeartbeatData(nodeID string, latencyMs, packetLossRate, jitterMs float64) *HeartbeatData {
	return &HeartbeatData{
		NodeID:         nodeID,
		LatencyMs:      latencyMs,
		PacketLossRate: packetLossRate,
		JitterMs:       jitterMs,
		Timestamp:      time.Now().Format(time.RFC3339),
	}
}

// NewPulseAPIClient creates a new Pulse API client with TLS support
func NewPulseAPIClient(serverURL string, timeout time.Duration) *PulseAPIClient {
	// Create HTTP client with TLS config
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12, // Enforce TLS 1.2 or higher (NFR-SEC-001)
			InsecureSkipVerify: false,            // Require certificate validation
			// In production, you may want to add:
			// RootCAs: customCertPool,
		},
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return &PulseAPIClient{
		serverURL:  serverURL,
		httpClient: httpClient,
		timeout:    timeout,
	}
}

// NewHeartbeatReporter creates a new HeartbeatReporter with probe scheduler integration
func NewHeartbeatReporter(apiClient *PulseAPIClient, nodeID string, scheduler ProbeScheduler) *HeartbeatReporter {
	return &HeartbeatReporter{
		apiClient: apiClient,
		nodeID:    nodeID,
		scheduler: scheduler,
		reporting: false,
	}
}

// SendHeartbeat sends heartbeat data to Pulse server with latency measurement
func (c *PulseAPIClient) SendHeartbeat(data *HeartbeatData) error {
	// Measure upload latency (NFR-PERF-001)
	startTime := time.Now()

	// Serialize heartbeat data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat data: %w", err)
	}

	// Create HTTP POST request
	url := c.serverURL + "/api/v1/beacon/heartbeat"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request with timeout
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Measure elapsed time
	elapsed := time.Since(startTime)

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Read error response body for debugging
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pulse API returned error %d: %s", resp.StatusCode, string(body))
	}

	// Validate upload latency
	if elapsed > MaxUploadLatency {
		logger.WithFields(map[string]interface{}{"component": "reporter", "latency": elapsed.String(), "threshold": MaxUploadLatency.String()}).Warn("Heartbeat upload latency exceeds requirement")
	}

	logger.WithFields(map[string]interface{}{"component": "reporter", "latency": elapsed.String()}).Info("Heartbeat reported successfully")
	return nil
}

// AggregateMetrics aggregates metrics from TCP and UDP probe results
func (r *HeartbeatReporter) AggregateMetrics(tcpResults []*models.TCPProbeResult, udpResults []*models.UDPProbeResult) *HeartbeatData {
	var totalLatency, totalPacketLoss, totalJitter float64
	count := 0

	// Aggregate TCP probe results (only successful probes)
	for _, result := range tcpResults {
		if result.Success {
			totalLatency += result.RTTMs
			totalPacketLoss += result.PacketLossRate
			totalJitter += result.JitterMs
			count++
		}
	}

	// Aggregate UDP probe results (only successful probes)
	for _, result := range udpResults {
		if result.Success {
			totalLatency += result.RTTMs
			totalPacketLoss += result.PacketLossRate
			totalJitter += result.JitterMs
			count++
		}
	}

	// If no successful probe results, report 100% packet loss with 0 latency/jitter
	// This semantically indicates "all probes failed" which is different from
	// "low latency/jitter" but is the only valid JSON representation
	if count == 0 {
		return &HeartbeatData{
			NodeID:         r.nodeID,
			LatencyMs:      0,   // 0 with 100% loss means "no successful probes"
			PacketLossRate: 100, // All probes failed
			JitterMs:       0,   // 0 with 100% loss means "no successful probes"
			Timestamp:      time.Now().Format(time.RFC3339),
		}
	}

	// Calculate averages
	return &HeartbeatData{
		NodeID:         r.nodeID,
		LatencyMs:      totalLatency / float64(count),
		PacketLossRate: totalPacketLoss / float64(count),
		JitterMs:       totalJitter / float64(count),
		Timestamp:      time.Now().Format(time.RFC3339),
	}
}

// StartReporting starts the scheduled heartbeat reporting with context support
func (r *HeartbeatReporter) StartReporting(ctx context.Context) {
	r.mu.Lock()
	if r.reporting {
		logger.WithField("component", "reporter").Warn("Heartbeat reporter already running")
		r.mu.Unlock()
		return
	}

	r.reporting = true
	r.ticker = time.NewTicker(ReportInterval)

	// Create cancellable context
	ctx, r.cancel = context.WithCancel(ctx)
	r.mu.Unlock()

	logger.WithFields(map[string]interface{}{"component": "reporter", "interval": ReportInterval.String()}).Info("Starting heartbeat reporter")

	// Start reporting goroutine with proper synchronization
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		// Report immediately on start (synchronized)
		r.reportWithRetry()

		// Start scheduled reporting
		for {
			select {
			case <-r.ticker.C:
				r.reportWithRetry()
			case <-ctx.Done():
				r.ticker.Stop()
				logger.WithField("component", "reporter").Info("Heartbeat reporter stopped")
				return
			}
		}
	}()
}

// StopReporting gracefully stops the heartbeat reporter
func (r *HeartbeatReporter) StopReporting() {
	r.mu.Lock()
	if !r.reporting {
		r.mu.Unlock()
		return
	}

	if r.cancel != nil {
		r.cancel()
	}
	r.reporting = false
	r.mu.Unlock()

	// Wait for goroutine to finish
	r.wg.Wait()
}

// reportWithRetry sends heartbeat with retry mechanism (max 3 retries, exponential backoff)
func (r *HeartbeatReporter) reportWithRetry() {
	// Get latest probe results from scheduler
	tcpResults, udpResults := r.scheduler.GetLatestResults()

	// Aggregate metrics from actual probe results
	data := r.AggregateMetrics(tcpResults, udpResults)

	for attempt := 0; attempt < MaxRetries; attempt++ {
		err := r.apiClient.SendHeartbeat(data)
		if err == nil {
			return // Success
		}

		logger.WithFields(map[string]interface{}{"component": "reporter", "attempt": attempt + 1, "max_retries": MaxRetries, "error": err.Error()}).Error("Heartbeat report failed")

		if attempt < MaxRetries-1 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}

	logger.WithFields(map[string]interface{}{"component": "reporter", "attempts": MaxRetries}).Error("Heartbeat report failed after retries, giving up")
}
