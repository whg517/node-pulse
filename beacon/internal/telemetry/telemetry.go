// Package telemetry provides OpenTelemetry SDK initialisation and lifecycle management
// for distributed tracing in the Node-Pulse Beacon agent.
//
// When enabled, the Beacon:
//   - Creates a TracerProvider backed by either an OTLP gRPC exporter or a stdout exporter.
//   - Wraps its outbound HTTP client with otelhttp so that every request to the Pulse server
//     carries a W3C "traceparent" header.  This enables end-to-end trace correlation from the
//     Beacon probe loop all the way through Pulse's API handlers.
package telemetry

import (
	"context"
	"fmt"
	"time"

	"beacon/internal/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds OpenTelemetry tracing configuration for the Beacon.
type Config struct {
	// Enabled controls whether tracing is active.
	Enabled bool

	// ServiceName is the logical name for this agent in traces (default: "beacon").
	ServiceName string

	// ServiceVersion is the deployed version string (default: "unknown").
	ServiceVersion string

	// NodeID is included as a resource attribute so traces can be filtered per node.
	NodeID string

	// OTLPEndpoint is the gRPC endpoint of an OTLP-compatible collector
	// (e.g. "localhost:4317").  When empty and Enabled is true, spans are
	// written to stdout (useful for dev/debug).
	OTLPEndpoint string

	// SamplingRate controls the fraction of operations that are traced (0.0 – 1.0).
	SamplingRate float64
}

// Provider wraps the SDK TracerProvider and owns its lifecycle.
type Provider struct {
	tp  *sdktrace.TracerProvider
	cfg Config
}

// Init initialises the global OpenTelemetry TracerProvider for the Beacon.
// Call Shutdown when the process is about to exit.
//
// When cfg.Enabled is false a no-op provider is installed so all tracing
// calls remain safe but produce no overhead.
func Init(ctx context.Context, cfg Config) (*Provider, error) {
	if !cfg.Enabled {
		logger.Info("Tracing disabled – using no-op provider")
		otel.SetTracerProvider(noop.NewTracerProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		return &Provider{cfg: cfg}, nil
	}

	applyDefaults(&cfg)

	res, err := buildResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("telemetry: build resource: %w", err)
	}

	exporter, err := buildExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("telemetry: build exporter: %w", err)
	}

	sampler := buildSampler(cfg.SamplingRate)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(256),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Infof("Tracing enabled: service=%s version=%s node=%s sampler=%.2f",
		cfg.ServiceName, cfg.ServiceVersion, cfg.NodeID, cfg.SamplingRate)

	return &Provider{tp: tp, cfg: cfg}, nil
}

// Shutdown flushes pending spans and shuts down the exporter.
// A 10-second timeout is applied automatically.
func (p *Provider) Shutdown(ctx context.Context) {
	if p.tp == nil {
		return
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := p.tp.Shutdown(shutdownCtx); err != nil {
		logger.Warnf("Telemetry shutdown error: %v", err)
	} else {
		logger.Info("Telemetry tracer provider shut down")
	}
}

// Tracer returns a named tracer from the global provider.
func Tracer(instrumentationName string) trace.Tracer {
	return otel.Tracer(instrumentationName)
}

func applyDefaults(cfg *Config) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "beacon"
	}
	if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "unknown"
	}
	if cfg.SamplingRate <= 0 || cfg.SamplingRate > 1 {
		cfg.SamplingRate = 1.0
	}
}

func buildResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
	}
	if cfg.NodeID != "" {
		attrs = append(attrs, semconv.HostName(cfg.NodeID))
	}
	return resource.New(ctx,
		resource.WithAttributes(attrs...),
		resource.WithHost(),
		resource.WithProcess(),
	)
}

func buildExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	if cfg.OTLPEndpoint != "" {
		logger.Infof("Connecting to OTLP collector at %s", cfg.OTLPEndpoint)
		conn, err := grpc.NewClient(cfg.OTLPEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("dial OTLP endpoint %s: %w", cfg.OTLPEndpoint, err)
		}
		return otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	}

	logger.Info("No OTLP endpoint configured – exporting traces to stdout")
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

func buildSampler(rate float64) sdktrace.Sampler {
	switch {
	case rate >= 1.0:
		return sdktrace.AlwaysSample()
	case rate <= 0.0:
		return sdktrace.NeverSample()
	default:
		return sdktrace.TraceIDRatioBased(rate)
	}
}
