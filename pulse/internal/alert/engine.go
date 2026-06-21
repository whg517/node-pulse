package alert

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/realtime"
	"github.com/whg517/node-pulse/pulse/internal/suppression"
	"github.com/whg517/node-pulse/pulse/internal/webhook"
)

// MetricData represents metric data from beacon heartbeat
type MetricData struct {
	NodeID         string
	LatencyMs      float64
	PacketLossRate float64
	JitterMs       float64
	Timestamp      time.Time
}

// AlertEngine handles alert evaluation and event creation
type AlertEngine struct {
	pool                     *pgxpool.Pool
	alertQuerier             db.AlertQuerier
	alertEventsQuerier       db.AlertEventsQuerier
	suppressionService       *suppression.Service
	webhookPushService       *webhook.PushService
	realtimeHub              *realtime.Hub
	metricChannel            chan *MetricData
	workerPoolSize           int
	ctx                      context.Context
	cancel                   context.CancelFunc
	wg                       sync.WaitGroup
	ruleCache                []*models.Alert
	ruleCacheMutex           sync.RWMutex
	ruleCacheLastRefresh     time.Time
	ruleCacheRefreshInterval time.Duration
}

// EngineConfig defines configuration for AlertEngine
type EngineConfig struct {
	WorkerPoolSize           int
	MetricChannelBufferSize  int
	RuleCacheRefreshInterval time.Duration
}

// DefaultEngineConfig returns default engine configuration
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		WorkerPoolSize:           10,
		MetricChannelBufferSize:  1000,
		RuleCacheRefreshInterval: 60 * time.Second,
	}
}

// NewAlertEngine creates a new AlertEngine
func NewAlertEngine(
	pool *pgxpool.Pool,
	alertQuerier db.AlertQuerier,
	config EngineConfig,
) *AlertEngine {
	ctx, cancel := context.WithCancel(context.Background())

	alertEventsQuerier := db.NewAlertEventsQuerier(pool)
	suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)
	suppressionService := suppression.NewService(suppressionQuerier)

	// Get webhooks querier for webhook push
	webhookQuerier := db.NewWebhookQuerier(pool)
	webhookLogsQuerier := db.NewWebhookLogsQuerier(pool)
	webhookPushService := webhook.NewPushService(webhookQuerier, webhookLogsQuerier, "http://localhost:6532")

	return &AlertEngine{
		pool:                     pool,
		alertQuerier:             alertQuerier,
		alertEventsQuerier:       alertEventsQuerier,
		suppressionService:       suppressionService,
		webhookPushService:       webhookPushService,
		metricChannel:            make(chan *MetricData, config.MetricChannelBufferSize),
		workerPoolSize:           config.WorkerPoolSize,
		ctx:                      ctx,
		cancel:                   cancel,
		ruleCache:                make([]*models.Alert, 0),
		ruleCacheRefreshInterval: config.RuleCacheRefreshInterval,
	}
}

// getPool returns the database pool
func (e *AlertEngine) getPool() *pgxpool.Pool {
	return e.pool
}

// WithWebhookURLValidator overrides the URL validator used by the internal webhook push service.
// Pass nil to disable URL validation (useful in tests with http:// servers).
func (e *AlertEngine) WithWebhookURLValidator(fn func(string) error) *AlertEngine {
	e.webhookPushService.WithURLValidator(fn)
	return e
}

// WithRealtimeHub enables websocket broadcasts for alert lifecycle events.
func (e *AlertEngine) WithRealtimeHub(hub *realtime.Hub) *AlertEngine {
	e.realtimeHub = hub
	return e
}

// Start starts the alert engine workers
func (e *AlertEngine) Start() {
	slog.Info("Starting alert engine", "worker_pool_size", e.workerPoolSize)

	// Initialize rule cache
	if err := e.refreshRuleCache(e.ctx); err != nil {
		slog.Error("Failed to initialize rule cache", "error", err)
	}

	// Start rule cache refresh goroutine
	e.wg.Add(1)
	go e.ruleCacheRefreshLoop()

	// Start worker pool
	for i := 0; i < e.workerPoolSize; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}
}

// Stop gracefully stops the alert engine
func (e *AlertEngine) Stop() {
	slog.Info("Stopping alert engine")
	e.cancel()
	close(e.metricChannel)
	e.wg.Wait()
	slog.Info("Alert engine stopped")
}

// EvaluateMetrics queues metric data for evaluation (non-blocking)
func (e *AlertEngine) EvaluateMetrics(data *MetricData) bool {
	select {
	case e.metricChannel <- data:
		return true
	default:
		// Channel full, log warning but don't block
		slog.Warn("Alert engine metric channel full, dropping metric",
			"node_id", data.NodeID,
			"timestamp", data.Timestamp)
		return false
	}
}

// worker processes metrics from the channel
func (e *AlertEngine) worker(id int) {
	defer e.wg.Done()

	slog.Debug("Starting alert engine worker", "worker_id", id)

	for {
		select {
		case <-e.ctx.Done():
			slog.Debug("Stopping alert engine worker", "worker_id", id)
			return
		case data, ok := <-e.metricChannel:
			if !ok {
				// Channel closed
				return
			}
			e.evaluateMetric(data)
		}
	}
}

// evaluateMetric evaluates a single metric against all applicable rules
func (e *AlertEngine) evaluateMetric(data *MetricData) {
	startTime := time.Now()
	alertsCreated := 0

	// Get cached rules
	e.ruleCacheMutex.RLock()
	rules := make([]*models.Alert, len(e.ruleCache))
	copy(rules, e.ruleCache)
	e.ruleCacheMutex.RUnlock()

	// Evaluate each rule
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Check if rule applies to this node
		if rule.NodeID != nil && *rule.NodeID != data.NodeID {
			// Node-specific rule for different node
			continue
		}

		// Evaluate rule
		if alertEvent := e.evaluateRule(rule, data); alertEvent != nil {
			// Check suppression before creating alert event
			ctx, cancel := context.WithTimeout(e.ctx, 5*time.Second)
			defer cancel()

			suppressed, err := e.suppressionService.ShouldSuppress(ctx, data.NodeID, rule.Metric)
			if err != nil {
				slog.Error("Failed to check suppression",
					"node_id", data.NodeID,
					"metric", rule.Metric,
					"error", err)
				// Continue with alert creation on error (fail open)
			} else if suppressed {
				slog.Info("Alert suppressed",
					"node_id", data.NodeID,
					"metric", rule.Metric,
					"level", rule.Level,
					"threshold", rule.Threshold,
					"current_value", alertEvent.CurrentValue)
				continue // Skip creating alert event
			}

			// Create alert event in database
			err = e.alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
			if err != nil {
				slog.Error("Failed to create alert event",
					"node_id", data.NodeID,
					"metric", rule.Metric,
					"error", err)
			} else {
				alertsCreated++
				slog.Info("Alert event created",
					"alert_id", alertEvent.ID,
					"node_id", alertEvent.NodeID,
					"metric", alertEvent.Metric,
					"threshold", alertEvent.Threshold,
					"current_value", alertEvent.CurrentValue,
					"level", alertEvent.Level)

				// Create alert record for lifecycle tracking (Story 6.1)
				alertRecord := &models.AlertRecord{
					AlertEventID: alertEvent.ID,
					NodeID:       alertEvent.NodeID,
					Metric:       alertEvent.Metric,
					Level:        alertEvent.Level,
					Status:       "pending", // Initial status
				}
				if recordErr := db.CreateAlertRecord(ctx, e.getPool(), alertRecord); recordErr != nil {
					slog.Error("Failed to create alert record",
						"alert_event_id", alertEvent.ID,
						"node_id", data.NodeID,
						"metric", rule.Metric,
						"error", recordErr)
					// Don't fail the alert if record creation fails
				} else {
					slog.Debug("Alert record created",
						"record_id", alertRecord.ID,
						"alert_event_id", alertEvent.ID)
					if e.realtimeHub != nil {
						e.realtimeHub.BroadcastAlertRecord(realtime.EventAlertNew, alertRecord)
					}
				}

				// Record suppression for future alerts
				err = e.suppressionService.RecordDefaultSuppression(ctx, data.NodeID, rule.Metric)
				if err != nil {
					slog.Error("Failed to record suppression",
						"node_id", data.NodeID,
						"metric", rule.Metric,
						"error", err)
					// Don't fail the alert if suppression recording fails
				}

				// Send webhook notifications asynchronously (non-blocking)
				go func(event *models.AlertEvent) {
					webhookCtx, webhookCancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer webhookCancel()

					if err := e.webhookPushService.SendAlert(webhookCtx, event); err != nil {
						slog.Error("Webhook push failed",
							"alert_id", event.ID,
							"error", err)
					}
				}(alertEvent)
			}
		}
	}

	duration := time.Since(startTime)
	if duration > 100*time.Millisecond {
		slog.Warn("Slow alert evaluation",
			"node_id", data.NodeID,
			"duration_ms", duration.Milliseconds(),
			"alerts_created", alertsCreated)
	}
}

// evaluateRule evaluates a single rule against metric data
func (e *AlertEngine) evaluateRule(rule *models.Alert, data *MetricData) *models.AlertEvent {
	var currentValue float64
	var exceedsThreshold bool

	switch rule.Metric {
	case "latency":
		currentValue = data.LatencyMs
		exceedsThreshold = currentValue > rule.Threshold
	case "packet_loss_rate":
		currentValue = data.PacketLossRate
		exceedsThreshold = currentValue > rule.Threshold
	case "jitter":
		currentValue = data.JitterMs
		exceedsThreshold = currentValue > rule.Threshold
	default:
		slog.Warn("Unknown metric type in alert rule", "metric", rule.Metric)
		return nil
	}

	if !exceedsThreshold {
		return nil
	}

	// Create alert event
	return &models.AlertEvent{
		NodeID:       data.NodeID,
		Metric:       rule.Metric,
		Threshold:    rule.Threshold,
		CurrentValue: currentValue,
		Level:        rule.Level,
		CreatedAt:    time.Now(),
	}
}

// ruleCacheRefreshLoop periodically refreshes the rule cache
func (e *AlertEngine) ruleCacheRefreshLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.ruleCacheRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			if err := e.refreshRuleCache(e.ctx); err != nil {
				slog.Error("Failed to refresh rule cache", "error", err)
			}
		}
	}
}

// refreshRuleCache refreshes the cached alert rules
func (e *AlertEngine) refreshRuleCache(ctx context.Context) error {
	rules, err := e.alertQuerier.GetAlerts(ctx, nil) // nil = get all rules (no node filter)
	if err != nil {
		return err
	}

	// Filter enabled rules only
	enabledRules := make([]*models.Alert, 0)
	for _, rule := range rules {
		if rule.Enabled {
			enabledRules = append(enabledRules, rule)
		}
	}

	e.ruleCacheMutex.Lock()
	e.ruleCache = enabledRules
	e.ruleCacheLastRefresh = time.Now()
	e.ruleCacheMutex.Unlock()

	slog.Debug("Rule cache refreshed",
		"total_rules", len(rules),
		"enabled_rules", len(enabledRules))

	return nil
}

// GetStats returns alert engine statistics
func (e *AlertEngine) GetStats() map[string]interface{} {
	e.ruleCacheMutex.RLock()
	defer e.ruleCacheMutex.RUnlock()

	return map[string]interface{}{
		"cached_rules":            len(e.ruleCache),
		"rule_cache_last_refresh": e.ruleCacheLastRefresh.Format(time.RFC3339),
		"metric_channel_depth":    len(e.metricChannel),
		"metric_channel_capacity": cap(e.metricChannel),
	}
}
