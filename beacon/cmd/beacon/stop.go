package beacon

import (
	"fmt"

	"github.com/spf13/cobra"

	"beacon/internal/config"
	"beacon/internal/process"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Beacon agent",
	Long:  `Stop the running Beacon agent gracefully.`,
	RunE:  runStop,
}

func runStop(cmd *cobra.Command, args []string) error {
	fmt.Println("[INFO] Stopping Beacon...")

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create process manager
	procMgr := process.NewManager(cfg)

	// Stop the process
	if err := procMgr.Stop(); err != nil {
		return fmt.Errorf("failed to stop beacon: %w", err)
	}

	fmt.Println("[INFO] Beacon stopped successfully")
	return nil
}
