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
	Status      string            `json:"status"`
	Checks      map[string]string `json:"checks"`
	Scheduler   *SchedulerStatus  `json:"scheduler,omitempty"`
	Time        string            `json:"timestamp"`
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
	db        Checker
	scheduler scheduler.Scheduler
}

// New creates a new health checker
func New(db Checker, sched scheduler.Scheduler) *HealthChecker {
	return &HealthChecker{
		db:        db,
		scheduler: sched,
	}
}

// Handler returns a Gin handler for health check
func (h *HealthChecker) Handler(c *gin.Context) {
	ctx := c.Request.Context()
	isHealthy := true
	checks := make(map[string]string)
	var schedulerStatus *SchedulerStatus

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

	status := "healthy"
	if !isHealthy {
		status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:    status,
			Checks:    checks,
			Scheduler: schedulerStatus,
			Time:      time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:    status,
		Checks:    checks,
		Scheduler: schedulerStatus,
		Time:      time.Now().UTC().Format(time.RFC3339),
	})
}
