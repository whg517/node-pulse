# Node-Pulse E2E

End-to-end tests for Node-Pulse using [Playwright](https://playwright.dev/).

All tests run against a **real stack** (frontend + Pulse API + PostgreSQL). There
is no mocked-API mode: the suite exists to verify the full system behaves
correctly, so every assertion exercises the real backend and database.

## Prerequisites

- Node.js 18+
- A running Node-Pulse stack:
  - PostgreSQL (default port 5432)
  - Pulse API (default port 6532)
  - Frontend (default port 5173)
- An admin account seeded (`admin` / `Admin123` by default).

## Quick Start (Docker stack — recommended)

The bundled compose file brings up Postgres + Pulse + frontend together:

```bash
cd e2e
npm install
npm run install:browsers     # one-time: installs Playwright browsers

npm run docker:up            # start the full stack (builds images on first run)
# wait for healthchecks to go green, then:
npm run test:smoke:fast      # quick sanity (Chromium only)

npm run docker:down          # stop when done
```

## Quick Start (existing local stack)

If you already run Pulse + frontend + DB locally:

```bash
cd e2e
npm install
npm run install:browsers
npm run test:smoke
```

Playwright can also start the frontend for you (it does **not** start Pulse or
the database — those must already be up):

```bash
E2E_START_FRONTEND=1 npm run test:smoke
```

## Test Suites

| Script | Coverage |
|--------|----------|
| `npm test` | All suites |
| `npm run test:smoke` | Critical path: login → dashboard → sidebar nav |
| `npm run test:smoke:fast` | Smoke, Chromium only, 4 workers |
| `npm run test:auth` | Login form, protected-route redirects, deep-link return |
| `npm run test:navigation` | Every protected route renders its shell |
| `npm run test:nodes` | Node management, probes, beacon config |
| `npm run test:alerts` | Alert rules, records, history |
| `npm run test:integrations` | Webhooks, system health |
| `npm run test:reports` | Reports, export, performance |
| `npm run test:settings` | Preferences, sessions, users |
| `npm run test:ui` | Playwright interactive UI mode |

`test:webhooks` / `test:export` / `test:rbac` are aliases pointing at the
integrations / reports / settings suites respectively.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `E2E_BASE_URL` | `http://localhost:5173` | Frontend base URL |
| `E2E_START_FRONTEND` | unset | Set to `1` to let Playwright launch the Vite dev server |
| `E2E_ADMIN_USER` | `admin` | Admin username for sign-in |
| `E2E_ADMIN_PASS` | `Admin123` | Admin password for sign-in |

## Test Layout

```text
tests/
  auth/          login and protected-route behavior
  navigation/    protected route inventory and legacy aliases
  smoke/         fast critical-path checks
  nodes/         node, probe, and beacon config flows
  alerts/        alert rules, records, and history
  integrations/  webhooks and system health
  reports/       reports, export, and performance pages
  settings/      preferences, sessions, users
  fixtures/      shared Playwright fixtures (authenticatedPage)
  support/       auth helper and shared selectors
```

The shared `authenticatedPage` fixture performs a real sign-in once and reuses
the session, so suites assert on post-login pages without re-authenticating.

## CI

In CI environments (`CI=true`):
- 2 retries on failure
- 2 workers for throughput
- Traces captured on first retry
- Screenshots and video retained on failure
