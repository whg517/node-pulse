# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Structure

This is a monorepo for **Node-Pulse**, a distributed network monitoring system with three main components:

- **`beacon/`** - Go-based monitoring agent that runs on nodes, performs TCP/UDP probes, and reports metrics
- **`pulse/`** - Go-based backend API server that receives metrics, manages nodes, and serves the frontend
- **`frontend/`** - React + TypeScript web UI for visualization and management

## Common Commands

### Beacon (Monitoring Agent)

```bash
cd beacon

# Build for Linux AMD64
make build

# Build for current platform
make build-local

# Run tests
make test
make test-coverage

# Run linter
make lint
make lint-fix

# Run the beacon
make run
```

Beacon requires a `beacon.yaml` configuration file. See `beacon.yaml.example` for reference.

### Pulse (Backend API Server)

```bash
cd pulse

# Download dependencies
make deps

# Run the server (default port 6532)
make run

# Build binary
make build

# Run tests (requires Docker for test database)
make test              # Full test suite with DB setup/teardown
make test-unit         # Unit tests only
make test-integration  # Integration tests only
make test-quick        # All tests (assumes DB is already running)

# Generate Swagger documentation
make swag

# Run linter
make lint

# Format code
make fmt
```

Configuration is via `.env` file. Copy `.env.example` and set required variables (especially `PULSE_DATABASE_URL`).

### Frontend (React)

```bash
cd frontend

# Install dependencies
npm install

# Run dev server (default port 5173)
npm run dev

# Build for production
npm run build

# Run tests
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
```

## Important Patterns

### Error Handling
- Beacon: Exponential backoff retry (1s, 2s, 4s) for 5xx and network errors
- Pulse: Distinguishes retryable (5xx, network) vs non-retryable (4xx) errors
- Frontend: Custom error classes with automatic JWT refresh on 401

### Security
- JWT secrets auto-generated (512-bit) if not provided
- TLS 1.2+ enforcement for Beacon→Pulse communication
- SHA-256 hashing for API keys and refresh tokens
- Rate limiting: 5 login attempts per IP per minute
- Account lockout: 5 failed attempts = 10 minute lockout
- Role-based access control (admin, operator, viewer, beacon)

### Performance
- Async batch writes to DB (non-blocking for heartbeat endpoint)
- Memory cache for real-time queries (ring buffer, 60 points per node)
- Worker pools for alert evaluation (10 workers)
- Pre-fetch token refresh (2 minutes before expiration)
- Non-blocking metric writes (drop on overflow)

## Testing

### Beacon Tests
Located in `beacon/internal/*/test.go` files. Run with `make test`.

### Pulse Tests
- Unit tests: `pulse/internal/**/*_test.go`
- Integration tests: `pulse/tests/integration/`
- Requires Docker test database: `make setup-test-db`

### Frontend Tests
Located in `frontend/src/**/*.test.tsx`. Run with `npm run test`.

## Configuration

### Beacon Configuration (`beacon.yaml`)
Required:
- `pulse_server` - Pulse server URL
- `node_id` - Unique node identifier
- `node_name` - Human-readable name
- `api_key` - API key for authentication

Optional:
- `region`, `tags`, `probes`, `reconnect`, `metrics_enabled`, `logging`

### Pulse Environment Variables (`.env`)
Required:
- `PULSE_DATABASE_URL` - PostgreSQL connection string

Optional (with defaults):
- `PULSE_SERVER_PORT=6532`
- `PULSE_LOG_LEVEL=info`
- `PULSE_CORS_ALLOWED_ORIGINS=http://localhost:5173`
- `PULSE_ADMIN_USERNAME=admin`, `PULSE_ADMIN_PASSWORD=Admin123`
- `PULSE_JWT_SECRET`, `PULSE_SESSION_SECRET` (auto-generated if empty)

## Development Notes

1. Configuration priority: Environment variables > config file > defaults
2. Beacon supports config hot-reload (no restart required)
3. Graceful shutdown: All components handle SIGTERM/SIGINT
4. Token rotation: Refresh tokens are one-time use
5. Concurrent safety: All shared state protected by mutexes
6. Prometheus metrics: Beacon exposes `/metrics` on port 2112
