// Package reporter provides heartbeat reporting functionality for Beacon.
// It aggregates probe metrics and reports them to the Pulse server via HTTP/HTTPS.
package reporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/whg517/node-pulse/beacon/internal/logger"
	"github.com/whg517/node-pulse/beacon/internal/models"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	// ReportInterval is the interval between heartbeat reports (60 seconds)
	ReportInterval = 60 * time.Second
	// defaultMaxRetries is the default maximum number of retry attempts for failed reports
	defaultMaxRetries = 3
	// MaxUploadLatency is the maximum acceptable upload latency (NFR-PERF-001)
	MaxUploadLatency = 5 * time.Second
)

// HeartbeatOutcomeListener is notified of heartbeat report outcomes. The Beacon's
// config.ModeManager satisfies this to drive degraded-mode state transitions
// (G16). Defined locally to avoid a reporter→config import cycle.
type HeartbeatOutcomeListener interface {
	RecordHeartbeatSuccess()
	RecordHeartbeatFailure()
}

// ResumeCache stores heartbeat payloads that could not be delivered and replays
// them once connectivity is restored (G18). reporter.PriorityCache satisfies this.
type ResumeCache interface {
	Add(entry *CacheEntry) error
	Remove(id string) bool
	GetAllEntriesForUpload() []*CacheEntry
}

// ResumeUploadRecorder records replayed bytes for Prometheus metrics (G18).
type ResumeUploadRecorder interface {
	RecordResumeUpload(bytes int64)
}

// emptyOutcomeListener is a no-op listener used when none is configured.
type emptyOutcomeListener struct{}

func (emptyOutcomeListener) RecordHeartbeatSuccess() {}
func (emptyOutcomeListener) RecordHeartbeatFailure() {}


// HeartbeatData represents the heartbeat data structure for reporting to Pulse
type HeartbeatData struct {
	NodeID         string  `json:"node_id"`           // UUID from Pulse registration
	LatencyMs      float64 `json:"latency_ms"`        // RTT mean in milliseconds
	PacketLossRate float64 `json:"packet_loss_rate"`  // Packet loss rate (0-100%)
	JitterMs       float64 `json:"jitter_ms"`         // Delay jitter in milliseconds
	Timestamp      string  `json:"timestamp"`         // ISO 8601 timestamp
}

// TokenProvider defines interface for getting JWT tokens
type TokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
	GetNodeID() string
	InvalidateToken()
}

// PulseAPIClient handles HTTP/HTTPS communication with Pulse server
type PulseAPIClient struct {
	serverURL        string
	httpClient       *http.Client
	timeout          time.Duration
	jwtClient        TokenProvider
	compressionOn    bool
	compressionLevel int
}

// compressedHeartbeatRequest is the wire shape expected by Pulse's
// POST /api/v1/beacon/heartbeat/compressed handler: base64(gzip(payload)) plus a
// CRC32 checksum of the compressed bytes.
type compressedHeartbeatRequest struct {
	Data     string `json:"data"`
	Checksum uint32 `json:"checksum"`
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
	ctx        context.Context // Store context for cancellation
	cancel     context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.Mutex
	reporting bool

	// Retry behavior (G19). Defaults preserve the legacy 3-attempt exponential
	// backoff when no ReconnectConfig is supplied.
	maxRetries int
	retryBase  time.Duration
	backoff    string // "exponential" | "linear" | "constant"

	// Optional collaborators (G16/G18). When nil, a no-op listener / no cache.
	outcomeListener HeartbeatOutcomeListener
	cache           ResumeCache
	resumeRecorder  ResumeUploadRecorder
}

// ReporterOption configures a HeartbeatReporter.
type ReporterOption func(*HeartbeatReporter)

// WithReconnectConfig applies the Beacon's reconnect settings to retry behavior.
// Zero-value fields fall back to safe defaults (G19).
func WithReconnectConfig(maxRetries int, retryIntervalSeconds int, backoff string) ReporterOption {
	return func(r *HeartbeatReporter) {
		if maxRetries > 0 {
			r.maxRetries = maxRetries
		}
		if retryIntervalSeconds > 0 {
			r.retryBase = time.Duration(retryIntervalSeconds) * time.Second
		}
		if backoff == "exponential" || backoff == "linear" || backoff == "constant" {
			r.backoff = backoff
		}
	}
}

// WithOutcomeListener wires a degraded-mode listener (G16).
func WithOutcomeListener(l HeartbeatOutcomeListener) ReporterOption {
	return func(r *HeartbeatReporter) {
		if l != nil {
			r.outcomeListener = l
		}
	}
}

// WithResumeCache wires the failed-payload cache + replay recorder (G18).
func WithResumeCache(cache ResumeCache, recorder ResumeUploadRecorder) ReporterOption {
	return func(r *HeartbeatReporter) {
		r.cache = cache
		if recorder != nil {
			r.resumeRecorder = recorder
		}
	}
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

// NewPulseAPIClient creates a new Pulse API client with TLS and JWT support
func NewPulseAPIClient(serverURL string, timeout time.Duration, jwtClient TokenProvider) *PulseAPIClient {
	return newPulseAPIClient(serverURL, timeout, jwtClient, false, 0)
}

// NewPulseAPIClientWithCompression creates a client that sends compressed
// heartbeats when enabled (G17).
func NewPulseAPIClientWithCompression(serverURL string, timeout time.Duration, jwtClient TokenProvider, level int) *PulseAPIClient {
	return newPulseAPIClient(serverURL, timeout, jwtClient, true, level)
}

func newPulseAPIClient(serverURL string, timeout time.Duration, jwtClient TokenProvider, compressionOn bool, level int) *PulseAPIClient {
	// Create HTTP client with TLS config
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12, // Enforce TLS 1.2 or higher (NFR-SEC-001)
			InsecureSkipVerify: false,            // Require certificate validation
			// In production, you may want to add:
			// RootCAs: customCertPool,
		},
	}

	// Wrap the transport with otelhttp so that every outbound request to Pulse
	// automatically carries a W3C "traceparent" header.  This propagates the
	// active span context from the Beacon probe loop into Pulse's API handler,
	// enabling end-to-end distributed trace correlation.
	tracingTransport := otelhttp.NewTransport(transport)

	httpClient := &http.Client{
		Transport: tracingTransport,
		Timeout:   timeout,
	}

	return &PulseAPIClient{
		serverURL:        serverURL,
		httpClient:       httpClient,
		timeout:          timeout,
		jwtClient:        jwtClient,
		compressionOn:    compressionOn,
		compressionLevel: level,
	}
}

// NewHeartbeatReporter creates a new HeartbeatReporter with probe scheduler integration.
// Optional behavior (reconnect config, degraded-mode listener, resume cache) can be
// supplied via ReporterOption (G16/G18/G19).
func NewHeartbeatReporter(apiClient *PulseAPIClient, scheduler ProbeScheduler, opts ...ReporterOption) *HeartbeatReporter {
	r := &HeartbeatReporter{
		apiClient:       apiClient,
		nodeID:          apiClient.jwtClient.GetNodeID(),
		scheduler:       scheduler,
		reporting:       false,
		maxRetries:      defaultMaxRetries,
		retryBase:       1 * time.Second,
		backoff:         "exponential",
		outcomeListener: emptyOutcomeListener{},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// SendHeartbeat sends heartbeat data to Pulse server with JWT authentication
func (c *PulseAPIClient) SendHeartbeat(ctx context.Context, data *HeartbeatData) error {
	// Measure upload latency (NFR-PERF-001)
	startTime := time.Now()

	// Get valid access token
	accessToken, err := c.jwtClient.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Serialize heartbeat data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat data: %w", err)
	}

	// Create HTTP POST request
	url := c.serverURL + "/api/v1/beacon/heartbeat"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Send request with timeout
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle authentication errors
	if resp.StatusCode == http.StatusUnauthorized {
		c.jwtClient.InvalidateToken()
		return fmt.Errorf("authentication failed: invalid or expired token")
	}

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

// SendCompressedHeartbeat gzip-compresses the JSON payload and POSTs it to the
// /api/v1/beacon/heartbeat/compressed endpoint (G17). The wire format matches
// Pulse's HandleCompressedHeartbeat: { "data": base64(gzip(json)), "checksum": crc32(gzip) }.
func (c *PulseAPIClient) SendCompressedHeartbeat(ctx context.Context, data *HeartbeatData) error {
	startTime := time.Now()

	accessToken, err := c.jwtClient.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat data: %w", err)
	}

	level := c.compressionLevel
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}
	gz, err := CompressWithLevel(jsonData, level)
	if err != nil {
		return fmt.Errorf("failed to compress heartbeat: %w", err)
	}

	checksum := crc32.ChecksumIEEE(gz)
	body := compressedHeartbeatRequest{
		Data:     base64.StdEncoding.EncodeToString(gz),
		Checksum: checksum,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal compressed request: %w", err)
	}

	url := c.serverURL + "/api/v1/beacon/heartbeat/compressed"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		c.jwtClient.InvalidateToken()
		return fmt.Errorf("authentication failed: invalid or expired token")
	}

	elapsed := time.Since(startTime)
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pulse API returned error %d: %s", resp.StatusCode, string(respBody))
	}

	if elapsed > MaxUploadLatency {
		logger.WithFields(map[string]interface{}{"component": "reporter", "latency": elapsed.String(), "threshold": MaxUploadLatency.String()}).Warn("Compressed heartbeat upload latency exceeds requirement")
	}

	logger.WithFields(map[string]interface{}{"component": "reporter", "latency": elapsed.String(), "original_bytes": len(jsonData), "compressed_bytes": len(gz)}).Info("Compressed heartbeat reported successfully")
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

	// Store context and create cancellable context
	r.ctx, r.cancel = context.WithCancel(ctx)
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
			case <-r.ctx.Done():
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

// reportWithRetry sends heartbeat with retry mechanism. Behavior:
//   - G18: on a fresh report, first drain any cached payloads from prior failures.
//   - G19: retry count and backoff follow the configured ReconnectConfig.
//   - G16: on terminal success/failure the outcome listener is notified (mode manager).
//   - G18: on terminal failure the payload is cached for later replay (when a cache is wired).
func (r *HeartbeatReporter) reportWithRetry() {
	// Replay previously failed payloads before sending the fresh heartbeat.
	r.drainCache()

	// Get latest probe results from scheduler
	tcpResults, udpResults := r.scheduler.GetLatestResults()

	// Aggregate metrics from actual probe results
	data := r.AggregateMetrics(tcpResults, udpResults)
	payload, err := json.Marshal(data)
	if err != nil {
		logger.WithFields(map[string]interface{}{"component": "reporter", "error": err.Error()}).Error("Failed to marshal heartbeat for caching")
		payload = nil
	}

	for attempt := 0; attempt < r.maxRetries; attempt++ {
		err := r.send(data)
		if err == nil {
			r.outcomeListener.RecordHeartbeatSuccess()
			return // Success
		}

		logger.WithFields(map[string]interface{}{"component": "reporter", "attempt": attempt + 1, "max_retries": r.maxRetries, "error": err.Error()}).Error("Heartbeat report failed")

		if attempt < r.maxRetries-1 {
			time.Sleep(r.computeBackoff(attempt))
		}
	}

	logger.WithFields(map[string]interface{}{"component": "reporter", "attempts": r.maxRetries}).Error("Heartbeat report failed after retries, giving up")
	r.outcomeListener.RecordHeartbeatFailure()
	r.cacheFailedPayload(payload, data.Timestamp)
}

// send dispatches the heartbeat via the compressed or plain endpoint.
func (r *HeartbeatReporter) send(data *HeartbeatData) error {
	if r.apiClient.compressionOn {
		return r.apiClient.SendCompressedHeartbeat(r.ctx, data)
	}
	return r.apiClient.SendHeartbeat(r.ctx, data)
}

// computeBackoff returns the backoff for a given 0-based attempt based on the
// configured strategy (G19).
func (r *HeartbeatReporter) computeBackoff(attempt int) time.Duration {
	switch r.backoff {
	case "linear":
		return r.retryBase * time.Duration(attempt+1)
	case "constant":
		return r.retryBase
	default: // "exponential" (also the fallback for unknown/empty)
		return r.retryBase * time.Duration(1<<uint(attempt))
	}
}

// drainCache replays cached failed payloads once connectivity is restored (G18).
func (r *HeartbeatReporter) drainCache() {
	if r.cache == nil {
		return
	}
	entries := r.cache.GetAllEntriesForUpload()
	for _, entry := range entries {
		if err := r.replayEntry(entry); err != nil {
			logger.WithFields(map[string]interface{}{"component": "reporter", "cache_id": entry.ID, "error": err.Error()}).Warn("Failed to drain cached heartbeat; stopping drain")
			return
		}
		r.cache.Remove(entry.ID)
		if r.resumeRecorder != nil {
			r.resumeRecorder.RecordResumeUpload(int64(len(entry.Data)))
		}
	}
}

// replayEntry re-POSTs a cached payload. The cache stores raw HeartbeatData JSON.
func (r *HeartbeatReporter) replayEntry(entry *CacheEntry) error {
	var data HeartbeatData
	if err := json.Unmarshal(entry.Data, &data); err != nil {
		return fmt.Errorf("decode cached payload: %w", err)
	}
	// Send up to 3 attempts; this is best-effort drain, not the full retry policy.
	for attempt := 0; attempt < defaultMaxRetries; attempt++ {
		if err := r.send(&data); err == nil {
			return nil
		} else if attempt == defaultMaxRetries-1 {
			return err
		}
		time.Sleep(r.computeBackoff(attempt))
	}
	return fmt.Errorf("drain failed")
}

// cacheFailedPayload stores the payload that could not be delivered (G18). Heartbeat
// payloads use CacheP2 (P1 is rejected by the cache by design).
func (r *HeartbeatReporter) cacheFailedPayload(payload []byte, timestamp string) {
	if r.cache == nil || len(payload) == 0 {
		return
	}
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		ts = time.Now()
	}
	entry := &CacheEntry{
		ID:        r.nodeID + "-" + timestamp,
		Priority:  CacheP2,
		Data:      payload,
		Checksum:  crc32.ChecksumIEEE(payload),
		Size:      int64(len(payload)),
		Timestamp: ts,
	}
	if err := r.cache.Add(entry); err != nil {
		logger.WithFields(map[string]interface{}{"component": "reporter", "error": err.Error()}).Warn("Failed to cache failed heartbeat payload")
	} else {
		logger.WithField("component", "reporter").Info("Cached failed heartbeat payload for later replay")
	}
}
