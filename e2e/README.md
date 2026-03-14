# Node-Pulse E2E Tests

End-to-end tests for the Node-Pulse distributed network monitoring system.

## Quick Start

```bash
# Install dependencies
npm install

# Install Playwright browsers
npx playwright install

# Run all tests
npm test

# Run smoke tests (fast feedback)
npm run test:smoke

# Run visual regression tests
npm run test:visual
```

### 2. Start Backend Server

```bash
cd pulse
make run
```

The backend should be running on `http://localhost:6532`.

### 3. Start Frontend Server

```bash
cd frontend
npm install
npm run dev
```

The frontend should be running on `http://localhost:5173`.

### 4. Install E2E Dependencies

```bash
cd e2e
npm install
npx playwright install chromium
```

## Running Tests

### Run All Tests

```bash
npm test
```

### Run Specific Test Suites

```bash
# Auth tests
npm run test:auth

# RBAC tests
npm run test:rbac

# Node tests
npm run test:nodes

# Alert tests
npm run test:alerts

# Webhook tests
npm run test:webhooks

# Export tests
npm run test:export
```

### Generate Visual Baselines

Visual regression baselines are handled by a dedicated GitHub Actions workflow.

- Run the Visual Snapshot Baselines workflow manually from GitHub Actions.
- Download the artifact and commit the generated snapshot directories under `e2e/tests/visual`.
- After the baselines are committed, the Visual Regression workflow will validate them on PRs and on pushes to `main`.

### Run with UI Mode

```bash
npm run test:ui
```

### Run in Debug Mode

```bash
npm run test:debug
```

### Run in Headed Mode

```bash
npm run test:headed
```

### View Test Report

```bash
npm run report
```

## Test Structure

```
e2e/
├── playwright.config.ts     # Playwright configuration
├── package.json             # Dependencies and scripts
├── global-setup.ts          # Database seeding and auth state setup
├── global-teardown.ts       # Cleanup after tests
├── fixtures/
│   └── auth.fixture.ts      # Auth fixtures for different roles
├── pages/                   # Page Object Model classes
│   ├── LoginPage.ts
│   ├── DashboardPage.ts
│   ├── NodesPage.ts
│   └── ...
└── tests/
    ├── auth/                # Authentication tests
    ├── rbac/                # Role-based access control tests
    ├── dashboard/           # Dashboard page tests
    ├── nodes/               # Node management tests
    ├── alerts/              # Alert tests
    ├── webhooks/            # Webhook tests
    ├── export/              # Data export tests
    ├── performance/         # Performance dashboard tests
    └── sessions/            # Session management tests
```

## Test Users

The global setup creates test users with different roles:

| Username | Password | Role |
|----------|----------|------|
| admin | Admin123 | admin |
| e2e_operator | E2eOperator123! | operator |
| e2e_viewer | E2eViewer123! | viewer |

## Acceptance Criteria

The tests verify the following acceptance criteria:

### Authentication
- AC-1: Valid login redirects to dashboard
- AC-2: Invalid credentials show error
- AC-3: Account lockout after 5 failed attempts
- AC-4: Rate limiting (5 requests per minute)
- AC-5: Logout redirects to login page
- AC-6: Session restored on page refresh
- AC-7: Expired token auto-refreshes
- AC-8: Cross-tab logout sync

### RBAC
- AC-9: Admin can access webhooks
- AC-10: Operator sees "Admin-only" on webhooks
- AC-11: Operator can CRUD nodes
- AC-12: Operator sees "Admin-only" on export
- AC-13: Viewer sees read-only UI

### CRUD Operations
- AC-14: Admin can create nodes
- AC-15: Operator can create nodes
- AC-16: Admin can create webhooks
- AC-17: Operator can create alert rules
- AC-18: Viewer gets 403 on write operations

### Page Functionality
- AC-19: Dashboard auto-refreshes every 5 seconds
- AC-20: Node detail time range selector works
- AC-21: Admin can submit data export
- AC-22: Session revocation works

## Troubleshooting

### Tests Fail with "Backend not healthy"

Ensure the backend is running:
```bash
cd pulse
make run
```

### Tests Fail with "Auth state not available"

The global setup may have failed. Check:
1. Backend is running
2. Default admin user exists (admin/Admin123)
3. Database is accessible

### Database Connection Errors

Ensure the test database is running:
```bash
cd pulse
docker-compose -f docker-compose.test.yml up -d
```

### Flaky Tests

Some tests may be flaky due to:
- Timing-sensitive operations (auto-refresh, token refresh)
- Cross-tab synchronization
- Rate limiting state

If tests are flaky, try:
1. Running with `--retries=3`
2. Running in serial mode: `--workers=1`
3. Checking network latency

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| TEST_DB_URL | postgresql://testuser:testpass123@localhost:5432/nodepulse_test | Test database URL |
| API_BASE_URL | http://localhost:6532 | Backend API URL |
| FRONTEND_BASE_URL | http://localhost:5173 | Frontend URL |
