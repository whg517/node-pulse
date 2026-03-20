package telemetry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/whg517/node-pulse/pulse/pkg/telemetry"
)

func TestInit_Disabled(t *testing.T) {
	cfg := telemetry.Config{Enabled: false}
	p, err := telemetry.Init(context.Background(), cfg)
	require.NoError(t, err)
	assert.NotNil(t, p)

	// No-op provider must not panic when used
	tracer := otel.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()
	assert.NotNil(t, ctx)
}

func TestInit_Enabled_StdoutExporter(t *testing.T) {
	// OTLPEndpoint empty → stdout exporter (no external service needed)
	cfg := telemetry.Config{
		Enabled:        true,
		ServiceName:    "pulse-test",
		ServiceVersion: "0.0.1",
		Environment:    "test",
		SamplingRate:   1.0,
		// OTLPEndpoint intentionally empty to use stdout
	}
	p, err := telemetry.Init(context.Background(), cfg)
	require.NoError(t, err)
	require.NotNil(t, p)
	defer p.Shutdown(context.Background())

	// Global tracer provider should now be installed
	tracer := otel.Tracer("pulse/api")
	ctx, span := tracer.Start(context.Background(), "test-request")
	defer span.End()
	assert.NotNil(t, ctx)
}

func TestInit_SamplingDefaults(t *testing.T) {
	tests := []struct {
		name         string
		samplingRate float64
	}{
		{"always sample", 1.0},
		{"never sample", 0.0},
		{"partial sample", 0.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := telemetry.Config{
				Enabled:      true,
				SamplingRate: tc.samplingRate,
			}
			p, err := telemetry.Init(context.Background(), cfg)
			require.NoError(t, err)
			require.NotNil(t, p)
			p.Shutdown(context.Background())
		})
	}
}

func TestTracer(t *testing.T) {
	tracer := telemetry.Tracer("pulse/test")
	assert.NotNil(t, tracer)
}
