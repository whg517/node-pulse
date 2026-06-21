# Product Requirements Document - NodePulse

**Owner:** Kevin
**Date:** 2026-06-21
**Version:** 4.0
**Status:** Current implementation baseline and next-phase product design

## 1. Product Context

NodePulse is a distributed network monitoring system for overseas infrastructure. Beacon agents run on monitored nodes, collect network metrics and route traces, and report to the Pulse platform. Pulse aggregates data, evaluates alert rules, serves operational APIs, and provides a web UI for monitoring, diagnosis, reporting, and configuration.

The product goal is not only to show whether a node is online. It should help operators answer three questions quickly:

- Is the problem local to one node, regional, or cross-border?
- Which route hop or network segment is likely responsible?
- What should the operator do next, and how can the result be shared?

## 2. Requirement Status Model

This PRD uses explicit status labels so product scope, roadmap, and code stay aligned.

- **Supported:** Implemented in the current codebase and usable through an API, Beacon behavior, or UI path.
- **Partially supported:** Core pieces exist, but the end-to-end workflow is incomplete or not production-ready.
- **Planned:** Desired in the next product phase but not yet implemented.
- **Deferred:** Deliberately outside the current roadmap.

## 3. Current Product Scope

### 3.1 Supported Capabilities

- User authentication, sessions, refresh-token rotation, RBAC, admin user management, and API key management.
- Beacon authentication by API key exchange for JWT.
- Beacon standalone and registered modes.
- Local Beacon probe configuration with hot reload.
- Server-managed Beacon probe configuration with version history and config acknowledgements.
- TCP, UDP, and MTR probe execution in Beacon.
- Heartbeat reporting from Beacon to Pulse.
- MTR result upload, persistence, latest-route query, and history query.
- Real-time metrics query and historical metrics query.
- Multi-node comparison and rule-based diagnosis API.
- Dashboard overview with metrics cards, world map, top anomalies, trend charts, and alert stream UI.
- Node management, node detail, route visualization, and comparison UI.
- Alert rules and alert records.
- Webhook CRUD, webhook payload rendering preview, retrying webhook delivery, and delivery log persistence.
- Data export as CSV.
- Report preview/print workflow using live metrics, latest MTR, alert context, and diagnosis output.
- System health page for database, scheduler, alert engine, webhook delivery, and suppression health.
- English and Simplified Chinese UI foundation with timezone preferences.
- Pulse Prometheus metrics and Beacon Prometheus metrics.
- OpenTelemetry tracing integration for Pulse and Beacon HTTP paths.

### 3.2 Partially Supported Capabilities

- **Cross-border transport optimization:** Compression, CRC, and priority cache utilities exist, and Pulse exposes a compressed heartbeat endpoint, but Beacon does not yet use compression/resume upload in its runtime heartbeat path.
- **Real-time alert push:** Frontend WebSocket and browser notification services exist, but Pulse does not expose a `/ws` or equivalent event stream endpoint.
- **Mobile alert handling:** A mobile alert detail component exists, but backend alert note routes are missing and status updates currently ignore note content.
- **Webhook operations:** Delivery retries exist, but manual test delivery, endpoint health state, queue depth enforcement, delivery history UI, success-rate UI, timeout metrics, and unhealthy recovery checks are incomplete.
- **Scheduled reports:** The Reports page can create schedules in local frontend state, but there is no server-side schedule persistence, execution, PDF generation job, or email delivery.
- **Excel export:** The UI acknowledges Excel as unavailable and backend export supports CSV only.
- **Configuration rollback:** Beacon config history can be viewed, but rollback-to-version is not implemented.
- **Internationalization:** The main framework is present, but some operational strings remain hardcoded.
- **Accessibility:** UI components generally use labels and semantic primitives, but WCAG 2.1 AA has not been fully audited or enforced.

### 3.3 Planned Capabilities

- End-to-end compressed and resumable Beacon upload.
- Pulse alert event streaming for dashboard and browser notifications.
- Alert notes and status timeline persistence.
- Webhook test send, health state, delivery statistics, and history view.
- Server-side report schedules with email delivery.
- XLSX export.
- Config rollback.
- Production hardening for TLS/mTLS defaults, health endpoints, backup, and load-test gates.
- Better diagnostic recommendations and owner attribution.

### 3.4 Deferred Capabilities

- Iperf3 throughput testing and MTU path discovery.
- AI-driven diagnosis.
- Multi-region Pulse deployment.
- Public SEO optimization.

## 4. Primary User Journeys

### 4.1 Operations Dashboard

An operator opens the dashboard, reviews global node health, checks the top anomaly list, drills into a degraded node, reviews recent metrics and MTR path data, and determines whether the issue is local, regional, or route-related.

**Success criteria:**

- Dashboard loads the current node and metric picture within 3 seconds in a normal deployment.
- Operator can reach node detail from dashboard map, table, or anomaly list.
- Node detail shows metrics history, diagnosis, and latest MTR data when available.

### 4.2 Alert Response

An on-call operator sees a new alert, opens the alert or node detail, updates status, adds an investigation note, and receives or shares subsequent recovery context.

**Success criteria:**

- New alert is visible in the dashboard stream within 5 seconds after Pulse records it.
- Alert status can move through pending, in progress, and resolved states.
- Notes are persisted with author and UTC timestamp.
- Mobile layout supports status update and note-taking.

### 4.3 Beacon Deployment And Configuration

A DevOps engineer creates or selects a node, configures probes, starts Beacon in standalone or registered mode, and verifies that metrics and config acknowledgements are visible.

**Success criteria:**

- Standalone Beacon runs local probes and exposes `/metrics` without Pulse authentication.
- Registered Beacon authenticates, pulls server-managed probe config, applies updates, and acknowledges the active version.
- Invalid probe config is rejected before it can degrade a Beacon.

### 4.4 Network Diagnosis And Reporting

A network specialist compares nodes across regions, uses MTR and metric history to identify likely ownership, and generates a report for handoff.

**Success criteria:**

- Multi-node comparison supports latency, packet loss, and jitter.
- Diagnosis output includes problem type, confidence, and recommendation.
- Report preview includes node info, key metrics, MTR path, baseline comparison, root cause summary, and event timeline.

### 4.5 Integrations

An operator configures webhooks, previews payloads, sends a test delivery, monitors success/failure, and relies on retry behavior during incidents.

**Success criteria:**

- Webhook URLs are validated and protected against SSRF.
- Operators can preview and test payloads before enabling a webhook.
- Delivery attempts are logged with status, retry count, and error.
- Unhealthy endpoints are visible and retried later.

## 5. Functional Requirements

### FR-1 Beacon Runtime

**Supported**

- Beacon must load `beacon.yaml` and validate node identity, Pulse server, API key requirements, metrics settings, logging settings, and probe configuration.
- Beacon must support standalone and registered modes.
- Beacon must support TCP, UDP, and MTR probes.
- Beacon must expose Prometheus metrics when enabled.
- Registered Beacon must authenticate with Pulse using an API key and JWT.
- Registered Beacon must report aggregated heartbeat metrics.
- Registered Beacon must fetch server-managed config periodically and acknowledge applied or failed versions.
- Beacon must upload MTR results to Pulse when MTR probes are configured.

**Planned**

- Beacon must expose `/healthz` with self-health status.
- Beacon must set mode/config-source metrics from real runtime state rather than defaults.
- Beacon must send compressed payloads when enabled.
- Beacon must persist failed uploads locally and resume them in order when Pulse is reachable.
- Beacon must expose cache, compression, corruption, and resume-upload metrics from active runtime providers.

### FR-2 Pulse Data And Diagnosis

**Supported**

- Pulse must persist nodes, metrics, alert rules, alert records, MTR results, users, sessions, API keys, exports, webhook configs, and webhook logs.
- Pulse must provide JSON APIs for latest metrics, history, comparison, diagnosis, latest MTR, and MTR history.
- Pulse must evaluate alert rules asynchronously from incoming heartbeat data.
- Pulse must run scheduled cleanup for metrics and alert suppression data.

**Planned**

- Pulse must expose event streaming for new alerts and status updates.
- Pulse must support alert notes with author, content, and timestamps.
- Pulse must support route-change cache freshness indicators when enough route history exists.
- Pulse must provide explicit diagnostic owner attribution: local node, regional link, carrier route, cross-border link, target service, or unknown.

### FR-3 Web UI

**Supported**

- The UI must provide authenticated operational pages for dashboard, nodes, probes, Beacon config, alerts, reports, exports, webhooks, system health, preferences, sessions, and users.
- The UI must use shared layout components, i18n, timezone preferences, and dark-mode compatible styling.
- The dashboard must include global map, summary metrics, trends, top anomalies, node table, and alert stream UI.
- Node detail must show metrics, diagnosis, MTR path visualization, and report navigation.
- Beacon config UI must support editing, validation, history view, and local templates.
- Reports UI must provide PDF preview/print and CSV export workflows.

**Planned**

- Replace hardcoded operational strings with locale keys.
- Add accessible text alternatives for chart-heavy views.
- Add alert notes/timeline UI backed by API persistence.
- Connect alert stream to Pulse event streaming.
- Add webhook delivery history/statistics and test-send UI.
- Add config rollback UI.

### FR-4 Alerts And Integrations

**Supported**

- Operators must manage alert rules for latency, packet loss, and jitter.
- Pulse must create alert records when rules are triggered.
- Operators must update alert record status.
- Pulse must send alert webhooks to enabled endpoints with retries.
- Operators must create, edit, delete, and preview webhook payloads.

**Planned**

- Alert record status update must support optional note content.
- Operators must add and view alert notes independently of status changes.
- Webhook configs must support custom headers and severity filters.
- Operators must send manual test payloads and see HTTP status, latency, and response summary.
- Pulse must mark repeatedly failing webhooks unhealthy and retry health checks later.

### FR-5 Reports And Exports

**Supported**

- Operators must export metrics as CSV for up to 50 nodes and up to 7 days.
- Operators must view export status and download completed CSV files.
- Operators must generate printable report previews from current frontend data and Pulse APIs.

**Planned**

- Pulse must generate XLSX exports.
- Pulse must persist report schedules and run daily, weekly, or monthly jobs.
- Pulse must send scheduled reports by email.
- Reports must include concise recommended actions and likely owner.

### FR-6 Administration And Security

**Supported**

- Pulse must support username/password login, refresh-token rotation, logout, session listing, session revocation, role-based permissions, and admin user management.
- Pulse must support API key lifecycle for Beacon authentication.
- Pulse must apply CSRF protection to selected mutation endpoints.
- Pulse must validate webhook URLs against SSRF rules before delivery.

**Planned**

- Production mode must default mTLS to strict unless explicitly disabled.
- Password reset must deliver reset emails through a configured provider.
- Mutation endpoints should consistently apply CSRF where browser-authenticated flows can reach them.
- Metrics endpoints must have documented production protection options.

## 6. Non-Functional Requirements

### NFR-1 Performance

- Pulse should support at least 50 concurrent Beacons reporting without sustained degradation.
- Pulse API P99 should be no more than 500 ms for heartbeat and ordinary operational queries under the target load.
- Dashboard data queries should normally return within 300 ms server time.
- Single-node history query should return within 1 second for the supported time window.
- Compression/decompression overhead should be no more than 50 ms per 1 MB after the compressed-upload path is active.

### NFR-2 Reliability

- Pulse should target 99.9% monthly availability for a single-region deployment.
- Heartbeat ingestion should avoid blocking on durable writes where possible.
- Beacon should keep probing in degraded network conditions and resume upload once Pulse is reachable.
- Pulse should gracefully shut down background workers, schedulers, exporters, and database connections.

### NFR-3 Security

- Production traffic should use TLS 1.2 or newer.
- Beacon-to-Pulse traffic should support mTLS and production defaults should be strict.
- API keys and refresh tokens must be stored hashed.
- Access tokens must not be stored in localStorage.
- Webhook delivery must reject private, loopback, link-local, and otherwise unsafe URLs.
- Password reset responses must avoid user enumeration.

### NFR-4 Observability

- Pulse must expose Prometheus metrics for API requests, response time, active connections, runtime resources, Beacon connectivity, webhook delivery, and auth events.
- Beacon must expose Prometheus metrics for availability, RTT, packet loss, jitter, active probes, probe counts, probe duration, failures, resource usage, and runtime mode.
- Pulse and Beacon should propagate W3C trace context on Beacon-to-Pulse HTTP calls.
- Health endpoints should distinguish healthy, degraded, and unhealthy states.

### NFR-5 Accessibility And Internationalization

- The UI should meet WCAG 2.1 AA for forms, dialogs, navigation, tables, and critical workflows.
- The UI must support English and Simplified Chinese.
- The UI must support at least 20 common IANA timezones.
- Timestamps in incident workflows should be available in local time and UTC.

### NFR-6 Data Retention And Backup

- Metrics retention defaults to 7 days.
- Authentication audit log retention target is 90 days.
- Export files are temporary artifacts and should be cleaned automatically.
- Configuration data should be backed up daily with a 30-day retention target in a later production-hardening phase.

## 7. Release Roadmap Baseline

The active roadmap is maintained in `docs/iteration-roadmap.md`. This PRD defines product intent and requirement status; the roadmap defines implementation order.

Immediate roadmap themes:

1. Make the current product safe and internally consistent.
2. Close partially supported workflows that already have UI/API fragments.
3. Add production-hardening gates for security, health, and performance.
4. Expand exports, reports, and integration workflows.

## 8. Version History

| Version | Date | Change |
| --- | --- | --- |
| 4.0 | 2026-06-21 | Rebuilt PRD around current implementation baseline, explicit status labels, and next-phase requirements. |
| 3.8 | 2026-06-14 | Previous synchronized PRD with mixed MVP and future-scope requirements. |
