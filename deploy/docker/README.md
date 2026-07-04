# deploy/docker

Deployment compose files for Node-Pulse. Dockerfiles live with their components
(`pulse/Dockerfile`, `beacon/Dockerfile`) and are referenced by the compose files
via relative paths.

## Files

- `docker-compose.prod.yml` — single-host production stack: PostgreSQL + Pulse.
  Pulse embeds the frontend (`//go:embed`), so one container serves both the SPA
  and the API. Beacon agents are deployed separately on each monitored node.

## Build context

`docker-compose.prod.yml` sets the Pulse build `context: ../..` (the repository
root), because `pulse/Dockerfile` copies the frontend sources (`frontend/`) and
embeds them into the Go binary. Run compose commands from the repository root:

```bash
cp .env.example .env   # then edit secrets
docker compose -f deploy/docker/docker-compose.prod.yml up -d --build
```

The `.env` file is read from the current working directory, so keep it at the
repository root.
