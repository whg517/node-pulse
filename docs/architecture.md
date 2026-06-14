# NodePulse Architecture

**Version:** 1.0
**Date:** 2026-06-14
**Status:** Current implementation

This document describes the current repository architecture after the frontend shadcn/ui rewrite and the backend monitoring feature work.

---

## 1. Repository Layout

```
node-pulse/
├── beacon/       # Go monitoring agent
├── pulse/        # Go API server and background workers
├── frontend/     # React + TypeScript web application
├── e2e/          # Playwright end-to-end environment and tests
└── docs/         # Product, architecture, auth, observability, UI design docs
```

### Component Responsibilities

| Component | Responsibility |
|-----------|----------------|
| `beacon` | Runs on monitored nodes, executes probes, exposes Prometheus metrics, and reports heartbeat/metrics to Pulse. |
| `pulse` | Receives metrics, manages nodes, users, sessions, probes, alerts, reports, webhooks, and health diagnostics. |
| `frontend` | Provides the authenticated operations UI for dashboards, node management, alerting, exports, reports, webhooks, and settings. |
| `e2e` | Starts a Docker-based PostgreSQL + Pulse + Frontend test environment and runs Playwright flows. |

---

## 2. Runtime Data Flow

```
Beacon nodes
    |
    | POST /heartbeat, metrics payloads
    v
Pulse API server
    |
    +--> Auth / API key validation
    +--> In-memory cache for recent metrics
    +--> Async persistence to PostgreSQL
    +--> Alert evaluation and suppression
    +--> Webhook dispatch
    +--> Cleanup and scheduled jobs
    |
    v
Frontend API clients
    |
    +--> Dashboard and live metrics
    +--> Node detail and diagnostics
    +--> Alert rule and record workflows
    +--> Reports, exports, webhooks, users, sessions
```

The heartbeat path is intentionally non-blocking where possible: Pulse accepts recent metric updates, updates cache, queues persistence/evaluation work, and keeps request latency low.

---

## 3. Beacon Architecture

```
beacon/internal/
├── api/          # Pulse client and transport integration
├── auth/         # API key / token handling
├── cli/          # Command entrypoints
├── config/       # YAML config loading and validation
├── diagnostics/  # Local diagnostic helpers
├── logger/       # Logging setup
├── metrics/      # Prometheus metrics
├── models/       # Shared beacon-side models
├── monitor/      # Probe scheduling and collection orchestration
├── probe/        # TCP/UDP probe implementations
├── process/      # Process/runtime helpers
├── reporter/     # Heartbeat and metric reporting
└── telemetry/    # OpenTelemetry integration
```

Beacon is configured with `beacon.yaml`. The current implementation focuses on node identity, Pulse server URL, API key authentication, probe target configuration, reporting, and Prometheus metric exposure.

Operational expectations:

- Network calls to Pulse use retry/backoff behavior.
- Prometheus metrics are exposed locally when enabled.
- Shared runtime state is protected by synchronization primitives.
- Configuration examples live beside the component as `beacon.yaml.example`.

---

## 4. Pulse Architecture

```
pulse/internal/
├── alert/        # Alert rule evaluation
├── api/          # API helpers and DTOs
├── auth/         # JWT, sessions, refresh tokens, API keys
├── cache/        # Recent metrics cache
├── cleanup/      # Retention cleanup jobs
├── config/       # YAML and environment configuration
├── csrf/         # CSRF middleware
├── db/           # PostgreSQL connection and migrations
├── diagnostic/   # System and network diagnostics
├── export/       # Data export and report jobs
├── health/       # Health endpoints
├── logger/       # Logging setup
├── models/       # Domain models
├── scheduler/    # Background scheduler
├── security/     # Validation and rate limiting
├── server/       # HTTP router and middleware
├── suppression/  # Alert suppression
└── webhook/      # Webhook delivery
```

Pulse is configured with `pulse.yaml` or `PULSE_` environment variables. Environment variables take priority over file configuration.

Primary runtime concerns:

- Authentication for users and beacons.
- CSRF protection on mutation endpoints.
- Rate limiting for sensitive flows such as login.
- Recent metrics cache for dashboard responsiveness.
- PostgreSQL persistence for durable data.
- Background jobs for cleanup, exports, alert evaluation, and webhook delivery.

---

## 5. Frontend Architecture

### 5.1 Stack

- React 19
- TypeScript 5
- Vite 7
- React Router 7
- Zustand 5
- TanStack Query 5 for polling/query workflows
- Axios API client
- Tailwind CSS 4
- shadcn/ui v4 primitives built on Radix UI
- Recharts and react-simple-maps for visualization
- i18next / react-i18next for English and Simplified Chinese
- Vitest and Testing Library

### 5.2 Source Layout

```
frontend/src/
├── api/          # Domain API clients and DTOs
├── components/
│   ├── alerts/   # Alert tables, forms, detail dialogs
│   ├── charts/   # Recharts chart components
│   ├── common/   # Cross-cutting helpers such as ProtectedRoute
│   ├── dashboard/# Dashboard cards, charts, maps, streams
│   ├── export/   # Export forms and history/status components
│   ├── layout/   # App shell, sidebar, header, breadcrumbs
│   ├── nodes/    # Node table, node dialog, MTR visualizations
│   ├── reports/  # Report generator and PDF-style report views
│   ├── sessions/ # Session list
│   ├── ui/       # shadcn/ui primitives
│   └── webhooks/ # Webhook dialog/form/table
├── config/       # Constants and remaining compatibility tokens
├── hooks/        # Domain hooks and query hooks
├── lib/          # query-client and utility helpers
├── locales/      # en.json and zh-CN.json
├── pages/        # Route components
├── services/     # Browser-side services
├── stores/       # Zustand stores
├── types/        # Shared frontend types
└── utils/        # Formatting, timezone, accessibility, health helpers
```

### 5.3 Frontend Runtime Flow

```
main.tsx
  -> wait for i18n initialization
  -> App.tsx
      -> initialize persisted theme before render
      -> QueryClientProvider
      -> BrowserRouter
      -> ProtectedLayout
          -> ProtectedRoute
          -> AppLayout
          -> route page component
```

Routes are lazy-loaded with `React.lazy` and rendered under a shared authenticated layout. Session restoration is guarded so React StrictMode does not trigger duplicated startup side effects.

### 5.4 Routes

| Path | Page |
|------|------|
| `/login` | Login |
| `/dashboard` | Dashboard |
| `/nodes` | Node management |
| `/nodes/:id` | Node detail |
| `/nodes/comparison` | Node comparison |
| `/nodes/probes` | Probe management |
| `/beacons/config` | Beacon configuration |
| `/alerts/rules` | Alert rules and routing rules |
| `/alerts/records` | Alert records |
| `/alerts/history` | Alert history |
| `/performance` | Performance dashboard |
| `/reports` | Reports and schedules |
| `/reports/history` | Data export history |
| `/integrations/webhooks` | Webhooks |
| `/integrations/health` | System health |
| `/settings/preferences` | Preferences |
| `/settings/sessions` | Sessions |
| `/settings/users` | Users |

Short aliases exist for legacy and e2e navigation: `/webhooks`, `/sessions`, and `/comparison`.

---

## 6. UI System Architecture

The frontend uses a shadcn/ui + Tailwind CSS 4 architecture:

```
CSS variables (:root / .dark)
  -> @theme inline tokens
  -> Tailwind semantic utilities
  -> shadcn/ui primitives
  -> domain components and pages
```

Current rules:

- Prefer semantic utilities such as `bg-card`, `text-foreground`, `border`, `text-muted-foreground`, and `bg-primary`.
- Do not wrap CSS variables manually in component code for ordinary styling.
- Dark mode is driven by CSS variables; use `dark:` only when a component genuinely needs structural dark-mode behavior.
- Dialogs, alert dialogs, buttons, inputs, labels, switches, cards, badges, sidebars, dropdowns, scroll areas, tooltips, and skeletons come from `frontend/src/components/ui/`.
- User-visible strings go through `useTranslation()` and must exist in both locale JSON files unless a deliberate fallback is provided.

See `docs/ui-design.md` for design rules and component usage.

---

## 7. State and Data Fetching

The frontend uses two complementary patterns:

- **Zustand stores** for application state with cross-page behavior: auth, settings, alerts, dashboard, nodes, exports, and webhooks.
- **TanStack Query hooks** for server data with polling/retry semantics in newer flows.

The API clients live under `frontend/src/api/` and should remain domain-oriented. Pages should compose domain hooks/components instead of embedding low-level API transport details.

---

## 8. Testing and Verification

Recommended verification for frontend changes:

```bash
cd frontend
npm run lint
npm run test -- --run
npm run build
```

Recommended verification for backend changes:

```bash
cd pulse
make test-unit
make test

cd ../beacon
make test
```

Integration and browser workflows live under `e2e/` and require the Docker-based environment described there.

---

## 9. Documentation Boundaries

- `docs/prd.md`: product requirements, user journeys, success criteria.
- `docs/architecture.md`: implementation architecture and repository structure.
- `docs/ui-design.md`: current frontend design system and UI component usage.
- `docs/authentication.md`: authentication, sessions, CSRF, and token flows.
- `docs/observability.md`: OpenTelemetry and metrics behavior.
