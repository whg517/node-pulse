package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/whg517/node-pulse/pulse/internal/config"
)

// Claims represents JWT custom claims
// Security: Does NOT include PII (username, email) - JWT can be decoded
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Jti    string `json:"jti"` // JWT ID for audit trail
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation
type JWTService struct {
	secret                         []byte
	accessTokenExpirationMinutes   int
	refreshTokenExpirationDays     int
}

var (
	jwtService     *JWTService
	jwtServiceOnce sync.Once
	jwtServiceErr  error
)

// NewJWTService creates a new JWT service (singleton pattern)
func NewJWTService() (*JWTService, error) {
	jwtServiceOnce.Do(func() {
		cfg := config.Get()

		// Validate JWT secret is at least 64 bytes (512 bits) per NIST standards
		if len(cfg.JWT.Secret) < 64 {
			jwtServiceErr = fmt.Errorf("JWT secret must be at least 64 bytes (512 bits) for security, got %d bytes", len(cfg.JWT.Secret))
			return
		}

		jwtService = &JWTService{
			secret:                         []byte(cfg.JWT.Secret),
			accessTokenExpirationMinutes:   15, // 15 minutes default
			refreshTokenExpirationDays:     7,  // 7 days default
		}
	})

	return jwtService, jwtServiceErr
}

// GenerateAccessToken generates a JWT access token
// Returns: token string, JWT ID (jti), error
func (s *JWTService) GenerateAccessToken(userID, role string) (string, string, error) {
	// Generate unique JWT ID for audit trail
	jti := generateJTI()

	now := time.Now()
	expiresAt := now.Add(time.Duration(s.accessTokenExpirationMinutes) * time.Minute)

	claims := Claims{
		UserID: userID,
		Role:   role,
		Jti:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "node-pulse",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, jti, nil
}

// ValidateAccessToken validates a JWT access token and returns the claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}

	// Check expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("access token expired")
	}

	return claims, nil
}

// GenerateRefreshToken generates a random refresh token
// Returns: token string, JWT ID (jti), error
func (s *JWTService) GenerateRefreshToken() (string, string, error) {
	// Generate 32-byte random token (256 bits)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Generate JTI for tracking
	jti := generateJTI()

	// Encode as base64 URL-safe
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	return token, jti, nil
}

// HashRefreshToken creates a SHA-256 hash of the refresh token for storage
func (s *JWTService) HashRefreshToken(token string) (string, error) {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:]), nil
}

// generateJTI generates a unique JWT ID for audit trail
func generateJTI() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
