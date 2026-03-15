package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// TestCleanupJob_Debug debugs the batch deletion issue
func TestCleanupJob_Debug(t *testing.T) {
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, t, pool)

	userID := uuid.New()
	now := time.Now()

	// Create user first
	hashedPassword, _ := HashPassword("testpass")
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", hashedPassword, "test@example.com", "admin")
	require.NoError(t, err)

	// Insert just 10 expired tokens
	for i := 0; i < 10; i++ {
		expiresAt := now.Add(-1 * time.Hour)
		maxValidUntil := now.Add(-30 * 24 * time.Hour)

		_, err := pool.Exec(ctx, `
			INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at, max_valid_until, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
		`, uuid.New(), HashTokenSHA256(fmt.Sprintf("token%d", i)), userID, expiresAt, maxValidUntil)
		require.NoError(t, err)
	}

	// Verify tokens were inserted
	var countBefore int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1", userID).Scan(&countBefore)
	require.NoError(t, err)
	t.Logf("Tokens before cleanup: %d", countBefore)

	// Check expires_at values
	rows, _ := pool.Query(ctx, "SELECT token_id, expires_at, NOW() as current_time, expires_at <= NOW() as is_expired FROM refresh_tokens WHERE user_id = $1 LIMIT 5", userID)
	defer rows.Close()
	for rows.Next() {
		var tokenID uuid.UUID
		var expiresAt time.Time
		var currentTime time.Time
		var isExpired bool
		_ = rows.Scan(&tokenID, &expiresAt, &currentTime, &isExpired)
		t.Logf("Token %s: expires_at=%v, current_time=%v, is_expired=%v", tokenID, expiresAt, currentTime, isExpired)
	}

	// Run cleanup
	job := NewCleanupJob(pool, 3600, 90)
	deleted, err := job.cleanupExpiredTokens(ctx)
	require.NoError(t, err)
	t.Logf("Tokens deleted: %d", deleted)

	// Verify cleanup
	var countAfter int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1", userID).Scan(&countAfter)
	require.NoError(t, err)
	t.Logf("Tokens after cleanup: %d", countAfter)

	if countAfter != 0 {
		// Show remaining tokens
		rows2, _ := pool.Query(ctx, "SELECT token_id, expires_at, NOW() as current_time, expires_at <= NOW() as is_expired FROM refresh_tokens WHERE user_id = $1", userID)
		defer rows2.Close()
		for rows2.Next() {
			var tokenID uuid.UUID
			var expiresAt time.Time
			var currentTime time.Time
			var isExpired bool
			_ = rows2.Scan(&tokenID, &expiresAt, &currentTime, &isExpired)
			t.Logf("Remaining token %s: expires_at=%v, current_time=%v, is_expired=%v", tokenID, expiresAt, currentTime, isExpired)
		}
	}
}
