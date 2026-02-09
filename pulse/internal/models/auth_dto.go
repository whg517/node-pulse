package models

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse represents successful token issuance response
type TokenResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	TokenType         string `json:"token_type"` // "Bearer"
	ExpiresIn         int    `json:"expires_in"`    // 900 (15 minutes)
	RefreshExpiresIn  int    `json:"refresh_expires_in"` // 604800 (7 days)
}

// SessionResponse represents a user session (refresh token)
type SessionResponse struct {
	SessionID     string    `json:"session_id"`
	CreatedAt     string    `json:"created_at"`
	LastUsedAt    *string   `json:"last_used_at,omitempty"`
	ExpiresAt     string    `json:"expires_at"`
	MaxValidUntil string    `json:"max_valid_until"`
	UserAgent     *string   `json:"user_agent,omitempty"`
	IPAddress     *string   `json:"ip_address,omitempty"`
}
