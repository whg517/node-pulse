package beacon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"beacon/internal/config"
	"beacon/internal/logger"
	"beacon/internal/metrics"
	"beacon/internal/monitor"
	"beacon/internal/process"
	"beacon/internal/probe"
	"beacon/internal/reporter"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Beacon agent",
	Long:  `Start the Beacon agent to perform network probes and report metrics to Pulse server.`,
	RunE:  runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Loading configuration...")

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Print node info immediately for user visibility
	fmt.Fprintf(cmd.OutOrStdout(), "Node ID: %s\n", cfg.NodeID)
	fmt.Fprintf(cmd.OutOrStdout(), "Node Name: %s\n", cfg.NodeName)

	// Initialize logger (Story 3.9)
	if err := logger.InitLogger(cfg); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Close()

	logger.WithFields(map[string]interface{}{
		"node_id":   cfg.NodeID,
		"node_name": cfg.NodeName,
		"config":    cfg.ConfigPath,
	}).Info("Configuration loaded successfully")

	// Create process manager
	procMgr := process.NewManager(cfg)

	// Write PID file
	if err := procMgr.WritePID(); err != nil {
		logger.WithError(err).Warn("Failed to write PID file")
	}
	defer procMgr.Cleanup()

	logger.Info("Starting probes...")

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

	logger.Info("Starting resource monitor...")

	// Create and start resource monitor (Story 3.11)
	var resourceMonitor monitor.Monitor
	if cfg.ResourceMonitor.Enabled {
		logAdapter := &monitor.LogrusLogger{}
		monitorCfg := &monitor.ResourceMonitorConfig{
			Enabled:              cfg.ResourceMonitor.Enabled,
			CheckIntervalSeconds: cfg.ResourceMonitor.CheckIntervalSeconds,
			Thresholds: monitor.ThresholdsConfig{
				CPUMicrocores: cfg.ResourceMonitor.Thresholds.CPUMicrocores,
				MemoryMB:      cfg.ResourceMonitor.Thresholds.MemoryMB,
			},
			Degradation: monitor.DegradationConfig{
				DegradedLevel: monitor.DegradationLevelConfig{
					CPUMicrocores:      cfg.ResourceMonitor.Degradation.DegradedLevel.CPUMicrocores,
					MemoryMB:           cfg.ResourceMonitor.Degradation.DegradedLevel.MemoryMB,
					IntervalMultiplier: cfg.ResourceMonitor.Degradation.DegradedLevel.IntervalMultiplier,
				},
				CriticalLevel: monitor.DegradationLevelConfig{
					CPUMicrocores:      cfg.ResourceMonitor.Degradation.CriticalLevel.CPUMicrocores,
					MemoryMB:           cfg.ResourceMonitor.Degradation.CriticalLevel.MemoryMB,
					IntervalMultiplier: cfg.ResourceMonitor.Degradation.CriticalLevel.IntervalMultiplier,
				},
				Recovery: monitor.RecoveryConfig{
					ConsecutiveNormalChecks: cfg.ResourceMonitor.Degradation.Recovery.ConsecutiveNormalChecks,
				},
			},
			Alerting: monitor.AlertingConfig{
				SuppressionWindowSeconds: cfg.ResourceMonitor.Alerting.SuppressionWindowSeconds,
			},
		}
		resourceMonitor, err = monitor.NewMonitor(monitorCfg, scheduler, logAdapter)
		if err != nil {
			logger.WithError(err).Warn("Failed to create resource monitor")
		} else {
			if err := resourceMonitor.Start(); err != nil {
				logger.WithError(err).Warn("Failed to start resource monitor")
			} else {
				defer resourceMonitor.Stop()
			}
		}
	}

	logger.Info("Starting metrics server...")

	// Create and start metrics server (Story 3.8)
	metricsServer := metrics.NewMetrics(cfg, scheduler)
	if err := metricsServer.Start(); err != nil {
		logger.WithError(err).Warn("Failed to start metrics server")
	}
	defer metricsServer.Stop()

	logger.Info("Starting heartbeat reporter...")

	// Create Pulse API client with 5 second timeout (NFR-PERF-001)
	apiClient := reporter.NewPulseAPIClient(cfg.PulseServer, 5*time.Second)

	// Create heartbeat reporter with scheduler integration
	heartbeatReporter := reporter.NewHeartbeatReporter(apiClient, cfg.NodeID, scheduler)

	// Start heartbeat reporting
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	heartbeatReporter.StartReporting(ctx)
	defer heartbeatReporter.StopReporting()

	logger.WithFields(map[string]interface{}{
		"node_id":   cfg.NodeID,
		"node_name": cfg.NodeName,
	}).Info("Beacon started successfully")
	logger.Info("Press Ctrl+C to stop...")

	// Wait for interrupt signal or context cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		// Signal received
	case <-ctx.Done():
		// Context cancelled (e.g., timeout in tests)
	}

	logger.Info("Shutting down gracefully...")

	return nil
}
