package alert

import (
	"context"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// Ports (interfaces) defining the collaborators the alert pipeline needs.
// The engine depends on these abstractions rather than concrete types, so
// each side-effect (persistence, suppression, notification, broadcasting)
// can be unit-tested and replaced independently.

// RuleSource provides the set of alert rules the engine evaluates against.
type RuleSource interface {
	GetAlerts(ctx context.Context, nodeID *string) ([]*models.Alert, error)
}

// SuppressionChecker decides whether an alert for a node/metric should be
// held back, and records a suppression window after an alert fires.
type SuppressionChecker interface {
	ShouldSuppress(ctx context.Context, nodeID string, metric string) (bool, error)
	RecordDefaultSuppression(ctx context.Context, nodeID string, metric string) error
}

// EventSink persists the lifecycle artefacts of a fired alert: the event
// itself and the tracking record. PersistAlert returns the created record
// (with its ID populated) so callers can broadcast it.
type EventSink interface {
	PersistAlert(ctx context.Context, event *models.AlertEvent) (*models.AlertRecord, error)
}

// Broadcaster pushes alert lifecycle events to realtime subscribers (e.g. the
// dashboard websocket stream). Implementations must be nil-safe.
type Broadcaster interface {
	BroadcastAlertRecord(eventType string, record *models.AlertRecord)
}

// WebhookPusher delivers an alert event to external webhook endpoints.
type WebhookPusher interface {
	SendAlert(ctx context.Context, alertEvent *models.AlertEvent) error
}

// EmailNotifier fans an alert out as email to every user whose server-side
// notification preferences subscribe them to the alert's severity level.
// Used by CompositeDispatcher as step 6 (after webhooks). Implementations
// must be nil-safe (a nil EmailNotifier skips email entirely).
type EmailNotifier interface {
	NotifyAlertSubscribers(ctx context.Context, event *models.AlertEvent) error
}

// Dispatcher is the single orchestration seam the engine calls when a rule
// fires. It coordinates suppression, persistence, broadcasting and webhook
// delivery, keeping the engine free of side-effect wiring.
type Dispatcher interface {
	Dispatch(ctx context.Context, event *models.AlertEvent, rule *models.Alert)
}

// noopDispatcher drops every alert; used as a safe default and in tests that
// only exercise evaluation/channel mechanics.
type noopDispatcher struct{}

func (noopDispatcher) Dispatch(context.Context, *models.AlertEvent, *models.Alert) {}
