# Node-Pulse Observability Design

This document describes the observability strategy for Node-Pulse, covering the three pillars of observability—**Metrics**, **Logs**, and **Traces**—and explains how they are implemented across the Pulse API server and Beacon monitoring agent.

## Table of Contents

1. [Overview](#overview)
2. [The Three Pillars](#the-three-pillars)
3. [Metrics](#metrics)
4. [Logging](#logging)
5. [Distributed Tracing](#distributed-tracing)
6. [End-to-End Correlation](#end-to-end-correlation)
7. [Configuration Reference](#configuration-reference)
8. [Deployment Integration](#deployment-integration)
9. [Development Quickstart](#development-quickstart)

---

## Overview

Observability is the ability to understand what is happening inside a system by examining its external outputs. Node-Pulse is a distributed monitoring system; applying observability to itself ensures that operators can detect performance regressions, trace errors, and understand the health of the platform without needing to redeploy or add new instrumentation ad hoc.

### Architecture at a Glance

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Observability Stack                           │
├──────────────┬──────────────────────┬───────────────────────────────┤
│   Pillar     │  Implementation      │  Collection Target             │
├──────────────┼──────────────────────┼───────────────────────────────┤
│ Metrics      │ Prometheus           │ Grafana / Alertmanager         │
│ Logs         │ logrus / stdlib log  │ Loki / ELK / CloudWatch        │
│ Traces       │ OpenTelemetry SDK    │ Jaeger / Grafana Tempo / OTLP  │
└──────────────┴──────────────────────┴───────────────────────────────┘
```

---

## The Three Pillars

### 1. Metrics — *What is the system doing right now?*

Metrics are numerical measurements aggregated over time. They answer questions like "how many requests per second?", "what is the P99 latency?", "how many beacons are connected?".

### 2. Logs — *What happened?*

Logs are time-stamped records of discrete events. They answer "what error occurred at 14:03:22?", "which user triggered this alert?".

### 3. Traces — *Why did it take so long?*

Traces follow a single request through a distributed system, showing how latency is distributed across services and database queries.

---

## Metrics

### Pulse Server Metrics

Prometheus metrics are exposed at `GET /metrics` (no authentication required). They are collected and updated by:

| Metric | Type | Description |
|--------|------|-------------|
| `pulse_server_memory_usage_bytes` | Gauge | Process RSS memory |
| `pulse_server_heap_alloc_bytes` | Gauge | Go heap allocation |
| `pulse_server_goroutines` | Gauge | Active goroutines |
| `pulse_server_cpu_usage_percent` | Gauge | GC CPU fraction (approximation) |
| `pulse_api_requests_total` | Counter | HTTP requests by endpoint + status |
| `pulse_api_response_time_seconds` | Histogram | HTTP response time per endpoint |
| `pulse_api_active_connections` | Gauge | Current open connections |
| `pulse_beacons_connected` | Gauge | Connected beacon agents |
| `pulse_beacons_disconnected_total` | Counter | Total beacon disconnections |
| `pulse_webhook_queue_depth` | Gauge | Pending webhook deliveries |
| `pulse_webhook_delivery_success_total` | Counter | Successful webhook deliveries |
| `pulse_webhook_delivery_failed_total` | Counter | Failed webhook deliveries |
| `pulse_compression_corruption_total` | Counter | Compressed payload corruption events |

Auth subsystem metrics (see `pulse/internal/auth/metrics.go`):

| Metric | Type | Description |
|--------|------|-------------|
| `pulse_auth_login_attempts_total` | Counter | Login attempts by result |
| `pulse_auth_refresh_rotations_total` | Counter | Token refresh rotations |
| `pulse_auth_blacklist_checks_total` | Counter | Token blacklist lookups |
| `pulse_auth_api_key_exchanges_total` | Counter | API key → JWT exchanges |

### Beacon Metrics

Exposed at `GET /metrics` on port `2112` (configurable via `metrics_port`):

| Metric | Type | Description |
|--------|------|-------------|
| `beacon_up` | GaugeVec | Probe availability (1=up, 0=down) |
| `beacon_rtt_seconds` | Gauge | Round-trip time in seconds |
| `beacon_packet_loss_rate` | Gauge | Packet loss rate (0–1) |
| `beacon_jitter_ms` | Gauge | RTT jitter in milliseconds |
| `beacon_active_probes` | Gauge | Number of active probe tasks |
| `beacon_compression_ratio` | Gauge | Payload compression ratio |
| `beacon_cache_size_bytes` | Gauge | Resume-upload cache size |
| `beacon_cache_evictions_total` | Counter | Cache evictions |
| `beacon_probe_duration_seconds` | Histogram | Probe execution time |
| `beacon_probe_failures_total` | CounterVec | Probe failures by type |
| `beacon_memory_usage_bytes` | Gauge | Agent RSS memory |
| `beacon_cpu_usage_percent` | Gauge | Agent CPU usage |

### Recommended Prometheus Scrape Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: pulse
    static_configs:
      - targets: ["pulse:6532"]

  - job_name: beacon
    static_configs:
      - targets: ["beacon-host:2112"]
```

### Recommended Alerting Rules

```yaml
# pulse-alerts.yml
groups:
  - name: pulse
    rules:
      - alert: PulseHighLatency
        expr: histogram_quantile(0.99, rate(pulse_api_response_time_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Pulse API P99 latency > 1 s"

      - alert: BeaconDown
        expr: beacon_up == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Beacon probe target is unreachable"
```

---

## Logging

### Pulse

Pulse uses the Go standard library `log` package with structured prefixes. Log level and format are configured in `pulse.yaml`:

```yaml
log:
  level: info      # debug | info | warn | error
  format: text     # text | json
  output: stdout
```

Set `format: json` in production to enable log aggregation with Loki, Fluentd, or CloudWatch Logs Insights.

**Trace ID injection**: when distributed tracing is enabled, `pulse/pkg/middleware.TraceIDMiddleware` prepends the active OTel trace ID and span ID to every log line during the request's lifetime:

```
[trace=4bf92f3577b34da6a3ce929d0e0e4736 span=00f067aa0ba902b7] [INFO] ...
```

This allows correlating a log line directly to its trace in Jaeger / Tempo.

### Beacon

Beacon uses [logrus](https://github.com/sirupsen/logrus) with rotating file support via lumberjack. Configuration in `beacon.yaml`:

```yaml
log_level: INFO        # DEBUG | INFO | WARN | ERROR
log_file: /var/log/beacon/beacon.log
log_max_size: 100      # MB
log_max_age: 7         # days
log_max_backups: 5
log_compress: true
log_to_console: false
```

---

## Distributed Tracing

### Technology Choice: OpenTelemetry

[OpenTelemetry](https://opentelemetry.io/) (OTel) is the CNCF standard for vendor-neutral distributed tracing. Choosing OTel means:

- **Vendor-agnostic**: export to Jaeger, Grafana Tempo, Zipkin, AWS X-Ray, Google Cloud Trace, Datadog — without changing application code.
- **W3C TraceContext**: spans carry `traceparent` / `tracestate` HTTP headers defined in [W3C Trace Context](https://www.w3.org/TR/trace-context/), enabling interoperability with any OTel-instrumented service.
- **Future-proof**: OpenMetrics + OTLP metrics export is on the OTel roadmap, eventually unifying all three pillars under one SDK.

### Implementation

#### Pulse (`pulse/pkg/telemetry`)

```
Init(ctx, Config) → *Provider
  ├── builds SDK TracerProvider
  │     ├── OTLP gRPC exporter  (when OTLPEndpoint is set)
  │     └── stdout exporter     (fallback for local dev)
  ├── installs W3C TraceContext + Baggage propagators globally
  └── configures sampler (AlwaysSample | TraceIDRatioBased | NeverSample)

Provider.Shutdown(ctx)          # flushes pending spans on exit
```

HTTP request tracing is provided by `otelgin` middleware in `pulse/pkg/middleware`:

```
OtelGinMiddleware("pulse")    # creates a span per HTTP request
TraceIDMiddleware()           # injects trace ID into X-Trace-Id response header + log prefix
```

#### Beacon (`beacon/internal/telemetry`)

Same SDK architecture as Pulse. The Beacon's HTTP client transport is wrapped with `otelhttp.NewTransport(...)` in `NewPulseAPIClient`, so every heartbeat request to Pulse:

1. Creates a child span under the active Beacon span context.
2. Injects `traceparent` and `tracestate` headers.
3. Pulse reads those headers and continues the trace as a child span.

#### Trace Flow

```
Beacon probe loop
   │
   ├─► [beacon/heartbeat] span
   │       ├── otelhttp Transport injects traceparent header
   │       │
   │       └─► Pulse HTTP server
   │               │
   │               ├─► [POST /api/v1/beacon/heartbeat] span  (otelgin)
   │               │       ├── JWT validation
   │               │       ├── cache write
   │               │       └── batch writer enqueue
   │               │
   │               └─► exported to OTLP collector
   │
   └─► exported to OTLP collector
```

### Sampler Strategy

| Environment | Recommended Rate | Rationale |
|-------------|-----------------|-----------|
| Development | 1.0 (always) | Full visibility during debugging |
| Staging | 1.0 | Verify tracing before production |
| Production (low traffic) | 1.0 | Complete picture, low overhead |
| Production (high traffic) | 0.05–0.1 | Reduce storage and CPU cost |

Configure via `sampling_rate` in `pulse.yaml` / `beacon.yaml`, or via the env-var `PULSE_TELEMETRY_SAMPLING_RATE`.

### X-Trace-Id Header

Every response from Pulse includes `X-Trace-Id: <traceID>` when tracing is enabled. Front-end clients can log this value alongside browser network errors, allowing support engineers to look up the exact backend trace immediately.

---

## End-to-End Correlation

With all three pillars active, an operator investigating a slow heartbeat can:

1. **Metrics**: notice `pulse_api_response_time_seconds{endpoint="/api/v1/beacon/heartbeat"}` P99 spike in Grafana.
2. **Trace**: open Jaeger and search for traces on the `/api/v1/beacon/heartbeat` endpoint in the same time window — find the outlier trace.
3. **Log**: copy the `traceID` from the trace, search Loki/Kibana for `trace=<id>` — see the exact error log line with stack trace.

This tight loop between metrics → traces → logs eliminates the need to reproduce incidents.

---

## Configuration Reference

### Pulse (`pulse.yaml`)

```yaml
telemetry:
  enabled: false                  # true to activate tracing
  service_name: pulse             # reported in traces
  service_version: unknown        # set to Git tag in CI/CD
  environment: development        # "development" | "staging" | "production"
  otlp_endpoint: ""               # gRPC OTLP collector (leave empty for stdout)
  sampling_rate: 1.0              # 0.0 – 1.0
```

Environment variable overrides (prefix `PULSE_TELEMETRY_`):

| Variable | Default |
|----------|---------|
| `PULSE_TELEMETRY_ENABLED` | `false` |
| `PULSE_TELEMETRY_SERVICE_NAME` | `pulse` |
| `PULSE_TELEMETRY_SERVICE_VERSION` | `unknown` |
| `PULSE_TELEMETRY_ENVIRONMENT` | `development` |
| `PULSE_TELEMETRY_OTLP_ENDPOINT` | `` |
| `PULSE_TELEMETRY_SAMPLING_RATE` | `1.0` |

### Beacon (`beacon.yaml`)

```yaml
telemetry:
  enabled: false
  service_name: beacon
  service_version: unknown
  otlp_endpoint: ""
  sampling_rate: 1.0
```

---

## Deployment Integration

### OpenTelemetry Collector (recommended)

Running an [OTel Collector](https://opentelemetry.io/docs/collector/) sidecar provides buffering, retry, and fan-out to multiple backends:

```yaml
# docker-compose snippet
services:
  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    ports:
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
    volumes:
      - ./otel-collector.yaml:/etc/otelcol-contrib/config.yaml

  pulse:
    environment:
      PULSE_TELEMETRY_ENABLED: "true"
      PULSE_TELEMETRY_OTLP_ENDPOINT: "otel-collector:4317"
      PULSE_TELEMETRY_ENVIRONMENT: "production"
      PULSE_TELEMETRY_SAMPLING_RATE: "0.1"
```

### Jaeger All-in-One (development)

```bash
docker run -d --name jaeger \
  -p 4317:4317 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest

# Then enable tracing in pulse:
export PULSE_TELEMETRY_ENABLED=true
export PULSE_TELEMETRY_OTLP_ENDPOINT=localhost:4317
```

Open Jaeger UI at `http://localhost:16686`.

### Grafana Tempo

```yaml
# docker-compose snippet
services:
  tempo:
    image: grafana/tempo:latest
    ports:
      - "4317:4317"
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
```

---

## Development Quickstart

**Enable stdout tracing (no external services needed):**

```bash
# Pulse
export PULSE_TELEMETRY_ENABLED=true
# No PULSE_TELEMETRY_OTLP_ENDPOINT → spans printed to stdout

# Beacon
# In beacon.yaml:
# telemetry:
#   enabled: true
```

**Verify traces appear:**

```
{
  "Name": "POST /api/v1/beacon/heartbeat",
  "SpanContext": {
    "TraceID": "4bf92f3577b34da6a3ce929d0e0e4736",
    "SpanID": "00f067aa0ba902b7",
    ...
  },
  "Duration": "2.345ms",
  ...
}
```

**Check the response header from any API call:**

```bash
curl -si http://localhost:6532/api/v1/health | grep X-Trace-Id
# X-Trace-Id: 4bf92f3577b34da6a3ce929d0e0e4736
```
