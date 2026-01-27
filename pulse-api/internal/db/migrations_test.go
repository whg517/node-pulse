package db

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/kevin/node-pulse/pulse-api/internal/testutil"
)

// TestUsersTableCreation tests users table exists and has correct structure
func TestUsersTableCreation(t *testing.T) {
	ctx := context.Background()

	// Arrange - Create test database connection
	// Note: This test assumes a test database is available
	// In practice, use test database or mock

	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Act - Check users table exists
	var tableName string
	err = pool.QueryRow(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'users'
	`).Scan(&tableName)

	if err != nil {
		t.Fatalf("Users table does not exist: %v", err)
	}

	// Assert - Table exists
	assert.Equal(t, "users", tableName)

	// Act - Check table has required columns
	rows, err := pool.Query(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'users'
		ORDER BY ordinal_position
	`)
	if err != nil {
		t.Fatalf("Failed to query columns: %v", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	expectedColumns := []string{"user_id", "username", "password_hash", "role",
		"failed_login_attempts", "locked_until", "created_at", "updated_at"}

	for rows.Next() {
		var columnName string
		var dataType string
		if err := rows.Scan(&columnName, &dataType, nil); err != nil {
			t.Fatalf("Failed to scan column: %v", err)
		}
		columns[columnName] = true
	}

	// Assert - All expected columns exist
	for _, col := range expectedColumns {
		assert.True(t, columns[col], "Missing column: %s", col)
	}

	// Act - Check indexes exist
	indexRows, err := pool.Query(ctx, `
		SELECT indexname
		FROM pg_indexes
		WHERE tablename = 'users' AND schemaname = 'public'
		ORDER BY indexname
	`)
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	defer indexRows.Close()

	indexes := make(map[string]bool)
	expectedIndexes := []string{"users_pkey", "idx_users_username"}

	for indexRows.Next() {
		var indexName string
		if err := indexRows.Scan(&indexName); err != nil {
			t.Fatalf("Failed to scan index: %v", err)
		}
		indexes[indexName] = true
	}

	// Assert - Required indexes exist
	for _, idx := range expectedIndexes {
		assert.True(t, indexes[idx], "Missing index: %s", idx)
	}
}

// TestSessionsTableCreation tests sessions table exists and has correct structure
func TestSessionsTableCreation(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Act - Check sessions table exists
	var tableName string
	err = pool.QueryRow(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'sessions'
	`).Scan(&tableName)

	if err != nil {
		t.Fatalf("Sessions table does not exist: %v", err)
	}

	// Assert - Table exists
	assert.Equal(t, "sessions", tableName)

	// Act - Check table has required columns
	rows, err := pool.Query(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'sessions'
		ORDER BY ordinal_position
	`)
	if err != nil {
		t.Fatalf("Failed to query columns: %v", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	expectedColumns := []string{"session_id", "user_id", "role",
		"expired_at", "created_at"}

	for rows.Next() {
		var columnName string
		var dataType string
		var isNullable string
		if err := rows.Scan(&columnName, &dataType, &isNullable); err != nil {
			t.Fatalf("Failed to scan column: %v", err)
		}
		columns[columnName] = true
	}

	// Assert - All expected columns exist
	for _, col := range expectedColumns {
		assert.True(t, columns[col], "Missing column: %s", col)
	}

	// Act - Check foreign key constraint exists
	var foreignKey string
	err = pool.QueryRow(ctx, `
		SELECT constraint_name
		FROM information_schema.table_constraints
		WHERE table_schema = 'public' AND table_name = 'sessions'
		  AND constraint_type = 'FOREIGN KEY'
	`).Scan(&foreignKey)

	if err != nil {
		t.Fatalf("Foreign key constraint does not exist: %v", err)
	}

	// Assert - Foreign key exists
	assert.NotNil(t, foreignKey, "Foreign key constraint should exist")

	// Act - Check indexes exist
	indexRows, err := pool.Query(ctx, `
		SELECT indexname
		FROM pg_indexes
		WHERE tablename = 'sessions' AND schemaname = 'public'
		ORDER BY indexname
	`)
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	defer indexRows.Close()

	indexes := make(map[string]bool)
	expectedIndexes := []string{"sessions_pkey", "idx_sessions_user_id",
		"idx_sessions_expired_at", "idx_sessions_user_expired"}

	for indexRows.Next() {
		var indexName string
		if err := indexRows.Scan(&indexName); err != nil {
			t.Fatalf("Failed to scan index: %v", err)
		}
		indexes[indexName] = true
	}

	// Assert - Required indexes exist
	for _, idx := range expectedIndexes {
		assert.True(t, indexes[idx], "Missing index: %s", idx)
	}
}

// TestSeedAdminUser tests that admin user is created with correct properties
func TestSeedAdminUser(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	// Act - Check admin user exists
	var userCount int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM users
		WHERE username = $1
	`, adminUsername).Scan(&userCount)

	if err != nil {
		t.Fatalf("Failed to query admin user: %v", err)
	}

	// Assert - Admin user exists
	assert.Equal(t, 1, userCount, "Admin user should exist")

	// Act - Check admin user has correct properties
	var userID, passwordHash, role string
	var failedAttempts int
	var lockedUntil sql.NullTime
	err = pool.QueryRow(ctx, `
		SELECT user_id, password_hash, role, failed_login_attempts, locked_until
		FROM users
		WHERE username = $1
	`, adminUsername).Scan(&userID, &passwordHash, &role, &failedAttempts, &lockedUntil)

	if err != nil {
		t.Fatalf("Failed to query admin user details: %v", err)
	}

	// Assert - User properties are correct
	assert.NotEqual(t, uuid.Nil, userID, "User ID should be a valid UUID")
	assert.NotEmpty(t, passwordHash, "Password hash should not be empty")
	assert.Equal(t, "admin", role, "Role should be admin")
	assert.Equal(t, 0, failedAttempts, "Failed attempts should be 0")
	assert.False(t, lockedUntil.Valid, "User should not be locked")

	// Assert - Password hash is bcrypt format (starts with $2a$, $2b$, or $2y$)
	assert.True(t, len(passwordHash) >= 60, "Bcrypt hash should be at least 60 characters")
	assert.True(t, passwordHash[0:4] == "$2a$" || passwordHash[0:4] == "$2b$" || passwordHash[0:4] == "$2y$",
		"Password hash should be in bcrypt format")
}

// TestCompositeIndexPerformance tests composite index works correctly
func TestCompositeIndexPerformance(t *testing.T) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testutil.GetTestDBURL())
	if err != nil {
		t.Skipf("Skipping test: no database connection: %v", err)
		return
	}
	defer pool.Close()

	// Arrange - Create a test user and session
	userID := uuid.New()
	sessionID := uuid.New()
	expiredAt := time.Now().Add(24 * time.Hour)
	testUsername := "test_" + userID.String()[:8]

	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, userID, testUsername, "hashedpassword", "viewer")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM users WHERE user_id = $1", userID)

	_, err = pool.Exec(ctx, `
		INSERT INTO sessions (session_id, user_id, role, expired_at, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, sessionID, userID, "viewer", expiredAt)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM sessions WHERE session_id = $1", sessionID)

	// Act - Query using composite index
	query := `
		SELECT session_id, user_id, role, expired_at
		FROM sessions
		WHERE user_id = $1 AND expired_at > $2
		ORDER BY expired_at DESC
		LIMIT 1
	`

	var resultSessionID, resultUserID, resultRole string
	var resultExpiredAt time.Time
	err = pool.QueryRow(ctx, query, userID, time.Now()).Scan(
		&resultSessionID, &resultUserID, &resultRole, &resultExpiredAt,
	)

	if err != nil {
		t.Fatalf("Composite index query failed: %v", err)
	}

	// Assert - Query returned correct results
	assert.Equal(t, sessionID.String(), resultSessionID, "Should return the test session")
	assert.Equal(t, userID.String(), resultUserID, "Should return the test user")
	assert.Equal(t, "viewer", resultRole, "Role should be viewer")
	assert.True(t, resultExpiredAt.After(time.Now().Add(23*time.Hour)), "Session should not be expired")
}
