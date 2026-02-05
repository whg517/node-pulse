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

// GetExportStatusHandler handles GET /api/v1/data/export/:id
// Returns the status of an export task
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
