package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Role represents a user role in the system
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
	RoleBeacon   Role = "beacon"
)

// Resource represents a system resource type
type Resource string

const (
	ResourceUsers      Resource = "users"
	ResourceNodes      Resource = "nodes"
	ResourceProbes     Resource = "probes"
	ResourceAlerts     Resource = "alerts"
	ResourceWebhooks   Resource = "webhooks"
	ResourceExport     Resource = "export"
	ResourceSystem     Resource = "system"
	ResourceBeacon     Resource = "beacon"
	ResourceConfig     Resource = "config"
)

// Action represents a permission action
type Action string

const (
	ActionView   Action = "view"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionAdmin  Action = "admin"
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
)

// Permission represents a single permission
type Permission struct {
	Resource Resource
	Action   Action
}

// RBACService handles role-based access control
type RBACService struct {
	pool *pgxpool.Pool
}

// NewRBACService creates a new RBAC service
func NewRBACService(pool *pgxpool.Pool) *RBACService {
	return &RBACService{
		pool: pool,
	}
}

// RolePermissions maps roles to their permissions
// Based on design section 3.2 permissions matrix
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		// Users - admin only
		{ResourceUsers, ActionView},
		{ResourceUsers, ActionCreate},
		{ResourceUsers, ActionUpdate},
		{ResourceUsers, ActionDelete},
		{ResourceUsers, ActionAdmin},
		// Nodes - all actions
		{ResourceNodes, ActionView},
		{ResourceNodes, ActionCreate},
		{ResourceNodes, ActionUpdate},
		{ResourceNodes, ActionDelete},
		// Probes - all actions
		{ResourceProbes, ActionView},
		{ResourceProbes, ActionCreate},
		{ResourceProbes, ActionUpdate},
		{ResourceProbes, ActionDelete},
		// Alerts - all actions
		{ResourceAlerts, ActionView},
		{ResourceAlerts, ActionCreate},
		{ResourceAlerts, ActionUpdate},
		{ResourceAlerts, ActionDelete},
		{ResourceAlerts, ActionAdmin},
		// Webhooks - admin only
		{ResourceWebhooks, ActionView},
		{ResourceWebhooks, ActionCreate},
		{ResourceWebhooks, ActionUpdate},
		{ResourceWebhooks, ActionDelete},
		// Export - admin only
		{ResourceExport, ActionView},
		{ResourceExport, ActionCreate},
		// System - admin only
		{ResourceSystem, ActionView},
		{ResourceSystem, ActionAdmin},
		{ResourceConfig, ActionView},
		{ResourceConfig, ActionUpdate},
		// Beacon config
		{ResourceBeacon, ActionRead},
		{ResourceBeacon, ActionWrite},
	},
	RoleOperator: {
		// Nodes - operator can create/update/delete
		{ResourceNodes, ActionView},
		{ResourceNodes, ActionCreate},
		{ResourceNodes, ActionUpdate},
		{ResourceNodes, ActionDelete},
		// Probes - operator can create/update/delete
		{ResourceProbes, ActionView},
		{ResourceProbes, ActionCreate},
		{ResourceProbes, ActionUpdate},
		{ResourceProbes, ActionDelete},
		// Alerts - operator can manage rules
		{ResourceAlerts, ActionView},
		{ResourceAlerts, ActionCreate},
		{ResourceAlerts, ActionUpdate},
		{ResourceAlerts, ActionDelete},
		// Beacon config
		{ResourceBeacon, ActionRead},
		{ResourceBeacon, ActionWrite},
		// System metrics (view only)
		{ResourceSystem, ActionView},
	},
	RoleViewer: {
		// Read-only access to most resources
		{ResourceNodes, ActionView},
		{ResourceProbes, ActionView},
		{ResourceAlerts, ActionView},
		{ResourceSystem, ActionView},
	},
	RoleBeacon: {
		// Beacon role - limited permissions
		{ResourceBeacon, ActionWrite},  // heartbeat:write
		{ResourceConfig, ActionView},    // config:read
	},
}

// HasPermission checks if a role has a specific permission
func (s *RBACService) HasPermission(role Role, resource Resource, action Action) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm.Resource == resource && perm.Action == action {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if role has any of the specified permissions
func (s *RBACService) HasAnyPermission(role Role, permissions []Permission) bool {
	for _, perm := range permissions {
		if s.HasPermission(role, perm.Resource, perm.Action) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if role has all specified permissions
func (s *RBACService) HasAllPermissions(role Role, permissions []Permission) bool {
	for _, perm := range permissions {
		if !s.HasPermission(role, perm.Resource, perm.Action) {
			return false
		}
	}
	return true
}

// CheckResourceOwnership checks if user owns or has access to a specific resource
// This implements resource-level access control (design section 3.3)
func (s *RBACService) CheckResourceOwnership(ctx context.Context, userID string, resourceType Resource, resourceID string) (bool, error) {
	if s.pool == nil {
		return false, fmt.Errorf("database pool not initialized")
	}

	// Admin has access to all resources
	var userRole string
	err := s.pool.QueryRow(ctx, `
		SELECT role FROM users WHERE user_id = $1
	`, userID).Scan(&userRole)

	if err != nil {
		return false, fmt.Errorf("failed to lookup user role: %w", err)
	}

	if userRole == string(RoleAdmin) {
		return true, nil
	}

	// Check resource ownership based on type
	switch resourceType {
	case ResourceNodes:
		// Check if user created the node
		var exists bool
		err = s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM nodes
				WHERE node_id = $1 AND created_by = $2
			)
		`, resourceID, userID).Scan(&exists)
		return exists, err

	case ResourceProbes:
		// Check if user created the probe
		var exists bool
		err = s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM probes
				WHERE probe_id = $1 AND created_by = $2
			)
		`, resourceID, userID).Scan(&exists)
		return exists, err

	case ResourceAlerts:
		// Check if user created the alert rule
		var exists bool
		err = s.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM alert_rules
				WHERE rule_id = $1 AND created_by = $2
			)
		`, resourceID, userID).Scan(&exists)
		return exists, err

	default:
		// For other resources, operators and admins have access
		return userRole == string(RoleOperator) || userRole == string(RoleAdmin), nil
	}
}

// GetResourceOwner returns the owner of a resource
func (s *RBACService) GetResourceOwner(ctx context.Context, resourceType Resource, resourceID string) (string, error) {
	if s.pool == nil {
		return "", fmt.Errorf("database pool not initialized")
	}

	var ownerID string
	var err error

	switch resourceType {
	case ResourceNodes:
		err = s.pool.QueryRow(ctx, `
			SELECT created_by FROM nodes WHERE node_id = $1
		`, resourceID).Scan(&ownerID)

	case ResourceProbes:
		err = s.pool.QueryRow(ctx, `
			SELECT created_by FROM probes WHERE probe_id = $1
		`, resourceID).Scan(&ownerID)

	case ResourceAlerts:
		err = s.pool.QueryRow(ctx, `
			SELECT created_by FROM alert_rules WHERE rule_id = $1
		`, resourceID).Scan(&ownerID)

	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	if err != nil {
		return "", fmt.Errorf("failed to get resource owner: %w", err)
	}

	return ownerID, nil
}

// CanModifyResource checks if a user can modify a specific resource
// Takes into account both role permissions and resource ownership
func (s *RBACService) CanModifyResource(ctx context.Context, userID string, resourceType Resource, resourceID string, action Action) (bool, error) {
	if s.pool == nil {
		return false, fmt.Errorf("database pool not initialized")
	}

	// First check role-based permission
	var userRole string
	err := s.pool.QueryRow(ctx, `
		SELECT role FROM users WHERE user_id = $1
	`, userID).Scan(&userRole)

	if err != nil {
		return false, fmt.Errorf("failed to lookup user role: %w", err)
	}

	// Admin can do anything
	if userRole == string(RoleAdmin) {
		return true, nil
	}

	// Check if role has the required permission
	if !s.HasPermission(Role(userRole), resourceType, action) {
		return false, nil
	}

	// For non-admin users, check resource ownership
	// Operators can only modify resources they created
	if userRole == string(RoleOperator) {
		hasAccess, err := s.CheckResourceOwnership(ctx, userID, resourceType, resourceID)
		if err != nil {
			return false, err
		}
		return hasAccess, nil
	}

	// Viewers cannot modify anything
	return false, nil
}

// GetUserRole retrieves the role for a user
func (s *RBACService) GetUserRole(ctx context.Context, userID string) (Role, error) {
	if s.pool == nil {
		return "", fmt.Errorf("database pool not initialized")
	}

	var roleStr string
	err := s.pool.QueryRow(ctx, `
		SELECT role FROM users WHERE user_id = $1
	`, userID).Scan(&roleStr)

	if err != nil {
		return "", fmt.Errorf("failed to lookup user role: %w", err)
	}

	return Role(roleStr), nil
}

// IsValidRole checks if a role string is valid
func IsValidRole(role string) bool {
	switch Role(role) {
	case RoleAdmin, RoleOperator, RoleViewer, RoleBeacon:
		return true
	default:
		return false
	}
}

// GetPermissionsForRole returns all permissions for a given role
func GetPermissionsForRole(role Role) []Permission {
	permissions, exists := RolePermissions[role]
	if !exists {
		return []Permission{}
	}
	return permissions
}

// RoleHierarchy defines role precedence (higher index = higher privilege)
// Used for determining if one role can override another
var RoleHierarchy = map[Role]int{
	RoleViewer:   0,
	RoleBeacon:   1,
	RoleOperator: 2,
	RoleAdmin:    3,
}

// HasHigherOrEqualRole checks if role1 has equal or higher privilege than role2
func HasHigherOrEqualRole(role1, role2 Role) bool {
	level1, exists1 := RoleHierarchy[role1]
	level2, exists2 := RoleHierarchy[role2]

	if !exists1 || !exists2 {
		return false
	}

	return level1 >= level2
}

// IsAdminRole checks if the given role is admin
func IsAdminRole(role Role) bool {
	return role == RoleAdmin
}

// IsOperatorOrAdmin checks if the role is operator or admin
func IsOperatorOrAdmin(role Role) bool {
	return role == RoleOperator || role == RoleAdmin
}
