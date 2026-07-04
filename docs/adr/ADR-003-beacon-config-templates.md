# ADR-003: Beacon configuration templates (server vs. local)

- **Status**: Accepted (implemented v2.3)
- **Date:** 2026-07-04
- **Owners:** Kevin
- **Related:** `docs/user-journey.md` §8 J5, §17 G15 (and G8/G9 — preview & batch)

## Context

The Beacon config page (`frontend/src/pages/BeaconConfigPage.tsx`) lets users
"Save as template" and apply templates. Today templates live **only in the
browser** via `settingsStore` localStorage. They don't sync across devices or
team members, and are lost when browser data is cleared. This is gap G15.

Related facts:

- Beacon config itself **is** server-managed: `POST /beacons/:id/config` writes
  a versioned config (`beacon_configs` + `beacon_config_history`), and the Beacon
  acks applied/failed versions. So the *target* config is durable; only the
  *template* library is local.
- Two related backend capabilities also lack UI: config preview
  (`POST /beacons/:id/config/preview`, G8) and group batch config
  (`POST /beacon-groups/:gid/config`, G9). Templates, preview, and batch deploy
  are naturally related workflows for operators managing many beacons.

## Decision (proposed)

Move templates server-side as a first-class resource, and frame them as the
foundation for batch deployment.

1. **New `beacon_config_templates` table** (uuid PK, owner user_id FK,
   name, description, probes jsonb, interval_seconds, timeout_seconds,
   created_at/updated_at). New CRUD endpoints
   `/api/v1/beacon-config-templates` (admin/operator).
2. **Frontend** swaps the localStorage template store for these endpoints.
   "Apply template" already calls `updateBeaconConfig`; that stays.
3. **(Follow-up, ties to G9 batch)** Add multi-select "apply template to these
   beacons" using the existing `POST /beacon-groups/:gid/config` (or a new
   multi-beacon endpoint). This turns templates into a real fleet-management
   tool rather than a one-at-a-time convenience.

## Consequences

- **Positive:** templates are shareable across the team and survive device
   changes; lays groundwork for batch deploy (G9) and config diff/preview (G8).
- **Negative:** another CRUD resource + permissions (who can edit a shared
   template?). Recommend owner-based edit, all-operator read, matching the
   `CheckResourceOwnership` pattern in `rbac.go`.
- **Operational:** applying a template is a config version bump per beacon, so
   each beacon still acks independently — no new failure mode beyond what config
   deploy already has.

## Alternatives considered

- **Keep templates local-only.** Rejected for the shareability/durability
   reasons above, but kept the UI working today with a warning banner
   (`beaconConfig.localOnlyTemplates`) pointing here. This avoids data loss for
   existing users while the server-side version is pending.
- **Store templates as rows in `beacon_config_history` with a `is_template`
   flag.** Rejected: mixes two concerns (audit history vs. reusable library) and
   complicates history queries.
- **Per-user templates only (no sharing).** Rejected: defeats the main benefit
   for ops teams standardizing across many beacons.

## Current state (as of 2026-07-04, implemented v2.3)

- **Implemented.** `0005_beacon_config_templates` migration + `BeaconConfigTemplatesRepository` (owner-scoped CRUD) + handlers (`/api/v1/beacon-config-templates`).
- Frontend `BeaconConfigPage` swapped from localStorage to the server API; warning banner removed. Save/list/apply/delete now hit the server; "apply template" still drives the existing `POST /beacons/:id/config` (new config version per beacon).
- G8 (config preview) and G9 (batch deploy) remain closed since v2.2.
