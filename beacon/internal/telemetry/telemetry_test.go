package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/whg517/node-pulse/beacon/internal/config"
	"github.com/whg517/node-pulse/beacon/internal/logger"
)

// initTestLogger bootstraps the global logger so telemetry Init/Shutdown log
// paths do not fail. Mirrors internal/metrics/metrics_additional_test.go.
func initTestLogger(t *testing.T) {
	t.Helper()
	if err := logger.InitLogger(&config.Config{
		LogLevel:      "INFO",
		LogFile:       "/tmp/test-telemetry.log",
		LogMaxSize:    10,
		LogMaxAge:     7,
		LogMaxBackups: 3,
		LogCompress:   false,
		LogToConsole:  false,
	}); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
}

// TestInit_DisabledInstallsNoopProvider verifies the disabled branch installs
// a no-op TracerProvider globally, returns a non-nil Provider, and performs no
// network I/O.
func TestInit_DisabledInstallsNoopProvider(t *testing.T) {
	initTestLogger(t)

	orig := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(orig) })

	prov, err := Init(context.Background(), Config{Enabled: false})
	require.NoError(t, err)
	require.NotNil(t, prov)
	assert.Nil(t, prov.tp, "disabled provider has no SDK tracer provider")

	// Global should now be a noop provider.
	_, ok := otel.GetTracerProvider().(noop.TracerProvider)
	assert.True(t, ok, "global TracerProvider should be the noop provider when disabled")

	// Shutdown must be safe (no-op) for the disabled provider.
	assert.NotPanics(t, func() { prov.Shutdown(context.Background()) })
}

// TestInit_EnabledWithStdoutExporter verifies the enabled path works without a
// real OTLP collector: an empty OTLPEndpoint uses the stdout exporter, so no
// network is required.
func TestInit_EnabledWithStdoutExporter(t *testing.T) {
	initTestLogger(t)

	orig := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(orig) })

	prov, err := Init(context.Background(), Config{
		Enabled:       true,
		NodeID:        "test-node",
		OTLPEndpoint:  "", // -> stdout exporter (no network)
		SamplingRate:  1.0,
	})
	require.NoError(t, err)
	require.NotNil(t, prov)
	require.NotNil(t, prov.tp, "enabled provider should hold an SDK tracer provider")

	// Cleanup: flush pending spans.
	prov.Shutdown(context.Background())
}

// TestApplyDefaults verifies default values are applied for empty fields and
// out-of-range sampling rates.
func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name             string
		input            Config
		wantServiceName  string
		wantVersion      string
		wantSamplingRate float64
	}{
		{
			name:             "all empty -> defaults",
			input:            Config{},
			wantServiceName:  "beacon",
			wantVersion:      "unknown",
			wantSamplingRate: 1.0,
		},
		{
			name:             "zero sampling rate clamped to 1.0",
			input:            Config{SamplingRate: 0},
			wantServiceName:  "beacon",
			wantVersion:      "unknown",
			wantSamplingRate: 1.0,
		},
		{
			name:             "negative sampling rate clamped to 1.0",
			input:            Config{SamplingRate: -0.5},
			wantServiceName:  "beacon",
			wantVersion:      "unknown",
			wantSamplingRate: 1.0,
		},
		{
			name:             "over-1 sampling rate clamped to 1.0",
			input:            Config{SamplingRate: 2.0},
			wantServiceName:  "beacon",
			wantVersion:      "unknown",
			wantSamplingRate: 1.0,
		},
		{
			name:             "valid values preserved",
			input:            Config{ServiceName: "my-beacon", ServiceVersion: "v1.2.3", SamplingRate: 0.5},
			wantServiceName:  "my-beacon",
			wantVersion:      "v1.2.3",
			wantSamplingRate: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.input
			applyDefaults(&cfg)
			assert.Equal(t, tt.wantServiceName, cfg.ServiceName)
			assert.Equal(t, tt.wantVersion, cfg.ServiceVersion)
			assert.Equal(t, tt.wantSamplingRate, cfg.SamplingRate)
		})
	}
}

// TestBuildSampler covers the three sampler branches plus boundaries.
func TestBuildSampler(t *testing.T) {
	tests := []struct {
		name string
		rate float64
		want string
	}{
		{"rate >= 1.0 -> AlwaysSample", 1.0, "AlwaysOnSampler"},
		{"rate > 1.0 -> AlwaysSample", 1.5, "AlwaysOnSampler"},
		{"rate <= 0.0 -> NeverSample", 0.0, "AlwaysOffSampler"},
		{"rate < 0.0 -> NeverSample", -1.0, "AlwaysOffSampler"},
		{"rate 0.5 -> TraceIDRatioBased", 0.5, "TraceIDRatioBased"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildSampler(tt.rate)
			assert.Contains(t, s.Description(), tt.want)
		})
	}
}

// TestShutdown_NilProviderIsSafe verifies Shutdown does not panic when the
// provider was created with a nil tp (disabled path).
func TestShutdown_NilProviderIsSafe(t *testing.T) {
	p := &Provider{} // tp == nil
	assert.NotPanics(t, func() { p.Shutdown(context.Background()) })
}

// TestTracerReturnsNonNil confirms Tracer returns a usable tracer.
func TestTracerReturnsNonNil(t *testing.T) {
	initTestLogger(t)
	orig := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(orig) })

	_, err := Init(context.Background(), Config{Enabled: false})
	require.NoError(t, err)

	tr := Tracer("test-instrument")
	assert.NotNil(t, tr)
}
