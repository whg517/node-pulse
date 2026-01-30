package beacon

import (
	"fmt"

	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Run Beacon in debug mode",
	Long:  `Run Beacon in debug mode with verbose logging.`,
	RunE:  runDebug,
}

func runDebug(cmd *cobra.Command, args []string) error {
	fmt.Println("[DEBUG] Debug mode enabled")
	fmt.Println("[DEBUG] (1/4) Loading configuration...")
	fmt.Println("[DEBUG] (2/4) Initializing probes...")
	fmt.Println("[DEBUG] (3/4) Starting heartbeat reporter...")
	fmt.Println("[DEBUG] (4/4) Debug mode active")
	return nil
}
