package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/whg517/node-pulse/beacon/internal/config"
	"github.com/whg517/node-pulse/beacon/internal/process"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Beacon agent",
	Long:  `Stop the running Beacon agent gracefully.`,
	RunE:  runStop,
}

func runStop(cmd *cobra.Command, args []string) error {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[INFO] Stopping Beacon...")

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create process manager
	procMgr := process.NewManager(cfg)

	// Stop the process
	if err := procMgr.Stop(); err != nil {
		// Check if it's because no PID file was found (not running)
		if strings.Contains(err.Error(), "no valid PID file found") {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[INFO] Beacon is not running")
			return nil
		}
		return fmt.Errorf("failed to stop beacon: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "[INFO] Beacon stopped successfully")
	return nil
}
