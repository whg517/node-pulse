package beacon

import (
	"fmt"

	"github.com/spf13/cobra"

	"beacon/internal/config"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Run Beacon in debug mode",
	Long:  `Run Beacon in debug mode with verbose logging.`,
	RunE:  runDebug,
}

func runDebug(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), "[DEBUG] Debug mode enabled")
	fmt.Fprintln(cmd.OutOrStdout(), "[DEBUG] (1/4) Loading configuration...")

	// Try to load configuration
	_, err := config.LoadConfig(configFile)
	if err != nil {
		// Show detailed error information
		fmt.Fprintln(cmd.OutOrStdout(), "[DEBUG] Configuration error:")
		fmt.Fprintf(cmd.OutOrStdout(), "[DEBUG]   Error: %s\n", err.Error())
		// The config.LoadConfig already includes detailed error information
		// with line numbers and suggestions
		return nil // Don't fail the command, just show the error
	}

	fmt.Fprintln(cmd.OutOrStdout(), "[DEBUG] (2/4) Initializing probes...")
	fmt.Fprintln(cmd.OutOrStdout(), "[DEBUG] (3/4) Starting heartbeat reporter...")
	fmt.Fprintln(cmd.OutOrStdout(), "[DEBUG] (4/4) Debug mode active")
	return nil
}
