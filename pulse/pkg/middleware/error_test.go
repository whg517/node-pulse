package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestErrorHandler_Success tests successful response passes through
func TestErrorHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestErrorHandler_AppError tests that AppError is formatted correctly
func TestErrorHandler_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		RespondWithError(c, ERR_INVALID_REQUEST, "Invalid input", http.StatusBadRequest)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Should return 400 because error was already handled
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Verify error response structure
	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != ERR_INVALID_REQUEST {
		t.Errorf("Expected error code %s, got %s", ERR_INVALID_REQUEST, resp.Code)
	}
	if resp.Message != "Invalid input" {
		t.Errorf("Expected message 'Invalid input', got '%s'", resp.Message)
	}
}

// TestErrorHandler_GinError tests that AppError in Gin context is formatted correctly
func TestErrorHandler_GinError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		// Add error to context without calling RespondWithError
		// This will be caught by ErrorHandler middleware
		c.Error(&AppError{
			Code:    ERR_DATABASE_ERROR,
			Message: "Database connection failed",
		})
		// Don't call c.JSON - let ErrorHandler handle it
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// ErrorHandler should have responded with error format
	if w.Code == http.StatusOK {
		t.Error("Expected error status code, got OK")
	}

	// Verify error response structure
	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != ERR_DATABASE_ERROR {
		t.Errorf("Expected error code %s, got %s", ERR_DATABASE_ERROR, resp.Code)
	}
	if resp.Message != "Database connection failed" {
		t.Errorf("Expected message 'Database connection failed', got '%s'", resp.Message)
	}
}

// TestErrorHandler_StandardError tests that standard errors are formatted
func TestErrorHandler_StandardError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		// Add standard error to context
		c.Error(http.ErrNotSupported)
		// Don't call c.JSON - let ErrorHandler handle it
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// ErrorHandler should have responded with error format
	if w.Code == http.StatusOK {
		t.Error("Expected error status code, got OK")
	}

	// Verify error response structure
	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Code != ERR_INTERNAL {
		t.Errorf("Expected error code %s, got %s", ERR_INTERNAL, resp.Code)
	}
}

// TestRespondWithSuccess tests success response helper
func TestRespondWithSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		RespondWithSuccess(c, gin.H{"key": "value"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify success response structure
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp["code"] != "SUCCESS" {
		t.Errorf("Expected code 'SUCCESS', got '%v'", resp["code"])
	}
	if resp["message"] != "OK" {
		t.Errorf("Expected message 'OK', got '%v'", resp["message"])
	}
}
