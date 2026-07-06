# NodePulse Documentation

This is the documentation index. Documents are organized by audience and
topic; each entry below names the file, who it's for, and what it covers so
you can jump straight to the right place.

> The root `README.md` covers project setup, the Docker stack, and a brief
> operations pointer. This file is the authoritative guide to **everything
> under `docs/`**.

## Product & scope

| Document | Audience | Covers |
|----------|----------|--------|
| [prd.md](prd.md) | Product, leads | Product context, the requirement status model (Supported / Partial / Planned / Deferred), functional & non-functional requirements (FR-1–FR-6, NFR-1–6), explicit success criteria with numbers, and the deferred-capabilities list. Source of truth for *what* the product must do and its quantitative targets. |
| [user-journey.md](user-journey.md) | Everyone (per role) | The authoritative, implementation-status-annotated user journeys: personas, the three-lifecycle overview (Deploy D1/D2 → Ops O1/O2 → Feature J1–J13), the B/F/U Implementation Layer Model, the §23 Implementation Gaps catalog, and cross-role playbooks. Read this to know *what users can actually do today and what's still gapped*. |

## Engineering

| Document | Audience | Covers |
|----------|----------|--------|
| [architecture.md](architecture.md) | Engineers | Implementation architecture and repository structure: repo layout, runtime data flow, Beacon/Pulse/Frontend internals, the UI system, state & data-fetching, testing. Read this to know *how the code is organized*. |
| [ui-design.md](ui-design.md) | Frontend engineers | Current frontend design system: design direction, theming architecture (Tailwind 4 + CSS variables), UI primitives, dialog/modal standards, layout, visualization, i18n, responsive behavior. |
| [authentication.md](authentication.md) | Engineers, security reviewers | Auth architecture: transport security, the login/refresh/logout flows, JWT service, session management, RBAC matrix, Beacon auth, security controls (rate limit, lockout, CSRF, 2FA), password requirements, DB schema. |
| [observability.md](observability.md) | SREs | Observability design: the three pillars (logs/metrics/traces), the full Prometheus metric catalog, logging config, OpenTelemetry tracing, end-to-end correlation, deployment integration. |
| [development-workflow.md](development-workflow.md) | All contributors | The mandatory development workflow: worktree-based branches, allowed commit types, completion gates (Go lint+build, frontend lint+build), squash-merge back to main. |

## Operations (under `docs/operation/`)

These are the operator-facing runbooks. `deploy/` ships the matching
artifacts (scripts, configs, service units); these documents explain how to
use them.

| Document | Audience | Covers |
|----------|----------|--------|
| [operation/operations.md](operation/operations.md) | SREs, on-call | The SRE runbook: health-endpoint triage table, common-incident playbooks (503, lockout, beacon offline, disk full, webhook delivery degraded), backup & restore procedures, config-change notes (Pulse restart vs Beacon SIGHUP), useful one-liners. |
| [operation/deployment-tls.md](operation/deployment-tls.md) | Deployers | TLS termination guide: common reverse-proxy requirements, quick-start for Caddy (recommended) and nginx + certbot, verification steps, Beacon `https://` notes. Reference configs live in `deploy/reverse-proxy/`. |
| [operation/upgrade.md](operation/upgrade.md) | Deployers, SREs | Upgrade & rollback: pre-upgrade checks (version, backup, changelog), the Docker-Compose upgrade flow, three rollback paths (code-only / migrate-down / backup-restore), per-release migration compatibility matrix, Beacon upgrade notes. |

## Decisions

| Document | Audience | Covers |
|----------|----------|--------|
| [adr/](adr/) | Architects, maintainers | Architecture Decision Records — non-obvious design decisions with context, alternatives considered, and consequences. Current set: ADR-001 report scheduling, ADR-002 alert routing, ADR-003 beacon config templates, ADR-004 unified config pattern. |

## How these relate

- **Scope question** ("should we build X?") → `prd.md`
- **Status question** ("can users do X today?") → `user-journey.md` (esp. §3 Implementation Layer Model and §23 Gaps)
- **How question** ("where is X in the code?") → `architecture.md`
- **Run question** ("how do I back up / upgrade / get TLS?") → `docs/operation/`
- **Why question** ("why did we pick X over Y?") → `adr/`

The two most common confusions this index exists to prevent:
1. `prd.md` §4 and `user-journey.md` both describe user journeys — `prd.md`
   carries the success criteria and requirement IDs, `user-journey.md`
   carries the implementation-status-annotated operation steps. They are
   complementary, not duplicative.
2. `architecture.md` §1-2 and `prd.md` §3 both "inventory the product" —
   `prd.md` inventories by capability (what it can do), `architecture.md`
   inventories by code component (how it's built). Different axes.
