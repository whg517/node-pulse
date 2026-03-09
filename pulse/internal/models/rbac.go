package models

import "github.com/jackc/pgx/v5/pgtype"

// Role represents a user role in the system
type Role struct {
	ID           int              `json:"id" db:"id"`
	Name         string           `json:"name" db:"name"`
	Description  *string          `json:"description,omitempty" db:"description"`
	IsSystemRole bool             `json:"is_system_role" db:"is_system_role"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
}

// Permission represents a specific permission for a resource
type Permission struct {
	ID          int     `json:"id" db:"id"`
	Resource    string  `json:"resource" db:"resource"` // nodes, probes, alerts, etc.
	Action      string  `json:"action" db:"action"`     // read, write, delete
	Description *string `json:"description,omitempty" db:"description"`
}

// RolePermission represents the junction between roles and permissions
type RolePermission struct {
	RoleID       int              `json:"role_id" db:"role_id"`
	PermissionID int              `json:"permission_id" db:"permission_id"`
	GrantedAt    pgtype.Timestamp `json:"granted_at" db:"granted_at"`
}

// RoleWithPermissions represents a role with its associated permissions
type RoleWithPermissions struct {
	Role        Role         `json:"role"`
	Permissions []Permission `json:"permissions"`
}

// UserWithRole represents a user with their role information
type UserWithRole struct {
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	Role      string  `json:"role"`
	Email     *string `json:"email,omitempty"`
	CreatedAt string  `json:"created_at"`
}
