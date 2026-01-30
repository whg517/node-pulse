package monitor

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// Start starts the resource monitoring
func (m *monitor) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return ErrAlreadyRunning
	}
	m.running = true
	m.mu.Unlock()

	m.wg.Add(1)
	go m.monitoringLoop()

	return nil
}

// Stop stops the resource monitoring
func (m *monitor) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	close(m.stopCh)
	m.wg.Wait()
}

// GetDegradationLevel returns the current degradation level
func (m *monitor) GetDegradationLevel() DegradationLevel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.level
}

// GetResourceUsage returns the latest resource usage
func (m *monitor) GetResourceUsage() *ResourceUsage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentUsage
}

// GetAlerts returns the history of alerts
func (m *monitor) GetAlerts() []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]Alert, len(m.alerts))
	copy(alerts, m.alerts)
	return alerts
}

// IsRunning returns whether the monitor is running
func (m *monitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// monitoringLoop is the main monitoring loop
func (m *monitor) monitoringLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Duration(m.cfg.CheckIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkResources()
		}
	}
}

// checkResources checks resource usage and triggers actions
func (m *monitor) checkResources() {
	usage, err := m.collectResourceUsage()
	if err != nil {
		m.logger.Errorf("Failed to collect resource usage: %v", err)
		return
	}

	m.mu.Lock()
	m.currentUsage = usage
	m.mu.Unlock()

	// Check thresholds and trigger alerts
	m.checkThresholds(usage)

	// Evaluate degradation level
	m.evaluateDegradation(usage)
}

// collectResourceUsage collects current resource usage statistics
func (m *monitor) collectResourceUsage() (*ResourceUsage, error) {
	// Get CPU usage percentage
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}

	// Convert to microcores (100% = 1000 microcores on single-core)
	cpuMicrocores := cpuPercent[0] * 10

	// Get memory usage
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	memoryMB := float64(memStat.Used) / 1024 / 1024

	return &ResourceUsage{
		CPUMicrocores: cpuMicrocores,
		MemoryMB:      memoryMB,
		Timestamp:     time.Now().Unix(),
	}, nil
}

// checkThresholds checks if resource usage exceeds thresholds and triggers alerts
func (m *monitor) checkThresholds(usage *ResourceUsage) {
	now := time.Now().Unix()

	// Check CPU threshold
	if usage.CPUMicrocores > float64(m.cfg.Thresholds.CPUMicrocores) {
		m.maybeTriggerAlert("cpu", usage.CPUMicrocores, float64(m.cfg.Thresholds.CPUMicrocores), now)
	}

	// Check memory threshold
	if usage.MemoryMB > float64(m.cfg.Thresholds.MemoryMB) {
		m.maybeTriggerAlert("memory", usage.MemoryMB, float64(m.cfg.Thresholds.MemoryMB), now)
	}
}

// maybeTriggerAlert triggers an alert if not within suppression window
func (m *monitor) maybeTriggerAlert(resourceType string, currentValue, threshold float64, now int64) {
	// Check suppression window
	m.mu.RLock()
	lastAlert, exists := m.lastAlertTime[resourceType]
	m.mu.RUnlock()

	suppressionWindow := int64(m.cfg.Alerting.SuppressionWindowSeconds)
	if exists && (now-lastAlert) < suppressionWindow {
		return // Within suppression window
	}

	// Determine alert level
	var level string
	if resourceType == "cpu" {
		if currentValue > float64(m.cfg.Degradation.CriticalLevel.CPUMicrocores) {
			level = "critical"
		} else {
			level = "degraded"
		}
	} else {
		if currentValue > float64(m.cfg.Degradation.CriticalLevel.MemoryMB) {
			level = "critical"
		} else {
			level = "degraded"
		}
	}

	// Record alert
	alert := Alert{
		ResourceType: resourceType,
		CurrentValue: currentValue,
		Threshold:    threshold,
		Level:        level,
		Timestamp:    now,
	}

	m.mu.Lock()
	m.alerts = append(m.alerts, alert)
	m.lastAlertTime[resourceType] = now
	m.mu.Unlock()

	// Log warning
	m.logger.Warnf("Resource usage exceeded: %s=%.2f (threshold=%.2f), level=%s",
		resourceType, currentValue, threshold, level)
}

// evaluateDegradation evaluates and updates degradation level
func (m *monitor) evaluateDegradation(usage *ResourceUsage) {
	var newLevel DegradationLevel

	// Check critical level
	if usage.CPUMicrocores > float64(m.cfg.Degradation.CriticalLevel.CPUMicrocores) ||
		usage.MemoryMB > float64(m.cfg.Degradation.CriticalLevel.MemoryMB) {
		newLevel = DegradationLevelCritical
	} else if usage.CPUMicrocores > float64(m.cfg.Degradation.DegradedLevel.CPUMicrocores) ||
		usage.MemoryMB > float64(m.cfg.Degradation.DegradedLevel.MemoryMB) {
		newLevel = DegradationLevelDegraded
	} else {
		newLevel = DegradationLevelNormal
	}

	m.mu.RLock()
	currentLevel := m.level
	requiredChecks := m.cfg.Degradation.Recovery.ConsecutiveNormalChecks
	m.mu.RUnlock()

	// Only transition to Normal if we've had consecutive normal checks
	if newLevel == DegradationLevelNormal && currentLevel != DegradationLevelNormal {
		// Increment consecutive normal checks counter
		m.mu.Lock()
		m.consecutiveNormalChecks++
		currentConsecutive := m.consecutiveNormalChecks
		m.mu.Unlock()

		// Only transition to Normal if we've met the required consecutive checks
		if currentConsecutive >= requiredChecks {
			m.updateDegradationLevel(newLevel)
		}
		return
	}

	// Immediate transition for non-normal levels or when staying in normal
	if newLevel != currentLevel {
		m.updateDegradationLevel(newLevel)
	}
}

// updateDegradationLevel updates the degradation level and notifies probe manager
func (m *monitor) updateDegradationLevel(newLevel DegradationLevel) {
	m.mu.Lock()
	oldLevel := m.level

	// Reset consecutive normal checks when leaving Normal level
	if newLevel != DegradationLevelNormal {
		m.consecutiveNormalChecks = 0
	}

	m.level = newLevel
	m.mu.Unlock()

	// Log level change
	m.logger.Infof("Degradation level changed: %s -> %s", oldLevel.String(), newLevel.String())

	// Update probe interval
	multiplier := m.getIntervalMultiplier(newLevel)
	if err := m.probeMgr.UpdateProbeInterval(multiplier); err != nil {
		m.logger.Errorf("Failed to update probe interval (multiplier=%d): %v", multiplier, err)
	}
}

// getIntervalMultiplier returns the interval multiplier for a degradation level
func (m *monitor) getIntervalMultiplier(level DegradationLevel) int {
	switch level {
	case DegradationLevelDegraded:
		return m.cfg.Degradation.DegradedLevel.IntervalMultiplier
	case DegradationLevelCritical:
		return m.cfg.Degradation.CriticalLevel.IntervalMultiplier
	default:
		return 1
	}
}
