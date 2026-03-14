# Node-Pulse

A distributed network monitoring system designed for overseas infrastructure. Beacon agents deployed on remote nodes collect multi-dimensional network metrics, which are aggregated and visualized by a centralized Pulse platform — enabling automatic differentiation between local node failures, cross-border link issues, and ISP routing anomalies.

## Architecture

```
Beacon (per node)          Pulse Server              PostgreSQL
     │                         │                          │
     │ POST /heartbeat          │                          │
     ├─────────────────────────►│                          │
     │  (JWT auth)              │                          │
     │                          │                          │
     │                          ├──► Memory Cache          │
     │                          │   (ring buffer)          │
     │                          │                          │
     │                          ├──► Batch Writer ────────►│
     │                          │   (async)                │
     │                          │                          │
     │                          ├──► Alert Engine          │
     │                          │   (worker pool)          │
     │                          │                          │
     │                          ├──► Webhook Dispatcher    │
     │                          │                          │
     │                          └──► Scheduler             │
     │                              (cleanup, suppression) │
```

### Components

| Component | Language | Description |
|-----------|----------|-------------|
| **`beacon/`** | Go | Monitoring agent; runs on each node, performs TCP/UDP probes, reports metrics |
| **`pulse/`** | Go | Backend API server; receives metrics, manages nodes, serves the frontend |
| **`frontend/`** | React + TypeScript | Web UI for visualization and management |
| **`e2e/`** | Playwright | End-to-end test suite |

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 15+
- Docker (for integration/E2E tests)

### 1. Start Pulse (Backend)

```bash
cd pulse

# Copy and edit configuration
cp config.example.yaml config.yaml
# Edit config.yaml: set database.url at minimum

# Download dependencies and run
make deps
make run
# Server starts on http://localhost:6532
# Swagger UI: http://localhost:6532/swagger/index.html
```

### 2. Start Frontend

```bash
cd frontend
npm install
npm run dev
# Dev server starts on http://localhost:5173
```

### 3. Deploy a Beacon Agent

```bash
cd beacon

# Copy and edit configuration
cp beacon.yaml.example beacon.yaml
# Edit beacon.yaml: set pulse_server, node_id, node_name, api_key

# Build and run
make build-local
make run
```

## Configuration

### Pulse (`pulse/config.yaml`)

Copy `config.example.yaml` → `config.yaml`. All settings can be overridden with `PULSE_`-prefixed environment variables (e.g. `PULSE_DATABASE_URL`).

| Key | Default | Description |
|-----|---------|-------------|
| `server.port` | `6532` | HTTP server port |
| `server.mode` | `debug` | `debug` or `release` |
| `database.url` | — | PostgreSQL connection string (**required**) |
| `log.level` | `info` | `debug`, `info`, `warn`, `error` |
| `cors.allowed_origins` | `http://localhost:3000,http://localhost:5173` | Allowed CORS origins |
| `admin.username` / `admin.password` | `admin` / `Admin123` | Initial admin credentials |
| `jwt.secret` | *(auto-generated)* | 512-bit secret; auto-generated if empty |
| `session.expiration_hours` | `24` | Session lifetime |
| `cleanup.retention_days` | `7` | Metric data retention |

**Production environment variables:**

```bash
PULSE_DATABASE_URL=postgres://user:pass@host:5432/nodepulse
PULSE_ADMIN_USERNAME=admin
PULSE_ADMIN_PASSWORD=<strong-password>
PULSE_JWT_SECRET=$(openssl rand -hex 32)
PULSE_SESSION_SECRET=$(openssl rand -hex 32)
PULSE_SERVER_MODE=release
```

### Beacon (`beacon/beacon.yaml`)

| Key | Required | Description |
|-----|----------|-------------|
| `pulse_server` | ✅ | Pulse server URL (HTTP/HTTPS) |
| `node_id` | ✅ | Unique node ID (alphanumeric, `-`, `_`) |
| `node_name` | ✅ | Human-readable name |
| `api_key` | ✅ | API key generated from the Pulse UI |
| `region` | — | Region label |
| `tags` | — | Custom string tags |
| `probes` | — | TCP/UDP probe targets |
| `reconnect` | — | Retry config (`max_retries`, `retry_interval`, `backoff`) |
| `metrics_enabled` | — | Expose Prometheus `/metrics` (default: `true`) |
| `metrics_port` | — | Prometheus metrics port (default: `2112`) |

## Development

### Beacon

```bash
cd beacon
make build          # Linux AMD64 static binary
make build-local    # Current platform
make test
make test-coverage
make lint
```

### Pulse

```bash
cd pulse
make build          # → bin/pulse-api
make build-debug    # → bin/pulse-api-debug (optimizations disabled)
make test           # Full suite (starts Docker test DB automatically)
make test-unit      # Unit tests only
make test-integration
make swag           # Regenerate Swagger docs
make lint
make fmt
```

### Frontend

```bash
cd frontend
npm run dev         # Dev server on :5173
npm run build       # Production build
npm run test        # Vitest unit/component tests
npm run lint
```

## Testing

### Unit & Integration Tests

```bash
# Beacon
cd beacon && make test

# Pulse (requires Docker)
cd pulse && make test

# Frontend
cd frontend && npm run test
```

### End-to-End Tests (Playwright)

```bash
cd e2e

# Option A: Docker environment (recommended)
npm run docker:up   # Starts PostgreSQL + Pulse + Frontend
npm test
npm run docker:down

# Option B: Against a running stack
npm install
npm test

# Targeted suites
npm run test:auth
npm run test:nodes
npm run test:alerts
npm run test:smoke       # Quick sanity checks
npm run test:smoke:fast  # Chromium only, 4 workers

# Debugging
npm run test:ui      # Playwright UI mode
npm run test:headed  # Browser visible
npm run report       # HTML report
```

## Docker

### Pulse

```bash
docker build -t node-pulse-api ./pulse
docker run -p 6532:6532 \
  -e PULSE_DATABASE_URL=postgres://user:pass@host:5432/nodepulse \
  -e PULSE_SERVER_MODE=release \
  node-pulse-api
```

### Full E2E Stack

```bash
docker-compose -f e2e/docker-compose.e2e.yml up -d
# Pulse:    http://localhost:6532
# Frontend: http://localhost:5173
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Beacon | Go 1.24, static binary |
| Backend | Go 1.24, Gin, pgx/PostgreSQL |
| Frontend | React 19, TypeScript 5, Vite 7 |
| Styling | Tailwind CSS 4 |
| State | Zustand 5 |
| Charts | ECharts 6 |
| i18n | i18next (EN + zh-CN) |
| Testing | Vitest, Playwright |

## Security Notes

- JWT access tokens are stored **in memory only** (never localStorage)
- JWT secrets are auto-generated (512-bit) if not configured
- API keys are stored as SHA-256 hashes
- Rate limiting: 5 login attempts/IP/minute; 10-minute lockout after 5 failures
- Role-based access control: `admin`, `operator`, `viewer`, `beacon`
- CSRF protection on all mutation endpoints
- TLS 1.2+ enforced for Beacon → Pulse communication

## API Documentation

Swagger UI is available when running in `debug` mode:

```
http://localhost:6532/swagger/index.html
```

Regenerate after API changes:

```bash
cd pulse && make swag
```
