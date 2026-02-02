package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockSessionValidator is a mock implementation of SessionValidator for testing
type mockSessionValidator struct {
	userID  string
	role    string
	err     error
	called  bool
	sessionID string
}

func (m *mockSessionValidator) GetSession(ctx context.Context, sessionID string) (string, string, error) {
	m.called = true
	m.sessionID = sessionID
	return m.userID, m.role, m.err
}

func TestAuthMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validator := &mockSessionValidator{
		userID: "user-123",
		role:   "admin",
		err:    nil,
	}

	router := gin.New()
	router.Use(AuthMiddleware(validator))
	router.GET("/test", func(c *gin.Context) {
		userID, ok := GetUserID(c)
		assert.True(t, ok)
		assert.Equal(t, "user-123", userID)

		role, ok := GetUserRole(c)
		assert.True(t, ok)
		assert.Equal(t, "admin", role)

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-abc"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, validator.called)
	assert.Equal(t, "session-abc", validator.sessionID)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validator := &mockSessionValidator{}

	router := gin.New()
	router.Use(AuthMiddleware(validator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.False(t, validator.called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_UNAUTHORIZED")
}

func TestAuthMiddleware_InvalidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validator := &mockSessionValidator{
		err: assert.AnError,
	}

	router := gin.New()
	router.Use(AuthMiddleware(validator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid-session"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, validator.called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_INVALID_SESSION")
}

func TestGetUserID_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	userID, ok := GetUserID(c)
	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestGetUserRole_NoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	role, ok := GetUserRole(c)
	assert.False(t, ok)
	assert.Empty(t, role)
}

func TestRequireAuth_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	userID, role, ok := RequireAuth(c)
	assert.False(t, ok)
	assert.Empty(t, userID)
	assert.Empty(t, role)
}
