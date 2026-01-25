package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Checker defines interface for health check components
type Checker interface {
	Check(ctx context.Context) error
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks"`
	Time    string            `json:"timestamp"`
}

// HealthChecker manages health checks
type HealthChecker struct {
	db Checker
}

// New creates a new health checker
func New(db Checker) *HealthChecker {
	return &HealthChecker{
		db: db,
	}
}

// Handler returns a Gin handler for health check
func (h *HealthChecker) Handler(c *gin.Context) {
	ctx := c.Request.Context()
	isHealthy := true
	checks := make(map[string]string)

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

	status := "healthy"
	if !isHealthy {
		status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:  status,
			Checks:  checks,
			Time:    time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:  status,
		Checks:  checks,
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}
