package beacon

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"beacon/internal/config"
	"beacon/internal/process"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Beacon agent status",
	Long:  `Display the current status of the Beacon agent.`,
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load configuration (optional for status)
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		// If config can't be loaded, show offline status
		status := map[string]interface{}{
			"status":         "offline",
			"last_heartbeat": nil,
			"config_version": "unknown",
			"error":          err.Error(),
		}
		jsonData, _ := json.MarshalIndent(status, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
		return nil
	}

	// Create process manager
	procMgr := process.NewManager(cfg)

	// Get process status
	procStatus := procMgr.GetStatus()

	// Build status response
	status := map[string]interface{}{
		"status":         "unknown",
		"node_id":        cfg.NodeID,
		"node_name":      cfg.NodeName,
		"last_heartbeat": time.Now().Format(time.RFC3339),
		"config_version": "1.0",
	}

	if procStatus.Running {
		status["status"] = "running"
		status["pid"] = procStatus.PID
	} else {
		status["status"] = "stopped"
		if procStatus.Error != "" {
			status["error"] = procStatus.Error
		}
	}

	// Output JSON
	jsonData, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(jsonData))
	return nil
}
