# deploy/docker

All docker-compose stacks for Node-Pulse live here. Dockerfiles stay with their
components (`pulse/Dockerfile`, `beacon/Dockerfile`) and are referenced by the
compose files via relative paths.

## Files

- `docker-compose.prod.yml` — single-host production stack: PostgreSQL + Pulse.
  Pulse embeds the frontend (`//go:embed`), so one container serves both the SPA
  and the API. Beacon agents are deployed separately on each monitored node.
- `docker-compose.e2e.yml` — E2E test stack (PostgreSQL + Pulse in debug mode).
  Driven by Playwright on the host via `npm --prefix e2e run docker:up`.
- `docker-compose.test.yml` — standalone PostgreSQL container for the Pulse
  unit/integration test suite (`make setup-test-db` / `cleanup-test-db` in
  `pulse/`).

## Build context

`docker-compose.prod.yml` and `docker-compose.e2e.yml` set the Pulse build
`context: ../..` (the repository root), because `pulse/Dockerfile` copies the
frontend sources (`frontend/`) and embeds them into the Go binary.
`docker-compose.test.yml` has no build section (it only runs a stock postgres
image), so its location does not affect the build.

## Running

Production (from the repository root — `.env` is read from the CWD):

```bash
cp .env.example .env   # then edit secrets
docker compose -f deploy/docker/docker-compose.prod.yml up -d --build
```

E2E (the npm scripts resolve the path from `e2e/`):

```bash
npm --prefix e2e run docker:up
```
