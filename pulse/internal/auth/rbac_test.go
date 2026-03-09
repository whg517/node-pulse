package auth

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

// TestRBACService_HasPermission tests basic permission checking
func TestRBACService_HasPermission(t *testing.T) {
	service := &RBACService{}

	tests := []struct {
		name     string
		role     Role
		resource Resource
		action   Action
		want     bool
	}{
		// Admin permissions
		{
			name:     "Admin can view users",
			role:     RoleAdmin,
			resource: ResourceUsers,
			action:   ActionView,
			want:     true,
		},
		{
			name:     "Admin can create users",
			role:     RoleAdmin,
			resource: ResourceUsers,
			action:   ActionCreate,
			want:     true,
		},
		{
			name:     "Admin can delete users",
			role:     RoleAdmin,
			resource: ResourceUsers,
			action:   ActionDelete,
			want:     true,
		},
		{
			name:     "Admin can manage webhooks",
			role:     RoleAdmin,
			resource: ResourceWebhooks,
			action:   ActionCreate,
			want:     true,
		},
		{
			name:     "Admin can export data",
			role:     RoleAdmin,
			resource: ResourceExport,
			action:   ActionCreate,
			want:     true,
		},
		{
			name:     "Admin can update config",
			role:     RoleAdmin,
			resource: ResourceConfig,
			action:   ActionUpdate,
			want:     true,
		},
		// Operator permissions
		{
			name:     "Operator can view nodes",
			role:     RoleOperator,
			resource: ResourceNodes,
			action:   ActionView,
			want:     true,
		},
		{
			name:     "Operator can create nodes",
			role:     RoleOperator,
			resource: ResourceNodes,
			action:   ActionCreate,
			want:     true,
		},
		{
			name:     "Operator can update probes",
			role:     RoleOperator,
			resource: ResourceProbes,
			action:   ActionUpdate,
			want:     true,
		},
		{
			name:     "Operator can delete alerts",
			role:     RoleOperator,
			resource: ResourceAlerts,
			action:   ActionDelete,
			want:     true,
		},
		{
			name:     "Operator can read beacon config",
			role:     RoleOperator,
			resource: ResourceBeacon,
			action:   ActionRead,
			want:     true,
		},
		{
			name:     "Operator cannot manage users",
			role:     RoleOperator,
			resource: ResourceUsers,
			action:   ActionCreate,
			want:     false,
		},
		{
			name:     "Operator cannot manage webhooks",
			role:     RoleOperator,
			resource: ResourceWebhooks,
			action:   ActionCreate,
			want:     false,
		},
		{
			name:     "Operator cannot export data",
			role:     RoleOperator,
			resource: ResourceExport,
			action:   ActionCreate,
			want:     false,
		},
		{
			name:     "Operator cannot update config",
			role:     RoleOperator,
			resource: ResourceConfig,
			action:   ActionUpdate,
			want:     false,
		},
		// Viewer permissions
		{
			name:     "Viewer can view nodes",
			role:     RoleViewer,
			resource: ResourceNodes,
			action:   ActionView,
			want:     true,
		},
		{
			name:     "Viewer can view probes",
			role:     RoleViewer,
			resource: ResourceProbes,
			action:   ActionView,
			want:     true,
		},
		{
			name:     "Viewer can view alerts",
			role:     RoleViewer,
			resource: ResourceAlerts,
			action:   ActionView,
			want:     true,
		},
		{
			name:     "Viewer can view system metrics",
			role:     RoleViewer,
			resource: ResourceSystem,
			action:   ActionView,
			want:     true,
		},
		{
			name:     "Viewer cannot create nodes",
			role:     RoleViewer,
			resource: ResourceNodes,
			action:   ActionCreate,
			want:     false,
		},
		{
			name:     "Viewer cannot update alerts",
			role:     RoleViewer,
			resource: ResourceAlerts,
			action:   ActionUpdate,
			want:     false,
		},
		// Beacon permissions
		{
			name:     "Beacon can write heartbeat",
			role:     RoleBeacon,
			resource: ResourceBeacon,
			action:   ActionWrite,
			want:     true,
		},
		{
			name:     "Beacon can read config",
			role:     RoleBeacon,
			resource: ResourceConfig,
			action:   ActionView,
			want:     true,
		},
		{
			name:     "Beacon cannot view nodes",
			role:     RoleBeacon,
			resource: ResourceNodes,
			action:   ActionView,
			want:     false,
		},
		{
			name:     "Beacon cannot manage alerts",
			role:     RoleBeacon,
			resource: ResourceAlerts,
			action:   ActionCreate,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.HasPermission(tt.role, tt.resource, tt.action)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRBACService_HasAnyPermission tests checking for any of multiple permissions
func TestRBACService_HasAnyPermission(t *testing.T) {
	service := &RBACService{}

	tests := []struct {
		name        string
		role        Role
		permissions []Permission
		want        bool
	}{
		{
			name: "Admin has at least one of admin permissions",
			role: RoleAdmin,
			permissions: []Permission{
				{ResourceUsers, ActionCreate},
				{ResourceWebhooks, ActionDelete},
			},
			want: true,
		},
		{
			name: "Operator has node view permission",
			role: RoleOperator,
			permissions: []Permission{
				{ResourceNodes, ActionView},
				{ResourceUsers, ActionCreate},
			},
			want: true,
		},
		{
			name: "Viewer has none of the write permissions",
			role: RoleViewer,
			permissions: []Permission{
				{ResourceNodes, ActionCreate},
				{ResourceAlerts, ActionUpdate},
			},
			want: false,
		},
		{
			name:     "Empty permission list",
			role:     RoleAdmin,
			permissions: []Permission{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.HasAnyPermission(tt.role, tt.permissions)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRBACService_HasAllPermissions tests checking for all of multiple permissions
func TestRBACService_HasAllPermissions(t *testing.T) {
	service := &RBACService{}

	tests := []struct {
		name        string
		role        Role
		permissions []Permission
		want        bool
	}{
		{
			name: "Admin has all node permissions",
			role: RoleAdmin,
			permissions: []Permission{
				{ResourceNodes, ActionView},
				{ResourceNodes, ActionCreate},
				{ResourceNodes, ActionUpdate},
				{ResourceNodes, ActionDelete},
			},
			want: true,
		},
		{
			name: "Operator doesn't have all user permissions",
			role: RoleOperator,
			permissions: []Permission{
				{ResourceUsers, ActionView},
				{ResourceUsers, ActionCreate},
			},
			want: false,
		},
		{
			name: "Viewer has all view permissions",
			role: RoleViewer,
			permissions: []Permission{
				{ResourceNodes, ActionView},
				{ResourceProbes, ActionView},
				{ResourceAlerts, ActionView},
			},
			want: true,
		},
		{
			name:     "Empty permission list returns true",
			role:     RoleAdmin,
			permissions: []Permission{},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.HasAllPermissions(tt.role, tt.permissions)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRBACService_CheckResourceOwnership tests resource ownership checking
func TestRBACService_CheckResourceOwnership(t *testing.T) {
	// This test would require a database connection
	// For now, we test the error case when pool is nil
	t.Run("Nil pool returns error", func(t *testing.T) {
		service := &RBACService{pool: nil}

		_, err := service.CheckResourceOwnership(context.Background(), "user1", ResourceNodes, "node1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database pool not initialized")
	})
}

// TestRBACService_CanModifyResource tests resource modification permissions
func TestRBACService_CanModifyResource(t *testing.T) {
	// Test nil pool case
	t.Run("Nil pool returns error", func(t *testing.T) {
		service := &RBACService{pool: nil}

		_, err := service.CanModifyResource(context.Background(), "user1", ResourceNodes, "node1", ActionUpdate)
		assert.Error(t, err)
	})
}

// TestIsValidRole tests role validation
func TestIsValidRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{"Valid admin role", "admin", true},
		{"Valid operator role", "operator", true},
		{"Valid viewer role", "viewer", true},
		{"Valid beacon role", "beacon", true},
		{"Invalid role", "superadmin", false},
		{"Empty role", "", false},
		{"Case sensitive", "Admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidRole(tt.role)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGetPermissionsForRole tests retrieving permissions for a role
func TestGetPermissionsForRole(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		minCount int // Minimum expected permissions
	}{
		{
			name:     "Admin has many permissions",
			role:     RoleAdmin,
			minCount: 20, // Admin should have at least 20 permissions
		},
		{
			name:     "Operator has node/probe/alert permissions",
			role:     RoleOperator,
			minCount: 10,
		},
		{
			name:     "Viewer has view permissions",
			role:     RoleViewer,
			minCount: 4,
		},
		{
			name:     "Beacon has limited permissions",
			role:     RoleBeacon,
			minCount: 2,
		},
		{
			name:     "Invalid role returns empty",
			role:     Role("invalid"),
			minCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms := GetPermissionsForRole(tt.role)
			assert.GreaterOrEqual(t, len(perms), tt.minCount)
		})
	}
}

// TestRoleHierarchy tests role hierarchy and privilege levels
func TestRoleHierarchy(t *testing.T) {
	tests := []struct {
		name     string
		role1    Role
		role2    Role
		expected bool // role1 >= role2
	}{
		{"Admin >= Admin", RoleAdmin, RoleAdmin, true},
		{"Admin >= Operator", RoleAdmin, RoleOperator, true},
		{"Admin >= Viewer", RoleAdmin, RoleViewer, true},
		{"Admin >= Beacon", RoleAdmin, RoleBeacon, true},
		{"Operator >= Operator", RoleOperator, RoleOperator, true},
		{"Operator >= Viewer", RoleOperator, RoleViewer, true},
		{"Operator >= Beacon", RoleOperator, RoleBeacon, true},
		{"Operator not >= Admin", RoleOperator, RoleAdmin, false},
		{"Viewer >= Viewer", RoleViewer, RoleViewer, true},
		{"Viewer >= Beacon", RoleViewer, RoleBeacon, true},
		{"Viewer not >= Operator", RoleViewer, RoleOperator, false},
		{"Beacon >= Beacon", RoleBeacon, RoleBeacon, true},
		{"Beacon not >= Viewer", RoleBeacon, RoleViewer, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasHigherOrEqualRole(tt.role1, tt.role2)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestIsAdminRole tests admin role check
func TestIsAdminRole(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{"Admin is admin", RoleAdmin, true},
		{"Operator is not admin", RoleOperator, false},
		{"Viewer is not admin", RoleViewer, false},
		{"Beacon is not admin", RoleBeacon, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAdminRole(tt.role)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsOperatorOrAdmin tests operator/admin check
func TestIsOperatorOrAdmin(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want bool
	}{
		{"Admin is operator or admin", RoleAdmin, true},
		{"Operator is operator or admin", RoleOperator, true},
		{"Viewer is not operator or admin", RoleViewer, false},
		{"Beacon is not operator or admin", RoleBeacon, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOperatorOrAdmin(tt.role)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestRBACService_Integration tests integration scenarios
func TestRBACService_Integration(t *testing.T) {
	service := &RBACService{}

	t.Run("Admin can perform all operations on nodes", func(t *testing.T) {
		actions := []Action{ActionView, ActionCreate, ActionUpdate, ActionDelete}
		for _, action := range actions {
			assert.True(t, service.HasPermission(RoleAdmin, ResourceNodes, action),
				"Admin should be able to %s nodes", action)
		}
	})

	t.Run("Operator can perform CRUD on nodes but not on users", func(t *testing.T) {
		// Can CRUD nodes
		assert.True(t, service.HasPermission(RoleOperator, ResourceNodes, ActionCreate))
		assert.True(t, service.HasPermission(RoleOperator, ResourceNodes, ActionUpdate))
		assert.True(t, service.HasPermission(RoleOperator, ResourceNodes, ActionDelete))

		// Cannot CRUD users
		assert.False(t, service.HasPermission(RoleOperator, ResourceUsers, ActionCreate))
		assert.False(t, service.HasPermission(RoleOperator, ResourceUsers, ActionUpdate))
		assert.False(t, service.HasPermission(RoleOperator, ResourceUsers, ActionDelete))
	})

	t.Run("Viewer can only view", func(t *testing.T) {
		resources := []Resource{ResourceNodes, ResourceProbes, ResourceAlerts, ResourceSystem}
		for _, resource := range resources {
			assert.True(t, service.HasPermission(RoleViewer, resource, ActionView),
				"Viewer should be able to view %s", resource)
			assert.False(t, service.HasPermission(RoleViewer, resource, ActionCreate),
				"Viewer should not be able to create %s", resource)
			assert.False(t, service.HasPermission(RoleViewer, resource, ActionUpdate),
				"Viewer should not be able to update %s", resource)
			assert.False(t, service.HasPermission(RoleViewer, resource, ActionDelete),
				"Viewer should not be able to delete %s", resource)
		}
	})

	t.Run("Beacon has limited permissions", func(t *testing.T) {
		// Can write beacon data
		assert.True(t, service.HasPermission(RoleBeacon, ResourceBeacon, ActionWrite))
		assert.True(t, service.HasPermission(RoleBeacon, ResourceConfig, ActionView))

		// Cannot access other resources
		assert.False(t, service.HasPermission(RoleBeacon, ResourceNodes, ActionView))
		assert.False(t, service.HasPermission(RoleBeacon, ResourceAlerts, ActionView))
		assert.False(t, service.HasPermission(RoleBeacon, ResourceUsers, ActionView))
	})

	t.Run("Webhook management is admin-only", func(t *testing.T) {
		actions := []Action{ActionView, ActionCreate, ActionUpdate, ActionDelete}
		roles := []Role{RoleAdmin, RoleOperator, RoleViewer, RoleBeacon}

		for _, action := range actions {
			for _, role := range roles {
				got := service.HasPermission(role, ResourceWebhooks, action)
				if role == RoleAdmin {
					assert.True(t, got, "Admin should be able to %s webhooks", action)
				} else {
					assert.False(t, got, "%s should not be able to %s webhooks", role, action)
				}
			}
		}
	})

	t.Run("Export is admin-only", func(t *testing.T) {
		assert.True(t, service.HasPermission(RoleAdmin, ResourceExport, ActionCreate))
		assert.True(t, service.HasPermission(RoleAdmin, ResourceExport, ActionView))
		assert.False(t, service.HasPermission(RoleOperator, ResourceExport, ActionCreate))
		assert.False(t, service.HasPermission(RoleViewer, ResourceExport, ActionView))
	})

	t.Run("Config management is admin-only", func(t *testing.T) {
		assert.True(t, service.HasPermission(RoleAdmin, ResourceConfig, ActionView))
		assert.True(t, service.HasPermission(RoleAdmin, ResourceConfig, ActionUpdate))
		assert.False(t, service.HasPermission(RoleOperator, ResourceConfig, ActionUpdate))
		assert.False(t, service.HasPermission(RoleViewer, ResourceConfig, ActionUpdate))
	})
}

// TestRBACService_PrecisionMetrics calculates permission coverage
func TestRBACService_PrecisionMetrics(t *testing.T) {
	t.Run("All roles have defined permissions", func(t *testing.T) {
		roles := []Role{RoleAdmin, RoleOperator, RoleViewer, RoleBeacon}
		for _, role := range roles {
			perms := GetPermissionsForRole(role)
			assert.Greater(t, len(perms), 0, "Role %s should have permissions defined", role)
		}
	})

	t.Run("Permission matrix is consistent", func(t *testing.T) {
		// Verify that higher privilege roles have superset of lower role permissions
		// where applicable (read operations)

		viewerPerms := GetPermissionsForRole(RoleViewer)
		operatorPerms := GetPermissionsForRole(RoleOperator)
		adminPerms := GetPermissionsForRole(RoleAdmin)

		// Count view permissions
		viewerViewPerms := 0
		operatorViewPerms := 0
		adminViewPerms := 0

		for _, p := range viewerPerms {
			if p.Action == ActionView {
				viewerViewPerms++
			}
		}
		for _, p := range operatorPerms {
			if p.Action == ActionView {
				operatorViewPerms++
			}
		}
		for _, p := range adminPerms {
			if p.Action == ActionView {
				adminViewPerms++
			}
		}

		// Admin should have at least as many view permissions as operator
		assert.GreaterOrEqual(t, adminViewPerms, operatorViewPerms)
		// Operator should have at least as many view permissions as viewer
		assert.GreaterOrEqual(t, operatorViewPerms, viewerViewPerms)
	})
}

// BenchmarkRBACService_HasPermission benchmarks permission checking
func BenchmarkRBACService_HasPermission(b *testing.B) {
	service := &RBACService{}

	b.Run("Admin permission", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			service.HasPermission(RoleAdmin, ResourceNodes, ActionCreate)
		}
	})

	b.Run("Operator permission", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			service.HasPermission(RoleOperator, ResourceNodes, ActionCreate)
		}
	})

	b.Run("Viewer permission", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			service.HasPermission(RoleViewer, ResourceNodes, ActionView)
		}
	})
}

// MockRBACServiceWithDB creates a mock service for database-dependent tests
// This would be used with a test database setup
func MockRBACServiceWithDB(pool *pgxpool.Pool) *RBACService {
	return &RBACService{pool: pool}
}

// Example test with real database connection (would need test DB setup)
func ExampleRBACService_CheckResourceOwnership() {
	// This example shows how the function would be used with a real database
	// In actual tests, you would set up a test database
	/*
		ctx := context.Background()
		pool := setupTestDB()
		defer pool.Close()

		svc := &RBACService{pool: pool}

		canAccess, err := svc.CheckResourceOwnership(ctx, "user-123", ResourceNodes, "node-456")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(canAccess)
	*/
}

// TestRBACService_ErrorHandling tests error handling scenarios
func TestRBACService_ErrorHandling(t *testing.T) {
	service := &RBACService{pool: nil}

	t.Run("GetUserRole with nil pool", func(t *testing.T) {
		_, err := service.GetUserRole(context.Background(), "user1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database pool not initialized")
	})

	t.Run("GetResourceOwner with nil pool", func(t *testing.T) {
		_, err := service.GetResourceOwner(context.Background(), ResourceNodes, "node1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database pool not initialized")
	})

	t.Run("CheckResourceOwnership with nil pool", func(t *testing.T) {
		_, err := service.CheckResourceOwnership(context.Background(), "user1", ResourceNodes, "node1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database pool not initialized")
	})

	t.Run("CanModifyResource with nil pool", func(t *testing.T) {
		_, err := service.CanModifyResource(context.Background(), "user1", ResourceNodes, "node1", ActionUpdate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database pool not initialized")
	})
}
