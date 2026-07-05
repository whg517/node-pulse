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

// JWTService handles JWT token generation and validation using RS256
type JWTService struct {
	privateKey        jwt.SigningMethod
	privateKeyPEM     []byte
	publicKeyPEM      []byte
	accessExpiration  time.Duration
	pool              *pgxpool.Pool
	issuer            string
	audience          []string
	keyID             string // Key identifier for key rotation (kid header)
	// Rotation window (O-G3): when previousPublicKeyPEM + previousKeyID are
	// set, ValidateAccessToken accepts tokens signed by the previous key as
	// well as the current one. New tokens are always signed with privateKey.
	previousPublicKeyPEM []byte
	previousKeyID        string
}

// NewJWTService creates a new JWT service instance with RS256 signing
// If privateKeyPEM is empty, falls back to HS256 for backward compatibility
func NewJWTService(privateKeyPEM, publicKeyPEM, keyID string, accessExpirationMinutes int, pool *pgxpool.Pool) *JWTService {
	var signingMethod jwt.SigningMethod = jwt.SigningMethodRS256

	// If no RSA keys provided, log warning but continue (will fail at key parse time)
	// This allows the service to be constructed and validated later
	return &JWTService{
		privateKey:        signingMethod,
		privateKeyPEM:     []byte(privateKeyPEM),
		publicKeyPEM:      []byte(publicKeyPEM),
		accessExpiration:  time.Duration(accessExpirationMinutes) * time.Minute,
		pool:              pool,
		issuer:            TokenIssuer,
		audience:          []string{TokenAudience},
		keyID:             keyID,
	}
}

// WithPreviousKey enables a key-rotation window (O-G3): tokens signed by
// the previous key remain valid until they expire, while new tokens are
// signed with the current key. Returns the receiver for chaining.
//
// Rotate by:
//  1. Generate a new key pair + keyID.
//  2. Move current → previous (set previous_* to the current values).
//  3. Set the new key as the current PrivateKey/PublicKey/KeyID.
//  4. Restart Pulse.
//  5. Once the longest access token (default 15 min) has expired, unset
//     previous_* and restart again.
//
// Passing empty previousPublicKeyPEM disables the window (no-op).
func (s *JWTService) WithPreviousKey(previousPublicKeyPEM, previousKeyID string) *JWTService {
	if previousPublicKeyPEM == "" {
		return s
	}
	s.previousPublicKeyPEM = []byte(previousPublicKeyPEM)
	s.previousKeyID = previousKeyID
	return s
}

// Claims represents JWT custom claims
type Claims struct {
	UserID    string   `json:"user_id"`
	Role      string   `json:"role"`
	SessionID string   `json:"session_id,omitempty"`
	Scope     []string `json:"scope,omitempty"`
	JTI       string   `json:"jti"` // JWT ID for blacklist tracking
	jwt.RegisteredClaims
}

// parsePrivateKey parses the RSA private key from PEM format
func (s *JWTService) parsePrivateKey() (interface{}, error) {
	if len(s.privateKeyPEM) == 0 {
		return nil, fmt.Errorf("private key not configured")
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(s.privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}
	return key, nil
}

// parsePublicKey parses the RSA public key from PEM format
func (s *JWTService) parsePublicKey() (interface{}, error) {
	if len(s.publicKeyPEM) == 0 {
		return nil, fmt.Errorf("public key not configured")
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(s.publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}
	return key, nil
}

// GenerateAccessToken generates a new RS256-signed access token with kid header
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

	privateKey, err := s.parsePrivateKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to parse private key for signing: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Set key ID (kid) header for key rotation support
	if s.keyID != "" {
		token.Header["kid"] = s.keyID
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, jti, nil
}

// GenerateAccessTokenWithSession generates a token with session ID and scopes
func (s *JWTService) GenerateAccessTokenWithSession(userID, role, sessionID string, scopes []string) (string, string, error) {
	jti := uuid.New().String()
	now := time.Now()

	claims := Claims{
		UserID:    userID,
		Role:      role,
		SessionID: sessionID,
		Scope:     scopes,
		JTI:       jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			Audience:  s.audience,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	privateKey, err := s.parsePrivateKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to parse private key for signing: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Set key ID (kid) header for key rotation support
	if s.keyID != "" {
		token.Header["kid"] = s.keyID
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, jti, nil
}

// GenerateAccessTokenWithExpiry generates a token with custom expiry time in minutes
func (s *JWTService) GenerateAccessTokenWithExpiry(userID, role string, expiryMinutes int) (string, string, error) {
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
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiryMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	privateKey, err := s.parsePrivateKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to parse private key for signing: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Set key ID (kid) header for key rotation support
	if s.keyID != "" {
		token.Header["kid"] = s.keyID
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, jti, nil
}

// ValidateAccessToken validates an RS256-signed access token and returns the claims
// Uses 60-second clock skew tolerance as specified in tech-spec
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	currentPublicKey, err := s.parsePublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key for verification: %w", err)
	}

	// Pre-parse the rotation-window public key once (O-G3), so the per-token
	// keyfunc below can select the right key without re-parsing on every request.
	var previousPublicKey interface{}
	if len(s.previousPublicKeyPEM) > 0 {
		if pk, perr := jwt.ParseRSAPublicKeyFromPEM(s.previousPublicKeyPEM); perr == nil {
			previousPublicKey = pk
		}
	}

	// Create parser with 60-second leeway for clock skew
	parser := jwt.NewParser(jwt.WithLeeway(60*time.Second))

	token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing algorithm - MUST be RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v (expected RS256)", token.Header["alg"])
		}

		// Resolve which key verifies this token. With a rotation window
		// (O-G3) a token's kid selects either the current or the previous
		// public key. Without one, behavior is unchanged from before:
		// current key only, kid (if present) must match.
		tokenKid, _ := token.Header["kid"].(string)

		if tokenKid != "" && s.keyID != "" && tokenKid == s.keyID {
			return currentPublicKey, nil
		}
		// Rotation window: accept a token signed by the previous key.
		if tokenKid != "" && s.previousKeyID != "" && tokenKid == s.previousKeyID && previousPublicKey != nil {
			return previousPublicKey, nil
		}
		// No kid in the token (older client) or kid matches the sole key:
		// fall through to the current key. When a rotation window is
		// configured and the kid doesn't match either key, reject.
		if tokenKid == "" || s.keyID == "" || s.previousKeyID == "" {
			if s.previousKeyID != "" && tokenKid != "" && tokenKid != s.keyID {
				return nil, fmt.Errorf("token key ID %s does not match current (%s) or previous (%s)", tokenKid, s.keyID, s.previousKeyID)
			}
			return currentPublicKey, nil
		}
		return nil, fmt.Errorf("token key ID %s does not match current key ID %s", tokenKid, s.keyID)
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

	result := &Claims{
		UserID: userID,
		Role:   role,
		JTI:    jti,
	}

	// Optional fields
	if sessionID, ok := claims["session_id"].(string); ok {
		result.SessionID = sessionID
	}

	if scopeList, ok := claims["scope"].([]interface{}); ok {
		scopes := make([]string, 0, len(scopeList))
		for _, s := range scopeList {
			if scopeStr, ok := s.(string); ok {
				scopes = append(scopes, scopeStr)
			}
		}
		result.Scope = scopes
	}

	return result, nil
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

// GetKeyID returns the current key ID
func (s *JWTService) GetKeyID() string {
	return s.keyID
}

// VerifyKeyID checks if a token's kid matches the current key ID
// Returns true if they match or if token has no kid
func (s *JWTService) VerifyKeyID(tokenKid string) bool {
	if s.keyID == "" {
		return true // No key rotation in use
	}
	return tokenKid == s.keyID
}
