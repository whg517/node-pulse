package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	pulseapi "beacon/internal/api"
	"beacon/internal/auth"
	"beacon/internal/config"
	"beacon/internal/logger"
	"beacon/internal/metrics"
	"beacon/internal/models"
	"beacon/internal/monitor"
	"beacon/internal/probe"
	"beacon/internal/process"
	"beacon/internal/reporter"
	"beacon/internal/telemetry"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Beacon agent",
	Long:  `Start the Beacon agent to perform network probes and report metrics to Pulse server.`,
	RunE:  runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Loading configuration...")

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Print node info immediately for user visibility
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Node ID: %s\n", cfg.NodeID)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Node Name: %s\n", cfg.NodeName)

	// Initialize logger (Story 3.9)
	if err := logger.InitLogger(cfg); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() { _ = logger.Close() }()

	logger.WithFields(map[string]interface{}{
		"node_id":   cfg.NodeID,
		"node_name": cfg.NodeName,
		"config":    cfg.ConfigPath,
	}).Info("Configuration loaded successfully")

	// Create context for canceling goroutines (must be before telemetry init)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize distributed tracing (opt-in via config).
	telemetryCfg := telemetry.Config{
		Enabled:        cfg.Telemetry.Enabled,
		ServiceName:    cfg.Telemetry.ServiceName,
		ServiceVersion: cfg.Telemetry.ServiceVersion,
		NodeID:         cfg.NodeID,
		OTLPEndpoint:   cfg.Telemetry.OTLPEndpoint,
		SamplingRate:   cfg.Telemetry.SamplingRate,
	}
	telemetryProvider, err := telemetry.Init(ctx, telemetryCfg)
	if err != nil {
		logger.WithError(err).Warn("Telemetry initialization failed – continuing without tracing")
	} else {
		defer telemetryProvider.Shutdown(context.Background())
	}

	// Create process manager
	procMgr := process.NewManager(cfg)

	// Write PID file
	if err := procMgr.WritePID(); err != nil {
		logger.WithError(err).Warn("Failed to write PID file")
	}
	defer func() { _ = procMgr.Cleanup() }()

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

	// Create config watcher for hot reload (Story 3.13)
	configWatcher, err := config.NewFileWatcher(cfg.ConfigPath, cfg, logger.GetLogger())
	if err != nil {
		logger.WithError(err).Warn("Failed to create config watcher, hot reload disabled")
	} else {
		// Register callback to reload probe config
		configWatcher.OnReload(func(newConfig *config.Config, changes []string) error {
			logger.WithField("changes", changes).Info("Reloading probe configuration...")
			if err := scheduler.ReloadConfig(newConfig.Probes); err != nil {
				return fmt.Errorf("failed to reload probe config: %w", err)
			}
			logger.Info("Probe configuration reloaded successfully")
			return nil
		})

		// Start config watcher in goroutine
		go func() {
			if err := configWatcher.Start(ctx); err != nil {
				logger.WithError(err).Error("Config watcher stopped with error")
			}
		}()

		logger.WithField("config_path", cfg.ConfigPath).Info("Config watcher started for hot reload")
	}

	logger.Info("Starting resource monitor...")

	// Create and start resource monitor (Story 3.11)
	var resourceMonitor monitor.Monitor
	if cfg.ResourceMonitor.Enabled {
		logAdapter := monitor.NewSlogLogger(logger.GetLogger())
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
	defer func() { _ = metricsServer.Stop() }()

	logger.Info("Starting heartbeat reporter...")

	if cfg.Mode.Mode == config.ModeStandalone {
		logger.Info("Beacon running in standalone mode; skipping Pulse authentication and heartbeat reporting")
		waitForStop(ctx, configWatcher)
		logger.Info("Shutting down gracefully...")
		return nil
	}

	// Validate API key configuration
	if cfg.APIKey == "" {
		return fmt.Errorf("required field 'api_key' is missing (JWT authentication required)")
	}

	// Create JWT client for authentication
	jwtClient, err := auth.NewJWTClient(cfg.PulseServer, cfg.APIKey, nil)
	if err != nil {
		return fmt.Errorf("failed to create JWT client: %w", err)
	}

	// Fetch initial token to validate configuration
	logger.Info("Authenticating with Pulse server...")
	if _, err := jwtClient.GetAccessToken(ctx); err != nil {
		return fmt.Errorf("failed to authenticate with Pulse server: %w", err)
	}
	logger.Info("Authentication successful")

	stopConfigSync := startServerConfigSync(ctx, cfg, jwtClient, scheduler)
	defer stopConfigSync()
	stopMTRReporting := startMTRResultReporting(ctx, cfg, jwtClient, scheduler)
	defer stopMTRReporting()

	// Create Pulse API client with 5 second timeout (NFR-PERF-001)
	apiClient := reporter.NewPulseAPIClient(cfg.PulseServer, 5*time.Second, jwtClient)

	// Create heartbeat reporter with scheduler integration
	heartbeatReporter := reporter.NewHeartbeatReporter(apiClient, scheduler)

	// Start heartbeat reporting (using existing context)
	heartbeatReporter.StartReporting(ctx)
	defer heartbeatReporter.StopReporting()

	logger.WithFields(map[string]interface{}{
		"node_id":   cfg.NodeID,
		"node_name": cfg.NodeName,
	}).Info("Beacon started successfully")
	logger.Info("Press Ctrl+C to stop...")

	waitForStop(ctx, configWatcher)

	logger.Info("Shutting down gracefully...")

	return nil
}

type accessTokenProvider interface {
	GetAccessToken(ctx context.Context) (string, error)
}

func startServerConfigSync(ctx context.Context, cfg *config.Config, tokenProvider accessTokenProvider, scheduler *probe.ProbeScheduler) context.CancelFunc {
	syncCtx, cancel := context.WithCancel(ctx)
	interval := time.Duration(cfg.Mode.ConfigCheckIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}

	go func() {
		currentVersion := 0
		syncConfig := func() {
			token, err := tokenProvider.GetAccessToken(syncCtx)
			if err != nil {
				logger.WithError(err).Warn("Failed to get access token for beacon config sync")
				return
			}

			client := pulseapi.NewPulseClient(cfg.PulseServer, token, nil)
			resp, err := client.GetBeaconConfig(syncCtx, cfg.NodeID)
			if err != nil {
				logger.WithError(err).Warn("Failed to fetch server beacon config")
				return
			}
			if resp.Data.Version <= currentVersion {
				return
			}

			probes := serverConfigToLocalProbes(resp.Data)
			if err := scheduler.ReloadConfig(probes); err != nil {
				logger.WithError(err).Warn("Failed to apply server beacon config")
				ackErr := client.AcknowledgeBeaconConfig(syncCtx, &pulseapi.BeaconConfigAckRequest{
					NodeID:       cfg.NodeID,
					Version:      resp.Data.Version,
					Status:       "failed",
					ErrorMessage: err.Error(),
				})
				if ackErr != nil {
					logger.WithError(ackErr).Warn("Failed to acknowledge rejected server beacon config")
				}
				return
			}
			if err := client.AcknowledgeBeaconConfig(syncCtx, &pulseapi.BeaconConfigAckRequest{
				NodeID:  cfg.NodeID,
				Version: resp.Data.Version,
				Status:  "applied",
			}); err != nil {
				logger.WithError(err).Warn("Failed to acknowledge applied server beacon config")
			}
			currentVersion = resp.Data.Version
			logger.WithFields(map[string]interface{}{
				"version":     currentVersion,
				"probe_count": len(probes),
			}).Info("Applied server beacon config")
		}

		syncConfig()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				syncConfig()
			case <-syncCtx.Done():
				return
			}
		}
	}()

	return cancel
}

func serverConfigToLocalProbes(serverConfig pulseapi.BeaconConfigData) []config.ProbeConfig {
	probes := make([]config.ProbeConfig, 0, len(serverConfig.Probes))
	for _, serverProbe := range serverConfig.Probes {
		probeType := strings.ToUpper(serverProbe.Type)
		localType := strings.ToLower(serverProbe.Type)
		switch probeType {
		case "TCP":
			localType = "tcp_ping"
		case "UDP":
			localType = "udp_ping"
		case "MTR":
			localType = "mtr"
		}

		interval := serverProbe.IntervalSeconds
		if interval == 0 {
			interval = serverConfig.IntervalSeconds
		}
		if interval < 60 {
			interval = 60
		}

		timeout := serverProbe.TimeoutSeconds
		if timeout == 0 {
			timeout = serverConfig.TimeoutSeconds
		}
		if timeout == 0 {
			timeout = 5
		}

		count := serverProbe.Count
		if count < 10 {
			count = 10
		}

		probes = append(probes, config.ProbeConfig{
			Type:           localType,
			Target:         serverProbe.Target,
			Port:           serverProbe.Port,
			TimeoutSeconds: timeout,
			Interval:       interval,
			Count:          count,
			MaxHops:        serverProbe.MaxHops,
			PacketSize:     serverProbe.PacketSize,
		})
	}

	return probes
}

type mtrResultProvider interface {
	GetLatestMTRResults() []*models.MTRResult
}

func startMTRResultReporting(ctx context.Context, cfg *config.Config, tokenProvider accessTokenProvider, resultProvider mtrResultProvider) context.CancelFunc {
	reportCtx, cancel := context.WithCancel(ctx)
	interval := time.Duration(cfg.Mode.ConfigCheckIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = reporter.ReportInterval
	}

	go func() {
		sent := make(map[string]time.Time)
		upload := func() {
			results := resultProvider.GetLatestMTRResults()
			if len(results) == 0 {
				return
			}

			token, err := tokenProvider.GetAccessToken(reportCtx)
			if err != nil {
				logger.WithError(err).Warn("Failed to get access token for MTR result upload")
				return
			}
			client := pulseapi.NewPulseClient(cfg.PulseServer, token, nil)

			for _, result := range results {
				if result == nil {
					continue
				}
				key := result.Target
				if lastSent, ok := sent[key]; ok && !result.CompletedAt.After(lastSent) {
					continue
				}

				if err := client.SendMTRResult(reportCtx, mtrResultToRequest(cfg.NodeID, result)); err != nil {
					logger.WithError(err).Warn("Failed to upload MTR result")
					continue
				}
				sent[key] = result.CompletedAt
			}
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				upload()
			case <-reportCtx.Done():
				return
			}
		}
	}()

	return cancel
}

func mtrResultToRequest(nodeID string, result *models.MTRResult) *pulseapi.MTRResultRequest {
	hops := make([]pulseapi.MTRHop, 0, len(result.Hops))
	for _, hop := range result.Hops {
		hops = append(hops, pulseapi.MTRHop{
			HopNumber:  hop.HopNumber,
			IP:         hop.IP,
			Hostname:   hop.Hostname,
			ASNumber:   hop.ASNumber,
			Sent:       hop.Sent,
			Received:   hop.Received,
			LossRate:   hop.LossRate,
			LastRTTMs:  hop.LastRTTMs,
			AvgRTTMs:   hop.AvgRTTMs,
			BestRTTMs:  hop.BestRTTMs,
			WorstRTTMs: hop.WorstRTTMs,
			StdDevMs:   hop.StdDevMs,
			Location:   hop.Location,
		})
	}

	return &pulseapi.MTRResultRequest{
		NodeID:       nodeID,
		Target:       result.Target,
		TotalHops:    result.TotalHops,
		Hops:         hops,
		CompletedAt:  result.CompletedAt.Format(time.RFC3339),
		Success:      result.Success,
		ErrorMessage: result.ErrorMessage,
	}
}

func waitForStop(ctx context.Context, configWatcher *config.FileWatcher) {
	// Wait for interrupt signal or context cancellation
	sigChan := make(chan os.Signal, 1)
	sighupChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	signal.Notify(sighupChan, syscall.SIGHUP)

	for {
		select {
		case <-sigChan:
			// Shutdown signal received
			logger.Info("Shutdown signal received...")
		case <-sighupChan:
			// SIGHUP received - trigger config reload
			logger.Info("SIGHUP received, reloading configuration...")
			if configWatcher != nil {
				if err := configWatcher.TriggerReload(); err != nil {
					logger.WithError(err).Error("Failed to reload configuration on SIGHUP")
				}
			} else {
				logger.Warn("SIGHUP received but config watcher is not enabled")
			}
			continue // Continue running after SIGHUP
		case <-ctx.Done():
			// Context cancelled (e.g., timeout in tests)
		}
		break // Exit loop for shutdown signals and context cancellation
	}
}
