package diagnostics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/beacon/internal/config"
	"github.com/whg517/node-pulse/beacon/internal/logger"
)

// ---------------------------------------------------------------------------
// Hand-written provider mocks (function-field-free, value-receiver style).
// These mirror internal/metrics/metrics_additional_test.go and implement the
// four provider interfaces defined in metrics.go.
// ---------------------------------------------------------------------------

type mockModeProvider struct {
	mode   config.OperatingMode
	source config.ConfigSource
}

func (m *mockModeProvider) GetMode() config.OperatingMode        { return m.mode }
func (m *mockModeProvider) GetConfigSource() config.ConfigSource { return m.source }

// mockModeStatusProvider additionally implements the anonymous
// modeStatusProvider interface checked in collectModeStatus.
type mockModeStatusProvider struct {
	mockModeProvider
	consecutiveFailures int
	lastSuccessTime     time.Time
	lastFailureTime     time.Time
}

func (m *mockModeStatusProvider) GetConsecutiveFailures() int    { return m.consecutiveFailures }
func (m *mockModeStatusProvider) GetLastSuccessTime() time.Time  { return m.lastSuccessTime }
func (m *mockModeStatusProvider) GetLastFailureTime() time.Time  { return m.lastFailureTime }

type mockCacheStatsProvider struct {
	size      int64
	count     int
	evictions int64
}

func (m *mockCacheStatsProvider) Size() int64      { return m.size }
func (m *mockCacheStatsProvider) Count() int       { return m.count }
func (m *mockCacheStatsProvider) Evictions() int64 { return m.evictions }

type mockCompressionProvider struct{ ratio float64 }

func (m *mockCompressionProvider) GetCompressionRatio() float64 { return m.ratio }

type mockProbeCountProvider struct{ count int }

func (m *mockProbeCountProvider) GetProbeCount() int { return m.count }

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// initTestLogger initialises the global logger so downstream formatting / log
// paths do not panic. Mirrors internal/metrics/metrics_additional_test.go.
func initTestLogger(t *testing.T) {
	t.Helper()
	if err := logger.InitLogger(&config.Config{
		LogLevel:      "INFO",
		LogFile:       "/tmp/test-diagnostics.log",
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   false,
		LogToConsole:  false,
	}); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
}

func newBaseConfig() *config.Config {
	return &config.Config{
		PulseServer: "http://localhost:6532",
		NodeID:      "test-node-1",
		NodeName:    "Test Node",
		Region:      "us-east",
		Tags:        []string{"tag1", "tag2"},
		LogLevel:    "INFO",
		DebugMode:   false,
		ConfigPath:  "",
		Mode:        config.ModeConfig{Mode: config.ModeRegistered},
	}
}

// ---------------------------------------------------------------------------
// Pure-function table-driven tests
// ---------------------------------------------------------------------------

func TestExtractHost(t *testing.T) {
	tests := []struct {
		name       string
		pulseServer string
		want       string
	}{
		{"http with default port", "http://example.com", "example.com:80"},
		{"https with default port", "https://example.com", "example.com:443"},
		{"explicit port preserved", "http://example.com:8080", "example.com:8080"},
		{"https explicit port preserved", "https://example.com:8443", "example.com:8443"},
		{"with path strips path", "http://example.com/some/path", "example.com:80"},
		{"plain host string (no scheme) returned as-is", "example.com:6532", "example.com:6532"},
		{"empty string returned as-is", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractHost(tt.pulseServer))
		})
	}
}

func TestContainsColon(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"host:8080", true},
		{"host", false},
		{":8080", true},
		{"", false},
		{"no-port-here", false},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			assert.Equal(t, tt.want, containsColon(tt.s))
		})
	}
}

func TestFormatProbeTarget(t *testing.T) {
	tests := []struct {
		name  string
		probe config.ProbeConfig
		want  string
	}{
		{"with port", config.ProbeConfig{Target: "1.2.3.4", Port: 80}, "1.2.3.4:80"},
		{"port zero omits port", config.ProbeConfig{Target: "example.com", Port: 0}, "example.com"},
		{"negative port omits port", config.ProbeConfig{Target: "example.com", Port: -1}, "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatProbeTarget(tt.probe))
		})
	}
}

// ---------------------------------------------------------------------------
// enhanceMetrics
// ---------------------------------------------------------------------------

func TestEnhanceMetrics_NoProvidersAppliesDefaults(t *testing.T) {
	c := &collector{}
	m := &PrometheusMetrics{}

	c.enhanceMetrics(m)

	assert.Equal(t, string(config.ModeRegistered), m.Mode)
	assert.Equal(t, string(config.SourceLocal), m.ConfigSource)
	assert.Equal(t, 0, m.ActiveProbes)
	assert.Equal(t, int64(0), m.CacheSizeBytes)
	assert.Equal(t, int64(0), m.CacheEvictions)
	assert.Equal(t, float64(0), m.CompressionRatio)
}

func TestEnhanceMetrics_WithProviders(t *testing.T) {
	c := &collector{
		modeProvider:        &mockModeProvider{mode: config.ModeStandalone, source: config.SourceServer},
		probeCountProvider:  &mockProbeCountProvider{count: 7},
		cacheStatsProvider:  &mockCacheStatsProvider{size: 1024, count: 5, evictions: 3},
		compressionProvider: &mockCompressionProvider{ratio: 55.5},
	}
	m := &PrometheusMetrics{}

	c.enhanceMetrics(m)

	assert.Equal(t, string(config.ModeStandalone), m.Mode)
	assert.Equal(t, string(config.SourceServer), m.ConfigSource)
	assert.Equal(t, 7, m.ActiveProbes)
	assert.Equal(t, int64(1024), m.CacheSizeBytes)
	assert.Equal(t, int64(3), m.CacheEvictions)
	assert.Equal(t, 55.5, m.CompressionRatio)
}

// ---------------------------------------------------------------------------
// collectModeStatus
// ---------------------------------------------------------------------------

func TestCollectModeStatus_NilProviderReturnsNil(t *testing.T) {
	c := &collector{}
	assert.Nil(t, c.collectModeStatus())
}

func TestCollectModeStatus_BasicProvider(t *testing.T) {
	c := &collector{
		modeProvider: &mockModeProvider{mode: config.ModeDegraded, source: config.SourceServer},
	}
	status := c.collectModeStatus()
	require.NotNil(t, status)
	assert.Equal(t, string(config.ModeDegraded), status.CurrentMode)
	assert.Equal(t, string(config.SourceServer), status.ConfigSource)
	assert.Equal(t, 0, status.ConsecutiveFailures, "basic provider does not expose failure info")
}

func TestCollectModeStatus_StatusProviderExtendsInfo(t *testing.T) {
	now := time.Now()
	c := &collector{
		modeProvider: &mockModeStatusProvider{
			mockModeProvider:    mockModeProvider{mode: config.ModeRegistered, source: config.SourceLocal},
			consecutiveFailures: 3,
			lastSuccessTime:     now,
			lastFailureTime:     now,
		},
	}
	status := c.collectModeStatus()
	require.NotNil(t, status)
	assert.Equal(t, 3, status.ConsecutiveFailures)
	assert.Equal(t, now, status.LastSuccessTime)
	assert.Equal(t, now, status.LastFailureTime)
}

// ---------------------------------------------------------------------------
// collectConfiguration / getConfigVersion
// ---------------------------------------------------------------------------

func TestCollectConfiguration(t *testing.T) {
	c := &collector{cfg: newBaseConfig()}
	c.cfg.ConfigPath = ""
	c.cfg.Compression.Enabled = true
	c.cfg.Compression.Level = 6
	c.cfg.Compression.MinSizeBytes = 2048
	c.cfg.Resume.Enabled = true
	c.cfg.Resume.MaxCacheSizeBytes = 10 * 1024 * 1024 // 10 MB
	c.cfg.MetricsEnabled = true
	c.cfg.MetricsPort = 2112

	info, err := c.collectConfiguration()
	require.NoError(t, err)
	assert.True(t, info.ConfigValid)
	assert.Equal(t, "INFO", info.LogLevel)
	assert.Equal(t, "unknown", info.ConfigVersion, "empty ConfigPath -> unknown version")
	assert.Equal(t, string(config.ModeRegistered), info.OperatingMode)
	assert.True(t, info.Compression.Enabled)
	assert.Equal(t, 6, info.Compression.Level)
	assert.Equal(t, float64(2), info.Compression.MinSizeKB) // 2048/1024
	assert.True(t, info.Resume.Enabled)
	assert.Equal(t, float64(10), info.Resume.MaxCacheSizeMB) // 10MB / 1024 / 1024

	// config_content should carry the metrics keys when enabled
	assert.Equal(t, true, info.ConfigContent["metrics_enabled"])
	assert.Equal(t, 2112, info.ConfigContent["metrics_port"])
}

func TestGetConfigVersion_EmptyPath(t *testing.T) {
	c := &collector{cfg: &config.Config{ConfigPath: ""}}
	assert.Equal(t, "unknown", c.getConfigVersion())
}

func TestGetConfigVersion_MissingFile(t *testing.T) {
	c := &collector{cfg: &config.Config{ConfigPath: "/nonexistent/path/beacon.yaml"}}
	assert.Equal(t, "unknown", c.getConfigVersion())
}

func TestGetConfigVersion_ExistingFile(t *testing.T) {
	// Use a known existing file (this test source itself) so os.Stat succeeds.
	c := &collector{cfg: &config.Config{ConfigPath: "config.go"}}
	v := c.getConfigVersion()
	assert.NotEqual(t, "unknown", v, "existing file should produce an RFC3339 timestamp")
	_, err := time.Parse(time.RFC3339, v)
	assert.NoError(t, err, "config version should be RFC3339-parseable")
}

// ---------------------------------------------------------------------------
// collectResourceMonitorInfo
// ---------------------------------------------------------------------------

func TestCollectResourceMonitorInfo_Disabled(t *testing.T) {
	c := &collector{cfg: &config.Config{ResourceMonitor: config.ResourceMonitorConfig{Enabled: false}}}
	info := c.collectResourceMonitorInfo()
	require.NotNil(t, info)
	assert.False(t, info.Enabled)
}

func TestCollectResourceMonitorInfo_Enabled(t *testing.T) {
	cfg := &config.Config{ResourceMonitor: config.ResourceMonitorConfig{
		Enabled:              true,
		CheckIntervalSeconds: 30,
		Alerting:             config.AlertingConfig{SuppressionWindowSeconds: 600},
		Thresholds:           config.ThresholdsConfig{CPUMicrocores: 250, MemoryMB: 512},
	}}
	c := &collector{cfg: cfg}
	info := c.collectResourceMonitorInfo()
	require.NotNil(t, info)
	assert.True(t, info.Enabled)
	assert.True(t, info.Configured)
	assert.Equal(t, 30, info.CheckIntervalSeconds)
	assert.Equal(t, 10, info.AlertSuppressionMinutes, "600s / 60 = 10 minutes")
	assert.Equal(t, 250, info.Thresholds.CPUMicrocores)
	assert.Equal(t, 512, info.Thresholds.MemoryMB)
}

// ---------------------------------------------------------------------------
// collectConnectionStatus
// ---------------------------------------------------------------------------

func TestCollectConnectionStatus(t *testing.T) {
	c := &collector{cfg: newBaseConfig()}
	status, err := c.collectConnectionStatus()
	require.NoError(t, err)
	// This is a documented stub pending reporter integration.
	assert.Equal(t, "unknown", status.Status)
	assert.Equal(t, "feature requires reporter integration", status.FailureReason)
	assert.Nil(t, status.LastSuccess)
	assert.Nil(t, status.LastFailure)
	assert.Equal(t, 0, status.RetryCount)
}

// ---------------------------------------------------------------------------
// collectProbeTasks
// ---------------------------------------------------------------------------

func TestCollectProbeTasks_NoProbes(t *testing.T) {
	c := &collector{cfg: newBaseConfig()} // no probes configured
	tasks, err := c.collectProbeTasks()
	require.NoError(t, err)
	assert.Equal(t, 0, tasks.TotalTasks)
	assert.Empty(t, tasks.Tasks)
}

func TestCollectProbeTasks_WithProbes(t *testing.T) {
	c := &collector{cfg: newBaseConfig()}
	c.cfg.Probes = []config.ProbeConfig{
		{Type: "tcp", Target: "10.0.0.1", Port: 80},
		{Type: "udp", Target: "8.8.8.8", Port: 0},
	}
	tasks, err := c.collectProbeTasks()
	require.NoError(t, err)
	assert.Equal(t, 2, tasks.TotalTasks)
	require.Len(t, tasks.Tasks, 2)
	assert.Equal(t, "tcp", tasks.Tasks[0].Type)
	assert.Equal(t, "10.0.0.1:80", tasks.Tasks[0].Target)
	assert.Equal(t, "unknown", tasks.Tasks[0].Status)
	assert.Equal(t, "8.8.8.8", tasks.Tasks[1].Target, "port 0 -> no :0 suffix")
}

// ---------------------------------------------------------------------------
// collectResourceUsage (real gopsutil reads -> assert no error, sane ranges)
// ---------------------------------------------------------------------------

func TestCollectResourceUsage_NoError(t *testing.T) {
	c := &collector{cfg: newBaseConfig()}
	usage, err := c.collectResourceUsage()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, usage.CPUPercent, 0.0)
	assert.GreaterOrEqual(t, usage.MemoryMB, 0.0)
	assert.GreaterOrEqual(t, usage.MemoryPercent, 0.0)
}

// ---------------------------------------------------------------------------
// collectNetworkStatus (uses httptest.Server as a fake Pulse server)
// ---------------------------------------------------------------------------

func TestCollectNetworkStatus_ReachableServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network status test in short mode")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := &collector{cfg: newBaseConfig()}
	c.cfg.PulseServer = ts.URL

	status, err := c.collectNetworkStatus()
	require.NoError(t, err)
	assert.True(t, status.PulseServerReachable, "local test server should be reachable")
	assert.Equal(t, ts.URL, status.PulseServerAddress)
	assert.Greater(t, status.RTTMs.Samples, 0)
	assert.LessOrEqual(t, status.PacketLossRate, 0.5, "most pings should succeed against a live local server")
}

func TestCollectNetworkStatus_UnreachableServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network status test in short mode")
	}
	// Point at a closed port on loopback to force failures.
	c := &collector{cfg: newBaseConfig()}
	c.cfg.PulseServer = "http://127.0.0.1:1" // port 1 is reserved/unused

	status, err := c.collectNetworkStatus()
	require.NoError(t, err, "collectNetworkStatus itself should not error; it records failures")
	assert.False(t, status.PulseServerReachable)
	assert.Equal(t, 1.0, status.PacketLossRate, "all 5 pings fail -> 100%% loss")
	assert.Equal(t, 0, status.RTTMs.Samples)
	assert.NotEmpty(t, status.RecentFailures)
}

// ---------------------------------------------------------------------------
// Collect / CollectJSON / CollectPretty (integration via the Collector interface)
// ---------------------------------------------------------------------------

func TestCollector_CollectFull(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Collect test in short mode (performs network dials)")
	}
	initTestLogger(t)
	cfg := newBaseConfig()
	cfg.PulseServer = "http://127.0.0.1:1" // force unreachable so it is deterministic & fast-failing
	c := NewCollector(cfg)

	info, err := c.Collect()
	require.NoError(t, err)
	assert.Equal(t, "test-node-1", info.NodeID)
	assert.Equal(t, "Test Node", info.NodeName)
	assert.Equal(t, "DEBUG", info.Level)
	assert.NotEmpty(t, info.Timestamp)
}

func TestCollector_CollectJSON_RoundTrips(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CollectJSON test in short mode (performs network dials)")
	}
	initTestLogger(t)
	cfg := newBaseConfig()
	cfg.PulseServer = "http://127.0.0.1:1"
	c := NewCollector(cfg)

	data, err := c.CollectJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// The JSON must unmarshal back into DiagnosticInfo.
	var info DiagnosticInfo
	require.NoError(t, json.Unmarshal(data, &info))
	assert.Equal(t, "test-node-1", info.NodeID)
}

func TestCollector_CollectPretty_ContainsHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CollectPretty test in short mode (performs network dials)")
	}
	initTestLogger(t)
	cfg := newBaseConfig()
	cfg.PulseServer = "http://127.0.0.1:1"
	c := NewCollector(cfg)

	out, err := c.CollectPretty()
	require.NoError(t, err)
	assert.Contains(t, out, "Beacon Diagnostic Information")
	assert.Contains(t, out, "Node ID: test-node-1")
	assert.Contains(t, out, "test-node-1")
	// Mode status should render the default line when no mode provider is set.
	assert.Contains(t, out, "Mode: registered (default)")
}

func TestCollector_SetProviders(t *testing.T) {
	c := NewCollector(newBaseConfig())

	// Setters must not panic and must store the providers.
	mode := &mockModeProvider{mode: config.ModeRegistered, source: config.SourceLocal}
	cache := &mockCacheStatsProvider{size: 1, count: 1, evictions: 1}
	comp := &mockCompressionProvider{ratio: 1.0}
	probe := &mockProbeCountProvider{count: 1}

	assert.NotPanics(t, func() {
		c.SetModeProvider(mode)
		c.SetCacheStatsProvider(cache)
		c.SetCompressionStatsProvider(comp)
		c.SetProbeCountProvider(probe)
	})

	// Verify via enhanceMetrics that providers were wired.
	m := &PrometheusMetrics{}
	c.(*collector).enhanceMetrics(m)
	assert.Equal(t, string(config.ModeRegistered), m.Mode)
	assert.Equal(t, int64(1), m.CacheSizeBytes)
	assert.Equal(t, float64(1.0), m.CompressionRatio)
	assert.Equal(t, 1, m.ActiveProbes)
}

// ---------------------------------------------------------------------------
// NewCollector
// ---------------------------------------------------------------------------

func TestNewCollector(t *testing.T) {
	cfg := newBaseConfig()
	c := NewCollector(cfg)
	require.NotNil(t, c)

	// The underlying collector should carry the config and a recent start time.
	col := c.(*collector)
	assert.Same(t, cfg, col.cfg)
	assert.False(t, col.startTime.IsZero(), "startTime should be set")
	assert.True(t, time.Since(col.startTime) < 5*time.Second, "startTime should be recent")
}
