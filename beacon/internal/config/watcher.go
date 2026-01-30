package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// FileWatcher monitors configuration file for changes and supports hot reloading
type FileWatcher struct {
	path     string
	config   atomic.Value // stores *Config
	debounce time.Duration
	logger   *logrus.Logger

	mu            sync.RWMutex
	callbacks     []func(*Config, []string) error // Added changes parameter
	version       int64
	timer         *time.Timer
	timerMu       sync.Mutex // Protects timer access
	reloadCount   int64       // Total reload count
	lastReload    time.Time   // Last successful reload timestamp
}

// NewFileWatcher creates a new configuration file watcher
func NewFileWatcher(path string, initialConfig *Config, logger *logrus.Logger) (*FileWatcher, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}

	fw := &FileWatcher{
		path:     path,
		debounce: 1 * time.Second, // 1 second debounce
		logger:   logger,
		version:  1,
	}
	fw.config.Store(initialConfig)

	return fw, nil
}

// Start begins monitoring the configuration file for changes
func (fw *FileWatcher) Start(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(fw.path); err != nil {
		return fmt.Errorf("failed to watch config file: %w", err)
	}

	fw.logger.WithFields(logrus.Fields{
		"path":    fw.path,
		"version": fw.version,
	}).Info("Config watcher started")

	var timer *time.Timer

	for {
		select {
		case <-ctx.Done():
			fw.logger.Info("Config watcher stopped")
			if timer != nil {
				timer.Stop()
			}
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only handle Write and Create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			fw.logger.WithField("event", event.Op.String()).Debug("File event detected")

			// Debounce: reset timer
			fw.timerMu.Lock()
			if fw.timer != nil {
				fw.timer.Stop()
			}
			fw.timer = time.AfterFunc(fw.debounce, func() {
				if err := fw.reloadConfig(); err != nil {
					fw.logger.WithError(err).Error("Failed to reload config")
				}
			})
			fw.timerMu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fw.logger.WithError(err).Error("File watcher error")
		}
	}
}

// reloadConfig reloads the configuration from file
func (fw *FileWatcher) reloadConfig() error {
	fw.logger.WithFields(logrus.Fields{
		"path":            fw.path,
		"current_version": fw.version,
	}).Info("Reloading configuration")

	// Check file permissions (security check)
	fileInfo, err := os.Stat(fw.path)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}
	// Warn if file is world-writable (permissions 0777 or others have write access)
	perms := fileInfo.Mode().Perm()
	if perms&0002 != 0 {
		fw.logger.WithField("permissions", perms).Warn("Config file is world-writable, this is a security risk")
	}

	// Load new config
	newConfig, err := LoadConfig(fw.path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate new config
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Get old config for diff
	oldConfig := fw.config.Load().(*Config)
	changes := fw.diffConfig(oldConfig, newConfig)

	if len(changes) == 0 {
		fw.logger.Info("No configuration changes detected")
		return nil
	}

	// Store old version for rollback and new version
	oldVersion := fw.version
	newVersion := fw.version + 1

	// Log summary of changes
	changeCount := len(changes)
	changeSummary := fmt.Sprintf("%d change(s)", changeCount)
	if changeCount <= 5 {
		// Log all changes if few
		fw.logger.WithFields(logrus.Fields{
			"changes":     changes,
			"new_version": newVersion,
		}).Info("Configuration changes detected")
	} else {
		// Log summary if many changes
		fw.logger.WithFields(logrus.Fields{
			"summary":     changeSummary,
			"new_version": newVersion,
		}).Info("Configuration changes detected")
		fw.logger.WithField("changes", changes).Debug("Detailed changes")
	}

	// Check for restart-required fields and warn
	for _, change := range changes {
		if strings.Contains(change, "node_id:") || strings.Contains(change, "node_name:") {
			fw.logger.Warn("Configuration change requires Beacon restart to take full effect")
		}
	}

	// Store old config for potential rollback
	oldConfigCopy := *oldConfig

	// Atomically switch config
	fw.config.Store(newConfig)
	fw.version = newVersion // Update version BEFORE callbacks

	// Execute callbacks
	fw.mu.RLock()
	callbacks := make([]func(*Config, []string) error, len(fw.callbacks))
	copy(callbacks, fw.callbacks)
	fw.mu.RUnlock()

	for _, callback := range callbacks {
		if err := callback(newConfig, changes); err != nil {
			fw.logger.WithError(err).Error("Config reload callback failed")

			// Rollback config AND version
			fw.config.Store(&oldConfigCopy)
			fw.version = oldVersion
			return fmt.Errorf("callback failed, config rolled back: %w", err)
		}
	}

	// Track reload statistics
	fw.reloadCount++
	fw.lastReload = time.Now()

	fw.logger.WithFields(logrus.Fields{
		"version":      fw.version,
		"reload_count": fw.reloadCount,
		"summary":      changeSummary,
	}).Info("Configuration reloaded successfully")

	return nil
}

// diffConfig computes differences between old and new configuration
func (fw *FileWatcher) diffConfig(old, new *Config) []string {
	var changes []string

	// Check pulse_server
	if old.PulseServer != new.PulseServer {
		changes = append(changes, fmt.Sprintf("pulse_server: %s -> %s", old.PulseServer, new.PulseServer))
	}

	// Check node_id (requires restart warning)
	if old.NodeID != new.NodeID {
		changes = append(changes, fmt.Sprintf("node_id: %s -> %s (WARNING: requires restart)", old.NodeID, new.NodeID))
	}

	// Check node_name (requires restart warning)
	if old.NodeName != new.NodeName {
		changes = append(changes, fmt.Sprintf("node_name: %s -> %s (WARNING: requires restart)", old.NodeName, new.NodeName))
	}

	// Check probes in detail (not just count)
	oldLen := len(old.Probes)
	newLen := len(new.Probes)

	if oldLen != newLen {
		changes = append(changes, fmt.Sprintf("probes: %d -> %d probes", oldLen, newLen))
	}

	// Compare probe configurations in detail
	maxProbes := oldLen
	if newLen > maxProbes {
		maxProbes = newLen
	}

	for i := 0; i < maxProbes; i++ {
		if i >= oldLen {
			// New probe added
			p := new.Probes[i]
			changes = append(changes, fmt.Sprintf("probes[%d]: ADDED %s probe to %s:%d (interval=%ds, timeout=%ds, count=%d)",
				i, p.Type, p.Target, p.Port, p.Interval, p.TimeoutSeconds, p.Count))
		} else if i >= newLen {
			// Probe removed
			p := old.Probes[i]
			changes = append(changes, fmt.Sprintf("probes[%d]: REMOVED %s probe to %s:%d",
				i, p.Type, p.Target, p.Port))
		} else {
			// Compare existing probe
			oldProbe := old.Probes[i]
			newProbe := new.Probes[i]

			if oldProbe.Type != newProbe.Type {
				changes = append(changes, fmt.Sprintf("probes[%d]: type %s -> %s", i, oldProbe.Type, newProbe.Type))
			}
			if oldProbe.Target != newProbe.Target {
				changes = append(changes, fmt.Sprintf("probes[%d]: target %s -> %s", i, oldProbe.Target, newProbe.Target))
			}
			if oldProbe.Port != newProbe.Port {
				changes = append(changes, fmt.Sprintf("probes[%d]: port %d -> %d", i, oldProbe.Port, newProbe.Port))
			}
			if oldProbe.Interval != newProbe.Interval {
				changes = append(changes, fmt.Sprintf("probes[%d]: interval %d -> %d seconds", i, oldProbe.Interval, newProbe.Interval))
			}
			if oldProbe.TimeoutSeconds != newProbe.TimeoutSeconds {
				changes = append(changes, fmt.Sprintf("probes[%d]: timeout %d -> %d seconds", i, oldProbe.TimeoutSeconds, newProbe.TimeoutSeconds))
			}
			if oldProbe.Count != newProbe.Count {
				changes = append(changes, fmt.Sprintf("probes[%d]: count %d -> %d", i, oldProbe.Count, newProbe.Count))
			}
		}
	}

	return changes
}

// GetConfig returns the current configuration (thread-safe)
func (fw *FileWatcher) GetConfig() *Config {
	return fw.config.Load().(*Config)
}

// GetConfigPath returns the configuration file path
func (fw *FileWatcher) GetConfigPath() string {
	return fw.path
}

// OnReload registers a callback to be invoked when config is reloaded
// The callback receives the new config and a list of changes
func (fw *FileWatcher) OnReload(callback func(*Config, []string) error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.callbacks = append(fw.callbacks, callback)
}

// GetVersion returns the current configuration version
func (fw *FileWatcher) GetVersion() int64 {
	return fw.version
}

// GetReloadCount returns the total number of successful reloads
func (fw *FileWatcher) GetReloadCount() int64 {
	return fw.reloadCount
}

// GetLastReloadTime returns the timestamp of the last successful reload
func (fw *FileWatcher) GetLastReloadTime() time.Time {
	return fw.lastReload
}
