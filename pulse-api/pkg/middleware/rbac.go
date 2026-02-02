package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RBACMiddleware validates user role permissions
func RBACMiddleware(requiredRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (should be set by AuthMiddleware)
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			})
			return
		}

		// Check if user has required role
		roleStr, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Invalid role format",
			})
			return
		}

		hasPermission := false
		for _, requiredRole := range requiredRoles {
			if roleStr == requiredRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "ERR_PERMISSION_DENIED",
				"message": "Permission denied",
				"details": map[string]interface{}{
					"required_roles": requiredRoles,
					"current_role":  roleStr,
				},
			})
			return
		}

		c.Next()
	}
}

// RequireRole checks if the current user has any of the required roles
// Returns true if the user has permission, false otherwise
func RequireRole(c *gin.Context, requiredRoles ...string) bool {
	role, exists := c.Get("role")
	if !exists {
		return false
	}

	roleStr, ok := role.(string)
	if !ok {
		return false
	}

	for _, requiredRole := range requiredRoles {
		if roleStr == requiredRole {
			return true
		}
	}

	return false
}

// HasRole checks if the current user has a specific role
func HasRole(c *gin.Context, role string) bool {
	return RequireRole(c, role)
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, "admin")
}

// IsOperator checks if the current user is an operator
func IsOperator(c *gin.Context) bool {
	return HasRole(c, "operator")
}
