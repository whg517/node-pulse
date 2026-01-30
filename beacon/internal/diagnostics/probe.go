package diagnostics

import (
	"fmt"
	"time"

	"beacon/internal/config"
)

// ProbeTaskInfo contains information about a single probe task
type ProbeTaskInfo struct {
	Type           string    `json:"type"`
	Target         string    `json:"target"`
	Status         string    `json:"status"`         // running, stopped, error, unknown
	LastExecution  *time.Time `json:"last_execution,omitempty"`
	NextExecution  *time.Time `json:"next_execution,omitempty"`
	LatencyMs      float64   `json:"latency_ms,omitempty"`
	PacketLossRate float64   `json:"packet_loss_rate,omitempty"`
}

// ProbeTasks contains probe task status information
type ProbeTasks struct {
	TotalTasks    int             `json:"total_tasks"`
	RunningTasks  int             `json:"running_tasks"`
	TotalExecs    int             `json:"total_executions"`    // Total probe executions
	SuccessExecs  int             `json:"success_executions"`  // Successful executions
	FailureExecs  int             `json:"failure_executions"`  // Failed executions
	Tasks         []ProbeTaskInfo `json:"tasks"`
}

// collectProbeTasks collects probe task status information
// NOTE: This requires integration with probe.Manager to get actual task status and execution statistics.
// For Story 3.10, this returns configuration-based information with unknown status.
func (c *collector) collectProbeTasks() (*ProbeTasks, error) {
	tasks := &ProbeTasks{
		TotalTasks:    len(c.cfg.Probes),
		RunningTasks:  0, // Requires probe.Manager integration
		TotalExecs:    0, // Requires probe.Manager integration
		SuccessExecs:  0, // Requires probe.Manager integration
		FailureExecs:  0, // Requires probe.Manager integration
		Tasks:         make([]ProbeTaskInfo, 0, len(c.cfg.Probes)),
	}

	for _, probe := range c.cfg.Probes {
		taskInfo := ProbeTaskInfo{
			Type:   probe.Type,
			Target: formatProbeTarget(probe),
			Status: "unknown", // Honest: requires probe manager to determine actual status
		}
		tasks.Tasks = append(tasks.Tasks, taskInfo)
	}

	return tasks, nil
}

// formatProbeTarget formats the probe target for display
func formatProbeTarget(probe config.ProbeConfig) string {
	if probe.Port > 0 {
		return fmt.Sprintf("%s:%d", probe.Target, probe.Port)
	}
	return probe.Target
}
