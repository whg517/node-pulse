package auth

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// BenchmarkJWT_ValidationThroughput benchmarks JWT validation with blacklist check
// Tech-Spec requirement: 1000 validations → verify < 10ms avg (includes blacklist check)
func BenchmarkJWT_ValidationThroughput(b *testing.B) {
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		b.Skip("Database not available")
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, &testing.T{}, pool)

	jwtService := NewJWTService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", 15, pool)

	userID := uuid.New().String()
	role := "admin"
	accessToken, _, _ := jwtService.GenerateAccessToken(userID, role)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jwtService.ValidateAccessToken(accessToken)
	}
}

// BenchmarkRefreshToken_Validation benchmarks refresh token validation
// Tech-Spec requirement: 100 validations → verify < 10ms avg
func BenchmarkRefreshToken_Validation(b *testing.B) {
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		b.Skip("Database not available")
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, &testing.T{}, pool)

	refreshService := NewRefreshTokenService(pool)
	userID := uuid.New()

	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, email, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
	`, userID, "testuser", "hash", "test@example.com", "admin")
	require.NoError(b, err)

	tokenPlain := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)
	maxValidUntil := time.Now().Add(30 * 24 * time.Hour)

	_, err = pool.Exec(ctx, `
		INSERT INTO refresh_tokens (token_id, token_hash, user_id, expires_at, max_valid_until, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, uuid.New(), HashTokenSHA256(tokenPlain), userID, expiresAt, maxValidUntil)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		refreshService.ValidateRefreshToken(ctx, tokenPlain)
	}
}

// BenchmarkConcurrentLoad benchmarks concurrent auth requests
// Tech-Spec requirement: 100 simultaneous requests → verify no deadlocks, no connection pool exhaustion
func BenchmarkConcurrentLoad(b *testing.B) {
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		b.Skip("Database not available")
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, &testing.T{}, pool)

	jwtService := NewJWTService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", 15, pool)
	userID := uuid.New().String()
	role := "admin"

	accessToken, _, _ := jwtService.GenerateAccessToken(userID, role)

	b.ResetTimer()

	numGoroutines := 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < b.N/numGoroutines; j++ {
				jwtService.ValidateAccessToken(accessToken)
			}
		}()
	}

	wg.Wait()
}

// BenchmarkBlacklist_LookupPerformance benchmarks blacklist lookup with 10K entries
// Tech-Spec requirement: 10,000 blacklist entries → verify < 5ms lookup
func BenchmarkBlacklist_LookupPerformance(b *testing.B) {
	testDSN := "postgres://testuser:testpass123@localhost:5432/nodepulse_test?sslmode=disable"
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		b.Skip("Database not available")
		return
	}
	defer pool.Close()

	cleanupTables(ctx, pool)
	createTestTables(ctx, &testing.T{}, pool)

	jwtService := NewJWTService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", 15, pool)

	// Insert 10,000 blacklist entries
	b.StopTimer()
	numEntries := 10000

	for i := 0; i < numEntries; i++ {
		jti := fmt.Sprintf("bench-jti-%c-%d", 'A'+(i%26), i)
		_, err := pool.Exec(ctx,
			"INSERT INTO token_blacklist (jti, revoked_at, expires_at) VALUES ($1, NOW(), NOW() + INTERVAL '1 hour')",
			jti)
		require.NoError(b, err)
	}

	var count int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM token_blacklist").Scan(&count)
	require.Equal(b, numEntries, count)

	b.StartTimer()

	jti := "bench-jti-A-5000"
	for i := 0; i < b.N; i++ {
		jwtService.CheckRevoked(ctx, jti)
	}
}
