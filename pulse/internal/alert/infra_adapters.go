package alert

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/realtime"
	"github.com/whg517/node-pulse/pulse/internal/suppression"
	"github.com/whg517/node-pulse/pulse/internal/webhook"
)

// This file holds the thin adapters that let the concrete infrastructure
// types satisfy the alert ports. They live in the alert package so the
// engine and its tests can compose them without leaking infrastructure
// concerns into the domain.

// --- RuleSource adapter -------------------------------------------------

// ruleSourceAdapter adapts db.AlertQuerier to the RuleSource port.
type ruleSourceAdapter struct {
	q db.AlertQuerier
}

func (a ruleSourceAdapter) GetAlerts(ctx context.Context, nodeID *string) ([]*models.Alert, error) {
	return a.q.GetAlerts(ctx, nodeID)
}

// NewRuleSource wraps a db.AlertQuerier as a RuleSource.
func NewRuleSource(q db.AlertQuerier) RuleSource { return ruleSourceAdapter{q: q} }

// --- SuppressionChecker adapter ----------------------------------------

// suppressionAdapter adapts *suppression.Service to the SuppressionChecker port.
type suppressionAdapter struct {
	s *suppression.Service
}

func (a suppressionAdapter) ShouldSuppress(ctx context.Context, nodeID, metric string) (bool, error) {
	return a.s.ShouldSuppress(ctx, nodeID, metric)
}

func (a suppressionAdapter) RecordDefaultSuppression(ctx context.Context, nodeID, metric string) error {
	return a.s.RecordDefaultSuppression(ctx, nodeID, metric)
}

// NewSuppressionChecker wraps a *suppression.Service as a SuppressionChecker.
func NewSuppressionChecker(s *suppression.Service) SuppressionChecker {
	return suppressionAdapter{s: s}
}

// --- EventSink adapter --------------------------------------------------

// eventSinkAdapter persists the alert event and its lifecycle tracking record.
// CreateAlertRecord mutates the record in place (sets ID/CreatedAt/UpdatedAt),
// which is why PersistAlert returns the same pointer.
type eventSinkAdapter struct {
	events db.AlertEventsQuerier
	pool   *pgxpool.Pool
}

func (a eventSinkAdapter) PersistAlert(ctx context.Context, event *models.AlertEvent) (*models.AlertRecord, error) {
	if err := a.events.CreateAlertEvent(ctx, event); err != nil {
		return nil, err
	}

	record := &models.AlertRecord{
		AlertEventID: event.ID,
		NodeID:       event.NodeID,
		Metric:       event.Metric,
		Level:        event.Level,
		Status:       "pending", // initial lifecycle status
	}
	if err := db.CreateAlertRecord(ctx, a.pool, record); err != nil {
		slog.Error("Failed to create alert record",
			"alert_event_id", event.ID, "node_id", event.NodeID, "metric", event.Metric, "error", err)
		// The event was persisted; return the partial record so the caller can
		// still broadcast it, matching the historical fail-and-continue behaviour.
		return record, nil
	}
	slog.Debug("Alert record created", "record_id", record.ID, "alert_event_id", event.ID)
	return record, nil
}

// NewEventSink wires the event + record writers behind a single EventSink port.
func NewEventSink(events db.AlertEventsQuerier, pool *pgxpool.Pool) EventSink {
	return eventSinkAdapter{events: events, pool: pool}
}

// --- Broadcaster adapter ------------------------------------------------

// broadcasterAdapter adapts *realtime.Hub to the Broadcaster port. The Hub's
// BroadcastAlertRecord is already nil-safe, so this just forwards.
type broadcasterAdapter struct {
	h *realtime.Hub
}

func (a broadcasterAdapter) BroadcastAlertRecord(eventType string, record *models.AlertRecord) {
	a.h.BroadcastAlertRecord(eventType, record)
}

// NewBroadcaster wraps a *realtime.Hub as a Broadcaster. Returns nil when the
// hub is nil so the dispatcher can skip broadcasting via a nil check.
func NewBroadcaster(h *realtime.Hub) Broadcaster {
	if h == nil {
		return nil
	}
	return broadcasterAdapter{h: h}
}

// --- WebhookPusher adapter ---------------------------------------------

// webhookAdapter adapts *webhook.PushService to the WebhookPusher port.
type webhookAdapter struct {
	p *webhook.PushService
}

func (a webhookAdapter) SendAlert(ctx context.Context, event *models.AlertEvent) error {
	return a.p.SendAlert(ctx, event)
}

// NewWebhookPusher wraps a *webhook.PushService as a WebhookPusher.
func NewWebhookPusher(p *webhook.PushService) WebhookPusher {
	return webhookAdapter{p: p}
}
