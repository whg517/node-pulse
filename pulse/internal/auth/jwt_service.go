package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// TokenIssuer identifies the issuer of JWT tokens
	TokenIssuer = "node-pulse"
	// TokenAudience identifies the intended audience for JWT tokens
	TokenAudience = "node-pulse-api"
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	secret           []byte
	accessExpiration time.Duration
	pool             *pgxpool.Pool
	issuer           string
	audience         []string
}

// NewJWTService creates a new JWT service instance
func NewJWTService(secret string, accessExpirationMinutes int, pool *pgxpool.Pool) *JWTService {
	return &JWTService{
		secret:           []byte(secret),
		accessExpiration: time.Duration(accessExpirationMinutes) * time.Minute,
		pool:             pool,
		issuer:           TokenIssuer,
		audience:         []string{TokenAudience},
	}
}

// Claims represents JWT custom claims
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	JTI    string `json:"jti"` // JWT ID for blacklist tracking
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token with issuer and audience claims
func (s *JWTService) GenerateAccessToken(userID, role string) (string, string, error) {
	jti := uuid.New().String()
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Role:   role,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			Audience:  s.audience,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, jti, nil
}

// ValidateAccessToken validates an access token and returns the claims
// Uses 60-second clock skew tolerance as specified in tech-spec
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	// Create parser with 60-second leeway for clock skew
	parser := jwt.NewParser(jwt.WithLeeway(60*time.Second))

	token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer claim
	tokenIssuer, ok := claims["iss"].(string)
	if !ok || tokenIssuer != s.issuer {
		return nil, fmt.Errorf("invalid token issuer: expected %s, got %s", s.issuer, tokenIssuer)
	}

	// Validate audience claim
	tokenAudience, ok := claims["aud"].(string)
	if !ok {
		// Check if audience is a list
		audList, ok := claims["aud"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid token audience")
		}
		audienceValid := false
		for _, aud := range audList {
			if audStr, ok := aud.(string); ok && audStr == TokenAudience {
				audienceValid = true
				break
			}
		}
		if !audienceValid {
			return nil, fmt.Errorf("invalid token audience")
		}
	} else if tokenAudience != TokenAudience {
		return nil, fmt.Errorf("invalid token audience: expected %s, got %s", TokenAudience, tokenAudience)
	}

	// Extract claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid role in token")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid jti in token")
	}

	return &Claims{
		UserID: userID,
		Role:   role,
		JTI:    jti,
	}, nil
}

// GetJTI extracts the JTI from a token string without full validation
// Used for blacklist operations
func (s *JWTService) GetJTI(tokenString string) (string, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	jti, ok := (*claims)["jti"].(string)
	if !ok {
		return "", fmt.Errorf("invalid jti in token")
	}

	return jti, nil
}

// CheckRevoked checks if a token's JTI is in the blacklist
// Returns error when pool is nil (fail-closed for security)
func (s *JWTService) CheckRevoked(ctx context.Context, jti string) (bool, error) {
	// Fail-closed: return error when pool is nil instead of accepting all tokens
	if s.pool == nil {
		return false, fmt.Errorf("database pool not initialized - cannot verify token revocation status")
	}

	var revoked bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM token_blacklist
			WHERE jti = $1
		)
	`, jti).Scan(&revoked)

	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return revoked, nil
}
