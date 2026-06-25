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

// AlertEngine evaluates beacon metrics against cached alert rules and, when a
// rule fires, hands the resulting event to a Dispatcher for side-effect
// orchestration (persistence, suppression, broadcast, webhooks).
//
// The engine itself owns only the worker pool, the metric channel and the rule
// cache; all side-effects live behind the Dispatcher port so they can be
// unit-tested and swapped independently.
type AlertEngine struct {
	pool                     *pgxpool.Pool
	ruleSource               RuleSource
	dispatcher               Dispatcher
	metricChannel            chan *MetricData
	workerPoolSize           int
	ctx                      context.Context
	cancel                   context.CancelFunc
	wg                       sync.WaitGroup
	ruleCache                []*models.Alert
	ruleCacheMutex           sync.RWMutex
	ruleCacheLastRefresh     time.Time
	ruleCacheRefreshInterval time.Duration

	// webhookPushService is retained so the With* builder methods can still
	// configure the webhook adapter created in NewAlertEngine.
	webhookPushService *webhook.PushService
}

// NewAlertEngine creates a new AlertEngine wired to the database-backed
// adapters. The public signature is unchanged from the pre-refactor engine.
func NewAlertEngine(
	pool *pgxpool.Pool,
	alertQuerier db.AlertQuerier,
	config EngineConfig,
) *AlertEngine {
	ctx, cancel := context.WithCancel(context.Background())

	alertEventsQuerier := db.NewAlertEventsQuerier(pool)
	suppressionQuerier := db.NewAlertSuppressionsQuerier(pool)
	suppressionService := suppression.NewService(suppressionQuerier)

	// Webhook push service: baseURL placeholder is overridden via
	// WithWebhookBaseURL from the configured server.base_url.
	webhookQuerier := db.NewWebhookQuerier(pool)
	webhookLogsQuerier := db.NewWebhookLogsQuerier(pool)
	webhookPushService := webhook.NewPushService(webhookQuerier, webhookLogsQuerier, "http://localhost:6532")

	dispatcher := &CompositeDispatcher{
		Suppression: NewSuppressionChecker(suppressionService),
		EventSink:   NewEventSink(alertEventsQuerier, pool),
		Broadcaster: nil, // set via WithRealtimeHub
		Webhook:     NewWebhookPusher(webhookPushService),
	}

	return &AlertEngine{
		pool:                     pool,
		ruleSource:               NewRuleSource(alertQuerier),
		dispatcher:               dispatcher,
		metricChannel:            make(chan *MetricData, config.MetricChannelBufferSize),
		workerPoolSize:           config.WorkerPoolSize,
		ctx:                      ctx,
		cancel:                   cancel,
		ruleCache:                make([]*models.Alert, 0),
		ruleCacheRefreshInterval: config.RuleCacheRefreshInterval,
		webhookPushService:       webhookPushService,
	}
}

// WithWebhookURLValidator overrides the URL validator used by the internal
// webhook push service. Pass nil to disable URL validation (useful in tests).
func (e *AlertEngine) WithWebhookURLValidator(fn func(string) error) *AlertEngine {
	e.webhookPushService.WithURLValidator(fn)
	return e
}

// WithRealtimeHub enables websocket broadcasts for alert lifecycle events.
func (e *AlertEngine) WithRealtimeHub(hub *realtime.Hub) *AlertEngine {
	if d, ok := e.dispatcher.(*CompositeDispatcher); ok {
		d.Broadcaster = NewBroadcaster(hub)
	}
	return e
}

// WithWebhookBaseURL sets the base URL used to render absolute links inside
// webhook payloads. Should be wired from config.Server.BaseURL so production
// deployments emit correct links instead of the localhost default.
func (e *AlertEngine) WithWebhookBaseURL(baseURL string) *AlertEngine {
	if baseURL != "" {
		e.webhookPushService.SetBaseURL(baseURL)
	}
	return e
}

// Start starts the alert engine workers.
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

// Stop gracefully stops the alert engine.
func (e *AlertEngine) Stop() {
	slog.Info("Stopping alert engine")
	e.cancel()
	close(e.metricChannel)
	e.wg.Wait()
	slog.Info("Alert engine stopped")
}

// EvaluateMetrics queues metric data for evaluation (non-blocking).
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

// worker processes metrics from the channel.
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

// evaluateMetric evaluates a single metric sample against all applicable rules
// and dispatches any resulting alert events. Side-effects are delegated to the
// configured Dispatcher.
func (e *AlertEngine) evaluateMetric(data *MetricData) {
	startTime := time.Now()
	alertsCreated := 0

	// Snapshot the cached rules under the read lock.
	e.ruleCacheMutex.RLock()
	rules := make([]*models.Alert, len(e.ruleCache))
	copy(rules, e.ruleCache)
	e.ruleCacheMutex.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		// Skip node-specific rules that do not target this node.
		if !RuleAppliesToNode(rule, data.NodeID) {
			continue
		}

		event := EvaluateRule(rule, data)
		if event == nil {
			continue
		}

		e.dispatcher.Dispatch(e.ctx, event, rule)
		alertsCreated++
	}

	if duration := time.Since(startTime); duration > 100*time.Millisecond {
		slog.Warn("Slow alert evaluation",
			"node_id", data.NodeID,
			"duration_ms", duration.Milliseconds(),
			"alerts_created", alertsCreated)
	}
}

// ruleCacheRefreshLoop periodically refreshes the rule cache.
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

// refreshRuleCache refreshes the cached alert rules.
func (e *AlertEngine) refreshRuleCache(ctx context.Context) error {
	if e.ruleSource == nil {
		panic("AlertEngine rule source not configured")
	}

	rules, err := e.ruleSource.GetAlerts(ctx, nil) // nil = get all rules (no node filter)
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

// GetStats returns alert engine statistics.
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
