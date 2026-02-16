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

### Authentication

**Beacon Authentication (JWT with API Key):**
1. Beacon configured with `api_key` in `beacon.yaml`
2. Beacon calls `POST /api/v1/beacon/token` with API key
3. Pulse validates SHA-256 hash of API key, returns JWT (15min expiration, role="beacon")
4. Beacon includes JWT in `Authorization: Bearer <token>` header for heartbeat requests
5. Beacon auto-refreshes token before expiration

**User Authentication (JWT with Refresh Tokens):**
1. User login via `POST /api/v1/auth/login` with username/password
2. Pulse validates bcrypt password hash, returns access token (15min) + sets refresh token cookie (7 days)
3. Frontend stores access token in Zustand store (`src/stores/authStore.ts`)
4. API client (`src/api/client.ts`) includes token in requests
5. On 401, frontend auto-refreshes via `POST /api/v1/auth/refresh`
6. Refresh tokens use rotation (one-time use, consumed after refresh)

### Key Components

**Beacon (`beacon/internal/`):**
- `config/` - YAML configuration with hot-reload support
- `probe/` - TCP/UDP probe scheduler with metrics calculation (RTT, jitter, packet loss)
- `monitor/` - Resource monitoring with adaptive degradation (reduces probe frequency under load)
- `auth/jwt_client.go` - JWT token lifecycle management
- `reporter/heartbeat_reporter.go` - Sends metrics to Pulse every 60 seconds

**Pulse (`pulse/internal/`):**
- `api/` - HTTP handlers for all endpoints
  - `beacon_handler.go` - `POST /api/v1/beacon/heartbeat` (requires beacon JWT)
  - `beacon_token_handler.go` - `POST /api/v1/beacon/token` (API key auth)
  - `routes.go` - Route registration with RBAC middleware
- `auth/` - JWT service and authentication handlers
  - `jwt_service.go` - Token generation (access: 15min, refresh: 7 days)
  - `auth_handler.go` - Login, refresh, logout, me endpoints
- `cache/` - High-performance metric storage
  - `memory_cache.go` - Ring buffer (60 points/node) for real-time queries
  - `batch_writer.go` - Async batch writes to PostgreSQL (buffer 1000, batch 100)
- `alert/engine.go` - Worker pool for real-time alert evaluation
- `models/` - Data models (Node, Probe, Alert, BeaconToken, User, etc.)

**Frontend (`frontend/src/`):**
- `api/client.ts` - Unified API client with JWT refresh interceptor
- `stores/` - Zustand state management (auth, nodes, alerts, dashboard, webhooks)
- `pages/` - Route components (dashboard, nodes, alerts, webhooks)
- `components/` - Reusable UI components

## Database

PostgreSQL is used for persistent storage. Key tables:
- `nodes` - Monitored nodes with API key hashes
- `probes` - Probe configurations per node
- `metric_records` - Time-series metrics (written by batch writer)
- `users` - User accounts with bcrypt password hashes
- `refresh_tokens` - Refresh token hashes for user sessions
- `beacon_tokens` - API key hashes for beacon authentication
- `alerts` - Alert rule definitions
- `alert_records` - Alert trigger history
- `webhooks` - Webhook configurations for alert notifications

## API Endpoints Summary

**Public:**
- `GET /health` - Health check
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/beacon/token` - Exchange API key for JWT (beacon auth)

**Authenticated (All roles):**
- `GET /api/v1/nodes` - List nodes
- `GET /api/v1/probes` - List probes
- `GET /api/v1/alerts/rules` - List alert rules
- `GET /api/v1/data/metrics` - Get real-time metrics (from memory cache)
- `GET /api/v1/data/history` - Get historical metrics (from PostgreSQL)

**Admin/Operator:**
- `POST/PUT/DELETE /api/v1/nodes` - Manage nodes
- `POST/PUT/DELETE /api/v1/probes` - Manage probes
- `POST/PUT/DELETE /api/v1/alerts/rules` - Manage alert rules

**Admin only:**
- `GET /api/v1/config` - Get configuration
- `POST /api/v1/data/export` - Create data export
- `POST/PUT/DELETE /api/v1/webhooks` - Manage webhooks

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
