# ADR-001: Report scheduling (server-side schedules + email vs. local-only)

- **Status**: Accepted (implemented v2.3)
- **Date:** 2026-07-04
- **Owners:** Kevin
- **Related:** `docs/user-journey.md` §11 J8, §17 G13

## Context

The Reports page (`frontend/src/pages/Reports.tsx`) lets users create "schedules"
(daily / weekly / monthly, format, time). However, today these schedules are
stored **only in the browser** via `settingsStore` (`localStorage`, key
`settings-store`). The server has no knowledge of them and never executes them.
A user switching devices, clearing browser data, or simply never revisiting the
tab gets no scheduled report. This is a "fake-server" capability (gap G13) that
risks misleading users.

Related facts from the codebase:

- The scheduler infrastructure exists: `pulse/internal/scheduler/scheduler.go`
  runs periodic jobs registered in `internal/server/registry.go` and started in
  `server.go:41`. Cleanup and suppression jobs already use it.
- `ExportTask` was just made durable (ADR-adjacent: §17 G3 added
  `0002_export_tasks` migration + repository). A scheduled report is essentially
  a time-triggered export + (optionally) a delivery step.
- PDF report generation today is **client-side** (`HealthReportPDF` +
  `window.print()`); the server has no PDF renderer.

## Decision (proposed)

Implement server-side report scheduling in three stages. Do **not** ship stage 1
until the data model and UI warnings (already added, see below) are in place.

1. **Persist schedules server-side.** New `report_schedules` table (uuid PK,
   owner user_id FK, name, frequency `daily|weekly|monthly`, time-of-day tz-aware,
   node_ids jsonb, metrics jsonb, format, enabled, last_run_at, next_run_at,
   created_at/updated_at) mirroring the existing migration conventions. New
   `GET/POST/PUT/DELETE /api/v1/reports/schedules` CRUD (admin/operator).
   Frontend replaces the localStorage store with these endpoints.

2. **Server-side generation + delivery.** A scheduler job iterates due schedules,
   reuses the existing export pipeline to produce the artifact, and delivers it.
   PDF rendering must move server-side (e.g. `chromedp` headless or a Go-native
   PDF library) — this is the largest piece of work. Email delivery requires a
   configured provider (SMTP or a transactional API).

3. **Failure handling & observability.** Per-schedule `last_run_at`,
   `last_status`, `last_error`, retry-with-backoff, and Prometheus metrics for
   schedule execution. Surface last-run status in the UI.

## Consequences

- **Positive:** schedules survive device/browser changes; reports arrive without
  a user opening the UI; auditable and observable.
- **Negative:** introduces a server-side PDF dependency (heavier image, longer
  build) and an external email provider dependency (secret management).
- **Operational:** email delivery needs provider config, rate limits, bounce
  handling, and timezone correctness (store schedule time with explicit tz;
  evaluate in the server's tz or per-user tz). Retries must not flood an inbox.

## Alternatives considered

- **Keep local-only and disable the UI.** Rejected: removes a useful capability
  and breaks existing user workflows. Instead we kept the UI but added an honest
  "Local-only" warning banner (this iteration) pointing to this ADR.
- **Client-side scheduling via a long-open browser tab.** Rejected: unreliable;
  a closed tab means no report; can't run server-side retries or email.
- **Cron + filesystem artifacts, no DB.** Rejected: loses ownership, audit, and
  per-user state; harder to expose in the UI.

## Current state (as of 2026-07-04, implemented v2.3)

- **Implemented.** `0003_report_schedules` migration + `ReportScheduleRepository` + CRUD handlers (`/api/v1/reports/schedules`).
- Scheduler task (`ReportScheduleRunner`) polls due schedules every minute, generates CSV (via `ExportService`) or PDF (via server-side `GeneratePDF` using `gopdf`), emails the artifact to `recipient_email` (or owner), and advances `last_run_at`/`next_run_at`.
- Email delivery via the new `notify` SMTP package (`SMTPSender`, falls back to log-only `NoopSender` when unconfigured).
- Frontend `Reports.tsx` swapped from localStorage to the server API; warning banner removed.
- XLSX format remains client-side-only (PDF + CSV are server-generated).
