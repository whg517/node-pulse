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

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// PasswordResetConfirmRequest represents a password reset confirmation
type PasswordResetConfirmRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// ChangePasswordResponse represents password change response
type ChangePasswordResponse struct {
	Message         string `json:"message"`
	SessionsRevoked bool   `json:"sessions_revoked"`
}

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	UserID    *string
	EventType *string
	IPAddress *string
	StartTime *string
	EndTime   *string
	Page      int
	Limit     int
	Offset    int
}
