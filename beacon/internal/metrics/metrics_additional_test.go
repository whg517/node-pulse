package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/beacon/internal/config"
	"github.com/whg517/node-pulse/beacon/internal/logger"
	"github.com/whg517/node-pulse/beacon/internal/models"
	"github.com/whg517/node-pulse/beacon/internal/probe"
)

// mockModeProvider implements ModeProvider for testing
type mockModeProvider struct {
	mode   config.OperatingMode
	source config.ConfigSource
}

func (m *mockModeProvider) GetMode() config.OperatingMode {
	return m.mode
}

func (m *mockModeProvider) GetConfigSource() config.ConfigSource {
	return m.source
}

// mockCacheStatsProvider implements CacheStatsProvider for testing
type mockCacheStatsProvider struct {
	size      int64
	count     int
	evictions int64
}

func (m *mockCacheStatsProvider) Size() int64    { return m.size }
func (m *mockCacheStatsProvider) Count() int     { return m.count }
func (m *mockCacheStatsProvider) Evictions() int64 { return m.evictions }

// mockCompressionProvider implements CompressionStatsProvider for testing
type mockCompressionProvider struct {
	ratio float64
}

func (m *mockCompressionProvider) GetCompressionRatio() float64 { return m.ratio }

func initAdditionalTestLogger(t *testing.T) {
	t.Helper()
	if err := logger.InitLogger(&config.Config{
		LogLevel:      "INFO",
		LogFile:       "/tmp/test-metrics-additional.log",
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   false,
		LogToConsole:  false,
	}); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
}

func newTestMetrics(t *testing.T, port int) (*Metrics, *probe.ProbeScheduler) {
	t.Helper()
	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          port,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)
	return NewMetrics(cfg, scheduler), scheduler
}

// TestMetrics_SetModeProvider tests SetModeProvider
func TestMetrics_SetModeProvider(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19200)

	provider := &mockModeProvider{
		mode:   config.ModeRegistered,
		source: config.SourceLocal,
	}
	m.SetModeProvider(provider)

	// Verify provider is set
	m.mu.RLock()
	assert.NotNil(t, m.modeProvider)
	m.mu.RUnlock()
}

// TestMetrics_SetCacheStatsProvider tests SetCacheStatsProvider
func TestMetrics_SetCacheStatsProvider(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19201)

	provider := &mockCacheStatsProvider{size: 1024, count: 5, evictions: 2}
	m.SetCacheStatsProvider(provider)

	m.mu.RLock()
	assert.NotNil(t, m.cacheStatsProvider)
	m.mu.RUnlock()
}

// TestMetrics_SetCompressionStatsProvider tests SetCompressionStatsProvider
func TestMetrics_SetCompressionStatsProvider(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19202)

	provider := &mockCompressionProvider{ratio: 0.6}
	m.SetCompressionStatsProvider(provider)

	m.mu.RLock()
	assert.NotNil(t, m.compressionProvider)
	m.mu.RUnlock()
}

// TestMetrics_UpdateMode tests UpdateMode
func TestMetrics_UpdateMode(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19203)

	// Should not panic
	m.UpdateMode(config.ModeRegistered)
	m.UpdateMode(config.ModeDegraded)
	m.UpdateMode(config.ModeStandalone)
}

// TestMetrics_UpdateConfigSource tests UpdateConfigSource
func TestMetrics_UpdateConfigSource(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19204)

	// Should not panic
	m.UpdateConfigSource(config.SourceLocal)
	m.UpdateConfigSource(config.SourceServer)
	m.UpdateConfigSource(config.SourceCached)
}

// TestMetrics_UpdateActiveProbes tests UpdateActiveProbes
func TestMetrics_UpdateActiveProbes(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19205)

	// Should not panic
	m.UpdateActiveProbes(0)
	m.UpdateActiveProbes(5)
}

// TestMetrics_UpdateCompressionRatio tests UpdateCompressionRatio
func TestMetrics_UpdateCompressionRatio(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19206)

	// Should not panic
	m.UpdateCompressionRatio(0.6)
	m.UpdateCompressionRatio(0.0)
}

// TestMetrics_UpdateCacheSize tests UpdateCacheSize
func TestMetrics_UpdateCacheSize(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19207)

	// Should not panic
	m.UpdateCacheSize(1024)
	m.UpdateCacheSize(0)
}

// TestMetrics_UpdateConfigVersion tests UpdateConfigVersion
func TestMetrics_UpdateConfigVersion(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19208)

	// Should not panic
	m.UpdateConfigVersion(1)
	m.UpdateConfigVersion(42)
}

// TestMetrics_RecordProbe tests RecordProbe
func TestMetrics_RecordProbe(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19209)

	// Should not panic
	m.RecordProbe("tcp_ping", 0.001)
	m.RecordProbe("udp_ping", 0.002)
}

// TestMetrics_RecordProbeFailure tests RecordProbeFailure
func TestMetrics_RecordProbeFailure(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19210)

	// Should not panic
	m.RecordProbeFailure("tcp_ping", "timeout")
	m.RecordProbeFailure("udp_ping", "connection_refused")
}

// TestMetrics_UpdateMemoryUsage tests UpdateMemoryUsage
func TestMetrics_UpdateMemoryUsage(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19211)

	// Should not panic
	m.UpdateMemoryUsage(1024 * 1024)
	m.UpdateMemoryUsage(0)
}

// TestMetrics_UpdateCPUUsage tests UpdateCPUUsage
func TestMetrics_UpdateCPUUsage(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19212)

	// Should not panic
	m.UpdateCPUUsage(25.5)
	m.UpdateCPUUsage(0.0)
}

// TestMetrics_RecordResumeUpload tests RecordResumeUpload
func TestMetrics_RecordResumeUpload(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19213)

	// Should not panic
	m.RecordResumeUpload(512)
	m.RecordResumeUpload(1024)
}

// TestMetrics_RecordCacheEviction tests RecordCacheEviction
func TestMetrics_RecordCacheEviction(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19214)

	// Should not panic
	m.RecordCacheEviction()
	m.RecordCacheEviction()
}

// TestMetrics_RecordCompressionCorruption tests RecordCompressionCorruption
func TestMetrics_RecordCompressionCorruption(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19215)

	// Should not panic
	m.RecordCompressionCorruption()
}

// TestMetrics_StartWithModeProvider tests Start with mode provider
func TestMetrics_StartWithModeProvider(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	m, _ := newTestMetrics(t, 19216)

	// Set providers before start
	m.SetModeProvider(&mockModeProvider{
		mode:   config.ModeRegistered,
		source: config.SourceServer,
	})
	m.SetCacheStatsProvider(&mockCacheStatsProvider{size: 512, count: 2, evictions: 0})
	m.SetCompressionStatsProvider(&mockCompressionProvider{ratio: 0.7})

	if err := m.Start(); err != nil {
		t.Fatalf("Start() with providers failed: %v", err)
	}
	defer func() { _ = m.Stop() }()

	assert.True(t, m.IsRunning())
}

// TestMetrics_UpdateMetrics_WithSuccessfulResults tests updateMetrics with successful results
func TestMetrics_UpdateMetrics_WithSuccessfulResults(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19218,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Manually set results for the scheduler (simulate probe results)
	// by calling updateMetrics directly
	m.updateMetrics()
	// No panic - covered
}

// TestMetrics_UpdateEnhancedMetrics_WithProviders tests updateEnhancedMetrics with all providers
func TestMetrics_UpdateEnhancedMetrics_WithProviders(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19219,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Set all providers
	m.SetModeProvider(&mockModeProvider{
		mode:   config.ModeDegraded,
		source: config.SourceCached,
	})
	m.SetCacheStatsProvider(&mockCacheStatsProvider{
		size:      2048,
		count:     10,
		evictions: 5,
	})
	m.SetCompressionStatsProvider(&mockCompressionProvider{ratio: 0.8})

	// Call updateEnhancedMetrics
	m.updateEnhancedMetrics()
	// No panic - covered
}

// TestMetrics_UpdateEnhancedMetrics_NoProviders tests updateEnhancedMetrics without providers
func TestMetrics_UpdateEnhancedMetrics_NoProviders(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19220,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)
	// No providers set
	m.updateEnhancedMetrics()
	// No panic - covered
}

// TestMetrics_UpdateEnhancedMetrics_NilScheduler tests updateEnhancedMetrics with nil scheduler
func TestMetrics_UpdateEnhancedMetrics_NilScheduler(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19221,
		MetricsUpdateSeconds: 10,
	}
	m := NewMetrics(cfg, nil)
	m.updateEnhancedMetrics()
	// No panic - covered
}

// TestMetrics_UpdateMetrics_AllProbesFailed tests updateMetrics when all probes failed
func TestMetrics_UpdateMetrics_AllProbesFailed(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19222,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// With no probe results, updateMetrics should handle the "all failed" case
	m.updateMetrics()
	// No panic - covered
}

// TestMetrics_UpdateMetrics_WithTCPSuccessResults tests updateMetrics with successful TCP results
func TestMetrics_UpdateMetrics_WithTCPSuccessResults(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19230,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Inject successful TCP results
	tcpResults := []*models.TCPProbeResult{
		{
			Success:        true,
			RTTMs:          2.5,
			JitterMs:       0.1,
			PacketLossRate: 0.0,
			SampleCount:    10,
		},
	}
	scheduler.SetLatestResultsForTest(tcpResults, nil)

	// Should hit the count > 0 path
	m.updateMetrics()
}

// TestMetrics_UpdateMetrics_WithUDPSuccessResults tests updateMetrics with successful UDP results
func TestMetrics_UpdateMetrics_WithUDPSuccessResults(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19231,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Inject successful UDP results
	udpResults := []*models.UDPProbeResult{
		{
			Success:         true,
			RTTMs:           3.0,
			JitterMs:        0.2,
			PacketLossRate:  5.0,
			SentPackets:     10,
			ReceivedPackets: 9,
		},
	}
	scheduler.SetLatestResultsForTest(nil, udpResults)

	// Should hit the count > 0 path
	m.updateMetrics()
}

// TestMetrics_UpdateMetrics_WithMixedResults tests updateMetrics with mixed results (some nil/failed)
func TestMetrics_UpdateMetrics_WithMixedResults(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19232,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Inject mixed results (some successful, some not)
	tcpResults := []*models.TCPProbeResult{
		{Success: true, RTTMs: 2.0, JitterMs: 0.1, PacketLossRate: 0.0},
		{Success: false, RTTMs: 0, ErrorMessage: "timeout"},
		nil,
	}
	scheduler.SetLatestResultsForTest(tcpResults, nil)

	// Should handle mixed results
	m.updateMetrics()
}

// TestMetrics_UpdateMetrics_AllFailed_WithResults tests when results exist but all failed
func TestMetrics_UpdateMetrics_AllFailed_WithResults(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19233,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)

	// Inject results where all probes failed
	tcpResults := []*models.TCPProbeResult{
		{Success: false, RTTMs: 0, ErrorMessage: "connection refused"},
	}
	scheduler.SetLatestResultsForTest(tcpResults, nil)

	// Should hit the "all probes failed" branch (count == 0, total > 0)
	m.updateMetrics()
}

// TestMetrics_InitializeEnhancedMetrics_WithModeProvider tests initializeEnhancedMetrics with mode provider
func TestMetrics_InitializeEnhancedMetrics_WithModeProvider(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19223,
		MetricsUpdateSeconds: 10,
	}
	scheduler, err := probe.NewProbeScheduler([]config.ProbeConfig{})
	require.NoError(t, err)

	m := NewMetrics(cfg, scheduler)
	m.SetModeProvider(&mockModeProvider{
		mode:   config.ModeStandalone,
		source: config.SourceLocal,
	})

	// Start to trigger initializeEnhancedMetrics
	if err := m.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer func() { _ = m.Stop() }()

	assert.True(t, m.IsRunning())
}

// TestMetrics_StartWithNilScheduler tests Start with nil scheduler
func TestMetrics_StartWithNilScheduler(t *testing.T) {
	initAdditionalTestLogger(t)
	defer func() { _ = logger.Close() }()

	cfg := &config.Config{
		NodeID:               "test-node",
		NodeName:             "test",
		MetricsEnabled:       true,
		MetricsPort:          19217,
		MetricsUpdateSeconds: 10,
	}
	m := NewMetrics(cfg, nil)

	if err := m.Start(); err != nil {
		t.Fatalf("Start() with nil scheduler failed: %v", err)
	}
	defer func() { _ = m.Stop() }()

	assert.True(t, m.IsRunning())
}
