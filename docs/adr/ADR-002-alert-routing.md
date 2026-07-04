# ADR-002: Alert routing rules (per-webhook dispatch filtering)

- **Status**: Accepted (implemented v2.3)
- **Date:** 2026-07-04
- **Owners:** Kevin
- **Related:** `docs/user-journey.md` §5 J2, §12 J9, §17 G14

## Context

The Alert Rules page has a "Routing Rules" tab
(`frontend/src/pages/AlertRulesPage.tsx`) that lets users define rules matching
metric / severity / node-group and route to a webhook or email. Today these
rules are stored **only in the browser** (`settingsStore` localStorage) and the
server does **not** evaluate them. On the server, *every enabled webhook receives
every alert* — see `pulse/internal/webhook/push_service.go:107` `SendAlert`,
where the only filter is `webhook.Enabled`.

Relevant dispatch-path facts (confirmed during planning):

- The injection point is `PushService.SendAlert` (`push_service.go:107-120`).
- The `*models.AlertEvent` already carries `Metric`, `Level` (P0/P1/P2), and
  `NodeID` at dispatch time — so metric/severity/node-based routing needs **no
  interface signature change** to `WebhookPusher.SendAlert(ctx, *AlertEvent)`.
- Delivery is already async (`dispatcher.go:103` `go deliverWebhook`) with a 30s
  timeout and concurrent per-webhook fan-out.
- There is **no node-group concept** anywhere in models or migrations today.
- `models.Webhook` has only `ID, URL, EventFormat, Enabled, CreatedAt` — no
  routing fields.

This is gap G14: a UI that implies server-enforced routing that doesn't exist.

## Decision (proposed)

Split routing into two independently shippable tiers based on what data already
flows to the dispatch point.

### Tier 1 — metric / severity / node routing (ship first)

- New `alert_routing_rules` table (uuid PK, owner, name, enabled,
  `metric` nullable, `severity` nullable array, `node_id` nullable,
  `webhook_id` FK → webhooks, created_at/updated_at).
- New CRUD endpoints under `/api/v1/alerts/routing-rules` (admin/operator).
- Inject a `Router`/`RouteMatcher` in `PushService.SendAlert` between fetching
  webhooks and iterating them. For each candidate webhook, load its active
  routing rules and keep the webhook only if at least one rule matches the event
  (null rule fields = wildcard). Webhooks with no rules keep current behavior
  (receive everything), preserving backward compatibility.
- Frontend replaces the localStorage routing store with the new endpoints.

### Tier 2 — node-group routing (ship later)

- Requires a new `node_groups` concept (table + node membership) that does not
  exist yet. Defer until there's a concrete operator need; design it together
  with any future node-tagging work.

## Consequences

- **Positive:** operators can stop alerting noise (e.g. only send latency alerts
  to the network team's webhook); honest UI; backward compatible.
- **Negative:** adds a DB read on the hot dispatch path. Mitigate with an
  in-memory rule cache (mirror the alert-engine rule-cache pattern in
  `internal/alert/engine.go`) refreshed on change.
- **Operational:** routing rules must be auditable (log which rule matched which
  event into `webhook_logs` or a debug field) so operators can diagnose "why
  didn't this alert reach my webhook".

## Alternatives considered

- **Route fields on the webhook itself** (e.g. `webhook.metrics[]`,
  `webhook.severities[]`) instead of a separate rules table. Rejected for tier 1
  because a single webhook may need different criteria per metric, and because
  it couples routing to the webhook CRUD. A separate table is more flexible and
  matches the PRD FR-4 "custom headers and severity filters" direction.
- **Evaluate routing in the dispatcher** (`alert/dispatcher.go`) instead of the
  push service. Rejected: the dispatcher is alert-engine-shaped and shouldn't
  know about webhook configuration; the push service already owns webhook
  selection, so filtering there is the smallest, most cohesive change.
- **Keep local-only.** Rejected: misleading. Kept the UI but added a warning
  banner (`alerts.localOnlyRouting`) pointing here.

## Current state (as of 2026-07-04, implemented v2.3)

- **Implemented (Tier 1).** `0004_alert_routing_rules` migration + `AlertRoutingRulesRepository` + CRUD handlers (`/api/v1/alerts/routing-rules`).
- `RouteMatcher` injected into `PushService.SendAlert` via `WithRouter`; webhooks with no rules still receive everything; webhooks with rules receive only matching alerts (metric/severity/node_id AND-match, nulls = wildcard).
- Wired in `alert.NewAlertEngine` via `webhook.NewRuleRouter`.
- Frontend `AlertRulesPage` routing tab swapped from localStorage to the server API; warning banner removed.
- Tier 2 (node-group-based routing) remains deferred — no node-group concept exists yet.
