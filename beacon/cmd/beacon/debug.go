package beacon

import (
	"fmt"

	"github.com/spf13/cobra"

	"beacon/internal/config"
	"beacon/internal/diagnostics"
)

var (
	debugCmd = &cobra.Command{
		Use:   "debug",
		Short: "Show detailed diagnostic information",
		Long: `Display comprehensive diagnostic information for troubleshooting.

This command outputs detailed information about:
- Network status and connectivity
- Configuration details
- Connection retry status
- Resource usage
- Probe task status
- Prometheus metrics summary

Output is in JSON format by default. Use --pretty for human-readable output.`,
		RunE: runDebug,
	}
)

func runDebug(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Create diagnostic collector
	collector := diagnostics.NewCollector(cfg)

	// Get pretty flag from command
	prettyPrint, _ := cmd.Flags().GetBool("pretty")

	// Collect and output diagnostic information
	if prettyPrint {
		output, err := collector.CollectPretty()
		if err != nil {
			return fmt.Errorf("error collecting diagnostics: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), output)
	} else {
		data, err := collector.CollectJSON()
		if err != nil {
			return fmt.Errorf("error collecting diagnostics: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	}

	return nil
}

func init() {
	debugCmd.Flags().BoolP("pretty", "p", false, "Pretty print output in human-readable format")
}
