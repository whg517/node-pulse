package alert

import (
	"context"
	"log/slog"
	"time"

	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/realtime"
)

// alertDispatchTimeout is the per-alert budget for the synchronous side of
// dispatch (suppression check + persistence + broadcast + suppression record).
// Webhook delivery uses its own longer timeout and runs asynchronously.
const alertDispatchTimeout = 5 * time.Second

// webhookDispatchTimeout bounds each asynchronous webhook delivery attempt.
const webhookDispatchTimeout = 30 * time.Second

// CompositeDispatcher is the default Dispatcher implementation. It coordinates
// the side-effects of a fired alert in the same order the engine historically
// performed them inline:
//
//  1. suppression check (fail-open on error)
//  2. persist event + tracking record
//  3. broadcast the record to realtime subscribers
//  4. record the suppression window for subsequent alerts
//  5. deliver webhooks asynchronously (non-blocking)
//
// Each collaborator is optional: a nil SuppressionChecker suppresses nothing,
// a nil Broadcaster skips the broadcast, and a nil WebhookPusher skips
// delivery. This mirrors the original nil-guarded behaviour.
type CompositeDispatcher struct {
	Suppression SuppressionChecker
	EventSink   EventSink
	Broadcaster Broadcaster
	Webhook     WebhookPusher
}

// Dispatch executes the alert side-effect pipeline for a single fired event.
// The rule is passed so its metric/node context can be used for suppression
// bookkeeping; the event already carries the evaluated values.
func (d *CompositeDispatcher) Dispatch(ctx context.Context, event *models.AlertEvent, rule *models.Alert) {
	dispatchCtx, cancel := context.WithTimeout(ctx, alertDispatchTimeout)
	defer cancel()

	// 1. Suppression check — fail open on error (alert still fires).
	if d.Suppression != nil {
		suppressed, err := d.Suppression.ShouldSuppress(dispatchCtx, event.NodeID, rule.Metric)
		if err != nil {
			slog.Error("Failed to check suppression",
				"node_id", event.NodeID,
				"metric", rule.Metric,
				"error", err)
			// Continue with alert creation on error (fail open)
		} else if suppressed {
			slog.Info("Alert suppressed",
				"node_id", event.NodeID,
				"metric", rule.Metric,
				"level", rule.Level,
				"threshold", rule.Threshold,
				"current_value", event.CurrentValue)
			return // Skip creating the alert event
		}
	}

	// 2. Persist the alert event + tracking record.
	record, err := d.EventSink.PersistAlert(dispatchCtx, event)
	if err != nil {
		slog.Error("Failed to persist alert",
			"node_id", event.NodeID,
			"metric", rule.Metric,
			"error", err)
		return
	}

	slog.Info("Alert event created",
		"alert_id", event.ID,
		"node_id", event.NodeID,
		"metric", event.Metric,
		"threshold", event.Threshold,
		"current_value", event.CurrentValue,
		"level", event.Level)

	// 3. Broadcast the new record to realtime subscribers.
	if d.Broadcaster != nil && record != nil {
		d.Broadcaster.BroadcastAlertRecord(realtime.EventAlertNew, record)
	}

	// 4. Record the suppression window for subsequent alerts of this node/metric.
	if d.Suppression != nil {
		if err := d.Suppression.RecordDefaultSuppression(dispatchCtx, event.NodeID, rule.Metric); err != nil {
			slog.Error("Failed to record suppression",
				"node_id", event.NodeID,
				"metric", rule.Metric,
				"error", err)
			// Don't fail the alert if suppression recording fails
		}
	}

	// 5. Deliver webhooks asynchronously so the worker is never blocked.
	if d.Webhook != nil {
		go d.deliverWebhook(event)
	}
}

// deliverWebhook sends the alert to all configured webhooks with a bounded
// timeout. Failures are logged but never propagate.
func (d *CompositeDispatcher) deliverWebhook(event *models.AlertEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), webhookDispatchTimeout)
	defer cancel()

	if err := d.Webhook.SendAlert(ctx, event); err != nil {
		slog.Error("Webhook push failed",
			"alert_id", event.ID,
			"error", err)
	}
}
