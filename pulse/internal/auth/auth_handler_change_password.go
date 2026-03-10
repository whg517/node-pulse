package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// ChangePassword handles password changes
// @Summary Change password
// @Description Change user password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ChangePasswordRequest true "Password change request"
// @Success 200 {object} models.ChangePasswordResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/auth/password/change [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user from context (set by JWT middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authorization required",
		})
		return
	}

	// Validate Content-Type header (prevent content-type confusion attacks)
	contentType := c.GetHeader("Content-Type")
	if contentType != "application/json" {
		c.JSON(http.StatusUnsupportedMediaType, models.ErrorResponse{
			Code:    "UNSUPPORTED_MEDIA_TYPE",
			Message: "Content-Type must be application/json",
		})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
		})
		return
	}

	// Get current password hash from database
	var currentPasswordHash string
	err := h.pool.QueryRow(ctx, `
		SELECT password_hash FROM users WHERE user_id = $1
	`, userID).Scan(&currentPasswordHash)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "USER_LOOKUP_FAILED",
			Message: "Failed to retrieve user information",
		})
		return
	}

	// Verify current password
	err = VerifyPassword(req.CurrentPassword, currentPasswordHash)
	if err != nil {
		// Log failed password change attempt
		h.logAuditEvent(ctx, "password_change_failed", &userID, c.ClientIP(), nil)

		constantAuthDelay()

		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    "INVALID_CURRENT_PASSWORD",
			Message: "Current password is incorrect",
		})
		return
	}

	// Validate new password strength
	err = ValidatePassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "WEAK_PASSWORD",
			Message: err.Error(),
		})
		return
	}

	// Check new password is not the same as current
	err = VerifyPassword(req.NewPassword, currentPasswordHash)
	if err == nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    "SAME_PASSWORD",
			Message: "New password must be different from current password",
		})
		return
	}

	// Hash new password
	newPasswordHash, err := HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "PASSWORD_HASH_FAILED",
			Message: "Failed to hash new password",
		})
		return
	}

	// Update password in database
	_, err = h.pool.Exec(ctx, `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE user_id = $2
	`, newPasswordHash, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "PASSWORD_UPDATE_FAILED",
			Message: "Failed to update password",
		})
		return
	}

	// Revoke all other sessions (keep current session)
	// Get current JTI from access token
	accessToken := c.GetHeader("Authorization")
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	_, err = h.jwtService.GetJTI(accessToken)
	if err == nil {
		// Revoke all refresh tokens except current session
		// We identify current session by checking recent tokens with this JTI
		_, err = h.pool.Exec(ctx, `
			UPDATE refresh_tokens
			SET revoked_at = NOW(), updated_at = NOW()
			WHERE user_id = $1
			  AND token_id NOT IN (
				SELECT token_id FROM refresh_tokens
				WHERE user_id = $1
				  AND created_at >= NOW() - INTERVAL '1 minute'
				LIMIT 1
			  )
			  AND revoked_at IS NULL
		`, userID)

		if err != nil {
			// Log error but don't fail the password change
			fmt.Printf("WARN: failed to revoke other sessions: %v\n", err)
		}
	}

	// Log successful password change
	h.logAuditEvent(ctx, "password_changed", &userID, c.ClientIP(), nil)

	c.JSON(http.StatusOK, models.ChangePasswordResponse{
		Message:         "Password changed successfully",
		SessionsRevoked: true,
	})
}
