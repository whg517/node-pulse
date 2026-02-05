package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRBACMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate authenticated user
		c.Set("user_id", "user-123")
		c.Set("role", "admin")
		c.Next()
	})
	router.Use(RBACMiddleware([]string{"admin", "operator"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRBACMiddleware_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate authenticated user with wrong role
		c.Set("user_id", "user-123")
		c.Set("role", "viewer")
		c.Next()
	})
	router.Use(RBACMiddleware([]string{"admin", "operator"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_PERMISSION_DENIED")
}

func TestRBACMiddleware_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RBACMiddleware([]string{"admin"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_UNAUTHORIZED")
}

func TestRBACMiddleware_InvalidRoleType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate invalid role type
		c.Set("role", 123)
		c.Next()
	})
	router.Use(RBACMiddleware([]string{"admin"}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_INTERNAL")
}

func TestRequireRole_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("role", "admin")

	assert.True(t, RequireRole(c, "admin"))
	assert.True(t, RequireRole(c, "admin", "operator"))
	assert.False(t, RequireRole(c, "operator"))
}

func TestHasRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("role", "admin")

	assert.True(t, HasRole(c, "admin"))
	assert.False(t, HasRole(c, "operator"))
}

func TestIsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("role", "admin")

	assert.True(t, IsAdmin(c))
	assert.False(t, IsOperator(c))
}

func TestIsOperator(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("role", "operator")

	assert.False(t, IsAdmin(c))
	assert.True(t, IsOperator(c))
}

func TestRequireRole_NoRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	assert.False(t, RequireRole(c, "admin"))
}
