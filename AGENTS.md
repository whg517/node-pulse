# AGENTS.md

Guidance for AI agents and contributors working in this repository.

## ⛔ Hard Rules (read before any edit)

1. **Never edit on `main`.** Create a worktree first: `git worktree add -b <type>-<name> .worktree/<type>-<name> main`. All work — including docs-only changes — happens in `.worktree/<type>-<name>/`.
2. **No `chore` type** for branches or commits. Use `feat | fix | docs | refactor | test | perf | build | ci | revert`.
3. **Completion gates are mandatory**, including for docs changes: `golangci-lint` (pulse + beacon), Go build (pulse + beacon), frontend lint, frontend build.
4. **Squash-merge back to `main`**, then remove the worktree.
5. Full procedure: `docs/development-workflow.md` (7 steps). When in doubt, that file wins over this summary.

> **AI agents:** before any edit, read `docs/development-workflow.md` and lay out its 7 steps as a todo list. "User said to commit" means "run the full workflow through squash-merge", not a single `git commit`. Report gate results faithfully — do not claim a failed gate passed.

## Repository Structure

Monorepo for **Node-Pulse**, a distributed network monitoring system:

- **`beacon/`** — Go monitoring agent (TCP/UDP probes, reports metrics)
- **`pulse/`** — Go backend API server (receives metrics, serves frontend)
- **`frontend/`** — React + TypeScript web UI
- **`e2e/`** — Playwright end-to-end tests
- **`docs/`** — Project documentation (PRD, auth/UI design, dev workflow)

## Development Workflow

Authoritative source: `docs/development-workflow.md`. Checklist form (copy into your todo list):

1. **Worktree** — `git worktree add -b <type>-<name> .worktree/<type>-<name> main` (branch from `main`).
2. **Type** — one of `feat | fix | docs | refactor | test | perf | build | ci | revert` (never `chore`).
3. **Commit** — Conventional Commit style, imperative summary, e.g. `docs: simplify AGENTS.md`.
4. **Gates** — run all four before merging (see "Completion Gates" below).
5. **Merge** — `git switch main && git merge --squash <type>-<name> && git commit`.
6. **Cleanup** — `git worktree remove .worktree/<type>-<name> && git branch -D <type>-<name>`.

### Completion Gates

```bash
(cd pulse   && make lint && make build)        # Go lint + build
(cd beacon  && make lint && make build-local)  # Go lint + build
(cd frontend && npm run lint && npm run build) # frontend lint + build
```

If a gate fails on `main` for reasons unrelated to your change, report it honestly and fix it in a separate branch — don't claim it passed.

## Common Commands

The component Makefiles (`pulse/`, `beacon/`) and `frontend/package.json` are the source of truth; the root Makefile only orchestrates them. Per-component targets like `make lint-pulse`, `make build-beacon` also exist.

### Root (mirrors CI)

```bash
make ci-local      # lint + build + test for all components
make lint          # lint pulse + beacon + frontend
make build         # build all components
make tidy          # go mod tidy on both Go modules (regenerates pulse swagger)
make docker-build  # build all production Docker images
make docker-up     # start production compose stack (needs .env)
```

### Beacon (`cd beacon`)

`make build` (Linux AMD64 static) · `make build-local` · `make run` (needs `beacon.yaml`) · `make test` · `make test-coverage` · `make lint` · `make lint-fix` · `make clean`

### Pulse (`cd pulse`)

`make run` (port 6532) · `make build` / `build-debug` · `make test` (full, needs Docker) · `make test-unit` / `test-integration` / `test-quick` · `make setup-test-db` / `cleanup-test-db` · `make migrate-up` / `migrate-down` / `migrate-version` / `migrate-create NAME=` · `make swag` / `swag-force` · `make lint` · `make fmt` · `make help`

Migrations auto-apply on startup; manual targets need the `migrate` CLI (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`). Config via `pulse.yaml` (see `pulse.yaml.example`); `PULSE_`-prefixed env vars override.

### Frontend (`cd frontend`)

`npm install` · `npm run dev` (port 5173) · `npm run build` · `npm run preview` · `npm run test` (Vitest) · `npm run lint`

### E2E (`cd e2e`)

`npm test` (needs Pulse + Frontend + DB) · `npm run test:smoke` / `test:smoke:fast` (only `smoke/` is implemented; auth, rbac, nodes, alerts, webhooks, export, dashboard, performance, reports, sessions, visual suites are planned) · `npm run docker:up` / `docker:down` / `docker:logs` / `docker:reset` (recommended env via `docker-compose.e2e.yml`) · `npm run test:ui` / `test:debug` / `test:headed` / `report`

## Architecture Overview

```
Beacon (per node) ──POST /heartbeat (JWT)──► Pulse Server ──► PostgreSQL
                                               ├── Memory Cache (ring buffer)
                                               ├── Batch Writer (async) ──► DB
                                               ├── Alert Engine (worker pool)
                                               ├── Webhook Dispatcher
                                               └── Scheduler (cleanup, suppression)
```

**Pulse internal packages** (`pulse/internal/`): `alert`, `auth`, `cache`, `cleanup`, `config`, `csrf`, `db` (PostgreSQL + golang-migrate), `diagnostic`, `export`, `health`, `models`, `scheduler`, `security`, `server`, `suppression`, `webhook`.

**Frontend stack**: React 19 + TypeScript 5 + Vite 7, Tailwind CSS 4 (dark mode via `dark:`), React Router v7, Zustand 5, i18next (EN + zh-CN), Recharts 3, Vitest + Testing Library. Source under `frontend/src/`: `api/`, `components/` (`common/`, `layout/`, domain folders), `config/`, `hooks/`, `locales/`, `pages/`, `stores/`, `types/`.

## Important Patterns

- **Error handling**: Beacon retries 5xx/network with exponential backoff (1s, 2s, 4s). Pulse distinguishes retryable (5xx, network) vs non-retryable (4xx). Frontend auto-refreshes JWT on 401 via Axios interceptor; surface errors with `ErrorBanner`.
- **Frontend design system**: Use `PageContainer` + `PageHeader` for layouts; `ErrorBanner` / `ActionButton` / `ConfirmDialog` for standard interactions; `statusColors` / `statusClasses` from `src/config/designTokens.ts`. Always add `dark:` variants. All user-visible strings via `t()` from `useTranslation()`, with keys added to both `en.json` and `zh-CN.json`.
- **Security**: JWT secrets auto-generated (512-bit) if unset; TLS 1.2+ for Beacon→Pulse; SHA-256 hashing for API keys/refresh tokens; rate limiting (5 logins/IP/min); account lockout (5 fails = 10 min); RBAC (admin, operator, viewer, beacon); CSRF on mutations; access tokens in memory only.
- **Performance**: Async batch DB writes (non-blocking heartbeat); ring-buffer cache (60 points/node); 10-worker alert pool; access token 15 min expiry with refresh-on-401; drop-on-overflow metric writes.

## Testing

- **Beacon**: `beacon/internal/*/` and `beacon/tests/` — `make test`.
- **Pulse unit**: `pulse/internal/**/*_test.go`, `pulse/tests/api/`, `pulse/tests/cache/`. **Integration**: `pulse/tests/integration/` (needs Docker DB via `make setup-test-db`).
- **Frontend**: `frontend/src/**/*.test.tsx` and `__tests__/` — `npm run test`.
- **E2E**: `e2e/tests/` — only `smoke/` implemented; see commands above.

## Configuration

- **Beacon** (`beacon.yaml`, see `beacon.yaml.example`): required `pulse_server`, `node_id`, `node_name`, `api_key`; optional `region`, `tags`, `probes`, `reconnect`, `metrics_*` (Prometheus `/metrics` on port 2112 by default). Supports hot-reload.
- **Pulse** (`pulse.yaml`, see `pulse.yaml.example`): `server.port` (6532), `server.mode`, `database.url` (**required**), `log.*`, `cors.allowed_origins`, `admin.*` (default `admin`/`Admin123`), `jwt.*`, `session.*`, `cleanup.*` (retention 7 days). Priority: **env vars (`PULSE_` prefix) > pulse.yaml > defaults**.

## Operational Notes

- All components handle SIGTERM/SIGINT for graceful shutdown.
- Refresh tokens are one-time use; access tokens live in memory only.
- All shared state is mutex-protected.
- Swagger UI at `http://localhost:6532/swagger/index.html` in debug mode.
