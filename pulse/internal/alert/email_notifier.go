package alert

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/notify"
)

// EmailNotifierImpl is the F4 Phase 2 EmailNotifier. It queries each user's
// server-side notification preferences, resolves the destination email
// (preference override or the user's profile address), and sends one message
// per subscriber.
//
// It depends on two repositories + the notify.Sender:
//   - NotificationPrefsRepository: who wants email, at what severity floor
//   - UserQuerier: resolve user_id → profile email when no override is set
//   - Sender: the SMTP (or noop) transport
//
// When the Sender is not Configured (e.g. dev mode with no SMTP), the
// notifier short-circuits and logs a debug line — no error, since this is
// an expected state, not a failure.
type EmailNotifierImpl struct {
	prefs   db.NotificationPrefsRepository
	users   db.UserQuerier
	sender  notify.Sender
	baseURL string // included in the email body as a deep link
}

// NewEmailNotifier constructs the notifier. A nil sender produces a notifier
// that always no-ops (useful when SMTP isn't configured but wiring must exist).
func NewEmailNotifier(prefs db.NotificationPrefsRepository, users db.UserQuerier, sender notify.Sender, baseURL string) *EmailNotifierImpl {
	return &EmailNotifierImpl{prefs: prefs, users: users, sender: sender, baseURL: baseURL}
}

// NotifyAlertSubscribers implements EmailNotifier. It queries subscribers at
// or below the alert's severity, resolves each one's email, and sends a
// single message per recipient. Per-recipient failures are logged and
// skipped (one bad address shouldn't block the rest).
func (n *EmailNotifierImpl) NotifyAlertSubscribers(ctx context.Context, event *models.AlertEvent) error {
	if n.sender == nil || !n.sender.Configured() {
		slog.Debug("Email notifier skipping (SMTP not configured)", "alert_id", event.ID)
		return nil
	}

	subs, err := n.prefs.ListSubscribersForLevel(ctx, event.Level)
	if err != nil {
		return fmt.Errorf("query subscribers: %w", err)
	}
	if len(subs) == 0 {
		return nil // nobody opted in at this level
	}

	subject := fmt.Sprintf("[NodePulse] %s alert: %s on node %s", event.Level, event.Metric, event.NodeID)
	body := fmt.Sprintf(
		"Alert ID:    %s\nLevel:        %s\nNode ID:      %s\nMetric:       %s\nThreshold:    %.2f\nCurrent:      %.2f\nTriggered at: %s\n\nView: %s/nodes/%s\n",
		event.ID, event.Level, event.NodeID, event.Metric,
		event.Threshold, event.CurrentValue, event.CreatedAt.Format("2006-01-02 15:04:05 MST"),
		n.baseURL, event.NodeID,
	)

	sent := 0
	for _, sub := range subs {
		addr := n.resolveEmail(ctx, sub)
		if addr == "" {
			continue
		}
		if err := n.sender.Send(ctx, addr, subject, body); err != nil {
			slog.Error("Failed to send alert email",
				"alert_id", event.ID, "user_id", sub.UserID, "error", err)
			continue
		}
		sent++
	}

	if sent > 0 {
		slog.Info("Alert email fan-out complete",
			"alert_id", event.ID, "recipients", sent, "level", event.Level)
	}
	return nil
}

// resolveEmail picks the subscriber's override address if set, otherwise
// falls back to the profile email from the users table. Returns "" if no
// usable address exists (the user is silently skipped).
func (n *EmailNotifierImpl) resolveEmail(ctx context.Context, sub *models.NotificationPrefs) string {
	if sub.NotifyEmail != nil && strings.TrimSpace(*sub.NotifyEmail) != "" {
		return *sub.NotifyEmail
	}
	uid, err := uuid.Parse(sub.UserID)
	if err != nil {
		slog.Error("Failed to parse user_id for notification email",
			"user_id", sub.UserID, "error", err)
		return ""
	}
	user, err := n.users.GetUserByID(ctx, uid)
	if err != nil {
		slog.Error("Failed to look up user for notification email",
			"user_id", sub.UserID, "error", err)
		return ""
	}
	if user.Email == nil || *user.Email == "" {
		return ""
	}
	return *user.Email
}
