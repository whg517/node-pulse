# Architecture Decision Records (ADR)

This directory holds lightweight ADRs for NodePulse. Each record captures a
decision that is non-obvious from the code alone: the context, the alternatives
considered, and the consequences.

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [ADR-001](./ADR-001-report-scheduling.md) | Report scheduling (server-side schedules + email) | Accepted (implemented v2.3) |
| [ADR-002](./ADR-002-alert-routing.md) | Alert routing rules (per-webhook dispatch filtering) | Accepted (implemented v2.3) |
| [ADR-003](./ADR-003-beacon-config-templates.md) | Beacon configuration templates (server vs. local) | Accepted (implemented v2.3) |
| [ADR-004](./ADR-004-unified-config-pattern.md) | Unified configuration pattern across components | Accepted |

## When to add an ADR

Add one whenever a feature has more than one reasonable production shape and the
choice isn't self-evident — especially when a UI exists today but the backend
doesn't yet enforce the behavior (the "fake-server" cases catalogued in
`docs/user-journey.md` §17).

## Template

```
# ADR-NNNN: Title

- Status: Proposed | Accepted | Superseded by ADR-XXXX
- Date: YYYY-MM-DD
- Owners: <name>

## Context
## Decision
## Consequences
## Alternatives considered
```
