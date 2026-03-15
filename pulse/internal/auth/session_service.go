package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

const (
	// GracePeriodDuration is the grace period for concurrent refresh attempts (5 minutes)
	GracePeriodDuration = 5 * time.Minute

	// DefaultSessionExpiry is the default session expiration time
	DefaultSessionExpiry = 30 * 24 * time.Hour // 30 days

	// DefaultMaxValidity is the maximum validity period for a session
	DefaultMaxValidity = 90 * 24 * time.Hour // 90 days
)

// SessionService manages user sessions and refresh tokens with enhanced security
type SessionService struct {
	pool                *pgxpool.Pool
	refreshTokenService *RefreshTokenService
	sessionMutexes      map[string]*sessionMutex
	mu                  sync.Mutex
	cleanupDone         chan struct{}
}

// sessionMutex wraps a mutex with last-used timestamp for cleanup
type sessionMutex struct {
	mu     *sync.Mutex
	usedAt time.Time
	userID string
}

// NewSessionService creates a new session service
func NewSessionService(pool *pgxpool.Pool, refreshTokenService *RefreshTokenService) *SessionService {
	service := &SessionService{
		pool:                pool,
		refreshTokenService: refreshTokenService,
		sessionMutexes:      make(map[string]*sessionMutex),
		cleanupDone:         make(chan struct{}),
	}

	// Start background cleanup goroutine
	go service.cleanupMutexes()

	return service
}

// cleanupMutexes periodically removes unused mutexes (prevents memory leak)
func (s *SessionService) cleanupMutexes() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for userID, sm := range s.sessionMutexes {
				if sm.usedAt.Before(cutoff) {
					delete(s.sessionMutexes, userID)
				}
			}
			s.mu.Unlock()
		case <-s.cleanupDone:
			return
		}
	}
}

// Shutdown stops the cleanup goroutine
func (s *SessionService) Shutdown() {
	close(s.cleanupDone)
}

// getSessionMutex returns a mutex for the given user ID
func (s *SessionService) getSessionMutex(userID string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessionMutexes[userID]; !exists {
		s.sessionMutexes[userID] = &sessionMutex{
			mu:     &sync.Mutex{},
			usedAt: time.Now(),
			userID: userID,
		}
	} else {
		s.sessionMutexes[userID].usedAt = time.Now()
	}

	return s.sessionMutexes[userID].mu
}

// generateDeviceID creates a random device ID for browser fingerprinting
func generateDeviceID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate device ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession creates a new session with a refresh token for a user
func (s *SessionService) CreateSession(ctx context.Context, userID string, ipAddress, userAgent string, rememberMe bool) (*models.Session, string, *models.RefreshToken, error) {
	// Acquire mutex for this user
	userMutex := s.getSessionMutex(userID)
	userMutex.Lock()
	defer userMutex.Unlock()

	// Check session limit
	var activeCount int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
	`, userID).Scan(&activeCount)

	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to check active sessions: %w", err)
	}

	if activeCount >= MaxSessionsPerUser { // MaxSessionsPerUser is defined in auth_handler.go
		// Delete oldest session to make room (use subquery: PostgreSQL DELETE does not support ORDER BY/LIMIT directly)
		_, err = s.pool.Exec(ctx, `
			DELETE FROM refresh_tokens
			WHERE session_id IN (
				SELECT session_id FROM sessions
				WHERE user_id = $1 AND expires_at > NOW()
				ORDER BY created_at ASC
				LIMIT 1
			)
		`, userID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to cleanup old session tokens: %w", err)
		}

		_, err = s.pool.Exec(ctx, `
			DELETE FROM sessions
			WHERE session_id = (
				SELECT session_id FROM sessions
				WHERE user_id = $1 AND expires_at > NOW()
				ORDER BY created_at ASC
				LIMIT 1
			)
		`, userID)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to cleanup old sessions: %w", err)
		}
	}

	// Generate device ID
	deviceID, err := generateDeviceID()
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to generate device ID: %w", err)
	}

	// Calculate expiration
	now := time.Now()
	expiresAt := now.Add(DefaultSessionExpiry)
	if rememberMe {
		expiresAt = now.Add(DefaultMaxValidity)
	}
	maxValidUntil := now.Add(DefaultMaxValidity)

	// Handle empty IP address - sessions.ip_address is INET type, use netip.Addr
	var ipAddressPtr *netip.Addr
	if ipAddress != "" {
		if addr, err := netip.ParseAddr(ipAddress); err == nil {
			ipAddressPtr = &addr
		}
	}

	// Handle empty user agent
	var userAgentPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	// Create session
	var session models.Session
	sessionID := uuid.New()
	err = s.pool.QueryRow(ctx, `
		INSERT INTO sessions (session_id, user_id, device_id, ip_address, user_agent, remember_me, expires_at, max_valid_until, last_activity_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING session_id, user_id, device_id, ip_address, user_agent, remember_me, expires_at, max_valid_until, last_activity_at, created_at
	`, sessionID, userID, deviceID, ipAddressPtr, userAgentPtr, rememberMe, expiresAt, maxValidUntil).Scan(
		&session.SessionID,
		&session.UserID,
		&session.DeviceID,
		&session.IPAddress,
		&session.UserAgent,
		&session.RememberMe,
		&session.ExpiresAt,
		&session.MaxValidUntil,
		&session.LastActivityAt,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Create refresh token linked to this session
	maxValidityDays := int(DefaultMaxValidity / (24 * time.Hour))
	refreshTokenPlain, refreshToken, err := s.refreshTokenService.CreateRefreshTokenForSession(ctx, userID, session.SessionID, userAgent, ipAddress, maxValidityDays)
	if err != nil {
		// Rollback session creation if token creation fails
		_, _ = s.pool.Exec(ctx, "DELETE FROM sessions WHERE session_id = $1", sessionID)
		return nil, "", nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &session, refreshTokenPlain, refreshToken, nil
}

// ValidateSession validates a session and returns it if valid
func (s *SessionService) ValidateSession(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := s.pool.QueryRow(ctx, `
		SELECT session_id, user_id, device_id, ip_address, user_agent, remember_me, expires_at, max_valid_until, last_activity_at, created_at
		FROM sessions
		WHERE session_id = $1 AND expires_at > NOW()
	`, sessionID).Scan(
		&session.SessionID,
		&session.UserID,
		&session.DeviceID,
		&session.IPAddress,
		&session.UserAgent,
		&session.RememberMe,
		&session.ExpiresAt,
		&session.MaxValidUntil,
		&session.LastActivityAt,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("session not found or expired: %w", err)
	}

	return &session, nil
}

// RefreshSession rotates a refresh token and updates session activity
// Implements grace period for race conditions (5 min window, same IP allowed)
func (s *SessionService) RefreshSession(ctx context.Context, oldRefreshToken string, ipAddress, userAgent string) (string, *models.RefreshToken, error) {
	// Hash the old token to look up user ID first (for mutex acquisition)
	hash := sha256.Sum256([]byte(oldRefreshToken))
	tokenHash := hex.EncodeToString(hash[:])

	var userID uuid.UUID
	var sessionID *uuid.UUID
	var oldIPAddress *netip.Addr
	var revokedAt *time.Time

	err := s.pool.QueryRow(ctx, `
		SELECT user_id, session_id, ip_address, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(&userID, &sessionID, &oldIPAddress, &revokedAt)

	if err != nil {
		return "", nil, fmt.Errorf("token not found: %w", err)
	}

	// Check grace period for race conditions
	if revokedAt != nil {
		// Token was revoked - check if within grace period and same IP
		timeSinceRevocation := time.Since(*revokedAt)
		if timeSinceRevocation <= GracePeriodDuration {
			// Check if IP addresses match
			if oldIPAddress != nil && ipAddress != "" && oldIPAddress.String() == ipAddress {
				// Within grace period and same IP - allow but warn
				// This is likely a race condition from concurrent tabs
				// We'll create a new token but log this event
				goto createNewToken
			}
		}
		// Outside grace period or different IP - potential token reuse attack
		return "", nil, fmt.Errorf("token already used")
	}

createNewToken:
	// Acquire mutex for this user
	userMutex := s.getSessionMutex(userID.String())
	userMutex.Lock()
	defer userMutex.Unlock()

	// Double-check token status (concurrent request might have revoked it)
	err = s.pool.QueryRow(ctx, `
		SELECT revoked_at FROM refresh_tokens WHERE token_hash = $1 FOR UPDATE
	`, tokenHash).Scan(&revokedAt)

	if err != nil {
		return "", nil, fmt.Errorf("token not found: %w", err)
	}

	if revokedAt != nil {
		// Check grace period again after acquiring lock
		timeSinceRevocation := time.Since(*revokedAt)
		if timeSinceRevocation > GracePeriodDuration {
			return "", nil, fmt.Errorf("token already used")
		}
		if oldIPAddress != nil && ipAddress != "" && oldIPAddress.String() != ipAddress {
			// Different IP outside grace period - revoke entire token family
			_ = s.revokeTokenFamily(ctx, userID, tokenHash)
			return "", nil, fmt.Errorf("token reuse detected from different IP")
		}
	}

	// Rotate the refresh token
	newToken, newTokenRecord, err := s.rotateRefreshTokenWithSession(ctx, oldRefreshToken, sessionID, userAgent, ipAddress)
	if err != nil {
		return "", nil, fmt.Errorf("failed to rotate token: %w", err)
	}

	// Update session activity
	if sessionID != nil {
		_, err = s.pool.Exec(ctx, `
			UPDATE sessions
			SET last_activity_at = NOW()
			WHERE session_id = $1
		`, sessionID)
		if err != nil {
			// Non-fatal error, log but continue
			fmt.Printf("Warning: failed to update session activity: %v\n", err)
		}
	}

	return newToken, newTokenRecord, nil
}

// rotateRefreshTokenWithSession rotates a refresh token and links it to a session
func (s *SessionService) rotateRefreshTokenWithSession(ctx context.Context, oldToken string, sessionID *uuid.UUID, userAgent, ipAddress string) (string, *models.RefreshToken, error) {
	// Hash the old token
	hash := sha256.Sum256([]byte(oldToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Get old token info
	var oldTokenInfo models.RefreshToken
	err = tx.QueryRow(ctx, `
		SELECT id, token_id, user_id, expires_at, max_valid_until, revoked_at, replaced_by, user_agent, ip_address, created_at, updated_at
		FROM refresh_tokens
		WHERE token_hash = $1 FOR UPDATE
	`, tokenHash).Scan(
		&oldTokenInfo.ID,
		&oldTokenInfo.TokenID,
		&oldTokenInfo.UserID,
		&oldTokenInfo.ExpiresAt,
		&oldTokenInfo.MaxValidUntil,
		&oldTokenInfo.RevokedAt,
		&oldTokenInfo.ReplacedBy,
		&oldTokenInfo.UserAgent,
		&oldTokenInfo.IPAddress,
		&oldTokenInfo.CreatedAt,
		&oldTokenInfo.UpdatedAt,
	)

	if err != nil {
		return "", nil, fmt.Errorf("token not found: %w", err)
	}

	// Mark old token as revoked
	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $1, updated_at = $1
		WHERE token_hash = $2
	`, now, tokenHash)

	if err != nil {
		return "", nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Calculate new expiration (sliding window)
	newExpiresAt := now.Add(7 * 24 * time.Hour)
	if newExpiresAt.After(oldTokenInfo.MaxValidUntil.Time) {
		newExpiresAt = oldTokenInfo.MaxValidUntil.Time
	}

	// Handle empty IP address
	var ipAddressPtr *string
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}

	// Handle empty user agent
	var userAgentPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	// Create new token linked to session
	newTokenID := uuid.New()
	newTokenPlain := newTokenID.String()
	newHash := sha256.Sum256([]byte(newTokenPlain))
	newTokenHash := hex.EncodeToString(newHash[:])

	var newToken models.RefreshToken
	err = tx.QueryRow(ctx, `
		INSERT INTO refresh_tokens (token_id, token_hash, user_id, session_id, expires_at, max_valid_until, user_agent, ip_address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, token_id, user_id, expires_at, max_valid_until, revoked_at, replaced_by, user_agent, ip_address, created_at, updated_at
	`, newTokenID, newTokenHash, oldTokenInfo.UserID, sessionID, newExpiresAt, oldTokenInfo.MaxValidUntil, userAgentPtr, ipAddressPtr).Scan(
		&newToken.ID,
		&newToken.TokenID,
		&newToken.UserID,
		&newToken.ExpiresAt,
		&newToken.MaxValidUntil,
		&newToken.RevokedAt,
		&newToken.ReplacedBy,
		&newToken.UserAgent,
		&newToken.IPAddress,
		&newToken.CreatedAt,
		&newToken.UpdatedAt,
	)

	if err != nil {
		return "", nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return "", nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newTokenPlain, &newToken, nil
}

// revokeTokenFamily revokes all tokens in the same family (tokens with same replaced_by chain)
func (s *SessionService) revokeTokenFamily(ctx context.Context, userID uuid.UUID, tokenHash string) error {
	// Find all tokens in the family by following replaced_by chain
	// This is a recursive operation that marks all descendants as revoked
	_, err := s.pool.Exec(ctx, `
		WITH RECURSIVE token_family AS (
			-- Start with the reused token
			SELECT token_id, token_hash, replaced_by
			FROM refresh_tokens
			WHERE token_hash = $1 AND user_id = $2

			UNION ALL

			-- Find all descendants (tokens that replace any token in the family)
			SELECT rt.token_id, rt.token_hash, rt.replaced_by
			FROM refresh_tokens rt
			INNER JOIN token_family tf ON rt.replaced_by = tf.token_id
			WHERE rt.revoked_at IS NULL
		)
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE token_hash IN (SELECT token_hash FROM token_family)
	`, tokenHash, userID)

	return err
}

// RevokeSession revokes a specific session and all its tokens
func (s *SessionService) RevokeSession(ctx context.Context, userID string, sessionID uuid.UUID) error {
	// Revoke all tokens for this session
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE session_id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke session tokens: %w", err)
	}

	// Delete the session
	_, err = s.pool.Exec(ctx, `
		DELETE FROM sessions WHERE session_id = $1 AND user_id = $2
	`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// RevokeAllSessions revokes all sessions for a user
func (s *SessionService) RevokeAllSessions(ctx context.Context, userID string) error {
	// Revoke all tokens for this user
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all tokens: %w", err)
	}

	// Delete all sessions for this user
	_, err = s.pool.Exec(ctx, `
		DELETE FROM sessions WHERE user_id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}

// GetUserSessions returns all active sessions for a user
func (s *SessionService) GetUserSessions(ctx context.Context, userID string) ([]models.Session, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT session_id, user_id, device_id, ip_address, user_agent, remember_me, expires_at, max_valid_until, last_activity_at, created_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.DeviceID,
			&session.IPAddress,
			&session.UserAgent,
			&session.RememberMe,
			&session.ExpiresAt,
			&session.MaxValidUntil,
			&session.LastActivityAt,
			&session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions from the database
func (s *SessionService) CleanupExpiredSessions(ctx context.Context, retentionDays int) error {
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	_, err := s.pool.Exec(ctx, `
		DELETE FROM sessions
		WHERE expires_at < $1
	`, cutoff)

	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}
