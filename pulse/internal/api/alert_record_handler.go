package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/realtime"
)

// AlertRecordHandler handles alert record-related HTTP requests
type AlertRecordHandler struct {
	pool        *pgxpool.Pool
	realtimeHub *realtime.Hub
}

// NewAlertRecordHandler creates a new alert record handler
func NewAlertRecordHandler(pool *pgxpool.Pool, realtimeHub ...*realtime.Hub) *AlertRecordHandler {
	var hub *realtime.Hub
	if len(realtimeHub) > 0 {
		hub = realtimeHub[0]
	}

	return &AlertRecordHandler{
		pool:        pool,
		realtimeHub: hub,
	}
}

// GetAlertRecordsHandler retrieves alert records with optional filtering
// @Summary		List alert records
// @Description	Retrieves alert records with optional filtering by node_id, level, status, and time range.
// @Tags			Alert Records
// @Accept			json
// @Produce		json
// @Param			node_id		query		string	false	"Filter by node UUID"
// @Param			level		query		string	false	"Filter by level (P0, P1, P2)"					Enums(P0, P1, P2)
// @Param			status		query		string	false	"Filter by status (pending, in_progress, resolved)"	Enums(pending, in_progress, resolved)
// @Param			start_time	query		string	false	"Filter from this time (ISO 8601)"
// @Param			end_time	query		string	false	"Filter until this time (ISO 8601)"
// @Param			limit		query		int		false	"Maximum records to return (1-100)"		default(50)
// @Param			offset		query		int		false	"Records to skip"						default(0)
// @Success		200	{object}	map[string]interface{}	"List of alert records"
// @Failure		400	{object}	map[string]interface{}	"Invalid query parameters"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		500	{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/alerts/records [get]
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
// @Summary		Update alert record status
// @Description	Updates the status of an alert record. Valid status transitions: pending→in_progress, in_progress→resolved.
// @Tags			Alert Records
// @Accept			json
// @Produce		json
// @Param			id		path		string					true	"Alert record ID"
// @Param			request	body		object					true	"Status update request"
// @Success		200		{object}	map[string]interface{}	"Alert record status updated"
// @Failure		400		{object}	map[string]interface{}	"Invalid request or status transition"
// @Failure		401		{object}	map[string]interface{}	"Unauthorized"
// @Failure		404		{object}	map[string]interface{}	"Alert record not found"
// @Failure		500		{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/alerts/records/{id}/status [put]
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
		Note   string `json:"note,omitempty"`
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

	var createdNote *models.AlertNote
	if strings.TrimSpace(req.Note) != "" {
		userID := c.GetString("user_id")
		note, err := db.CreateAlertNote(ctx, pool, recordID, &userID, req.Note)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "ERR_CREATE_NOTE",
				"message": "Status was updated, but failed to create alert note",
				"details": gin.H{"error": err.Error()},
			})
			return
		}
		createdNote = note
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

	notes, err := db.GetAlertNotes(ctx, pool, recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_GET_NOTES",
			"message": "Failed to retrieve alert notes",
			"details": gin.H{"error": err.Error()},
		})
		return
	}
	record.Notes = notes
	h.broadcastRecordStatus(record)
	if createdNote != nil && h.realtimeHub != nil {
		h.realtimeHub.BroadcastAlertNote(createdNote)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      record,
		"message":   "Alert record status updated successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// AddAlertNoteHandler creates an operator note for an alert record.
// @Summary		Add alert record note
// @Description	Adds an operator note to an alert record without changing its status.
// @Tags			Alert Records
// @Accept			json
// @Produce		json
// @Param			id		path		string	true	"Alert record ID"
// @Param			request	body		object	true	"Note request"
// @Success		200	{object}	map[string]interface{}	"Alert note added"
// @Failure		400	{object}	map[string]interface{}	"Invalid request"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		404	{object}	map[string]interface{}	"Alert record not found"
// @Failure		500	{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/alerts/records/{id}/notes [post]
func (h *AlertRecordHandler) AddAlertNoteHandler(c *gin.Context) {
	recordID := c.Param("id")
	if recordID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_MISSING_RECORD_ID",
			"message": "Missing alert record ID",
		})
		return
	}

	var req struct {
		Note string `json:"note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INVALID_BODY",
			"message": "Invalid request body. Note field is required",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	userID := c.GetString("user_id")
	note, err := db.CreateAlertNote(c.Request.Context(), h.pool, recordID, &userID, req.Note)
	if err != nil {
		if errors.Is(err, db.ErrAlertRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "ERR_RECORD_NOT_FOUND",
				"message": "Alert record not found",
			})
			return
		}
		if errors.Is(err, db.ErrAlertNoteEmpty) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "ERR_EMPTY_NOTE",
				"message": "Note content cannot be empty",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_CREATE_NOTE",
			"message": "Failed to create alert note",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      note,
		"message":   "Alert note added successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
	if h.realtimeHub != nil {
		h.realtimeHub.BroadcastAlertNote(note)
	}
}

// GetAlertNotesHandler lists notes for an alert record.
// @Summary		List alert record notes
// @Description	Retrieves notes for an alert record ordered oldest-first.
// @Tags			Alert Records
// @Produce		json
// @Param			id	path	string	true	"Alert record ID"
// @Success		200	{object}	map[string]interface{}	"Alert notes"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		404	{object}	map[string]interface{}	"Alert record not found"
// @Failure		500	{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/alerts/records/{id}/notes [get]
func (h *AlertRecordHandler) GetAlertNotesHandler(c *gin.Context) {
	recordID := c.Param("id")
	if recordID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_MISSING_RECORD_ID",
			"message": "Missing alert record ID",
		})
		return
	}

	notes, err := db.GetAlertNotes(c.Request.Context(), h.pool, recordID)
	if err != nil {
		if errors.Is(err, db.ErrAlertRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "ERR_RECORD_NOT_FOUND",
				"message": "Alert record not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_GET_NOTES",
			"message": "Failed to retrieve alert notes",
			"details": gin.H{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      notes,
		"message":   "Alert notes retrieved successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (h *AlertRecordHandler) broadcastRecordStatus(record *models.AlertRecord) {
	if h.realtimeHub == nil || record == nil {
		return
	}

	eventType := realtime.EventAlertUpdated
	if record.Status == "resolved" {
		eventType = realtime.EventAlertResolved
	}
	h.realtimeHub.BroadcastAlertRecord(eventType, record)
}
