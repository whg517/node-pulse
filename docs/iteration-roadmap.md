# NodePulse Iteration Roadmap

**Date:** 2026-06-21
**Status:** Active implementation roadmap
**Source of truth:** `docs/prd.md` v4.0

This roadmap is based on the current implementation baseline. It prioritizes closing partially supported workflows before adding new surface area.

## 1. Planning Principles

- Keep PRD, roadmap, tests, and UI messaging aligned.
- Prefer finishing already-started workflows over adding new ones.
- Treat production security and data integrity as release blockers.
- Implement one coherent workflow per iteration.
- Add focused tests at API, component, or service boundaries touched by the change.
- Keep unsupported features honest in the UI until backend support exists.

## 2. Phase 0 - Baseline Alignment

**Goal:** Make requirements and roadmap trustworthy again.

**Status:** Completed.

**Scope:**

- Replace the old mixed-scope PRD with a current implementation baseline.
- Replace the old roadmap with a phased plan tied to PRD v4.0.
- Identify first implementation slice from already-partial workflows.

**Acceptance:**

- PRD uses Supported, Partially supported, Planned, and Deferred labels.
- Roadmap priorities map to current code gaps.
- First implementation tasks are small enough to test locally.

## 3. Phase 1 - Safety And Workflow Closure

### P1.1 Production mTLS Mode Alignment

**Status:** Completed.

**Problem:** Production mode in config is `release`, but mTLS default strict-mode detection checks for `production`.

**Scope:**

- Treat `release` as production for mTLS default behavior.
- Preserve explicit `PULSE_MTLS_ENABLED` overrides.
- Add or update middleware tests.

**Acceptance:**

- `PULSE_SERVER_MODE=release` defaults mTLS to strict when no explicit mTLS mode is set.
- Debug/test modes continue to default to disabled.

### P1.2 Alert Notes API

**Status:** Completed.

**Problem:** Frontend alert mobile/detail flows support notes, but Pulse has no `/alerts/records/:id/notes` routes and status update ignores `note`.

**Scope:**

- Add `alert_notes` migration.
- Add DB helpers for create/list notes.
- Add `POST /api/v1/alerts/records/:id/notes`.
- Add `GET /api/v1/alerts/records/:id/notes`.
- Persist optional note during status update.
- Return notes in shapes compatible with current frontend DTOs.
- Add focused backend tests.

**Acceptance:**

- Updating alert status with `note` stores a note.
- Adding a note without changing status stores a note.
- Listing notes returns notes newest or oldest in a documented order.
- Missing alert record returns 404.

### P1.3 Webhook Test Delivery

**Status:** Completed.

**Problem:** Payload preview exists, but operators cannot send a manual test delivery before enabling a webhook.

**Scope:**

- Add backend endpoint for webhook test delivery using the same renderer and SSRF validation path.
- Return HTTP status, response time, and concise error.
- Add frontend action in the webhook dialog/page.
- Add tests for success and validation failure.

**Acceptance:**

- Operator can preview payload and send a test payload.
- Invalid URLs fail inline.
- Test delivery does not require an existing alert.

### P1.4 Regression Gate

**Status:** Completed for P1 implementation slice.

**Scope:**

- Run focused Go tests for touched Pulse packages.
- Run focused frontend tests for alert records/webhooks if UI changes.
- Run frontend build after UI work.

## 4. Phase 2 - Real-Time Incident Flow

### P2.1 Pulse Alert Event Stream

**Status:** Completed for the first production path.

**Problem:** Frontend connects to `/ws`, but Pulse does not expose an event stream.

**Scope:**

- Choose WebSocket or SSE for first production path.
- Broadcast alert-created, alert-status-updated, and alert-note-created events.
- Authenticate stream connections.
- Connect dashboard alert stream to live backend events with polling fallback.

**Acceptance:**

- New alert appears on dashboard without manual refresh.
- Browser notification can be triggered from backend event data.
- Connection failure degrades gracefully.

### P2.2 Alert Timeline UX

**Status:** Partially completed.

**Scope:**

- Display persisted status changes and notes together.
- Show UTC and user timezone.
- Ensure mobile alert detail uses persisted backend data.

**Completed:**

- Persist alert status-change history.
- Expose merged alert timeline API for created, status-change, and note events.
- Display merged timeline in desktop alert detail with local and UTC timestamps.

**Remaining:**

- Wire mobile alert detail to the persisted timeline API.

## 5. Phase 3 - Beacon Transport Integrity

### P3.1 Compressed Heartbeat Path

**Problem:** Compression utilities and server endpoint exist, but Beacon runtime does not use them.

**Scope:**

- Wire Beacon compression settings into heartbeat reporting.
- Send compressed payloads to `/api/v1/beacon/heartbeat/compressed` when enabled and beneficial.
- Fall back to ordinary heartbeat on compression failure if configured.
- Expose active compression metrics.

**Acceptance:**

- Compressed heartbeat is accepted by Pulse and produces the same metric effects.
- CRC mismatch is rejected and counted.
- Compression ratio metric reflects runtime traffic.

### P3.2 Resume Upload Cache

**Scope:**

- Persist failed non-heartbeat metric payloads locally.
- Apply FIFO and priority behavior from PRD.
- Upload cached entries after reconnection.
- Expose cache size, eviction, and resume byte metrics.

## 6. Phase 4 - Reports And Exports

### P4.1 XLSX Export

**Scope:**

- Add XLSX generation in Pulse export service.
- Re-enable Excel in export/report UI.
- Add format-specific tests.

### P4.2 Server-Side Report Scheduling

**Scope:**

- Persist report schedules.
- Execute daily, weekly, and monthly schedules.
- Generate report artifacts server-side.
- Add email delivery after email provider configuration exists.

### P4.3 Report Recommendations

**Scope:**

- Strengthen root cause and recommended action summaries.
- Highlight likely owner and confidence.

## 7. Phase 5 - Production Hardening

### P5.1 Health Endpoints

**Scope:**

- Add Pulse `/healthz` alias or document current `/api/v1/health` as canonical.
- Add Beacon `/healthz`.
- Add `beacon_self_health_status`.

### P5.2 Backup Automation

**Scope:**

- Add config-data backup target abstraction.
- Support local filesystem first; S3-compatible target later.
- Add retention cleanup.

### P5.3 Performance Gates

**Scope:**

- Add repeatable heartbeat load test for 50 Beacons.
- Track P95/P99 latency for heartbeat and dashboard queries.
- Document baseline hardware and test command.

### P5.4 Accessibility And i18n Sweep

**Scope:**

- Remove remaining hardcoded operational strings.
- Audit dialogs, forms, tables, icon buttons, charts, and mobile alert detail.
- Add textual alternatives for chart-heavy views.

## 8. Deferred Work

- Iperf3 throughput testing.
- MTU path discovery.
- AI-driven diagnosis.
- Multi-region Pulse deployment.

### 8.1 Backend concurrency / slow query under load (surfaced by e2e)

**Status:** Open. Surfaced by the e2e CRUD suite (`tests/nodes/`, `tests/alerts/`)
running serially against the containerized docker-compose stack. Individual
requests are fast (curl `/api/v1/nodes` returns 200 in ~2 ms), but under
concurrent load the backend degrades badly and returns 500.

**Observed symptoms (from pulse request logs during a full e2e run):**

```
GET /api/v1/nodes          500  20.7s   ← should be ~2ms
GET /api/v1/data/history   500  25.7s   ← node_id + 2 metrics, 24h range
GET /nodes                 200  25.6s   ← SPA navigation blocked waiting on the above
GET /ws?token=...          200  25.8s   ← websocket held open the whole time
```

A single request is fine; the failure appears once the websocket stream and
several `/data/*` calls are in flight at the same time.

**Root cause (diagnosed, not yet fixed):**

1. **N+1 query in `GetHistoryHandler`** —
   `pulse/internal/api/data_handler.go:298` loops `for nodeID { for metric {
   queryMetricHistory(...) }}`, issuing one DB query per (node × metric) pair
   **serially**. With N nodes × M metrics that is N×M round trips on one request.
2. **No request-level timeout** — the handler uses
   `ctx := context.Background()` (`data_handler.go:295`) instead of the gin
   request context, so a slow query can hold a pool connection indefinitely.
3. **Unconfigured connection pool** — `db.New()` (`internal/db/conn.go`) calls
   `pgxpool.New(url)` with no explicit `MaxConns`; pgx defaults to
   `max(4, runtime.NumCPU())`, which is small on the container and is exhausted
   once several long queries + the websocket compete for connections.

**Why it returns 500 (not just slow):** when the pool is starved, queued
queries eventually error (context deadline / connection acquisition failure),
and the handler surfaces that as `500 "Failed to query historical data"`.

**Reproduce locally:**
```bash
cd e2e && npm run docker:up
E2E_BASE_URL=http://localhost:6532 npx playwright test tests/nodes/ --project=chromium --workers=1
# watch pulse logs: docker logs nodepulse-e2e-api -f | grep -E "500|20\.\d+s"
```

**Proposed fix (separate PR):**
- Replace the N+1 loop with a single batched/`IN (...)` query for history.
- Use the gin request context (`c.Request.Context()`) with a per-handler
  timeout, not `context.Background()`.
- Configure the pgxpool explicitly (`MaxConns`, `MaxConnLifetime`,
  `MaxConnIdleTime`) from config, and default higher than `NumCPU`.

This is a backend capacity/correctness concern, independent of the e2e suite
that exposed it; the e2e tests themselves are correct (each passes in isolation).

## 9. Current Sprint

Completed implementation slices:

1. P1.1 Production mTLS Mode Alignment.
2. P1.2 Alert Notes API.
3. P1.3 Webhook Test Delivery.
4. P2.1 Pulse Alert Event Stream.
5. P2.2 Alert Timeline UX backend and desktop detail slice.

Next implementation slice:

1. P2.2 Mobile Alert Timeline Wiring.
2. Then continue to Phase 3 Beacon Transport Integrity.
