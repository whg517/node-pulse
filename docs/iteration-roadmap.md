# NodePulse Iteration Roadmap

**Date:** 2026-06-21
**Status:** Active iteration plan

This roadmap captures the remaining product work identified from the recent code, UI, documentation, and implementation reviews. Each item is intended to be delivered as an independent feature or fix commit.

---

## 1. Iteration Principles

- Ship one functional change per commit.
- Prefer existing API, store, component, i18n, and design-token patterns.
- Keep operations UI dense, predictable, and task-focused.
- Add tests where behavior, user-visible text, or API contracts change.
- Keep visible strings in both `en.json` and `zh-CN.json`.
- Treat unsupported capabilities honestly in UI until the backend supports them.

---

## 2. High Priority

### H1. Webhook Payload Preview and Test Send

**Problem:** Webhook event format templates are now rendered by the backend, but operators cannot preview or test a configured payload before relying on it for alert delivery.

**Scope:**
- Add a backend endpoint to render or send a webhook test event.
- Add frontend actions in the webhook workflow for previewing the payload and triggering a test delivery.
- Show success/failure feedback without leaving the dialog/page.
- Reuse the existing webhook form JSON editor and i18n structure.

**Acceptance:**
- A user can preview the rendered JSON for the current template.
- A user can send a test payload to a configured webhook URL.
- Invalid JSON or invalid URLs produce clear inline errors.
- Unit tests cover rendering and UI feedback.

### H2. Beacon Config Validation

**Problem:** Beacon probe configuration can be edited, but the UI does not yet guide users strongly enough on invalid target, port, interval, timeout, or count values.

**Scope:**
- Add client-side validation for probe rows and global settings.
- Keep errors close to the field that caused them.
- Prevent save while the config is invalid.
- Preserve current compact table-like layout.

**Acceptance:**
- Empty targets are rejected.
- Ports must be `1..65535`.
- Intervals, timeouts, and counts must respect configured minimums.
- Tests cover invalid and valid save flows.

### H3. Frontend i18n Completion

**Problem:** Some user-visible strings remain hardcoded in English or Chinese, especially on operational pages and reusable components.

**Scope:**
- Continue page-by-page cleanup for `Retry`, `Loading...`, system health labels, export history, login, session badges, report table labels, and alert mobile notes.
- Add missing keys to both locales.
- Avoid changing domain protocol labels such as TCP, UDP, CSV, PDF, MFA, and metric IDs.

**Acceptance:**
- Targeted pages render locale-backed labels.
- Locale JSON parses.
- Existing tests are updated where visible text changes.

### H4. Full Regression Pass

**Problem:** Recent iterations touched backend webhook delivery, frontend reports, exports, node dialogs, and Beacon config UI.

**Scope:**
- Run frontend build and lint.
- Run focused frontend tests for touched pages/components.
- Run relevant Go package tests.
- Run smoke/e2e if the local Docker environment is available.

**Acceptance:**
- Test results are documented in the final iteration summary.
- Any failure is either fixed or recorded with a concrete blocker.

---

## 3. Medium Priority

### M1. Excel Export Backend Support

**Problem:** Excel export is intentionally disabled in the frontend because the backend currently supports CSV only.

**Scope:**
- Add XLSX generation in the backend export pipeline.
- Re-enable Excel in export/report UI when the API supports it.
- Add format-specific tests.

**Acceptance:**
- CSV and Excel exports both produce downloadable files.
- Unsupported format errors are no longer triggered for Excel.

### M2. Diagnostic Report Recommendations

**Problem:** Reports include metrics, diagnosis, MTR, and alert context, but the operator-facing action summary can be stronger.

**Scope:**
- Add concise root-cause and recommended-action summaries.
- Highlight likely owner: local node, cross-border link, carrier route, or unknown.
- Keep PDF/report UI readable and scan-friendly.

**Acceptance:**
- Reports include a clear action summary when enough evidence exists.
- Low-confidence cases are labeled as such.

### M3. Legacy Beacon Token Cleanup

**Problem:** `pulse/internal/db/beacon_tokens.go` appears unused after the migration to `api_keys`.

**Scope:**
- Confirm no runtime path references the legacy querier.
- Remove or archive the dead code and related tests if safe.
- Update docs if the old terminology appears.

**Acceptance:**
- The codebase has one clear API-key based Beacon authentication path.
- Go tests still compile.

### M4. Password Reset Email Delivery

**Problem:** Password reset token generation exists, but email sending is still a TODO.

**Scope:**
- Choose/configure an email provider or SMTP abstraction.
- Send password reset emails with safe, localized content.
- Add configuration and failure logging.

**Acceptance:**
- Password reset flow can deliver email in a configured environment.
- Missing email configuration fails clearly without exposing secrets.

---

## 4. Lower Priority

### L1. Accessibility Sweep

**Scope:**
- Verify focus states, aria labels, dialog titles, table semantics, and keyboard navigation.
- Prioritize forms, dialogs, tables, and icon-only buttons.

### L2. Documentation Sync

**Scope:**
- Update architecture and UI docs after each completed feature group.
- Keep PRD scope aligned with actual supported capabilities.

### L3. E2E Coverage Expansion

**Scope:**
- Add smoke flows for webhook test delivery, Beacon config validation, export format behavior, and report generation.

---

## 5. Recommended Next Sequence

1. H1 Webhook payload preview and test send.
2. H2 Beacon config validation.
3. H3 i18n cleanup pass for the most visible remaining pages.
4. H4 full regression pass.
5. M1 Excel export support or M2 report recommendations, depending on operator priority.
