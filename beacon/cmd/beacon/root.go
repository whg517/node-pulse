package beacon

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
	debugFlag  bool
	rootCmd    *cobra.Command
)

// Initialize the root command
func init() {
	rootCmd = &cobra.Command{
		Use:   "beacon",
		Short: "Beacon - Network monitoring agent",
		Long:  `Beacon is a lightweight network monitoring agent that performs TCP/UDP probes and reports metrics to Pulse server.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "beacon.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "enable debug mode")

	// Add subcommands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(debugCmd)
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// GetRootCmd returns the root command (for testing)
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// GetConfigFile returns the config file path
func GetConfigFile() string {
	return configFile
}

// IsDebug returns whether debug mode is enabled
func IsDebug() bool {
	return debugFlag
}
