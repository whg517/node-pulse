package notify

import (
	"context"
	"log/slog"
)

// NoopSender is a log-only Sender used when SMTP is not configured (dev/test or
// environments without outbound mail). It never errors so callers can stay simple.
type NoopSender struct{}

// Configured is always false for the noop sender.
func (n *NoopSender) Configured() bool { return false }

// Send logs the message and returns nil.
func (n *NoopSender) Send(ctx context.Context, to, subject, body string, attachments ...Attachment) error {
	slog.Warn("Email not sent (SMTP not configured); message logged only",
		"component", "notify", "to", to, "subject", subject,
		"attachments", len(attachments), "backend", "noop")
	return nil
}
