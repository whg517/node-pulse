package diagnostic

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Logger interface for baseline calculator logging
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
}

// Baseline represents calculated baseline metrics
type Baseline struct {
	LatencyMs      float64   `json:"latency_ms"`
	PacketLossRate float64   `json:"packet_loss_rate"`
	JitterMs       float64   `json:"jitter_ms"`
	CalculatedAt   time.Time `json:"calculated_at"`
	NodeCount      int       `json:"node_count"`
	DataPointCount int       `json:"data_point_count"`
}

// BaselineCache provides thread-safe caching for baselines
type BaselineCache struct {
	mu        sync.RWMutex
	baseline  *Baseline
	expiresAt time.Time
}

// Get returns the cached baseline if not expired
func (c *BaselineCache) Get() *Baseline {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.baseline == nil || time.Now().After(c.expiresAt) {
		return nil
	}
	return c.baseline
}

// Set updates the cached baseline with expiration
func (c *BaselineCache) Set(baseline *Baseline, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.baseline = baseline
	c.expiresAt = time.Now().Add(ttl)
}

// Clear removes the cached baseline
func (c *BaselineCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.baseline = nil
	c.expiresAt = time.Time{}
}

// BaselineCalculator calculates and caches 7-day moving average baselines
type BaselineCalculator struct {
	db             *sql.DB
	cache          *BaselineCache
	updateInterval time.Duration
	cacheTTL       time.Duration
	logger         Logger
	stopChan       chan struct{}
	wg             sync.WaitGroup
	running        bool
	mu             sync.Mutex
}

// BaselineCalculatorOption configures the baseline calculator
type BaselineCalculatorOption func(*BaselineCalculator)

// WithUpdateInterval sets the background update interval
func WithUpdateInterval(d time.Duration) BaselineCalculatorOption {
	return func(c *BaselineCalculator) {
		c.updateInterval = d
	}
}

// WithCacheTTL sets the cache time-to-live
func WithCacheTTL(d time.Duration) BaselineCalculatorOption {
	return func(c *BaselineCalculator) {
		c.cacheTTL = d
	}
}

// WithLogger sets the logger
func WithLogger(logger Logger) BaselineCalculatorOption {
	return func(c *BaselineCalculator) {
		c.logger = logger
	}
}

// NewBaselineCalculator creates a new baseline calculator
func NewBaselineCalculator(db *sql.DB, opts ...BaselineCalculatorOption) *BaselineCalculator {
	c := &BaselineCalculator{
		db:             db,
		cache:          &BaselineCache{},
		updateInterval: 15 * time.Minute,
		cacheTTL:       15 * time.Minute,
		logger:         &noopLogger{},
		stopChan:       make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// CalculateBaselines calculates baselines from 7-day historical data
func (c *BaselineCalculator) CalculateBaselines(ctx context.Context) (*Baseline, error) {
	query := `
		SELECT
			COALESCE(AVG(latency_ms), 0) as avg_latency,
			COALESCE(AVG(packet_loss_rate), 0) as avg_packet_loss,
			COALESCE(AVG(jitter_ms), 0) as avg_jitter,
			COUNT(*) as data_points,
			COUNT(DISTINCT node_id) as node_count
		FROM metrics
		WHERE timestamp >= NOW() - INTERVAL '7 days'
		  AND timestamp < NOW() - INTERVAL '1 hour'
		  AND latency_ms > 0
		  AND packet_loss_rate < 0.5
	`

	var latency, packetLoss, jitter float64
	var dataPoints, nodeCount int

	err := c.db.QueryRowContext(ctx, query).Scan(
		&latency,
		&packetLoss,
		&jitter,
		&dataPoints,
		&nodeCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query baselines: %w", err)
	}

	// If no data points, return default baselines
	if dataPoints == 0 {
		c.logger.Warn("No historical data available for baseline calculation, using defaults")
		return &Baseline{
			LatencyMs:      50.0,
			PacketLossRate: 0.01,
			JitterMs:       2.0,
			CalculatedAt:   time.Now(),
			NodeCount:      0,
			DataPointCount: 0,
		}, nil
	}

	baseline := &Baseline{
		LatencyMs:      latency,
		PacketLossRate: packetLoss,
		JitterMs:       jitter,
		CalculatedAt:   time.Now(),
		NodeCount:      nodeCount,
		DataPointCount: dataPoints,
	}

	c.logger.Info("Calculated baselines from 7-day history",
		"latency_ms", latency,
		"packet_loss_rate", packetLoss,
		"jitter_ms", jitter,
		"node_count", nodeCount,
		"data_points", dataPoints,
	)

	return baseline, nil
}

// GetBaselines returns cached baselines, recalculating if expired
func (c *BaselineCalculator) GetBaselines(ctx context.Context) (*Baseline, error) {
	// Try to get from cache first
	if baseline := c.cache.Get(); baseline != nil {
		return baseline, nil
	}

	// Cache miss or expired - recalculate
	baseline, err := c.CalculateBaselines(ctx)
	if err != nil {
		// If calculation fails, try to return any stale cached value
		c.cache.mu.RLock()
		staleBaseline := c.cache.baseline
		c.cache.mu.RUnlock()

		if staleBaseline != nil {
			c.logger.Warn("Using stale baseline due to calculation error", "error", err)
			return staleBaseline, nil
		}

		return nil, err
	}

	// Update cache
	c.cache.Set(baseline, c.cacheTTL)

	return baseline, nil
}

// Start begins the background worker for periodic baseline updates
func (c *BaselineCalculator) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("baseline calculator already running")
	}
	c.running = true
	c.stopChan = make(chan struct{})
	c.mu.Unlock()

	c.wg.Add(1)
	go c.runBackgroundWorker(ctx)

	c.logger.Info("Started baseline calculator background worker",
		"update_interval", c.updateInterval,
		"cache_ttl", c.cacheTTL)

	return nil
}

// Stop halts the background worker
func (c *BaselineCalculator) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return nil
	}

	close(c.stopChan)
	c.wg.Wait()
	c.running = false

	c.logger.Info("Stopped baseline calculator background worker")
	return nil
}

// runBackgroundWorker periodically updates baselines
func (c *BaselineCalculator) runBackgroundWorker(ctx context.Context) {
	defer c.wg.Done()

	// Initial calculation (skip if no database)
	if c.db != nil {
		if err := c.updateBaselines(ctx); err != nil {
			c.logger.Error("Initial baseline calculation failed", "error", err)
		}
	}

	ticker := time.NewTicker(c.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Skip if no database configured
			if c.db == nil {
				continue
			}
			if err := c.updateBaselines(ctx); err != nil {
				c.logger.Error("Periodic baseline calculation failed", "error", err)
			}
		}
	}
}

// updateBaselines calculates and caches new baselines
func (c *BaselineCalculator) updateBaselines(ctx context.Context) error {
	baseline, err := c.CalculateBaselines(ctx)
	if err != nil {
		return err
	}

	c.cache.Set(baseline, c.cacheTTL)
	return nil
}

// noopLogger is a default no-op logger
type noopLogger struct{}

func (l *noopLogger) Info(msg string, args ...any)  {}
func (l *noopLogger) Error(msg string, args ...any) {}
func (l *noopLogger) Warn(msg string, args ...any)  {}
