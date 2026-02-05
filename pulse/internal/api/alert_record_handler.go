package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/db"
)

// AlertRecordHandler handles alert record-related HTTP requests
type AlertRecordHandler struct {
	pool *pgxpool.Pool
}

// NewAlertRecordHandler creates a new alert record handler
func NewAlertRecordHandler(pool *pgxpool.Pool) *AlertRecordHandler {
	return &AlertRecordHandler{
		pool: pool,
	}
}

// GetAlertRecordsHandler retrieves alert records with optional filtering
func (h *AlertRecordHandler) GetAlertRecordsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	pool := h.pool

	// Parse query parameters
	var filters db.AlertRecordFilters
	filters.Limit = 50 // Default limit
	filters.Offset = 0 // Default offset

	// Parse node_id filter
	if nodeID := c.Query("node_id"); nodeID != "" {
		filters.NodeID = &nodeID
	}

	// Parse level filter
	if level := c.Query("level"); level != "" {
		// Validate level
		if level != "P0" && level != "P1" && level != "P2" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ERR_INVALID_LEVEL",
				"message": "Invalid level parameter. Must be P0, P1, or P2",
			})
			return
		}
		filters.Level = &level
	}

	// Parse status filter
	if status := c.Query("status"); status != "" {
		// Validate status
		if status != "pending" && status != "in_progress" && status != "resolved" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ERR_INVALID_STATUS",
				"message": "Invalid status parameter. Must be pending, in_progress, or resolved",
			})
			return
		}
		filters.Status = &status
	}

	// Parse start_time filter
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ERR_INVALID_START_TIME",
				"message": "Invalid start_time format. Use ISO 8601 format (e.g., 2026-01-01T00:00:00Z)",
				"details": gin.H{"error": err.Error()},
			})
			return
		}
		filters.StartTime = &startTime
	}

	// Parse end_time filter
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ERR_INVALID_END_TIME",
				"message": "Invalid end_time format. Use ISO 8601 format (e.g., 2026-01-01T00:00:00Z)",
				"details": gin.H{"error": err.Error()},
			})
			return
		}
		filters.EndTime = &endTime
	}

	// Parse limit filter
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ERR_INVALID_LIMIT",
				"message": "Invalid limit parameter. Must be an integer between 1 and 100",
			})
			return
		}
		filters.Limit = limit
	}

	// Parse offset filter
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ERR_INVALID_OFFSET",
				"message": "Invalid offset parameter. Must be a non-negative integer",
			})
			return
		}
		filters.Offset = offset
	}

	// Query alert records
	records, err := db.GetAlertRecords(ctx, pool, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_QUERY_ALERT_RECORDS",
			"message": "Failed to query alert records",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      records,
		"message":   "Alert records retrieved successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// UpdateAlertRecordStatusHandler updates the status of an alert record
func (h *AlertRecordHandler) UpdateAlertRecordStatusHandler(c *gin.Context) {
	ctx := c.Request.Context()
	pool := h.pool

	// Get record ID from URL parameter
	recordID := c.Param("id")
	if recordID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_MISSING_RECORD_ID",
			"message": "Missing alert record ID",
		})
		return
	}

	// Parse request body
	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_BODY",
			"message": "Invalid request body. Status field is required",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	// Validate status
	if req.Status != "pending" && req.Status != "in_progress" && req.Status != "resolved" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_STATUS",
			"message": "Invalid status. Must be pending, in_progress, or resolved",
		})
		return
	}

	// Update alert record status
	if err := db.UpdateAlertRecordStatus(ctx, pool, recordID, req.Status); err != nil {
		// Check if it's a "not found" error
		if errors.Is(err, db.ErrAlertRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "ERR_RECORD_NOT_FOUND",
				"message": "Alert record not found",
			})
			return
		}

		// Check if it's an invalid transition error
		if errors.Is(err, db.ErrInvalidStatusTransition) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ERR_INVALID_STATUS_TRANSITION",
				"message": err.Error(),
			})
			return
		}

		// Other errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_UPDATE_STATUS",
			"message": "Failed to update alert record status",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	// Get updated record
	record, err := db.GetAlertRecordByID(ctx, pool, recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_GET_RECORD",
			"message": "Failed to retrieve updated alert record",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      record,
		"message":   "Alert record status updated successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
