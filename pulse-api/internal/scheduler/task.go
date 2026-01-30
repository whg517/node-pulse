package scheduler

import (
	"context"
	"time"
)

// Task defines the interface for scheduled tasks
type Task interface {
	// Name returns the task name
	Name() string

	// Execute runs the task with the given context
	Execute(ctx context.Context) error

	// Interval returns the execution interval
	Interval() time.Duration
}

// Scheduler defines the interface for task scheduler
type Scheduler interface {
	// Start starts the scheduler
	Start(ctx context.Context) error

	// Stop stops the scheduler (waits for current task to complete)
	Stop() error

	// RegisterTask registers a task
	RegisterTask(task Task) error

	// GetTaskStatus gets task status
	GetTaskStatus(taskName string) (*TaskStatus, error)
}

// TaskStatus represents the current status of a task
type TaskStatus struct {
	Name         string        `json:"name"`
	IsRunning    bool          `json:"is_running"`
	LastRun      time.Time     `json:"last_run"`
	NextRun      time.Time     `json:"next_run"`
	LastDuration time.Duration `json:"last_duration"`
	LastError    string        `json:"last_error,omitempty"`
	RunCount     int64         `json:"run_count"`
}
