package middleware

import (
	"context"
)

// JWTService interface defines methods needed by middleware
// This interface allows mocking without importing internal/auth
type JWTService interface {
	// ValidateAccessToken validates a JWT access token and returns claims
	ValidateAccessToken(tokenString string) (*JWTClaims, error)

	// CheckRevoked checks if a token has been revoked (is in blacklist)
	CheckRevoked(ctx context.Context, jti string) (bool, error)
}

// JWTClaims represents JWT custom claims (simplified version of auth.Claims)
type JWTClaims struct {
	UserID string
	Role   string
	JTI    string
}

// MockJWTService is a mock implementation of JWTService for testing
type MockJWTService struct {
	ValidateFunc func(tokenString string) (*JWTClaims, error)
	CheckRevokedFunc func(ctx context.Context, jti string) (bool, error)
}

// ValidateAccessToken mocks validation
func (m *MockJWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(tokenString)
	}
	return &JWTClaims{
		UserID: "test-user-id",
		Role:   "admin",
		JTI:    "test-jti",
	}, nil
}

// CheckRevoked mocks blacklist check
func (m *MockJWTService) CheckRevoked(ctx context.Context, jti string) (bool, error) {
	if m.CheckRevokedFunc != nil {
		return m.CheckRevokedFunc(ctx, jti)
	}
	return false, nil
}
