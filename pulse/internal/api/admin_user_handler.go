package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/whg517/node-pulse/pulse/internal/auth"
	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/pkg/middleware"
)

var (
	// ErrUsernameRequired is returned when username is missing
	ErrUsernameRequired = "ERR_USERNAME_REQUIRED"
	// ErrPasswordRequired is returned when password is missing
	ErrPasswordRequired = "ERR_PASSWORD_REQUIRED"
	// ErrRoleRequired is returned when role is missing
	ErrRoleRequired = "ERR_ROLE_REQUIRED"
	// ErrInvalidRole is returned when role is invalid
	ErrInvalidRole = "ERR_INVALID_ROLE"
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFoundCode = "ERR_USER_NOT_FOUND"
	// ErrCannotDeleteSelf is returned when attempting to delete own account
	ErrCannotDeleteSelf = "ERR_CANNOT_DELETE_SELF"
	// ErrLastAdminCode is returned when attempting to delete last admin
	ErrLastAdminCode = "ERR_LAST_ADMIN"
	// ErrPasswordInvalid is returned when password doesn't meet requirements
	ErrPasswordInvalid = "ERR_PASSWORD_INVALID"
)

// AdminUserHandler handles admin user API requests
type AdminUserHandler struct {
	userQuerier db.UserQuerier
	auditLogger *auth.AuditLogger
}

// NewAdminUserHandler creates a new AdminUserHandler
func NewAdminUserHandler(userQuerier db.UserQuerier, auditLogger *auth.AuditLogger) *AdminUserHandler {
	return &AdminUserHandler{
		userQuerier: userQuerier,
		auditLogger: auditLogger,
	}
}

// ListUsers handles GET /api/v1/admin/users
// @Summary		List all users
// @Description	Retrieves a paginated list of users. Admin role required.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Param			limit		query		int		false	"Number of items per page"	default(100)	maximum(100)
// @Param			offset		query		int		false	"Number of items to skip"	default(0)
// @Success		200		{object}	models.ListUsersResponse	"List of users"
// @Failure		401		{object}	models.ErrorResponse	"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse	"Forbidden (requires admin role)"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Security		BearerAuth
// @Router			/admin/users [get]
func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	// RBAC is handled by middleware - only admin can reach this handler

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 100
	}
	if limit > 100 {
		limit = 100 // Max 100 items per page
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	ctx := context.Background()

	// Get users from database
	users, total, err := h.userQuerier.ListUsers(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "Failed to retrieve users",
			Details: err.Error(),
		})
		return
	}

	// Calculate pagination info
	hasNext := offset+limit < total
	hasPrev := offset > 0

	c.JSON(http.StatusOK, models.ListUsersResponse{
		Data: struct {
			Users      []*models.User `json:"users"`
			Total      int            `json:"total"`
			Limit      int            `json:"limit"`
			Offset     int            `json:"offset"`
			Pagination struct {
				HasNext bool `json:"has_next"`
				HasPrev bool `json:"has_prev"`
			} `json:"pagination"`
		}{
			Users:  users,
			Total:  total,
			Limit:  limit,
			Offset: offset,
			Pagination: struct {
				HasNext bool `json:"has_next"`
				HasPrev bool `json:"has_prev"`
			}{
				HasNext: hasNext,
				HasPrev: hasPrev,
			},
		},
		Message:   "Users retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetUser handles GET /api/v1/admin/users/:id
// @Summary		Get user by ID
// @Description	Retrieves detailed information about a specific user. Admin role required.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Param			id		path		string	true	"User UUID"
// @Success		200		{object}	models.GetUserResponse	"User details"
// @Failure		400		{object}	models.ErrorResponse	"Invalid UUID format"
// @Failure		401		{object}	models.ErrorResponse	"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse	"Forbidden (requires admin role)"
// @Failure		404		{object}	models.ErrorResponse	"User not found"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Security		BearerAuth
// @Router			/admin/users/{id} [get]
func (h *AdminUserHandler) GetUser(c *gin.Context) {
	// RBAC is handled by middleware - only admin can reach this handler

	// Parse UUID from path parameter
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "Invalid user ID format",
			Details: map[string]interface{}{
				"user_id": idParam,
				"error":   err.Error(),
			},
		})
		return
	}

	ctx := context.Background()

	// Get user from database
	user, err := h.userQuerier.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrUserNotFoundCode,
				Message: "User not found",
				Details: map[string]interface{}{
					"user_id": idParam,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "Failed to retrieve user",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.GetUserResponse{
		Data: struct {
			User *models.User `json:"user"`
		}{
			User: user,
		},
		Message:   "User retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// CreateUser handles POST /api/v1/admin/users
// @Summary		Create a new user
// @Description	Creates a new user with the specified role. Admin role required.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Param			request	body		models.CreateUserRequest	true	"User creation request"
// @Success		201		{object}	models.CreateUserResponse	"User created successfully"
// @Failure		400		{object}	models.ErrorResponse	"Invalid request parameters"
// @Failure		401		{object}	models.ErrorResponse	"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse	"Forbidden (requires admin role)"
// @Failure		409		{object}	models.ErrorResponse	"Username already exists"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Security		BearerAuth
// @Router			/admin/users [post]
func (h *AdminUserHandler) CreateUser(c *gin.Context) {
	// RBAC is handled by middleware - only admin can reach this handler

	// Get requesting user ID for audit logging
	requestingUserID := c.GetString("user_id")

	// Parse request body
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrUsernameRequired,
			Message: "Username is required",
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrPasswordRequired,
			Message: "Password is required",
		})
		return
	}

	if req.Role == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrRoleRequired,
			Message: "Role is required",
		})
		return
	}

	// Validate role
	if !auth.IsValidRole(req.Role) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrInvalidRole,
			Message: "Invalid role. Must be one of: admin, operator, viewer",
			Details: map[string]interface{}{
				"provided_role": req.Role,
				"valid_roles":   []string{"admin", "operator", "viewer"},
			},
		})
		return
	}

	// Validate password strength
	if err := auth.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrPasswordInvalid,
			Message: "Password does not meet requirements",
			Details: map[string]interface{}{
				"requirements": "Must be 8-32 characters with uppercase, lowercase, and digit",
				"error":        err.Error(),
			},
		})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_HASH_ERROR",
			Message: "Failed to hash password",
			Details: err.Error(),
		})
		return
	}

	ctx := context.Background()

	// Create user
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
	}

	err = h.userQuerier.CreateUser(ctx, user, passwordHash)
	if err != nil {
		if errors.Is(err, db.ErrUserExists) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Code:    "ERR_USER_EXISTS",
				Message: "Username already exists",
				Details: map[string]interface{}{
					"username": req.Username,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "Failed to create user",
			Details: err.Error(),
		})
		return
	}

	// Audit logging
	if h.auditLogger != nil {
		requestingUserUUID, _ := uuid.Parse(requestingUserID)
		createdUserUUID, _ := uuid.Parse(user.UserID)
		_ = h.auditLogger.LogEvent(ctx, "user_created", &requestingUserUUID, c.ClientIP(), map[string]interface{}{
			"created_user_id": createdUserUUID.String(),
			"username":        user.Username,
			"role":            user.Role,
		})
	}

	c.JSON(http.StatusCreated, models.CreateUserResponse{
		Data: struct {
			User *models.User `json:"user"`
		}{
			User: user,
		},
		Message:   "User created successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// UpdateUser handles PUT /api/v1/admin/users/:id
// @Summary		Update a user
// @Description	Updates an existing user. Admin role required.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Param			id		path		string	true	"User UUID"
// @Param			request	body		models.UpdateUserRequest	true	"User update request"
// @Success		200		{object}	models.UpdateUserResponse	"User updated successfully"
// @Failure		400		{object}	models.ErrorResponse	"Invalid request parameters"
// @Failure		401		{object}	models.ErrorResponse	"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse	"Forbidden (requires admin role)"
// @Failure		404		{object}	models.ErrorResponse	"User not found"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Security		BearerAuth
// @Router			/admin/users/{id} [put]
func (h *AdminUserHandler) UpdateUser(c *gin.Context) {
	// RBAC is handled by middleware - only admin can reach this handler

	// Get requesting user ID for audit logging
	requestingUserID := c.GetString("user_id")

	// Parse UUID from path parameter
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "Invalid user ID format",
			Details: map[string]interface{}{
				"user_id": idParam,
				"error":   err.Error(),
			},
		})
		return
	}

	// Parse request body
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "Invalid request parameters",
			Details: err.Error(),
		})
		return
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Username != nil {
		if *req.Username == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrUsernameRequired,
				Message: "Username cannot be empty",
			})
			return
		}
		updates["username"] = *req.Username
	}

	if req.Email != nil {
		updates["email"] = req.Email
	}

	if req.Password != nil {
		if *req.Password == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrPasswordRequired,
				Message: "Password cannot be empty",
			})
			return
		}

		// Validate password strength
		if err := auth.ValidatePassword(*req.Password); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrPasswordInvalid,
				Message: "Password does not meet requirements",
				Details: map[string]interface{}{
					"requirements": "Must be 8-32 characters with uppercase, lowercase, and digit",
					"error":        err.Error(),
				},
			})
			return
		}

		// Hash password
		passwordHash, err := auth.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:    "ERR_HASH_ERROR",
				Message: "Failed to hash password",
				Details: err.Error(),
			})
			return
		}
		updates["password"] = passwordHash
	}

	if req.Role != nil {
		// Validate role
		if !auth.IsValidRole(*req.Role) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrInvalidRole,
				Message: "Invalid role. Must be one of: admin, operator, viewer",
				Details: map[string]interface{}{
					"provided_role": *req.Role,
					"valid_roles":   []string{"admin", "operator", "viewer"},
				},
			})
			return
		}
		updates["role"] = *req.Role
	}

	// Check if there are any updates
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "No fields to update",
		})
		return
	}

	ctx := context.Background()

	// Update user
	err = h.userQuerier.UpdateUser(ctx, userID, updates)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrUserNotFoundCode,
				Message: "User not found",
				Details: map[string]interface{}{
					"user_id": idParam,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "Failed to update user",
			Details: err.Error(),
		})
		return
	}

	// Fetch updated user
	user, err := h.userQuerier.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "Failed to retrieve updated user",
			Details: err.Error(),
		})
		return
	}

	// Audit logging
	if h.auditLogger != nil {
		requestingUserUUID, _ := uuid.Parse(requestingUserID)
		_ = h.auditLogger.LogEvent(ctx, "user_updated", &requestingUserUUID, c.ClientIP(), map[string]interface{}{
			"updated_user_id": userID.String(),
			"username":        user.Username,
			"role":            user.Role,
		})
	}

	c.JSON(http.StatusOK, models.UpdateUserResponse{
		Data: struct {
			User *models.User `json:"user"`
		}{
			User: user,
		},
		Message:   "User updated successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// DeleteUser handles DELETE /api/v1/admin/users/:id
// @Summary		Delete a user
// @Description	Deletes a user. Cannot delete own account or the last admin. Admin role required.
// @Tags			admin
// @Accept			json
// @Produce		json
// @Param			id		path		string	true	"User UUID"
// @Success		200		{object}	models.DeleteUserResponse	"User deleted successfully"
// @Failure		400		{object}	models.ErrorResponse	"Invalid request (e.g., deleting self or last admin)"
// @Failure		401		{object}	models.ErrorResponse	"Unauthorized"
// @Failure		403		{object}	models.ErrorResponse	"Forbidden (requires admin role)"
// @Failure		404		{object}	models.ErrorResponse	"User not found"
// @Failure		500		{object}	models.ErrorResponse	"Internal server error"
// @Security		BearerAuth
// @Router			/admin/users/{id} [delete]
func (h *AdminUserHandler) DeleteUser(c *gin.Context) {
	// RBAC is handled by middleware - only admin can reach this handler

	// Get requesting user ID for validation and audit logging
	requestingUserID := c.GetString("user_id")

	// Parse UUID from path parameter
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    middleware.ERR_INVALID_REQUEST,
			Message: "Invalid user ID format",
			Details: map[string]interface{}{
				"user_id": idParam,
				"error":   err.Error(),
			},
		})
		return
	}

	// Cannot delete own account
	if idParam == requestingUserID {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Code:    ErrCannotDeleteSelf,
			Message: "Cannot delete your own account",
			Details: map[string]interface{}{
				"user_id": idParam,
			},
		})
		return
	}

	ctx := context.Background()

	// Get user to check role before deletion
	user, err := h.userQuerier.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrUserNotFoundCode,
				Message: "User not found",
				Details: map[string]interface{}{
					"user_id": idParam,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "Failed to retrieve user",
			Details: err.Error(),
		})
		return
	}

	// Delete user
	err = h.userQuerier.DeleteUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Code:    ErrUserNotFoundCode,
				Message: "User not found",
				Details: map[string]interface{}{
					"user_id": idParam,
				},
			})
			return
		}

		if errors.Is(err, db.ErrLastAdmin) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Code:    ErrLastAdminCode,
				Message: "Cannot delete the last admin user",
				Details: map[string]interface{}{
					"user_id": idParam,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    "ERR_DATABASE_ERROR",
			Message: "Failed to delete user",
			Details: err.Error(),
		})
		return
	}

	// Audit logging
	if h.auditLogger != nil {
		requestingUserUUID, _ := uuid.Parse(requestingUserID)
		_ = h.auditLogger.LogEvent(ctx, "user_deleted", &requestingUserUUID, c.ClientIP(), map[string]interface{}{
			"deleted_user_id": userID.String(),
			"username":        user.Username,
			"role":            user.Role,
		})
	}

	c.JSON(http.StatusOK, models.DeleteUserResponse{
		Message:   "User deleted successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}
