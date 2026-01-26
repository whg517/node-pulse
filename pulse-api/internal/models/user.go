package models

import "github.com/jackc/pgx/v5/pgtype"

// User represents a user in system
type User struct {
	UserID              string              `json:"user_id" db:"user_id"`
	Username            string              `json:"username" db:"username"`
	PasswordHash        string              `json:"-" db:"password_hash"` // Never expose in JSON
	Role                string              `json:"role" db:"role"`
	FailedLoginAttempts int                 `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *pgtype.Timestamp `json:"locked_until,omitempty" db:"locked_until"`
	CreatedAt           pgtype.Timestamp  `json:"created_at" db:"created_at"`
	UpdatedAt           pgtype.Timestamp  `json:"updated_at" db:"updated_at"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents successful login response
type LoginResponse struct {
	Data struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}
