// Package notify provides outbound email delivery for features that need to
// reach users out-of-band (password-reset links, scheduled report attachments).
//
// It defines a Sender interface backed by SMTP in production and a log-only
// NoopSender when SMTP is not configured, so the rest of the system can depend
// on Sender unconditionally without branching on availability.
package notify

import "context"

// Attachment is a file attached to an outbound message.
type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string // e.g. "text/csv", "application/pdf"
}

// Sender delivers an email message to one or more recipients.
type Sender interface {
	// Send delivers a message. `to` may be a comma-separated list of addresses.
	// Attachments are optional. Returns nil on accepted-for-delivery.
	Send(ctx context.Context, to, subject, body string, attachments ...Attachment) error
	// Configured reports whether a real transport is wired (vs the noop sender).
	Configured() bool
}

// New returns a Sender backed by SMTP when host is set, otherwise a NoopSender
// that logs the would-be message. This keeps callers simple: they always call
// Send and trust it to do the right thing per environment.
func New(cfg SMTPConfig) Sender {
	if cfg.Host == "" {
		return &NoopSender{}
	}
	return NewSMTPSender(cfg)
}
