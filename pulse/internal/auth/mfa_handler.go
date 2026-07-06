package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/whg517/node-pulse/pulse/internal/csrf"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// MFALoginHandler completes a two-factor login. The client posts the
// mfa_ticket from Login's `mfa_required` response plus a TOTP code; on
// success it receives the same access/refresh tokens a normal login returns.
//
// @Summary		Complete 2FA login
// @Description	Exchange an MFA ticket + TOTP code for access/refresh tokens
// @Tags			auth,mfa
// @Accept		json
// @Produce		json
// @Param		body	body	models.MFALoginRequest	true	"ticket + code"
// @Success		200	{object}	models.LoginResponse
// @Failure		400	{object}	models.ErrorResponse	"Invalid ticket or code"
// @Router		/auth/login/mfa [post]
func (h *AuthHandler) MFALoginHandler(c *gin.Context) {
	if h.mfaService == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Code: "MFA_DISABLED", Message: "MFA is not enabled on this server",
		})
		return
	}
	var req models.MFALoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "INVALID_REQUEST", Message: "Invalid request format"})
		return
	}

	ctx := c.Request.Context()
	ipAddress := c.ClientIP()

	userID, username, role, err := h.mfaService.VerifyLogin(ctx, req.MFATicket, req.Code)
	if err != nil {
		h.logAuditEvent(ctx, "mfa_login_failed", &userID, ipAddress, map[string]interface{}{"reason": err.Error()})
		constantAuthDelay()
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Code: "ERR_MFA_INVALID", Message: err.Error()})
		return
	}

	// Reset failed attempts now that the full login (password + MFA) succeeded.
	_, _ = h.pool.Exec(ctx, `UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE user_id = $1`, userID)

	accessToken, jti, err := h.jwtService.GenerateAccessToken(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "TOKEN_GENERATION_FAILED", Message: "Failed to generate access token"})
		return
	}
	userAgent := c.GetHeader("User-Agent")
	refreshToken, _, err := h.refreshTokenService.CreateRefreshToken(ctx, userID, userAgent, ipAddress, h.maxValidityDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "TOKEN_GENERATION_FAILED", Message: "Failed to generate refresh token"})
		return
	}

	h.logAuditEvent(ctx, "mfa_login", &userID, ipAddress, map[string]interface{}{"jti": jti, "user_agent": userAgent})

	csrfToken, err := csrf.GenerateCSRFToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "CSRF_GENERATION_FAILED", Message: "Failed to generate CSRF token"})
		return
	}
	csrf.SetCSRFCookie(c, csrfToken, h.cookieSecure)
	sameSiteMode := http.SameSiteLaxMode
	if h.cookieSecure {
		sameSiteMode = http.SameSiteStrictMode
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name: "refresh_token", Value: refreshToken, Path: "/api/v1/auth",
		HttpOnly: true, Secure: h.cookieSecure, SameSite: sameSiteMode,
		MaxAge: int((time.Duration(h.refreshExpirationDays*24) * time.Hour).Seconds()),
	})

	resp := models.LoginResponse{}
	resp.Data.UserID = userID
	resp.Data.Username = username
	resp.Data.Role = role
	resp.Data.AccessToken = accessToken
	resp.Data.CSRFToken = csrfToken
	resp.Message = "Login successful"
	resp.Timestamp = time.Now().Format(time.RFC3339)
	c.JSON(http.StatusOK, resp)
}

// MFASetupHandler begins enabling 2FA for the authenticated user. Returns a
// TOTP secret + otpauth URI (render as a QR) and a setup ticket the verify
// step consumes. The secret is not persisted until MFAVerifySetupHandler.
//
// @Summary		Begin 2FA enrollment
// @Description	Generate a TOTP secret + otpauth URI for the user's authenticator
// @Tags			auth,mfa
// @Produce		json
// @Success		200	{object}	models.MFASetupResponse
// @Failure		409	{object}	models.ErrorResponse	"MFA already enabled"
// @Security		BearerAuth
// @Router		/auth/mfa/setup [post]
func (h *AuthHandler) MFASetupHandler(c *gin.Context) {
	if h.mfaService == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "MFA_DISABLED", Message: "MFA is not enabled on this server"})
		return
	}
	userID := c.GetString("user_id")
	username := c.GetString("username")
	if username == "" {
		// Fallback: look the username up so the authenticator shows a real label.
		_ = h.pool.QueryRow(c.Request.Context(), `SELECT username FROM users WHERE user_id = $1`, userID).Scan(&username)
	}
	setup, err := h.mfaService.GenerateSecret(c.Request.Context(), userID, username)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{Code: "MFA_SETUP_FAILED", Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": setup, "message": "Scan the QR, then verify with a code", "timestamp": time.Now().Format(time.RFC3339)})
}

// MFAVerifySetupHandler confirms enrollment: the user enters a code from their
// authenticator, we validate it against the pending secret, and persist
// mfa_enabled=true. Single-use ticket.
//
// @Summary		Confirm 2FA enrollment
// @Description	Verify a TOTP code and persist the MFA secret
// @Tags			auth,mfa
// @Accept		json
// @Produce		json
// @Param		body	body	map[string]string	true	"setup ticket + code"
// @Success		200	{object}	map[string]interface{}
// @Failure		400	{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router		/auth/mfa/verify [post]
func (h *AuthHandler) MFAVerifySetupHandler(c *gin.Context) {
	if h.mfaService == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "MFA_DISABLED", Message: "MFA is not enabled on this server"})
		return
	}
	var req struct {
		Ticket string `json:"ticket" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "INVALID_REQUEST", Message: "Invalid request format"})
		return
	}
	if err := h.mfaService.VerifySetup(c.Request.Context(), req.Ticket, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "MFA_VERIFY_FAILED", Message: err.Error()})
		return
	}
	userID := c.GetString("user_id")
	h.logAuditEvent(c.Request.Context(), "mfa_enabled", &userID, c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"message": "MFA enabled", "timestamp": time.Now().Format(time.RFC3339)})
}

// MFADisableHandler turns MFA off for the authenticated user. Requires the
// current password to prevent a hijacked session from silently weakening the
// account. Admins can also disable via the admin user endpoint family.
//
// @Summary		Disable 2FA
// @Description	Turn MFA off for the current user (requires current password)
// @Tags			auth,mfa
// @Accept		json
// @Produce		json
// @Param		body	body	map[string]string	true	"current password"
// @Success		200	{object}	map[string]interface{}
// @Failure		400	{object}	models.ErrorResponse
// @Security		BearerAuth
// @Router		/auth/mfa/disable [post]
func (h *AuthHandler) MFADisableHandler(c *gin.Context) {
	if h.mfaService == nil {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{Code: "MFA_DISABLED", Message: "MFA is not enabled on this server"})
		return
	}
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "INVALID_REQUEST", Message: "Password required"})
		return
	}
	userID := c.GetString("user_id")
	ctx := c.Request.Context()

	var hash string
	if err := h.pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE user_id = $1`, userID).Scan(&hash); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "DB_ERROR", Message: "Failed to verify"})
		return
	}
	if err := VerifyPassword(req.Password, hash); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Code: "ERR_INVALID_CREDENTIALS", Message: "Current password is incorrect"})
		return
	}
	if err := h.mfaService.Disable(ctx, userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Code: "DB_ERROR", Message: err.Error()})
		return
	}
	h.logAuditEvent(ctx, "mfa_disabled", &userID, c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"message": "MFA disabled", "timestamp": time.Now().Format(time.RFC3339)})
}

// MFAStatusHandler reports whether the authenticated user has MFA enabled.
// Used by the preferences page to render the enable/disable card.
//
// @Summary		Get 2FA status
// @Description	Whether the current user has MFA enabled
// @Tags			auth,mfa
// @Produce		json
// @Success		200	{object}	map[string]interface{}
// @Security		BearerAuth
// @Router		/auth/mfa/status [get]
func (h *AuthHandler) MFAStatusHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	enabled := false
	if h.mfaService != nil {
		if v, err := h.mfaService.IsEnabled(c.Request.Context(), userID); err == nil {
			enabled = v
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"enabled": enabled}, "timestamp": time.Now().Format(time.RFC3339)})
}
