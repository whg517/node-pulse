// Package telemetry provides OpenTelemetry SDK initialization and lifecycle management
// for distributed tracing in the Node-Pulse Pulse server.
//
// The three pillars of observability that this package contributes to:
//   - Traces  – distributed request tracing via OpenTelemetry SDK + OTLP exporter
//   - Metrics – already covered by Prometheus (pulse/pkg/metrics)
//   - Logs    – already covered by standard log package; trace IDs are injected into
//               log lines via the TraceIDMiddleware (pulse/pkg/middleware)
//
// Usage:
//
//	cfg := telemetry.Config{Enabled: true, OTLPEndpoint: "localhost:4317", ServiceName: "pulse"}
//	tp, err := telemetry.Init(ctx, cfg)
//	if err != nil { ... }
//	defer tp.Shutdown(ctx)
package telemetry

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
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

// Config holds OpenTelemetry configuration.
type Config struct {
	// Enabled controls whether tracing is active. When false a no-op tracer is used.
	Enabled bool `yaml:"enabled"`

	// ServiceName is the logical name of this service (default: "pulse").
	ServiceName string `yaml:"service_name"`

	// ServiceVersion is the deployed version string (default: "unknown").
	ServiceVersion string `yaml:"service_version"`

	// Environment is the deployment environment, e.g. "production", "staging", "development".
	Environment string `yaml:"environment"`

	// OTLPEndpoint is the gRPC endpoint of an OTLP-compatible collector
	// (e.g. "localhost:4317" for a local Jaeger / Grafana Tempo / OTel Collector).
	// When empty and Enabled is true, traces are written to stdout (useful for dev/debug).
	OTLPEndpoint string `yaml:"otlp_endpoint"`

	// SamplingRate controls the fraction of requests that are traced (0.0 – 1.0).
	// Defaults to 1.0 (always sample). Set to lower values in high-traffic production.
	SamplingRate float64 `yaml:"sampling_rate"`
}

// Provider wraps the SDK TracerProvider and owns its lifecycle.
type Provider struct {
	tp  *sdktrace.TracerProvider
	cfg Config
}

// Init initialises the global OpenTelemetry TracerProvider and TextMapPropagator.
// Call Shutdown on the returned Provider when the process is about to exit.
//
// When cfg.Enabled is false, a no-op provider is installed so all tracing calls
// are safe to use but produce no overhead or output.
func Init(ctx context.Context, cfg Config) (*Provider, error) {
	if !cfg.Enabled {
		log.Println("[INFO] [Telemetry] Tracing disabled – using no-op provider")
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
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Install the global provider and W3C TraceContext + Baggage propagators so
	// trace context is automatically propagated across process boundaries via HTTP
	// headers (traceparent / tracestate / baggage).
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Printf("[INFO] [Telemetry] Tracing enabled: service=%s version=%s env=%s sampler=%.2f",
		cfg.ServiceName, cfg.ServiceVersion, cfg.Environment, cfg.SamplingRate)

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
		log.Printf("[WARN] [Telemetry] Shutdown error: %v", err)
	} else {
		log.Println("[INFO] [Telemetry] Tracer provider shut down")
	}
}

// Tracer returns a named tracer from the global provider.
// serviceName should match the instrumentation library name, e.g. "pulse/api".
func Tracer(instrumentationName string) trace.Tracer {
	return otel.Tracer(instrumentationName)
}

// applyDefaults fills in zero-value config fields.
func applyDefaults(cfg *Config) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "pulse"
	}
	if cfg.ServiceVersion == "" {
		cfg.ServiceVersion = "unknown"
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.SamplingRate <= 0 || cfg.SamplingRate > 1 {
		cfg.SamplingRate = 1.0
	}
}

// buildResource creates an OTel resource describing this service instance.
func buildResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
		resource.WithHost(),
		resource.WithProcess(),
	)
}

// buildExporter creates either an OTLP gRPC exporter (when OTLPEndpoint is set)
// or a stdout exporter for local development/debugging.
func buildExporter(ctx context.Context, cfg Config) (sdktrace.SpanExporter, error) {
	if cfg.OTLPEndpoint != "" {
		log.Printf("[INFO] [Telemetry] Connecting to OTLP collector at %s", cfg.OTLPEndpoint)
		conn, err := grpc.NewClient(cfg.OTLPEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, fmt.Errorf("dial OTLP endpoint %s: %w", cfg.OTLPEndpoint, err)
		}
		return otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	}

	// Fallback: pretty-print spans to stdout (handy for local dev)
	log.Println("[INFO] [Telemetry] No OTLP endpoint – exporting traces to stdout")
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

// buildSampler selects the trace sampler based on the sampling rate.
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
