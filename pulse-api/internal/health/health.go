package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kevin/node-pulse/pulse-api/internal/scheduler"
)

// Checker defines interface for health check components
type Checker interface {
	Check(ctx context.Context) error
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status       string             `json:"status"`
	Checks       map[string]string  `json:"checks"`
	Scheduler    *SchedulerStatus   `json:"scheduler,omitempty"`
	AlertSystem  *AlertSystemStatus `json:"alert_system,omitempty"`
	Time         string             `json:"timestamp"`
}

// SchedulerStatus represents scheduler health status
type SchedulerStatus struct {
	Running bool                       `json:"running"`
	Tasks   map[string]TaskStatusInfo  `json:"tasks,omitempty"`
}

// TaskStatusInfo represents task status for health check
type TaskStatusInfo struct {
	IsRunning    bool   `json:"is_running"`
	LastRun      string `json:"last_run,omitempty"`
	RunCount     int64  `json:"run_count"`
	LastError    string `json:"last_error,omitempty"`
}

// HealthChecker manages health checks
type HealthChecker struct {
	db                  Checker
	scheduler           scheduler.Scheduler
	alertSystemChecker  *AlertSystemChecker
}

// New creates a new health checker
func New(db Checker, sched scheduler.Scheduler, alertSystemChecker *AlertSystemChecker) *HealthChecker {
	return &HealthChecker{
		db:                 db,
		scheduler:          sched,
		alertSystemChecker: alertSystemChecker,
	}
}

// Handler returns a Gin handler for health check
// @Summary		Health check
// @Description	Returns the health status of the API service, including database, scheduler, and alert system components.
// @Description
// @Description	**Status values:**
// @Description	- `healthy`: All components are operational
// @Description	- `degraded`: Some components are non-critically failing
// @Description	- `unhealthy`: Critical components are failing
// @Tags			health
// @Accept			json
// @Produce		json
// @Success		200	{object}	HealthResponse	"Health status"
// @Router			/health [get]
func (h *HealthChecker) Handler(c *gin.Context) {
	ctx := c.Request.Context()
	isHealthy := true
	isDegraded := false
	checks := make(map[string]string)
	var schedulerStatus *SchedulerStatus
	var alertSystemStatus *AlertSystemStatus

	// Check database - nil database is not an error, it's disabled
	if h.db == nil {
		checks["database"] = "disabled"
	} else {
		// Check database connection
		if err := h.db.Check(ctx); err != nil {
			isHealthy = false
			checks["database"] = "error: " + err.Error()
		} else {
			checks["database"] = "ok"
		}
	}

	// Check scheduler status
	if h.scheduler != nil {
		// Try to get cleanup task status
		if taskStatus, err := h.scheduler.GetTaskStatus("metrics-cleanup"); err == nil {
			schedulerStatus = &SchedulerStatus{
				Running: true,
				Tasks: map[string]TaskStatusInfo{
					"metrics-cleanup": {
						IsRunning: taskStatus.IsRunning,
						LastRun:   taskStatus.LastRun.Format(time.RFC3339),
						RunCount:  taskStatus.RunCount,
						LastError: taskStatus.LastError,
					},
				},
			}
		} else {
			// Scheduler exists but task not found or not registered
			schedulerStatus = &SchedulerStatus{
				Running: true,
				Tasks:   map[string]TaskStatusInfo{},
			}
		}
	}

	// Check alert system health
	if h.alertSystemChecker != nil {
		alertSystemStatus = &AlertSystemStatus{}

		// Check alert engine with timeout
		checkCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		if engineStatus, err := h.alertSystemChecker.CheckAlertEngine(checkCtx); err == nil {
			alertSystemStatus.AlertEngine = engineStatus
			checks["alert_engine"] = engineStatus.Status

			// Determine if degraded or unhealthy
			if engineStatus.Status == "full" {
				isHealthy = false
			} else if engineStatus.Status == "stale" {
				isDegraded = true
			}
		} else {
			checks["alert_engine"] = "error: " + err.Error()
		}
		cancel()

		// Check webhook delivery with timeout
		checkCtx, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
		if deliveryStatus, err := h.alertSystemChecker.CheckWebhookDelivery(checkCtx); err == nil {
			alertSystemStatus.WebhookDelivery = deliveryStatus
			checks["webhook_delivery"] = deliveryStatus.Status

			// Determine if degraded or unhealthy
			if deliveryStatus.Status == "unhealthy" {
				isDegraded = true
			} else if deliveryStatus.Status == "degraded" {
				isDegraded = true
			}
		} else {
			checks["webhook_delivery"] = "error: " + err.Error()
		}
		cancel()

		// Check alert suppression with timeout
		checkCtx, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
		if suppressionStatus, err := h.alertSystemChecker.CheckAlertSuppression(checkCtx); err == nil {
			alertSystemStatus.AlertSuppression = suppressionStatus
			checks["alert_suppression"] = suppressionStatus.Status
		} else {
			checks["alert_suppression"] = "error: " + err.Error()
		}
		cancel()
	}

	// Determine overall status
	status := "healthy"
	if !isHealthy {
		status = "unhealthy"
	} else if isDegraded {
		status = "degraded"
	}

	// Return appropriate HTTP status code
	httpStatus := http.StatusOK
	if status == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, HealthResponse{
		Status:      status,
		Checks:      checks,
		Scheduler:   schedulerStatus,
		AlertSystem: alertSystemStatus,
		Time:        time.Now().UTC().Format(time.RFC3339),
	})
}
