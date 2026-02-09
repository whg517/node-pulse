package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMiddleware_TokenValidation_MissingToken tests missing Authorization header
func TestMiddleware_TokenValidation_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock JWT service
	mockJWT := &MockJWTService{}

	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(mockJWT))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should return 401 without token")
}

// TestMiddleware_TokenValidation_InvalidFormat tests malformed token is rejected
func TestMiddleware_TokenValidation_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock JWT service
	mockJWT := &MockJWTService{}

	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(mockJWT))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should return 401 for invalid format")
}

// TestMiddleware_TokenValidation_ValidToken tests valid token is accepted
func TestMiddleware_TokenValidation_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock JWT service that returns valid claims
	mockJWT := &MockJWTService{
		ValidateFunc: func(tokenString string) (*JWTClaims, error) {
			return &JWTClaims{
				UserID: "user-123",
				Role:   "admin",
				JTI:    "jti-abc",
			}, nil
		},
		CheckRevokedFunc: func(ctx context.Context, jti string) (bool, error) {
			return false, nil
		},
	}

	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(mockJWT))
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"message": "success",
		})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should return 200 for valid token")
	assert.Contains(t, w.Body.String(), "user-123", "Should set user_id in context")
}

// TestMiddleware_TokenValidation_InvalidToken tests invalid token is rejected
func TestMiddleware_TokenValidation_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock JWT service that returns error
	mockJWT := &MockJWTService{
		ValidateFunc: func(tokenString string) (*JWTClaims, error) {
			return nil, errors.New("invalid token")
		},
	}

	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(mockJWT))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should return 401 for invalid token")
}

// TestMiddleware_TokenValidation_RevokedToken tests revoked token is rejected
func TestMiddleware_TokenValidation_RevokedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock JWT service that returns token as revoked
	mockJWT := &MockJWTService{
		ValidateFunc: func(tokenString string) (*JWTClaims, error) {
			return &JWTClaims{
				UserID: "user-123",
				Role:   "admin",
				JTI:    "jti-revoked",
			}, nil
		},
		CheckRevokedFunc: func(ctx context.Context, jti string) (bool, error) {
			return true, nil // Token is revoked
		},
	}

	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(mockJWT))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer revoked-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Should return 401 for revoked token")
	assert.Contains(t, w.Body.String(), "TOKEN_REVOKED", "Should return token revoked error")
}

// TestMiddleware_HelperFunctions tests context helper functions
func TestMiddleware_HelperFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test GetUserID when not set
	_, err := GetUserID(c)
	assert.Error(t, err, "Should return error when user_id not set")

	// Test GetUserRole when not set
	_, err = GetUserRole(c)
	assert.Error(t, err, "Should return error when role not set")

	// Test GetJTI when not set
	_, err = GetJTI(c)
	assert.Error(t, err, "Should return error when jti not set")

	// Set values and test retrieval
	c.Set("user_id", "user-123")
	c.Set("role", "admin")
	c.Set("jti", "jti-abc")

	userID, err := GetUserID(c)
	assert.NoError(t, err, "Should not return error when user_id is set")
	assert.Equal(t, "user-123", userID, "Should return correct user_id")

	role, err := GetUserRole(c)
	assert.NoError(t, err, "Should not return error when role is set")
	assert.Equal(t, "admin", role, "Should return correct role")

	jti, err := GetJTI(c)
	assert.NoError(t, err, "Should not return error when jti is set")
	assert.Equal(t, "jti-abc", jti, "Should return correct jti")
}

// BenchmarkMiddleware_TokenValidation benchmarks token validation performance
func BenchmarkMiddleware_TokenValidation(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockJWT := &MockJWTService{
		ValidateFunc: func(tokenString string) (*JWTClaims, error) {
			return &JWTClaims{
				UserID: "user-123",
				Role:   "admin",
				JTI:    "jti-abc",
			}, nil
		},
		CheckRevokedFunc: func(ctx context.Context, jti string) (bool, error) {
			return false, nil
		},
	}

	router := gin.New()
	router.Use(JWTAuthMiddlewareWithInterface(mockJWT))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token-123")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
