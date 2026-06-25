package alert

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/realtime"
)

// --- mock ports for dispatcher tests ------------------------------------

type fakeSuppression struct {
	mu           sync.Mutex
	suppressed   bool
	suppressErr  error
	recordErr    error
	shouldCalled int
	recordCalled int
	lastNode     string
	lastMetric   string
}

func (f *fakeSuppression) ShouldSuppress(_ context.Context, nodeID, metric string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.shouldCalled++
	f.lastNode = nodeID
	f.lastMetric = metric
	return f.suppressed, f.suppressErr
}

func (f *fakeSuppression) RecordDefaultSuppression(_ context.Context, nodeID, metric string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recordCalled++
	f.lastNode = nodeID
	f.lastMetric = metric
	return f.recordErr
}

type fakeEventSink struct {
	mu         sync.Mutex
	persist    int
	persistErr error
	records    []*models.AlertRecord
}

func (f *fakeEventSink) PersistAlert(_ context.Context, event *models.AlertEvent) (*models.AlertRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.persist++
	if f.persistErr != nil {
		return nil, f.persistErr
	}
	record := &models.AlertRecord{
		ID:           "rec-" + event.ID,
		AlertEventID: event.ID,
		NodeID:       event.NodeID,
		Metric:       event.Metric,
		Level:        event.Level,
		Status:       "pending",
	}
	f.records = append(f.records, record)
	return record, nil
}

type fakeBroadcaster struct {
	mu       sync.Mutex
	calls    int
	lastType string
	lastRec  *models.AlertRecord
}

func (f *fakeBroadcaster) BroadcastAlertRecord(eventType string, record *models.AlertRecord) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.lastType = eventType
	f.lastRec = record
}

type fakeWebhook struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (f *fakeWebhook) SendAlert(_ context.Context, _ *models.AlertEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.err
}

// --- tests --------------------------------------------------------------

// TestDispatcher_HappyPath verifies the full orchestration fires in order:
// suppression check → persist → broadcast → record suppression → webhook.
func TestDispatcher_HappyPath(t *testing.T) {
	sup := &fakeSuppression{}
	sink := &fakeEventSink{}
	bcast := &fakeBroadcaster{}
	wh := &fakeWebhook{}
	d := &CompositeDispatcher{Suppression: sup, EventSink: sink, Broadcaster: bcast, Webhook: wh}

	event := &models.AlertEvent{ID: "evt-1", NodeID: "node-1", Metric: MetricNameLatency, Level: "P1"}
	rule := &models.Alert{ID: "rule-1", Metric: MetricNameLatency, Threshold: 100}

	d.Dispatch(context.Background(), event, rule)

	// Give the async webhook goroutine time to run.
	waitFor(t, func() bool { return wh.calls == 1 })

	assert.Equal(t, 1, sup.shouldCalled, "suppression checked once")
	assert.Equal(t, 1, sink.persist, "event persisted once")
	assert.Equal(t, 1, bcast.calls, "broadcast once")
	assert.Equal(t, realtime.EventAlertNew, bcast.lastType)
	assert.Equal(t, "rec-evt-1", bcast.lastRec.ID, "broadcast carried the persisted record id")
	assert.Equal(t, 1, sup.recordCalled, "suppression window recorded")
	assert.Equal(t, 1, wh.calls, "webhook delivered once")
}

// TestDispatcher_Suppressed returns early: no persist, no broadcast, no webhook.
func TestDispatcher_Suppressed(t *testing.T) {
	sup := &fakeSuppression{suppressed: true}
	sink := &fakeEventSink{}
	bcast := &fakeBroadcaster{}
	wh := &fakeWebhook{}
	d := &CompositeDispatcher{Suppression: sup, EventSink: sink, Broadcaster: bcast, Webhook: wh}

	event := &models.AlertEvent{ID: "evt-1", NodeID: "node-1", Metric: MetricNameLatency}
	rule := &models.Alert{ID: "rule-1", Metric: MetricNameLatency}

	d.Dispatch(context.Background(), event, rule)

	assert.Equal(t, 1, sup.shouldCalled)
	assert.Equal(t, 0, sink.persist, "nothing persisted when suppressed")
	assert.Equal(t, 0, bcast.calls)
	assert.Equal(t, 0, wh.calls)
}

// TestDispatcher_SuppressionCheckError fails open: the alert still fires.
func TestDispatcher_SuppressionCheckError(t *testing.T) {
	sup := &fakeSuppression{suppressErr: assert.AnError}
	sink := &fakeEventSink{}
	bcast := &fakeBroadcaster{}
	wh := &fakeWebhook{}
	d := &CompositeDispatcher{Suppression: sup, EventSink: sink, Broadcaster: bcast, Webhook: wh}

	event := &models.AlertEvent{ID: "evt-1", NodeID: "node-1", Metric: MetricNameLatency}
	rule := &models.Alert{ID: "rule-1", Metric: MetricNameLatency}

	d.Dispatch(context.Background(), event, rule)
	waitFor(t, func() bool { return wh.calls == 1 })

	assert.Equal(t, 1, sink.persist, "fail-open: alert persisted despite suppression check error")
	assert.Equal(t, 1, wh.calls)
}

// TestDispatcher_PersistError aborts: no broadcast, no webhook.
func TestDispatcher_PersistError(t *testing.T) {
	sup := &fakeSuppression{}
	sink := &fakeEventSink{persistErr: assert.AnError}
	bcast := &fakeBroadcaster{}
	wh := &fakeWebhook{}
	d := &CompositeDispatcher{Suppression: sup, EventSink: sink, Broadcaster: bcast, Webhook: wh}

	event := &models.AlertEvent{ID: "evt-1", NodeID: "node-1", Metric: MetricNameLatency}
	rule := &models.Alert{ID: "rule-1", Metric: MetricNameLatency}

	d.Dispatch(context.Background(), event, rule)

	assert.Equal(t, 0, bcast.calls, "no broadcast when persist failed")
	assert.Equal(t, 0, wh.calls, "no webhook when persist failed")
	assert.Equal(t, 0, sup.recordCalled, "no suppression recorded when persist failed")
}

// TestDispatcher_NilCollaborators does not panic when optional ports are nil.
func TestDispatcher_NilCollaborators(t *testing.T) {
	sink := &fakeEventSink{}
	d := &CompositeDispatcher{EventSink: sink} // Suppression/Broadcaster/Webhook nil

	event := &models.AlertEvent{ID: "evt-1", NodeID: "node-1", Metric: MetricNameLatency}
	rule := &models.Alert{ID: "rule-1", Metric: MetricNameLatency}

	assert.NotPanics(t, func() {
		d.Dispatch(context.Background(), event, rule)
	})
	assert.Equal(t, 1, sink.persist, "event still persisted with nil optional ports")
}

// waitFor polls a condition for up to 200ms, useful for the async webhook step.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	require.Fail(t, "condition never became true within 200ms")
}
