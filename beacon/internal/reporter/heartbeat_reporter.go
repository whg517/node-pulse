package reporter

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"beacon/internal/models"
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

// HeartbeatReporter manages scheduled heartbeat reporting to Pulse
type HeartbeatReporter struct {
	apiClient *PulseAPIClient
	nodeID    string
	ticker    *time.Ticker
	stopChan  chan struct{}
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
			MinVersion: tls.VersionTLS12, // Enforce TLS 1.2 or higher
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

// NewHeartbeatReporter creates a new HeartbeatReporter
func NewHeartbeatReporter(apiClient *PulseAPIClient, nodeID string) *HeartbeatReporter {
	return &HeartbeatReporter{
		apiClient: apiClient,
		nodeID:    nodeID,
		reporting: false,
		stopChan:  nil,
	}
}

// SendHeartbeat sends heartbeat data to Pulse server
func (c *PulseAPIClient) SendHeartbeat(data *HeartbeatData) error {
	// Serialize heartbeat data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Create HTTP POST request
	url := c.serverURL + "/api/v1/beacon/heartbeat"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request with timeout
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pulse API returned error %d", resp.StatusCode)
	}

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

	// If no successful probe results, use default values
	if count == 0 {
		return &HeartbeatData{
			NodeID:         r.nodeID,
			LatencyMs:      0,
			PacketLossRate: 100, // All probes failed
			JitterMs:       0,
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

// StartReporting starts the scheduled heartbeat reporting (every 60 seconds)
func (r *HeartbeatReporter) StartReporting() {
	if r.reporting {
		log.Println("[WARN] Heartbeat reporter already running")
		return
	}

	r.reporting = true
	r.ticker = time.NewTicker(60 * time.Second)
	r.stopChan = make(chan struct{})

	log.Println("[INFO] Starting heartbeat reporter (interval: 60s)")

	// Report immediately on start
	go r.reportWithRetry()

	// Start scheduled reporting
	go func() {
		for {
			select {
			case <-r.ticker.C:
				r.reportWithRetry()
			case <-r.stopChan:
				r.ticker.Stop()
				log.Println("[INFO] Heartbeat reporter stopped")
				return
			}
		}
	}()
}

// StopReporting gracefully stops the heartbeat reporter
func (r *HeartbeatReporter) StopReporting() {
	if !r.reporting {
		return
	}

	close(r.stopChan)
	r.reporting = false
}

// reportWithRetry sends heartbeat with retry mechanism (max 3 retries, exponential backoff)
func (r *HeartbeatReporter) reportWithRetry() {
	// Create heartbeat data with default values (no probe results yet)
	data := NewHeartbeatData(r.nodeID, 0, 100, 0)

	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := r.apiClient.SendHeartbeat(data)
		if err == nil {
			log.Println("[INFO] Heartbeat reported successfully to Pulse")
			return // Success
		}

		log.Printf("[ERROR] Heartbeat report failed (attempt %d/%d): %v", attempt+1, maxRetries, err)

		if attempt < maxRetries-1 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}
	}

	log.Printf("[ERROR] Heartbeat report failed after %d attempts, giving up", maxRetries)
}
