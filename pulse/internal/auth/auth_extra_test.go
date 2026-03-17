package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	)

// ---- password_utils.go tests ----

func TestValidatePassword_Valid(t *testing.T) {
	testCases := []struct {
		password string
	}{
		{"Password1"},
		{"SecureP4ss"},
		{"abcABC123"},
		{"Xy1zzz56"},
	}

	for _, tc := range testCases {
		err := ValidatePassword(tc.password)
		assert.NoError(t, err, "Password %q should be valid", tc.password)
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	err := ValidatePassword("Abc1")
	assert.ErrorIs(t, err, ErrPasswordTooShort)
}

func TestValidatePassword_TooLong(t *testing.T) {
	err := ValidatePassword("AbcPassword123ThatIsVeryLongAndExceedsLimit!")
	assert.ErrorIs(t, err, ErrPasswordTooLong)
}

func TestValidatePassword_MissingUppercase(t *testing.T) {
	err := ValidatePassword("password1")
	assert.ErrorIs(t, err, ErrPasswordMissingUppercase)
}

func TestValidatePassword_MissingLowercase(t *testing.T) {
	err := ValidatePassword("PASSWORD1")
	assert.ErrorIs(t, err, ErrPasswordMissingLowercase)
}

func TestValidatePassword_MissingDigit(t *testing.T) {
	err := ValidatePassword("PasswordOnly")
	assert.ErrorIs(t, err, ErrPasswordMissingDigit)
}

// ---- AuditLogger tests ----

func TestNewAuditLogger(t *testing.T) {
	logger := NewAuditLogger(nil)
	assert.NotNil(t, logger)
	assert.Nil(t, logger.pool)
}

func TestAuditLogger_EventConstants(t *testing.T) {
	// Verify constants are defined (compilation check)
	assert.NotEmpty(t, EventLoginSuccess)
	assert.NotEmpty(t, EventLoginFailed)
	assert.NotEmpty(t, EventLoginLocked)
	assert.NotEmpty(t, EventRateLimitExceeded)
	assert.NotEmpty(t, EventTokenGenerated)
	assert.NotEmpty(t, EventTokenRefreshed)
	assert.NotEmpty(t, EventTokenRevoked)
	assert.NotEmpty(t, EventTokenBlacklisted)
	assert.NotEmpty(t, EventSessionCreated)
	assert.NotEmpty(t, EventSessionRevoked)
	assert.NotEmpty(t, EventAllSessionsRevoked)
	assert.NotEmpty(t, EventAPIKeyGenerated)
	assert.NotEmpty(t, EventAPIKeyUsed)
	assert.NotEmpty(t, EventAPIKeyRevoked)
	assert.NotEmpty(t, EventAdminRevokeAll)
}

// ---- EnhancedAuditLogger tests ----

func TestNewEnhancedAuditLogger(t *testing.T) {
	logger := NewEnhancedAuditLogger(nil)
	assert.NotNil(t, logger)
	assert.Nil(t, logger.pool)
}

func TestEnhancedAuditLogger_EventConstants(t *testing.T) {
	assert.NotEmpty(t, EventAccessDenied)
	assert.NotEmpty(t, EventPrivilegeEscalation)
	assert.NotEmpty(t, EventSuspiciousActivity)
}

// ---- ProgressiveRateLimiter tests ----

func TestNewProgressiveRateLimiter(t *testing.T) {
	limiter := NewProgressiveRateLimiter(nil)
	assert.NotNil(t, limiter)
}

func TestGetViolationDuration(t *testing.T) {
	tests := []struct {
		level    ViolationLevel
		expected time.Duration
	}{
		{NoViolation, 0},
		{FirstViolation, 60 * time.Second},
		{SecondViolation, 5 * time.Minute},
		{ThirdViolation, 1 * time.Hour},
		{FourthPlusViolation, 24 * time.Hour},
	}

	for _, tt := range tests {
		duration := getViolationDuration(tt.level)
		assert.Equal(t, tt.expected, duration, "ViolationLevel %d", tt.level)
	}
}

func TestGetWindowStart_Progressive(t *testing.T) {
	now := time.Now()
	windowStart := getWindowStart(now, WindowPerMinute)
	assert.Equal(t, now.Truncate(time.Minute), windowStart)

	windowStart2 := getWindowStart(now, WindowPerHour)
	assert.Equal(t, now.Truncate(time.Hour), windowStart2)
}

// ---- RBACService tests ----

func TestNewRBACService(t *testing.T) {
	service := NewRBACService(nil)
	assert.NotNil(t, service)
}

// ---- APIKeyService tests ----

func TestNewAPIKeyService(t *testing.T) {
	service := NewAPIKeyService(nil)
	assert.NotNil(t, service)
}

// ---- AuthMetrics tests ----

func TestAuthMetrics_RecordLoginAttempt(t *testing.T) {
	// Should not panic - just records Prometheus metrics
	assert.NotPanics(t, func() {
		RecordLoginAttempt("success", 100*time.Millisecond)
		RecordLoginAttempt("failure", 50*time.Millisecond)
		RecordLoginAttempt("locked", 10*time.Millisecond)
	})
}

func TestAuthMetrics_RecordTokenGeneration(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordTokenGeneration("access")
		RecordTokenGeneration("refresh")
	})
}

func TestAuthMetrics_RecordRefreshRotation(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordRefreshRotation("success", 100*time.Millisecond)
		RecordRefreshRotation("invalid_token", 10*time.Millisecond)
	})
}

func TestAuthMetrics_UpdateActiveUsers(t *testing.T) {
	assert.NotPanics(t, func() {
		UpdateActiveUsers(1.0)
		UpdateActiveRefreshTokens(5.0)
		UpdateBlacklistSize(10.0)
	})
}

func TestAuthMetrics_RecordRateLimitCheck(t *testing.T) {
	assert.NotPanics(t, func() {
		RecordRateLimitCheck("login", "exceeded")
		RecordRateLimitCheck("refresh", "allowed")
	})
}

// ---- CleanupJob tests ----

func TestNewCleanupJob(t *testing.T) {
	job := NewCleanupJob(nil, 3600, 30)
	assert.NotNil(t, job)
}

// ---- PasswordResetService tests ----

func TestNewPasswordResetService(t *testing.T) {
	svc := NewPasswordResetService(nil)
	assert.NotNil(t, svc)
}

// ---- Additional auth handler tests (no DB) ----

func TestLogEvent_NilPool(t *testing.T) {
	logger := NewAuditLogger(nil)
	userID := uuid.New()
	details := map[string]interface{}{"key": "value"}

	// With nil pool, LogEvent will panic
	assert.Panics(t, func() {
		_ = logger.LogEvent(context.Background(), EventLoginSuccess, &userID, "127.0.0.1", details)
	})
}

// ---- RBACService nil pool tests ----

func TestRBACService_CheckResourceOwnership_NilPool(t *testing.T) {
	service := NewRBACService(nil)

	owns, err := service.CheckResourceOwnership(context.Background(), "user-1", ResourceNodes, "node-1")
	assert.Error(t, err)
	assert.False(t, owns)
	assert.Contains(t, err.Error(), "database pool not initialized")
}

func TestRBACService_GetResourceOwner_NilPool(t *testing.T) {
	service := NewRBACService(nil)

	owner, err := service.GetResourceOwner(context.Background(), ResourceNodes, "node-1")
	assert.Error(t, err)
	assert.Empty(t, owner)
	assert.Contains(t, err.Error(), "database pool not initialized")
}

// ---- Auth handler constructor test ----

func TestNewAuthHandler(t *testing.T) {
	// Test that NewAuthHandler can be called without panicking
	assert.NotPanics(t, func() {
		handler := NewAuthHandler(nil, "", "", "", 15, 7, 30, false, RateLimitOptions{})
		assert.NotNil(t, handler)
	})
}
