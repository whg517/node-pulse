package alert

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/webhook"
)

// --- shared test doubles ------------------------------------------------

// ruleSourceFunc is a RuleSource backed by a closure, for rule-cache tests.
type ruleSourceFunc func(ctx context.Context, nodeID *string) ([]*models.Alert, error)

func (f ruleSourceFunc) GetAlerts(ctx context.Context, nodeID *string) ([]*models.Alert, error) {
	return f(ctx, nodeID)
}

// recordingDispatcher records every Dispatch call so evaluateMetric tests can
// assert how many alerts fired and with what events.
type recordingDispatcher struct {
	mu     sync.Mutex
	events []*models.AlertEvent
	rules  []*models.Alert
}

func (r *recordingDispatcher) Dispatch(_ context.Context, event *models.AlertEvent, rule *models.Alert) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	r.rules = append(r.rules, rule)
}

func (r *recordingDispatcher) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}

// newTestEngine builds an AlertEngine wired with a noop dispatcher, a rule
// source serving the given rules, and sensible defaults. It does not call Start.
func newTestEngine(rules []*models.Alert, dispatcher Dispatcher) *AlertEngine {
	if dispatcher == nil {
		dispatcher = noopDispatcher{}
	}
	if rules == nil {
		rules = make([]*models.Alert, 0)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &AlertEngine{
		ruleSource:               ruleSourceFunc(func(_ context.Context, _ *string) ([]*models.Alert, error) { return rules, nil }),
		dispatcher:               dispatcher,
		metricChannel:            make(chan *MetricData, 1000),
		workerPoolSize:           2,
		ctx:                      ctx,
		cancel:                   cancel,
		ruleCache:                rules,
		ruleCacheLastRefresh:     time.Now(),
		ruleCacheRefreshInterval: 60 * time.Second,
		webhookPushService:       webhook.NewPushService(nil, nil, "http://localhost:6532"),
	}
}

// --- config / stats tests (unchanged behaviour) -------------------------

func TestDefaultEngineConfig(t *testing.T) {
	cfg := DefaultEngineConfig()
	assert.Equal(t, 10, cfg.WorkerPoolSize)
	assert.Equal(t, 1000, cfg.MetricChannelBufferSize)
	assert.Equal(t, 60*time.Second, cfg.RuleCacheRefreshInterval)
}

func TestAlertEngine_GetStats(t *testing.T) {
	engine := &AlertEngine{
		ruleCache:            make([]*models.Alert, 0),
		ruleCacheLastRefresh: time.Time{},
		metricChannel:        make(chan *MetricData, 100),
	}
	stats := engine.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats["cached_rules"])
	assert.Equal(t, 0, stats["metric_channel_depth"])
	assert.Equal(t, 100, stats["metric_channel_capacity"])
}

func TestAlertEngine_GetStats_WithRules(t *testing.T) {
	nodeID := "node-1"
	engine := &AlertEngine{
		ruleCache: []*models.Alert{
			{ID: "rule-1", NodeID: &nodeID, Metric: MetricNameLatency, Threshold: 100, Level: "warning", Enabled: true},
			{ID: "rule-2", Metric: MetricNamePacketLoss, Threshold: 0.5, Level: "critical", Enabled: true},
		},
		ruleCacheLastRefresh: time.Now(),
		metricChannel:        make(chan *MetricData, 1000),
	}
	stats := engine.GetStats()
	assert.Equal(t, 2, stats["cached_rules"])
	assert.Equal(t, 1000, stats["metric_channel_capacity"])
}

// --- builder / channel tests -------------------------------------------

func TestAlertEngine_EvaluateMetrics_NonBlocking(t *testing.T) {
	engine := &AlertEngine{
		metricChannel: make(chan *MetricData, 10),
		dispatcher:    noopDispatcher{},
	}
	data := &MetricData{NodeID: "node-1", LatencyMs: 200, PacketLossRate: 0.1, JitterMs: 5, Timestamp: time.Now()}
	assert.True(t, engine.EvaluateMetrics(data))
}

func TestAlertEngine_EvaluateMetrics_ChannelFull(t *testing.T) {
	engine := &AlertEngine{
		metricChannel: make(chan *MetricData, 0),
		dispatcher:    noopDispatcher{},
	}
	data := &MetricData{NodeID: "node-1", LatencyMs: 200, Timestamp: time.Now()}
	assert.False(t, engine.EvaluateMetrics(data))
}

// --- evaluateMetric tests (now assert via the injected dispatcher) ------

func TestAlertEngine_evaluateMetric_NoRules(t *testing.T) {
	rec := &recordingDispatcher{}
	engine := newTestEngine(nil, rec)
	engine.evaluateMetric(&MetricData{NodeID: "node-1", LatencyMs: 50, Timestamp: time.Now()})
	assert.Equal(t, 0, rec.count(), "no dispatch with no rules")
}

func TestAlertEngine_evaluateMetric_DisabledRule(t *testing.T) {
	rule := &models.Alert{ID: "rule-disabled", Metric: MetricNameLatency, Threshold: 100, Level: "warning", Enabled: false}
	rec := &recordingDispatcher{}
	engine := newTestEngine([]*models.Alert{rule}, rec)
	engine.evaluateMetric(&MetricData{NodeID: "node-1", LatencyMs: 500, Timestamp: time.Now()})
	assert.Equal(t, 0, rec.count(), "disabled rules do not fire")
}

func TestAlertEngine_evaluateMetric_NodeSpecificRule_WrongNode(t *testing.T) {
	otherNode := "node-2"
	rule := &models.Alert{ID: "rule-1", NodeID: &otherNode, Metric: MetricNameLatency, Threshold: 100, Level: "warning", Enabled: true}
	rec := &recordingDispatcher{}
	engine := newTestEngine([]*models.Alert{rule}, rec)
	engine.evaluateMetric(&MetricData{NodeID: "node-1", LatencyMs: 500, Timestamp: time.Now()})
	assert.Equal(t, 0, rec.count(), "node-specific rule does not fire for another node")
}

func TestAlertEngine_evaluateMetric_FiresAndDispatches(t *testing.T) {
	rule := &models.Alert{ID: "rule-1", Metric: MetricNameLatency, Threshold: 100, Level: "warning", Enabled: true}
	rec := &recordingDispatcher{}
	engine := newTestEngine([]*models.Alert{rule}, rec)
	engine.evaluateMetric(&MetricData{NodeID: "node-1", LatencyMs: 500, Timestamp: time.Now()})

	require.Equal(t, 1, rec.count(), "rule fires once")
	assert.Equal(t, "node-1", rec.events[0].NodeID)
	assert.Equal(t, 500.0, rec.events[0].CurrentValue)
}

// --- rule cache tests ---------------------------------------------------

func TestAlertEngine_refreshRuleCache_WithMock(t *testing.T) {
	rules := []*models.Alert{{ID: "rule-1", Metric: MetricNameLatency, Enabled: true}}
	engine := &AlertEngine{
		ruleSource: ruleSourceFunc(func(_ context.Context, _ *string) ([]*models.Alert, error) { return rules, nil }),
		ruleCache:  make([]*models.Alert, 0),
		dispatcher: noopDispatcher{},
	}
	err := engine.refreshRuleCache(context.Background())
	require.NoError(t, err)
	assert.Len(t, engine.ruleCache, 1)
	assert.NotEqual(t, time.Time{}, engine.ruleCacheLastRefresh)
}

func TestAlertEngine_refreshRuleCache_Error(t *testing.T) {
	engine := &AlertEngine{
		ruleSource: ruleSourceFunc(func(_ context.Context, _ *string) ([]*models.Alert, error) { return nil, assert.AnError }),
		dispatcher: noopDispatcher{},
	}
	err := engine.refreshRuleCache(context.Background())
	assert.Error(t, err)
}

func TestAlertEngine_refreshRuleCache_NilRuleSource(t *testing.T) {
	engine := &AlertEngine{ruleCache: make([]*models.Alert, 0), dispatcher: noopDispatcher{}}
	assert.Panics(t, func() {
		_ = engine.refreshRuleCache(context.Background())
	})
}

// --- lifecycle test -----------------------------------------------------

func TestAlertEngine_Start_Stop(t *testing.T) {
	rules := []*models.Alert{{ID: "rule-1", Metric: MetricNameLatency, Threshold: 100, Level: "warning", Enabled: true}}
	engine := newTestEngine(rules, nil)

	assert.NotPanics(t, func() { engine.Start() })

	// Send a metric below threshold so no DB call occurs.
	engine.EvaluateMetrics(&MetricData{NodeID: "node-1", LatencyMs: 50, Timestamp: time.Now()})
	time.Sleep(10 * time.Millisecond)

	assert.NotPanics(t, func() { engine.Stop() })
}

// --- builder wiring test -----------------------------------------------

func TestAlertEngine_WithWebhookURLValidator(t *testing.T) {
	engine := newTestEngine(nil, nil)
	result := engine.WithWebhookURLValidator(func(_ string) error { return nil })
	assert.Equal(t, engine, result)
	result = engine.WithWebhookURLValidator(nil)
	assert.Equal(t, engine, result)
}
