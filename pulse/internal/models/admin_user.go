package models

import "github.com/jackc/pgx/v5/pgtype"

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Data struct {
		Users      []*User `json:"users"`
		Total      int     `json:"total"`
		Limit      int     `json:"limit"`
		Offset     int     `json:"offset"`
		Pagination struct {
			HasNext bool `json:"has_next"`
			HasPrev bool `json:"has_prev"`
		} `json:"pagination"`
	} `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// GetUserResponse represents the response for getting a single user
type GetUserResponse struct {
	Data struct {
		User *User `json:"user"`
	} `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Username string  `json:"username" binding:"required"`
	Email    *string `json:"email,omitempty"`
	Password string  `json:"password" binding:"required"`
	Role     string  `json:"role" binding:"required,oneof=admin operator viewer"`
}

// CreateUserResponse represents the response for creating a user
type CreateUserResponse struct {
	Data struct {
		User *User `json:"user"`
	} `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
	Role     *string `json:"role,omitempty" binding:"omitempty,oneof=admin operator viewer"`
}

// UpdateUserResponse represents the response for updating a user
type UpdateUserResponse struct {
	Data struct {
		User *User `json:"user"`
	} `json:"data"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// DeleteUserResponse represents the response for deleting a user
type DeleteUserResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// UserDetail represents a user with additional details
type UserDetail struct {
	UserID              string           `json:"user_id"`
	Username            string           `json:"username"`
	Email               *string          `json:"email,omitempty"`
	Role                string           `json:"role"`
	FailedLoginAttempts int              `json:"failed_login_attempts"`
	LockedUntil         *pgtype.Timestamp `json:"locked_until,omitempty"`
	MFAEnabled          bool             `json:"mfa_enabled"`
	CreatedAt           pgtype.Timestamp `json:"created_at"`
	UpdatedAt           pgtype.Timestamp `json:"updated_at"`
}
