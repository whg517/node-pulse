package beacon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"beacon/internal/config"
	"beacon/internal/metrics"
	"beacon/internal/probe"
	"beacon/internal/process"
	"beacon/internal/reporter"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Beacon agent",
	Long:  `Start the Beacon agent to perform network probes and report metrics to Pulse server.`,
	RunE:  runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	fmt.Println("[INFO] Loading configuration...")
	
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create process manager
	procMgr := process.NewManager(cfg)

	// Write PID file
	if err := procMgr.WritePID(); err != nil {
		log.Printf("[WARN] Failed to write PID file: %v", err)
	}
	defer procMgr.Cleanup()

	fmt.Println("[INFO] Starting probes...")

	// Create probe scheduler
	scheduler, err := probe.NewProbeScheduler(cfg.Probes)
	if err != nil {
		return fmt.Errorf("failed to create probe scheduler: %w", err)
	}

	// Start probe scheduler
	if err := scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start probe scheduler: %w", err)
	}
	defer scheduler.Stop()

	fmt.Println("[INFO] Starting metrics server...")

	// Create and start metrics server (Story 3.8)
	metricsServer := metrics.NewMetrics(cfg, scheduler)
	if err := metricsServer.Start(); err != nil {
		log.Printf("[WARN] Failed to start metrics server: %v", err)
	}
	defer metricsServer.Stop()

	fmt.Println("[INFO] Starting heartbeat reporter...")

	// Create Pulse API client with 5 second timeout (NFR-PERF-001)
	apiClient := reporter.NewPulseAPIClient(cfg.PulseServer, 5*time.Second)

	// Create heartbeat reporter with scheduler integration
	heartbeatReporter := reporter.NewHeartbeatReporter(apiClient, cfg.NodeID, scheduler)

	// Start heartbeat reporting
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	heartbeatReporter.StartReporting(ctx)
	defer heartbeatReporter.StopReporting()

	fmt.Println("[INFO] Beacon started successfully")
	fmt.Println("[INFO] Press Ctrl+C to stop...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("[INFO] Shutting down gracefully...")

	return nil
}
