# AGENTS.md

This file provides guidance to AI agents and contributors when working with code in this repository.

## Repository Structure

This is a monorepo for **Node-Pulse**, a distributed network monitoring system with three main components:

- **`beacon/`** - Go-based monitoring agent that runs on nodes, performs TCP/UDP probes, and reports metrics
- **`pulse/`** - Go-based backend API server that receives metrics, manages nodes, and serves the frontend
- **`frontend/`** - React + TypeScript web UI for visualization and management
- **`e2e/`** - Playwright end-to-end test suite
- **`docs/`** - Project documentation (PRD, auth design, UI design)

## Development Workflow

Follow `docs/development-workflow.md` for all code and documentation changes:

- Develop in Git worktrees under `.worktree/`.
- Start from `main`, then create a branch named `<type>-<name>`.
- Squash-merge completed work back to `main`.
- Required completion gates: `golangci-lint`, Go build, frontend lint, and frontend build.
- Use standardized Conventional Commit messages with the allowed types in the workflow document.
- Do not use the `chore` type for branches or commits.

## Common Commands

### Beacon (Monitoring Agent)

```bash
cd beacon

# Build for Linux AMD64 (static binary)
make build

# Build for current platform
make build-local

# Run the beacon (requires beacon.yaml)
make run

# Run tests
make test
make test-coverage

# Run linter
make lint
make lint-fix

# Clean build artifacts
make clean
```

Beacon requires a `beacon.yaml` configuration file. See `beacon.yaml.example` for reference.

### Pulse (Backend API Server)

```bash
cd pulse

# Download dependencies
make deps

# Run the server (default port 6532)
make run

# Build binary → bin/pulse-api
make build

# Build debug binary (optimizations disabled) → bin/pulse-api-debug
make build-debug

# Run tests (requires Docker for test database)
make test              # Full test suite: setup DB + run tests + cleanup DB
make test-unit         # Unit tests only (short mode)
make test-integration  # Integration tests only
make test-quick        # All tests (assumes DB is already running)

# Database management
make setup-test-db     # Start test PostgreSQL container
make cleanup-test-db   # Stop test PostgreSQL container

# Generate Swagger documentation (auto-installs swag if missing)
make swag
make swag-clean        # Remove generated docs
make swag-force        # Clean + regenerate

# Run linter (requires golangci-lint)
make lint

# Format Go code
make fmt

# Show all available targets
make help
```

Configuration is via `pulse.yaml`. Copy `pulse.yaml.example` and update values.
Environment variables with `PULSE_` prefix override config file settings.

### Frontend (React)

```bash
cd frontend

# Install dependencies
npm install

# Run dev server (default port 5173)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run unit/component tests (Vitest)
npm run test

# Lint
npm run lint
```

## Architecture Overview

### System Flow

```
Beacon (per node)          Pulse Server              PostgreSQL
     │                         │                          │
     │ POST /heartbeat         │                          │
     ├────────────────────────►│                          │
     │  (JWT auth)             │                          │
     │                         │                          │
     │                         ├──► Memory Cache          │
     │                         │   (ring buffer)          │
     │                         │                          │
     │                         ├──► Batch Writer ────────►│
     │                         │   (async)                │
     │                         │                          │
     │                         ├──► Alert Engine          │
     │                         │   (worker pool)          │
     │                         │                          │
     │                         ├──► Webhook Dispatcher    │
     │                         │                          │
     │                         └──► Scheduler             │
     │                             (cleanup, suppression) │
```

### Frontend Architecture

```
src/
├── api/           # API client functions (per domain: auth, nodes, alerts, etc.)
├── components/    # Reusable UI components
│   ├── common/    # Shared: PageContainer, PageHeader, ErrorBanner, ActionButton,
│   │              #         ConfirmDialog, LoadingSpinner, ProtectedRoute, etc.
│   ├── layout/    # AppLayout, Header, Sidebar, Breadcrumb, PageHeader
│   ├── alerts/    # Alert rules, records components
│   ├── charts/    # ECharts wrappers
│   ├── dashboard/ # MetricCard, NodeListTable
│   ├── nodes/     # Node-related components
│   ├── export/    # Data export components
│   ├── sessions/  # Session management components
│   └── webhooks/  # Webhook components
├── config/        # designTokens.ts, constants.ts
├── hooks/         # Custom hooks: useAuth, useDashboard, useNodeDetail, useTheme, etc.
├── locales/       # i18n: en.json, zh-CN.json
├── pages/         # Route-level page components
├── stores/        # Zustand state stores (authStore, nodesStore, alertsStore, etc.)
└── types/         # Shared TypeScript types
```

### Frontend Tech Stack

- **React 19** + **TypeScript 5** + **Vite 7**
- **Tailwind CSS 4** for styling (dark mode via `dark:` classes)
- **React Router v7** for client-side routing
- **Zustand 5** for state management
- **i18next** + **react-i18next** for internationalization (EN + zh-CN)
- **ECharts 6** for charts and data visualization
- **Vitest** + **@testing-library/react** for unit/component tests

### Frontend Routes

| Path | Component | Notes |
|------|-----------|-------|
| `/login` | LoginPage | Public |
| `/dashboard` | DashboardPage | |
| `/nodes` | NodeManagementPage | |
| `/nodes/:id` | NodeDetailPage | |
| `/nodes/comparison` | NodeComparisonPage | |
| `/alerts/rules` | AlertRulesPage | |
| `/alerts/records` | AlertRecordsPage | |
| `/alerts/history` | AlertHistoryPage | |
| `/reports` | ReportsPage | |
| `/reports/history` | DataExportPage | |
| `/integrations/webhooks` | WebhooksPage | |
| `/integrations/health` | SystemHealthPage | |
| `/settings/preferences` | PreferencesPage | |
| `/settings/sessions` | SessionsPage | |
| `/settings/users` | UsersPage | admin only |

### Pulse Internal Packages

```
pulse/internal/
├── alert/         # Alert rule evaluation engine
├── auth/          # JWT, session, API key management
├── cache/         # In-memory ring buffer cache
├── cleanup/       # Scheduled data cleanup (retention)
├── config/        # Configuration loading (YAML + env vars)
├── csrf/          # CSRF protection middleware
├── db/            # PostgreSQL connection + migrations
├── diagnostic/    # System health diagnostics
├── export/        # Data export functionality
├── health/        # Health check endpoints
├── models/        # Shared domain models
├── scheduler/     # Background job scheduler
├── security/      # Input validation, rate limiting
├── server/        # HTTP server setup, router, middleware
├── suppression/   # Alert suppression logic
└── webhook/       # Webhook delivery
```

## Important Patterns

### Error Handling
- Beacon: Exponential backoff retry (1s, 2s, 4s) for 5xx and network errors
- Pulse: Distinguishes retryable (5xx, network) vs non-retryable (4xx) errors
- Frontend: Axios interceptor auto-refreshes JWT on 401; `ErrorBanner` component for error display

### Frontend Design System
- Use `PageContainer` + `PageHeader` for all page layouts (replaces custom nav/header)
- Use `ErrorBanner` for error states, `ActionButton` for primary actions, `ConfirmDialog` for confirmations
- Use `statusColors`, `statusClasses` from `src/config/designTokens.ts` for status indicators
- Always add `dark:` class variants for dark mode support
- Use `t()` from `useTranslation()` for all user-visible strings; add keys to both `en.json` and `zh-CN.json`

### Security
- JWT secrets auto-generated (512-bit) if not provided in config
- TLS 1.2+ enforcement for Beacon→Pulse communication
- SHA-256 hashing for API keys and refresh tokens
- Rate limiting: 5 login attempts per IP per minute
- Account lockout: 5 failed attempts = 10 minute lockout
- Role-based access control (admin, operator, viewer, beacon)
- CSRF protection on mutation endpoints
- Access tokens stored in memory only (not localStorage)

### Performance
- Async batch writes to DB (non-blocking for heartbeat endpoint)
- Memory cache for real-time queries (ring buffer, 60 points per node)
- Worker pools for alert evaluation (10 workers)
- JWT: access token 15 min expiry; refresh via 401 interceptor (no timer)
- Non-blocking metric writes (drop on overflow)

## Testing

### Beacon Tests
Located in `beacon/internal/*/` and `beacon/tests/`. Run with `make test`.

### Pulse Tests
- Unit tests: `pulse/internal/**/*_test.go` and `pulse/tests/api/`, `pulse/tests/cache/`
- Integration tests: `pulse/tests/integration/`
- Requires Docker test database: `make setup-test-db`

### Frontend Tests
Located in `frontend/src/**/*.test.tsx` and `frontend/src/**/__tests__/`. Run with `npm run test` (uses Vitest).

### E2E Tests (Playwright)
Located in `e2e/tests/` organized by feature:
- `tests/auth/` - Login, logout, sessions, token refresh
- `tests/rbac/` - Role-based access control (admin, operator, viewer)
- `tests/nodes/` - Node list, detail, CRUD, comparison
- `tests/alerts/` - Alert rules, history, records
- `tests/webhooks/` - Webhook configuration
- `tests/export/` - Data export functionality
- `tests/dashboard/` - Dashboard metrics
- `tests/performance/` - Performance metrics
- `tests/reports/` - Reports
- `tests/sessions/` - Session management
- `tests/smoke/` - Smoke tests (quick sanity checks)
- `tests/visual/` - Visual regression tests

```bash
cd e2e

# Install dependencies
npm install

# Run all tests (requires running Pulse + Frontend + DB)
npm test

# Run specific test suites
npm run test:auth
npm run test:rbac
npm run test:nodes
npm run test:alerts
npm run test:webhooks
npm run test:export
npm run test:smoke        # Quick smoke tests
npm run test:smoke:fast   # Smoke tests (Chromium only, 4 workers)

# Docker-based E2E environment (recommended)
npm run docker:up      # Start Pulse, Frontend, PostgreSQL in Docker
npm run docker:down    # Stop containers
npm run docker:logs    # View logs
npm run docker:reset   # Reset volumes and restart

# Debugging
npm run test:ui        # Playwright UI mode
npm run test:debug     # Debug mode
npm run test:headed    # Run with browser visible
npm run report         # Show test report
```

The E2E environment uses `docker-compose.e2e.yml` which sets up:
- PostgreSQL test database (port 5432)
- Pulse backend API (port 6532)
- Frontend dev server (port 5173)

## Configuration

### Beacon Configuration (`beacon.yaml`)
Required:
- `pulse_server` - Pulse server URL
- `node_id` - Unique node identifier (alphanumeric, hyphens, underscores)
- `node_name` - Human-readable name
- `api_key` - API key for JWT authentication

Optional:
- `region` - Region label
- `tags` - Custom string tags
- `probes` - TCP/UDP probe targets (type, target, port, interval, count, timeout)
- `reconnect` - Retry config (max_retries, retry_interval, backoff strategy)
- `metrics_enabled` - Enable Prometheus `/metrics` endpoint (default: true)
- `metrics_port` - Prometheus metrics port (default: 2112)
- `metrics_update_seconds` - Metrics update interval (default: 10)

### Pulse Configuration (`pulse.yaml`)
Copy `pulse.yaml.example` to `pulse.yaml`. All fields can also be set via environment variables with `PULSE_` prefix (e.g. `PULSE_DATABASE_URL`).

Key sections:
- `server.port` (default: `6532`), `server.mode` (`debug`/`release`)
- `database.url` - PostgreSQL connection string (**required**)
- `log.level` (`debug`/`info`/`warn`/`error`), `log.format` (`text`/`json`)
- `cors.allowed_origins` (default: `http://localhost:4173,http://localhost:5173`)
- `admin.username` / `admin.password` (default: `admin` / `Admin123`)
- `jwt.secret` (auto-generated if empty), `jwt.expiration_hours` (default: 24)
- `session.secret` (auto-generated if empty), `session.expiration_hours` (default: 24)
- `cleanup.enabled`, `cleanup.retention_days` (default: 7)

Configuration priority: **Environment variables > pulse.yaml > defaults**

## Development Notes

1. Configuration priority: Environment variables > pulse.yaml > built-in defaults
2. Beacon supports config hot-reload (no restart required)
3. Graceful shutdown: All components handle SIGTERM/SIGINT
4. Token rotation: Refresh tokens are one-time use; access tokens live in memory only
5. Concurrent safety: All shared state protected by mutexes
6. Prometheus metrics: Beacon exposes `/metrics` on port 2112 (configurable)
7. Swagger UI available at `http://localhost:6532/swagger/index.html` when running in debug mode
