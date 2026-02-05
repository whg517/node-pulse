package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionService handles session CRUD operations
type SessionService struct {
	pool *pgxpool.Pool
}

// NewSessionService creates a new session service
func NewSessionService(pool *pgxpool.Pool) *SessionService {
	return &SessionService{pool: pool}
}

// CreateSession creates a new session for a user
func (s *SessionService) CreateSession(ctx context.Context, userID, role string) (string, error) {
	sessionID := uuid.New()
	expiredAt := time.Now().Add(24 * time.Hour)

	query := `
		INSERT INTO sessions (session_id, user_id, role, expired_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`

	_, err := s.pool.Exec(ctx, query, sessionID, userID, role, expiredAt)
	if err != nil {
		return "", err
	}

	return sessionID.String(), nil
}

// DeleteSession deletes a session by session ID
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE session_id = $1`
	_, err := s.pool.Exec(ctx, query, sessionID)
	return err
}

// GetSession retrieves a session by session ID
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (string, string, error) {
	var userID, role string
	err := s.pool.QueryRow(ctx, `
		SELECT user_id, role
		FROM sessions
		WHERE session_id = $1 AND expired_at > NOW()
	`, sessionID).Scan(&userID, &role)

	if err != nil {
		return "", "", err
	}

	return userID, role, nil
}

// DeleteExpiredSessions removes all expired sessions
func (s *SessionService) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expired_at <= NOW()`
	_, err := s.pool.Exec(ctx, query)
	return err
}
