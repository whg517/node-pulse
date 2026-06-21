package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func resetMTLSForTest() {
	mTLSConfig = nil
	mTLSConfigOnce = sync.Once{}
}

// TestMTLSConfig_ModeDisabled tests that mTLS is disabled in disabled mode
func TestMTLSConfig_ModeDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set mTLS to disabled mode
	config := &TLSConfig{Mode: "disabled"}
	SetMTLSConfig(config)

	router := gin.New()
	router.Use(MTLSAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Request without TLS should pass in disabled mode
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should allow request without TLS in disabled mode")
}

// TestMTLSConfig_ModeWarn_NoTLS tests that warn mode logs warnings but allows requests without TLS
func TestMTLSConfig_ModeWarn_NoTLS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set mTLS to warn mode
	config := &TLSConfig{Mode: "warn"}
	SetMTLSConfig(config)

	router := gin.New()
	router.Use(MTLSAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Request without TLS should pass with warning in warn mode
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should allow request without TLS in warn mode")
}

// TestMTLSConfig_ModeStrict_NoTLS tests that strict mode rejects requests without TLS
func TestMTLSConfig_ModeStrict_NoTLS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set mTLS to strict mode
	config := &TLSConfig{Mode: "strict"}
	SetMTLSConfig(config)

	router := gin.New()
	router.Use(MTLSAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Request without TLS should be rejected in strict mode
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code, "Should reject request without TLS in strict mode")
	assert.Contains(t, w.Body.String(), "MTLS_REQUIRED", "Should return MTLS_REQUIRED error")
}

// TestMTLSConfig_GetMTLSConfig tests the GetMTLSConfig helper function
func TestMTLSConfig_GetMTLSConfig(t *testing.T) {
	// Reset to nil
	resetMTLSForTest()

	config := GetMTLSConfig()
	assert.NotNil(t, config, "Should return non-nil config")
	assert.Equal(t, "disabled", config.Mode, "Should default to disabled mode")

	// Set a custom config
	customConfig := &TLSConfig{Mode: "warn"}
	SetMTLSConfig(customConfig)

	config = GetMTLSConfig()
	assert.Equal(t, "warn", config.Mode, "Should return the custom config")
}

func TestInitMTLSConfig_DefaultsStrictInReleaseMode(t *testing.T) {
	resetMTLSForTest()
	t.Setenv("PULSE_SERVER_MODE", "release")
	t.Setenv("PULSE_MTLS_ENABLED", "")

	InitMTLSConfig()

	assert.Equal(t, "strict", GetMTLSConfig().Mode, "release mode should default mTLS to strict")
}

func TestInitMTLSConfig_DefaultsDisabledInDebugMode(t *testing.T) {
	resetMTLSForTest()
	t.Setenv("PULSE_SERVER_MODE", "debug")
	t.Setenv("PULSE_MTLS_ENABLED", "")

	InitMTLSConfig()

	assert.Equal(t, "disabled", GetMTLSConfig().Mode, "debug mode should default mTLS to disabled")
}

func TestInitMTLSConfig_ExplicitModeOverridesReleaseDefault(t *testing.T) {
	resetMTLSForTest()
	t.Setenv("PULSE_SERVER_MODE", "release")
	t.Setenv("PULSE_MTLS_ENABLED", "warn")

	InitMTLSConfig()

	assert.Equal(t, "warn", GetMTLSConfig().Mode, "explicit mTLS mode should override release default")
}

// TestMTLSConfig_ModeValues tests valid mode values
func TestMTLSConfig_ModeValues(t *testing.T) {
	validModes := []string{"disabled", "warn", "strict"}

	for _, mode := range validModes {
		config := &TLSConfig{Mode: mode}
		SetMTLSConfig(config)

		retrieved := GetMTLSConfig()
		assert.Equal(t, mode, retrieved.Mode, "Mode should be preserved: "+mode)
	}
}
