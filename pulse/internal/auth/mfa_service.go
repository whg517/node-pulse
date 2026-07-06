package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// MFAService handles TOTP-based two-factor authentication.
//
// Lifecycle:
//  1. Setup: user opts in → GenerateSecret returns a secret + otpauth URI +
//     a short-lived setup ticket. The secret is NOT yet persisted.
//  2. Verify: user enters a code from their authenticator → VerifySetup
//     validates it and persists mfa_enabled=true + the secret.
//  3. Login: if mfa_enabled, Login returns an mfa_ticket instead of tokens;
//     VerifyLogin consumes the ticket + a fresh code to issue tokens.
//  4. Disable: admin or the user themselves turns MFA off.
//
// Pending tickets (login + setup) live in memory: they expire after 5 min,
// are single-use, and a restart drops them (the user simply retries). That's
// acceptable for 2FA — if Pulse restarts mid-login, the user logs in again.
type MFAService struct {
	pool *pgxpool.Pool

	mu      sync.Mutex
	tickets map[string]*mfaTicket // login-step-2 pending tickets
}

// mfaTicket is the in-memory state linking a login-step-1 success to the
// user info needed to finish login after the TOTP code checks out.
type mfaTicket struct {
	userID    string
	username  string
	role      string
	ipAddress string
	userAgent string
	issuedAt  time.Time
}

// TicketTTL is how long a login-step-2 ticket stays redeemable.
const TicketTTL = 5 * time.Minute

// NewMFAService constructs the service.
func NewMFAService(pool *pgxpool.Pool) *MFAService {
	return &MFAService{
		pool:   pool,
		tickets: make(map[string]*mfaTicket),
	}
}

// GenerateSecret creates a new TOTP secret for a user without persisting it.
// Returns the secret, an otpauth URI (encode into a QR), and a setup ticket
// that VerifySetup consumes. Issuer/Account appear in the user's authenticator.
func (s *MFAService) GenerateSecret(ctx context.Context, userID, username string) (*models.MFASetupResponse, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Node-Pulse",
		AccountName: username,
		// 30s period + 6 digits is the Google Authenticator default and what
		// every authenticator app expects; don't change without a strong reason.
		Period:    30,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("generate TOTP secret: %w", err)
	}

	// Confirm the user doesn't already have MFA on (don't let them stack).
	var enabled bool
	err = s.pool.QueryRow(ctx, `SELECT mfa_enabled FROM users WHERE user_id = $1`, userID).Scan(&enabled)
	if err != nil {
		return nil, fmt.Errorf("lookup mfa state: %w", err)
	}
	if enabled {
		return nil, errors.New("MFA already enabled")
	}

	// Hold the secret in the ticket so VerifySetup can persist it. We don't
	// write to DB yet — the user might scan the QR and never verify.
	ticket := s.issueTicket(userID, username, "", "", "")
	// Stash the proposed secret on the ticket via a sidecar map keyed by ticket.
	s.storePendingSecret(ticket, key.Secret())

	return &models.MFASetupResponse{
		Secret:     key.Secret(),
		OTPAuthURI: key.URL(),
		Ticket:     ticket,
	}, nil
}

// pendingSecrets holds setup-time secrets keyed by setup ticket, so the secret
// doesn't touch the DB until VerifySetup succeeds. Separate from login tickets.
var (
	pendingSecretsMu sync.Mutex
	pendingSecrets   = make(map[string]string)
)

func (s *MFAService) storePendingSecret(ticket, secret string) {
	pendingSecretsMu.Lock()
	defer pendingSecretsMu.Unlock()
	pendingSecrets[ticket] = secret
	// Best-effort cleanup: drop the secret once the ticket would be too old
	// to use. We don't spin a goroutine per secret; the verifier also bounds it.
	go func() {
		time.Sleep(TicketTTL + time.Minute)
		pendingSecretsMu.Lock()
		delete(pendingSecrets, ticket)
		pendingSecretsMu.Unlock()
	}()
}

// VerifySetup confirms the user's authenticator is correctly seeded by
// checking a live code, then persists mfa_enabled=true + the secret.
func (s *MFAService) VerifySetup(ctx context.Context, ticket, code string) error {
	pendingSecretsMu.Lock()
	secret, ok := pendingSecrets[ticket]
	pendingSecretsMu.Unlock()
	if !ok {
		return errors.New("invalid or expired setup ticket")
	}

	if !totp.Validate(code, secret) {
		return errors.New("invalid TOTP code")
	}

	// Resolve the user from the ticket.
	s.mu.Lock()
	t, ok := s.tickets[ticket]
	s.mu.Unlock()
	if !ok || time.Since(t.issuedAt) > TicketTTL {
		return errors.New("invalid or expired setup ticket")
	}

	_, err := s.pool.Exec(ctx, `
		UPDATE users SET mfa_enabled = TRUE, mfa_secret = $1, updated_at = NOW()
		WHERE user_id = $2 AND mfa_enabled = FALSE
	`, secret, t.userID)
	if err != nil {
		return fmt.Errorf("persist mfa secret: %w", err)
	}

	// Single-use ticket.
	s.consumeTicket(ticket)
	pendingSecretsMu.Lock()
	delete(pendingSecrets, ticket)
	pendingSecretsMu.Unlock()

	slog.Info("MFA enabled", "component", "mfa", "user_id", t.userID)
	return nil
}

// IsEnabled reports whether the user has MFA turned on.
func (s *MFAService) IsEnabled(ctx context.Context, userID string) (bool, error) {
	var enabled bool
	err := s.pool.QueryRow(ctx, `SELECT mfa_enabled FROM users WHERE user_id = $1`, userID).Scan(&enabled)
	return enabled, err
}

// IssueLoginTicket is called by Login after password verification when the
// account has MFA enabled. The returned ticket must be presented (with a
// valid TOTP code) to VerifyLogin to complete authentication.
func (s *MFAService) IssueLoginTicket(userID, username, role, ipAddress, userAgent string) string {
	return s.issueTicket(userID, username, role, ipAddress, userAgent)
}

// VerifyLogin completes a 2FA login: validates the code against the user's
// stored secret, consumes the ticket, and returns the identity for the caller
// to mint access/refresh tokens.
func (s *MFAService) VerifyLogin(ctx context.Context, ticket, code string) (userID, username, role string, err error) {
	s.mu.Lock()
	t, ok := s.tickets[ticket]
	s.mu.Unlock()
	if !ok || time.Since(t.issuedAt) > TicketTTL {
		return "", "", "", errors.New("invalid or expired MFA ticket")
	}

	var secret *string
	err = s.pool.QueryRow(ctx, `SELECT mfa_secret FROM users WHERE user_id = $1 AND mfa_enabled = TRUE`, t.userID).Scan(&secret)
	if err != nil || secret == nil {
		return "", "", "", errors.New("MFA not enabled for user")
	}

	if !totp.Validate(code, *secret) {
		return "", "", "", errors.New("invalid TOTP code")
	}

	s.consumeTicket(ticket)
	return t.userID, t.username, t.role, nil
}

// Disable turns MFA off for a user (clears the secret). Used by the user
// themselves and by admins.
func (s *MFAService) Disable(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE users SET mfa_enabled = FALSE, mfa_secret = NULL, updated_at = NOW()
		WHERE user_id = $1
	`, userID)
	return err
}

// --- ticket plumbing ---

func (s *MFAService) issueTicket(userID, username, role, ipAddress, userAgent string) string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	ticket := hex.EncodeToString(b)

	s.mu.Lock()
	s.tickets[ticket] = &mfaTicket{
		userID:    userID,
		username:  username,
		role:      role,
		ipAddress: ipAddress,
		userAgent: userAgent,
		issuedAt:  time.Now(),
	}
	s.mu.Unlock()
	return ticket
}

func (s *MFAService) consumeTicket(ticket string) {
	s.mu.Lock()
	delete(s.tickets, ticket)
	s.mu.Unlock()
}
