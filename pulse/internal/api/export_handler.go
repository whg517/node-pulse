package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/whg517/node-pulse/pulse/internal/export"
	"github.com/whg517/node-pulse/pulse/internal/models"
)

// ExportHandler handles data export API requests
type ExportHandler struct {
	exportService *export.ExportService
}

// NewExportHandler creates a new ExportHandler
func NewExportHandler(exportService *export.ExportService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
	}
}

// CreateExportRequest represents the request to create an export
type CreateExportRequest struct {
	NodeIDs   []string `form:"node_ids" binding:"required,min=1,max=50"`
	StartTime string   `form:"start_time" binding:"required"`
	EndTime   string   `form:"end_time" binding:"required"`
	Metrics   []string `form:"metrics" binding:"required,min=1"`
	Format    string   `form:"format"`
}

// CreateExportResponse represents the response when creating an export
type CreateExportResponse struct {
	Data      models.ExportTask `json:"data"`
	Message   string            `json:"message"`
	Timestamp string            `json:"timestamp"`
}

// CreateExportHandler handles POST /api/v1/data/export
// Creates a new export task and returns immediately
// @Summary		Create data export task
// @Description	Creates an asynchronous export task for metrics data. Admin role required.
// @Tags			Export
// @Accept			json
// @Produce		json
// @Param			node_ids	query		string					true	"Node UUIDs to export (1-50, repeatable)"
// @Param			start_time	query		string					true	"Start time in ISO 8601 format"
// @Param			end_time	query		string					true	"End time in ISO 8601 format"
// @Param			metrics		query		string					true	"Metrics to export: latency, packet_loss_rate, jitter (repeatable)"
// @Param			format		query		string					false	"Export format (only csv supported)"	default(csv)
// @Success		202	{object}	CreateExportResponse	"Export task created"
// @Failure		400	{object}	map[string]interface{}	"Invalid request parameters"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		403	{object}	map[string]interface{}	"Forbidden (requires admin role)"
// @Security		BearerAuth
// @Router			/data/export [post]
func (h *ExportHandler) CreateExportHandler(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"details": "User not authenticated",
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"details": "Invalid user ID format",
		})
		return
	}

	// Parse request parameters
	var req CreateExportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// Parse timestamps
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid start_time format",
			"details": "Must be ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid end_time format",
			"details": "Must be ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
		})
		return
	}

	// Validate metrics
	validMetrics := map[string]bool{
		"latency":          true,
		"packet_loss_rate": true,
		"jitter":           true,
	}
	for _, metric := range req.Metrics {
		if !validMetrics[metric] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid metric",
				"details": fmt.Sprintf("Metric '%s' is not valid. Valid metrics: latency, packet_loss_rate, jitter", metric),
			})
			return
		}
	}

	// Set default format
	if req.Format == "" {
		req.Format = "csv"
	}

	// Validate format
	if req.Format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid format",
			"details": fmt.Sprintf("Format '%s' is not supported. Only CSV is supported in MVP", req.Format),
		})
		return
	}

	// Create export request
	exportReq := &export.CreateExportRequest{
		UserID:    userIDStr,
		NodeIDs:   req.NodeIDs,
		StartTime: startTime,
		EndTime:   endTime,
		Metrics:   req.Metrics,
		Format:    req.Format,
	}

	// Create export task
	ctx := context.Background()
	task, err := h.exportService.CreateExport(ctx, exportReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to create export task",
			"details": err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusAccepted, CreateExportResponse{
		Data:      *task,
		Message:   "Export task created successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetExportStatusResponse represents the response for export status
type GetExportStatusResponse struct {
	Data      models.ExportTask `json:"data"`
	Message   string            `json:"message"`
	Timestamp string            `json:"timestamp"`
}

// ListExportsResponse represents the response for listing export tasks.
type ListExportsResponse struct {
	Data      []models.ExportTask `json:"data"`
	Message   string              `json:"message"`
	Timestamp string              `json:"timestamp"`
}

// ListExportsHandler handles GET /api/v1/data/export.
// @Summary		List export tasks
// @Description	Returns recent export tasks for the current admin user.
// @Tags			Export
// @Accept			json
// @Produce		json
// @Param			limit	query	int	false	"Maximum number of tasks to return"	default(50)
// @Success		200	{object}	ListExportsResponse	"Export task list"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		403	{object}	map[string]interface{}	"Forbidden (requires admin role)"
// @Security		BearerAuth
// @Router			/data/export [get]
func (h *ExportHandler) ListExportsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"details": "User not authenticated",
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"details": "Invalid user ID format",
		})
		return
	}

	limit := 50
	if rawLimit := c.Query("limit"); rawLimit != "" {
		if _, err := fmt.Sscanf(rawLimit, "%d", &limit); err != nil || limit < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid limit",
				"details": "limit must be a positive integer",
			})
			return
		}
	}

	tasks := h.exportService.ListExports(userIDStr, limit)
	responseTasks := make([]models.ExportTask, 0, len(tasks))
	for _, task := range tasks {
		responseTasks = append(responseTasks, *task)
	}

	c.JSON(http.StatusOK, ListExportsResponse{
		Data:      responseTasks,
		Message:   "Export tasks retrieved",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// GetExportStatusHandler handles GET /api/v1/data/export/:id
// Returns the status of an export task
// @Summary		Get export task status
// @Description	Returns the current status of an export task. Admin role required.
// @Tags			Export
// @Accept			json
// @Produce		json
// @Param			id	path		string					true	"Export task ID"
// @Success		200	{object}	GetExportStatusResponse	"Export task status"
// @Failure		400	{object}	map[string]interface{}	"Missing export ID"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		403	{object}	map[string]interface{}	"Forbidden (requires admin role)"
// @Failure		404	{object}	map[string]interface{}	"Export task not found"
// @Security		BearerAuth
// @Router			/data/export/{id} [get]
func (h *ExportHandler) GetExportStatusHandler(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing export ID",
			"details": "export_id is required",
		})
		return
	}

	// Get export task
	task, err := h.exportService.GetExport(exportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Export not found",
			"details": err.Error(),
		})
		return
	}

	// Return response
	message := "Export status retrieved"
	if task.IsCompleted() {
		message = "Export completed successfully"
	} else if task.IsFailed() {
		message = "Export failed"
	} else if task.IsProcessing() {
		message = "Export is being processed"
	}

	c.JSON(http.StatusOK, GetExportStatusResponse{
		Data:      *task,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// DownloadExportHandler handles GET /api/v1/data/export/:id/download
// Downloads the export file
// @Summary		Download export file
// @Description	Downloads the completed export file as CSV. Admin role required.
// @Tags			Export
// @Produce		text/csv
// @Param			id	path		string	true	"Export task ID"
// @Success		200	{file}		binary	"CSV file download"
// @Failure		400	{object}	map[string]interface{}	"Export not ready"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		403	{object}	map[string]interface{}	"Forbidden (requires admin role)"
// @Failure		404	{object}	map[string]interface{}	"Export task or file not found"
// @Security		BearerAuth
// @Router			/data/export/{id}/download [get]
func (h *ExportHandler) DownloadExportHandler(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing export ID",
			"details": "export_id is required",
		})
		return
	}

	// Get export task
	task, err := h.exportService.GetExport(exportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Export not found",
			"details": err.Error(),
		})
		return
	}

	// Check if export is completed
	if !task.IsCompleted() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Export not ready",
			"details": fmt.Sprintf("Export status is %s. Please wait for completion", task.Status),
		})
		return
	}

	// Check if file exists
	if task.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Export file not found",
			"details": "No file associated with this export",
		})
		return
	}

	// Set headers for file download
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=metrics_export_%s.csv", exportID))
	c.Header("Content-Length", fmt.Sprintf("%d", task.FileSize))

	// Send file
	c.File(task.FilePath)
}

// DeleteExportHandler deletes an export task and its generated file.
//
// @Summary		Delete export task
// @Description	Removes an export task record and the associated file. Admin only.
// @Tags			export
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"Export task ID"
// @Success		204	{object}	nil		"Deleted"
// @Failure		400	{object}	map[string]interface{}	"Missing export ID"
// @Failure		404	{object}	map[string]interface{}	"Export not found"
// @Failure		500	{object}	map[string]interface{}	"Internal error"
// @Security		BearerAuth
// @Router			/data/export/{id} [delete]
func (h *ExportHandler) DeleteExportHandler(c *gin.Context) {
	exportID := c.Param("id")
	if exportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Missing export ID",
			"details": "export_id is required",
		})
		return
	}

	if err := h.exportService.DeleteExport(exportID); err != nil {
		// DeleteExport only returns an error from the durable store; a missing
		// in-memory task is not an error (the row may still exist).
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete export",
			"details": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
