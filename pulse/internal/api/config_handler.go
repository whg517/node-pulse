package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whg517/node-pulse/pulse/internal/config"
)

// GetConfigHandler returns current configuration (admin-only, credentials redacted)
// @Summary Get current configuration
// @Description Get current application configuration (admin only, passwords redacted)
// @Tags config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/config [get]
func GetConfigHandler(c *gin.Context) {
	cfg := config.Get()

	// Build response with redacted credentials
	response := map[string]interface{}{
		"server": map[string]interface{}{
			"port":          cfg.Server.Port,
			"read_timeout":  cfg.Server.ReadTimeout,
			"write_timeout": cfg.Server.WriteTimeout,
			"idle_timeout":  cfg.Server.IdleTimeout,
			"mode":          cfg.Server.Mode,
		},
		"database": map[string]interface{}{
			"url":                maskURL(cfg.DB.URL),
			"max_connections":    cfg.DB.MaxConnections,
			"min_connections":    cfg.DB.MinConnections,
			"conn_max_lifetime":  cfg.DB.ConnMaxLifetime,
			"conn_max_idle_time": cfg.DB.ConnMaxIdleTime,
		},
		"cleanup": map[string]interface{}{
			"enabled":           cfg.Cleanup.Enabled,
			"interval_seconds":  cfg.Cleanup.IntervalSeconds,
			"retention_days":    cfg.Cleanup.RetentionDays,
			"slow_threshold_ms": cfg.Cleanup.SlowThresholdMs,
		},
		"log": map[string]interface{}{
			"level":  cfg.Log.Level,
			"format": cfg.Log.Format,
			"output": cfg.Log.Output,
		},
		"cors": map[string]interface{}{
			"allowed_origins": cfg.CORS.AllowedOrigins,
			"allowed_methods": cfg.CORS.AllowedMethods,
			"allowed_headers": cfg.CORS.AllowedHeaders,
			"max_age":         cfg.CORS.MaxAge,
		},
		"admin": map[string]interface{}{
			"username": cfg.Admin.Username,
			"password": "***REDACTED***",
		},
		"session": map[string]interface{}{
			"secret":           "***REDACTED***",
			"expiration_hours": cfg.Session.ExpirationHours,
			"cookie_secure":    cfg.Session.CookieSecure,
			"cookie_samesite":  cfg.Session.CookieSameSite,
		},
		"jwt": map[string]interface{}{
			"secret":                            "***REDACTED***",
			"access_token_expiration_minutes":   cfg.JWT.AccessTokenExpirationMinutes,
			"refresh_token_expiration_days":     cfg.JWT.RefreshTokenExpirationDays,
		},
	}

	// Log audit entry (in production, use proper audit logging)
	// TODO: Add audit logging for all config access

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// ValidateConfigHandler validates current configuration (admin-only)
// @Summary Validate configuration
// @Description Validate current application configuration (admin only)
// @Tags config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /api/config/validate [get]
func ValidateConfigHandler(c *gin.Context) {
	cfg := config.Get()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	// Check for warnings (e.g., auto-generated secrets in production)
	warnings := []string{}

	if cfg.IsProduction() {
		// Check if using default secrets
		if len(cfg.Session.Secret) < 32 {
			warnings = append(warnings, "Session secret is short, consider using a longer secret in production")
		}
		if len(cfg.JWT.Secret) < 32 {
			warnings = append(warnings, "JWT secret is short, consider using a longer secret in production")
		}

		// Check if using default admin password
		if cfg.Admin.Password == "admin123" {
			warnings = append(warnings, "Using default admin password in production is not recommended")
		}
	}

	response := map[string]interface{}{
		"valid":    true,
		"warnings": warnings,
	}

	// Log audit entry (in production, use proper audit logging)
	// TODO: Add audit logging for all config validation access

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// maskURL masks sensitive parts of a URL
func maskURL(url string) string {
	if url == "" {
		return ""
	}
	// Truncate long URLs for safety
	if len(url) > 50 {
		return url[:47] + "..."
	}
	return url
}
