package notify

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// SMTPConfig configures the SMTP transport. Loaded from the `notify.smtp`
// config section (env override prefix PULSE_NOTIFY_SMTP_*).
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	// TLS: when true (default for port 465/587), a STARTTLS/TLS handshake is
	// attempted before auth.
	UseTLS bool `yaml:"use_tls"`
}

// SMTPSender delivers mail via a plain SMTP server.
type SMTPSender struct {
	cfg SMTPConfig
}

// NewSMTPSender builds an SMTP-backed Sender. Defaults: port 587, TLS on.
func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	// Default TLS on unless explicitly disabled.
	cfg.UseTLS = true
	return &SMTPSender{cfg: cfg}
}

// Configured is always true for the SMTP sender.
func (s *SMTPSender) Configured() bool { return true }

// Send delivers one message. RFC 5322 via net/smtp.
func (s *SMTPSender) Send(ctx context.Context, to, subject, body string, attachments ...Attachment) error {
	if s.cfg.From == "" {
		return errors.New("notify: smtp.from is not configured")
	}
	if to == "" {
		return errors.New("notify: empty recipient")
	}

	msg, err := buildMessage(s.cfg.From, to, subject, body, attachments)
	if err != nil {
		return fmt.Errorf("notify: build message: %w", err)
	}

	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))
	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	// Dial with a timeout derived from the context deadline (fallback 30s).
	dialTimeout := 30 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		dialTimeout = time.Until(deadline)
	}
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return fmt.Errorf("notify: dial %s: %w", addr, err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("notify: smtp client: %w", err)
	}
	defer func() { _ = client.Quit() }()

	if s.cfg.UseTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: s.cfg.Host}); err != nil {
				return fmt.Errorf("notify: starttls: %w", err)
			}
		}
	}
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("notify: auth: %w", err)
		}
	}
	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("notify: MAIL FROM: %w", err)
	}
	recipients := splitAddresses(to)
	for _, r := range recipients {
		if err := client.Rcpt(r); err != nil {
			return fmt.Errorf("notify: RCPT TO %s: %w", r, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("notify: DATA: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("notify: write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("notify: close data: %w", err)
	}

	slog.Info("Email sent",
		"component", "notify", "to", to, "subject", subject,
		"attachments", len(attachments))
	return nil
}

func splitAddresses(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if a := strings.TrimSpace(p); a != "" {
			out = append(out, a)
		}
	}
	return out
}
