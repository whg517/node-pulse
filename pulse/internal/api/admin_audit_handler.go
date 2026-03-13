package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whg517/node-pulse/pulse/internal/auth"
	_ "github.com/whg517/node-pulse/pulse/internal/models" // imported for swagger type resolution
)

// AdminAuditHandler handles audit log query endpoints for admins
type AdminAuditHandler struct {
	pool        *pgxpool.Pool
	auditLogger *auth.AuditLogger
}

// NewAdminAuditHandler creates a new admin audit handler
func NewAdminAuditHandler(pool *pgxpool.Pool) *AdminAuditHandler {
	return &AdminAuditHandler{
		pool:        pool,
		auditLogger: auth.NewAuditLogger(pool),
	}
}

// GetAuditLogs retrieves audit logs with filtering and pagination
// @Summary Get audit logs (Admin)
// @Description Query audit logs with filters and pagination
// @Tags admin
// @Produce json
// @Security Bearer
// @Param event_type query string false "Filter by event type"
// @Param user_id query string false "Filter by user ID"
// @Param start_time query string false "Start time (RFC3339)"
// @Param end_time query string false "End time (RFC3339)"
// @Param limit query int false "Page size (default 50, max 100)"
// @Param offset query int false "Page offset (default 0)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/admin/audit/logs [get]
func (h *AdminAuditHandler) GetAuditLogs(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	filters := auth.AuditLogFilter{
		Limit:  50,
		Offset: 0,
	}

	if eventType := c.Query("event_type"); eventType != "" {
		filters.EventType = eventType
	}

	if userID := c.Query("user_id"); userID != "" {
		filters.UserID = &userID
	}

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "INVALID_DATE_FORMAT",
				"message": "start_time must be in RFC3339 format",
			})
			return
		}
		filters.StartTime = &startTime
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "INVALID_DATE_FORMAT",
				"message": "end_time must be in RFC3339 format",
			})
			return
		}
		filters.EndTime = &endTime
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 || limit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "INVALID_LIMIT",
				"message": "limit must be between 0 and 100",
			})
			return
		}
		filters.Limit = limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "INVALID_OFFSET",
				"message": "offset must be >= 0",
			})
			return
		}
		filters.Offset = offset
	}

	// Query audit logs
	logs, totalCount, err := h.auditLogger.QueryAuditLogs(ctx, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "QUERY_FAILED",
			"message": "Failed to query audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":        logs,
		"total_count": totalCount,
		"limit":       filters.Limit,
		"offset":      filters.Offset,
	})
}

// GetAuditLogByID retrieves a single audit log entry by ID
// @Summary Get audit log by ID (Admin)
// @Description Get a specific audit log entry
// @Tags admin
// @Produce json
// @Security Bearer
// @Param id path int true "Audit log ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/admin/audit/logs/:id [get]
func (h *AdminAuditHandler) GetAuditLogByID(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_ID",
			"message": "Invalid audit log ID",
		})
		return
	}

	log, err := h.auditLogger.GetAuditLogByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Audit log entry not found",
		})
		return
	}

	c.JSON(http.StatusOK, log)
}
