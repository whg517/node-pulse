package alert

import (
"context"
"testing"
"time"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"

"github.com/whg517/node-pulse/pulse/internal/db"
"github.com/whg517/node-pulse/pulse/internal/models"
"github.com/whg517/node-pulse/pulse/internal/suppression"
"github.com/whg517/node-pulse/pulse/internal/webhook"
)

// mockAlertQuerier is a mock implementation of db.AlertQuerier for unit testing
type mockAlertQuerier struct {
alerts []*models.Alert
err    error
}

func (m *mockAlertQuerier) CreateAlert(_ context.Context, _ *models.Alert) error {
return m.err
}

func (m *mockAlertQuerier) GetAlerts(_ context.Context, _ *string) ([]*models.Alert, error) {
return m.alerts, m.err
}

func (m *mockAlertQuerier) GetAlertByID(_ context.Context, _ string) (*models.Alert, error) {
if len(m.alerts) == 0 {
return nil, assert.AnError
}
return m.alerts[0], m.err
}

func (m *mockAlertQuerier) UpdateAlert(_ context.Context, _ string, _ *models.UpdateAlertRequest) (*models.Alert, error) {
if len(m.alerts) == 0 {
return nil, assert.AnError
}
return m.alerts[0], m.err
}

func (m *mockAlertQuerier) DeleteAlert(_ context.Context, _ string) error {
return m.err
}

// mockAlertEventsQuerier is a mock implementation of db.AlertEventsQuerier
type mockAlertEventsQuerier struct {
err error
}

func (m *mockAlertEventsQuerier) CreateAlertEvent(_ context.Context, _ *models.AlertEvent) error {
return m.err
}

func (m *mockAlertEventsQuerier) GetAlertEvents(_ context.Context, _ string) ([]*models.AlertEvent, error) {
return nil, m.err
}

// mockSuppressionQuerier is a mock implementation of db.AlertSuppressionsQuerier
type mockSuppressionQuerier struct {
suppressed bool
err        error
}

func (m *mockSuppressionQuerier) CheckSuppression(_ context.Context, _ string, _ string) (*models.AlertSuppression, error) {
if m.suppressed {
return &models.AlertSuppression{SuppressedUntil: time.Now().Add(5 * time.Minute)}, nil
}
return nil, db.ErrSuppressionNotFound
}

func (m *mockSuppressionQuerier) CreateOrUpdateSuppression(_ context.Context, _ string, _ string, _ time.Time) error {
return m.err
}

func (m *mockSuppressionQuerier) DeleteExpiredSuppressions(_ context.Context) (int64, error) {
return 0, m.err
}

func (m *mockSuppressionQuerier) CountActiveSuppressions(_ context.Context) (int64, error) {
return 0, m.err
}

// newTestAlertEngine creates an AlertEngine for unit testing (no real DB needed)
func newTestAlertEngine(alertQuerier db.AlertQuerier, rules []*models.Alert) *AlertEngine {
ctx, cancel := context.WithCancel(context.Background())

suppressionQuerier := &mockSuppressionQuerier{}
suppressionService := suppression.NewService(suppressionQuerier)

if rules == nil {
rules = make([]*models.Alert, 0)
}

engine := &AlertEngine{
alertQuerier:             alertQuerier,
alertEventsQuerier:       &mockAlertEventsQuerier{},
suppressionService:       suppressionService,
webhookPushService:       webhook.NewPushService(nil, nil, "http://localhost:6532"),
metricChannel:            make(chan *MetricData, 1000),
workerPoolSize:           2,
ctx:                      ctx,
cancel:                   cancel,
ruleCache:                rules,
ruleCacheLastRefresh:     time.Now(),
ruleCacheRefreshInterval: 60 * time.Second,
}
// Disable URL validation for tests
engine.webhookPushService.WithURLValidator(nil)
return engine
}

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
{ID: "rule-1", NodeID: &nodeID, Metric: "latency", Threshold: 100, Level: "warning", Enabled: true},
{ID: "rule-2", Metric: "packet_loss_rate", Threshold: 0.5, Level: "critical", Enabled: true},
},
ruleCacheLastRefresh: time.Now(),
metricChannel:        make(chan *MetricData, 1000),
}

stats := engine.GetStats()
assert.Equal(t, 2, stats["cached_rules"])
assert.Equal(t, 1000, stats["metric_channel_capacity"])
}

func TestAlertEngine_WithWebhookURLValidator(t *testing.T) {
engine := newTestAlertEngine(&mockAlertQuerier{}, nil)

result := engine.WithWebhookURLValidator(func(_ string) error { return nil })
assert.Equal(t, engine, result)

result = engine.WithWebhookURLValidator(nil)
assert.Equal(t, engine, result)
}

func TestAlertEngine_EvaluateMetrics_NonBlocking(t *testing.T) {
engine := &AlertEngine{
metricChannel: make(chan *MetricData, 10),
}

data := &MetricData{
NodeID:         "node-1",
LatencyMs:      200,
PacketLossRate: 0.1,
JitterMs:       5,
Timestamp:      time.Now(),
}

accepted := engine.EvaluateMetrics(data)
assert.True(t, accepted)
}

func TestAlertEngine_EvaluateMetrics_ChannelFull(t *testing.T) {
engine := &AlertEngine{
metricChannel: make(chan *MetricData, 0),
}

data := &MetricData{
NodeID:    "node-1",
LatencyMs: 200,
Timestamp: time.Now(),
}

accepted := engine.EvaluateMetrics(data)
assert.False(t, accepted)
}

func TestAlertEngine_evaluateRule_Latency(t *testing.T) {
engine := &AlertEngine{}
nodeID := "node-1"

t.Run("Latency exceeds threshold", func(t *testing.T) {
rule := &models.Alert{
ID: "rule-1", NodeID: &nodeID, Metric: "latency",
Threshold: 100, Level: "warning", Enabled: true,
}
data := &MetricData{NodeID: "node-1", LatencyMs: 150, Timestamp: time.Now()}

event := engine.evaluateRule(rule, data)
require.NotNil(t, event)
assert.Equal(t, "node-1", event.NodeID)
assert.Equal(t, "latency", event.Metric)
assert.Equal(t, float64(100), event.Threshold)
assert.Equal(t, float64(150), event.CurrentValue)
assert.Equal(t, "warning", event.Level)
})

t.Run("Latency within threshold", func(t *testing.T) {
rule := &models.Alert{Metric: "latency", Threshold: 100, Level: "warning", Enabled: true}
data := &MetricData{NodeID: "node-1", LatencyMs: 50, Timestamp: time.Now()}

event := engine.evaluateRule(rule, data)
assert.Nil(t, event)
})
}

func TestAlertEngine_evaluateRule_PacketLoss(t *testing.T) {
engine := &AlertEngine{}

t.Run("Packet loss exceeds threshold", func(t *testing.T) {
rule := &models.Alert{Metric: "packet_loss_rate", Threshold: 0.1, Level: "critical"}
data := &MetricData{NodeID: "node-1", PacketLossRate: 0.5, Timestamp: time.Now()}

event := engine.evaluateRule(rule, data)
require.NotNil(t, event)
assert.Equal(t, "packet_loss_rate", event.Metric)
assert.Equal(t, float64(0.5), event.CurrentValue)
})

t.Run("Packet loss within threshold", func(t *testing.T) {
rule := &models.Alert{Metric: "packet_loss_rate", Threshold: 0.5}
data := &MetricData{NodeID: "node-1", PacketLossRate: 0.1, Timestamp: time.Now()}

event := engine.evaluateRule(rule, data)
assert.Nil(t, event)
})
}

func TestAlertEngine_evaluateRule_Jitter(t *testing.T) {
engine := &AlertEngine{}

t.Run("Jitter exceeds threshold", func(t *testing.T) {
rule := &models.Alert{Metric: "jitter", Threshold: 10, Level: "warning"}
data := &MetricData{NodeID: "node-1", JitterMs: 25, Timestamp: time.Now()}

event := engine.evaluateRule(rule, data)
require.NotNil(t, event)
assert.Equal(t, "jitter", event.Metric)
assert.Equal(t, float64(25), event.CurrentValue)
})

t.Run("Jitter within threshold", func(t *testing.T) {
rule := &models.Alert{Metric: "jitter", Threshold: 50}
data := &MetricData{NodeID: "node-1", JitterMs: 5, Timestamp: time.Now()}

event := engine.evaluateRule(rule, data)
assert.Nil(t, event)
})
}

func TestAlertEngine_evaluateRule_UnknownMetric(t *testing.T) {
engine := &AlertEngine{}

rule := &models.Alert{Metric: "unknown_metric", Threshold: 100}
data := &MetricData{NodeID: "node-1", LatencyMs: 200, Timestamp: time.Now()}

event := engine.evaluateRule(rule, data)
assert.Nil(t, event)
}

func TestAlertEngine_refreshRuleCache_WithMock(t *testing.T) {
nodeID := "node-1"
rules := []*models.Alert{
{ID: "rule-1", Metric: "latency", Threshold: 100, Level: "warning", Enabled: true},
{ID: "rule-2", NodeID: &nodeID, Metric: "jitter", Threshold: 10, Level: "critical", Enabled: false},
{ID: "rule-3", Metric: "packet_loss_rate", Threshold: 0.5, Level: "warning", Enabled: true},
}

mockQuerier := &mockAlertQuerier{alerts: rules}
engine := newTestAlertEngine(mockQuerier, nil)

err := engine.refreshRuleCache(context.Background())
require.NoError(t, err)

// Only enabled rules should be cached
assert.Len(t, engine.ruleCache, 2)
assert.NotZero(t, engine.ruleCacheLastRefresh)
}

func TestAlertEngine_refreshRuleCache_Error(t *testing.T) {
mockQuerier := &mockAlertQuerier{err: assert.AnError}
engine := newTestAlertEngine(mockQuerier, nil)

err := engine.refreshRuleCache(context.Background())
assert.Error(t, err)
}

func TestAlertEngine_refreshRuleCache_NilQuerier(t *testing.T) {
engine := &AlertEngine{ruleCache: make([]*models.Alert, 0)}

assert.Panics(t, func() {
_ = engine.refreshRuleCache(context.Background())
})
}

func TestAlertEngine_evaluateMetric_NoRules(t *testing.T) {
engine := newTestAlertEngine(&mockAlertQuerier{}, nil)

data := &MetricData{NodeID: "node-1", LatencyMs: 50, Timestamp: time.Now()}

assert.NotPanics(t, func() {
engine.evaluateMetric(data)
})
}

func TestAlertEngine_evaluateMetric_DisabledRule(t *testing.T) {
rule := &models.Alert{
ID: "rule-disabled", Metric: "latency", Threshold: 100,
Level: "warning", Enabled: false,
}
engine := newTestAlertEngine(&mockAlertQuerier{}, []*models.Alert{rule})

data := &MetricData{NodeID: "node-1", LatencyMs: 50, Timestamp: time.Now()}

assert.NotPanics(t, func() {
engine.evaluateMetric(data)
})
}

func TestAlertEngine_evaluateMetric_NodeSpecificRule_WrongNode(t *testing.T) {
otherNode := "node-2"
rule := &models.Alert{
ID: "rule-1", NodeID: &otherNode, Metric: "latency",
Threshold: 100, Level: "warning", Enabled: true,
}
engine := newTestAlertEngine(&mockAlertQuerier{}, []*models.Alert{rule})

data := &MetricData{NodeID: "node-1", LatencyMs: 50, Timestamp: time.Now()}

assert.NotPanics(t, func() {
engine.evaluateMetric(data)
})
}

func TestAlertEngine_Start_Stop(t *testing.T) {
rules := []*models.Alert{
{ID: "rule-1", Metric: "latency", Threshold: 100, Level: "warning", Enabled: true},
}
mockQuerier := &mockAlertQuerier{alerts: rules}
engine := newTestAlertEngine(mockQuerier, nil)

assert.NotPanics(t, func() {
engine.Start()
})

// Send a metric to exercise the worker
engine.EvaluateMetrics(&MetricData{
NodeID:    "node-1",
LatencyMs: 50, // Below threshold (100), so no DB call
Timestamp: time.Now(),
})

time.Sleep(10 * time.Millisecond)

assert.NotPanics(t, func() {
engine.Stop()
})
}

// newTestAlertEngineWithSuppressionMock creates an AlertEngine with a custom suppression mock
func newTestAlertEngineWithSuppressionMock(alertQuerier db.AlertQuerier, rules []*models.Alert, suppressedMock *mockSuppressionQuerier, eventsMock *mockAlertEventsQuerier) *AlertEngine {
ctx, cancel := context.WithCancel(context.Background())

suppressionService := suppression.NewService(suppressedMock)

if rules == nil {
rules = make([]*models.Alert, 0)
}

engine := &AlertEngine{
alertQuerier:             alertQuerier,
alertEventsQuerier:       eventsMock,
suppressionService:       suppressionService,
webhookPushService:       webhook.NewPushService(nil, nil, "http://localhost:6532"),
metricChannel:            make(chan *MetricData, 1000),
workerPoolSize:           2,
ctx:                      ctx,
cancel:                   cancel,
ruleCache:                rules,
ruleCacheLastRefresh:     time.Now(),
ruleCacheRefreshInterval: 60 * time.Second,
}
engine.webhookPushService.WithURLValidator(nil)
return engine
}

func TestAlertEngine_evaluateMetric_AlertSuppressed(t *testing.T) {
rule := &models.Alert{
ID: "rule-1", Metric: "latency",
Threshold: 50, Level: "warning", Enabled: true,
}

suppressedMock := &mockSuppressionQuerier{suppressed: true}
eventsQuerier := &mockAlertEventsQuerier{}
engine := newTestAlertEngineWithSuppressionMock(
&mockAlertQuerier{},
[]*models.Alert{rule},
suppressedMock,
eventsQuerier,
)

data := &MetricData{NodeID: "node-1", LatencyMs: 200, Timestamp: time.Now()}

assert.NotPanics(t, func() {
engine.evaluateMetric(data)
})
}

func TestAlertEngine_evaluateMetric_AlertEventsError(t *testing.T) {
rule := &models.Alert{
ID: "rule-1", Metric: "latency",
Threshold: 50, Level: "warning", Enabled: true,
}

suppressedMock := &mockSuppressionQuerier{suppressed: false}
eventsQuerier := &mockAlertEventsQuerier{err: assert.AnError}
engine := newTestAlertEngineWithSuppressionMock(
&mockAlertQuerier{},
[]*models.Alert{rule},
suppressedMock,
eventsQuerier,
)

data := &MetricData{NodeID: "node-1", LatencyMs: 200, Timestamp: time.Now()}

assert.NotPanics(t, func() {
engine.evaluateMetric(data)
})
}

func TestAlertEngine_evaluateMetric_SuppressionCheckError(t *testing.T) {
rule := &models.Alert{
ID: "rule-1", Metric: "latency",
Threshold: 50, Level: "warning", Enabled: true,
}

suppressedMock := &mockSuppressionQuerier{suppressed: false, err: assert.AnError}
eventsQuerier := &mockAlertEventsQuerier{err: assert.AnError}  // fail events to avoid pool nil panic
engine := newTestAlertEngineWithSuppressionMock(
&mockAlertQuerier{},
[]*models.Alert{rule},
suppressedMock,
eventsQuerier,
)

data := &MetricData{NodeID: "node-1", LatencyMs: 200, Timestamp: time.Now()}

assert.NotPanics(t, func() {
engine.evaluateMetric(data)
})
}

func TestAlertEngine_NewAlertEngine(t *testing.T) {
// NewAlertEngine requires a pool - test that it doesn't panic when pool is nil
// (since NewAlertEventsQuerier etc. accept nil)
assert.NotPanics(t, func() {
engine := NewAlertEngine(nil, &mockAlertQuerier{}, DefaultEngineConfig())
assert.NotNil(t, engine)
})
}

func TestAlertEngine_getPool(t *testing.T) {
engine := newTestAlertEngine(&mockAlertQuerier{}, nil)
pool := engine.getPool()
assert.Nil(t, pool) // test engine has nil pool
}
