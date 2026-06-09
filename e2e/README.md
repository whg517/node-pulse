# Node-Pulse E2E Tests

End-to-end tests for Node-Pulse using [Playwright](https://playwright.dev/).

## Prerequisites

- Node.js 18+
- A running Node-Pulse stack (frontend + backend + database)

## Quick Start

```bash
cd e2e
npm install
npx playwright install --with-deps

# Run smoke tests against a local dev stack
npm run test:smoke
```

## Test Suites

| Script | Description |
|--------|-------------|
| `npm test` | All tests |
| `npm run test:smoke` | Login, dashboard, node list (quick sanity) |
| `npm run test:smoke:fast` | Smoke tests, Chromium only, 4 workers |
| `npm run test:auth` | Auth flows (login/logout/sessions/token refresh) |
| `npm run test:nodes` | Node management, detail, comparison |
| `npm run test:alerts` | Alert rules, records, history |
| `npm run test:webhooks` | Webhook configuration |
| `npm run test:export` | Data export |
| `npm run test:rbac` | Role-based access control |
| `npm run test:ui` | Playwright interactive UI mode |
| `npm run test:headed` | Run with browser visible |
| `npm run report` | View last test report |

## Docker-based E2E Environment

The recommended way to run a full suite is with Docker Compose:

```bash
cd e2e
npm run docker:up     # Start Pulse, Frontend, PostgreSQL
npm test              # Run all tests
npm run docker:down   # Stop containers
```

This uses `docker-compose.e2e.yml` in the project root which sets up:
- PostgreSQL test database (port 5432)
- Pulse backend API (port 6532)
- Frontend dev server (port 5173)

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_URL` | `http://localhost:5173` | Frontend base URL |
| `E2E_ADMIN_USER` | `admin` | Admin username |
| `E2E_ADMIN_PASS` | `Admin123` | Admin password |

## Configuration

See `playwright.config.ts` for full Playwright configuration including browser projects
(Chromium, Firefox, WebKit) and retry/trace settings.

## CI

In CI environments (`CI=true`):
- 2 retries on failure
- Single worker for stability
- Traces captured on first retry
- Screenshots on failure
