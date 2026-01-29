# API Integration Tests

This directory contains integration tests for the Pulse API.

## Prerequisites

### Test Database Setup

Integration tests require a PostgreSQL test database. You have two options:

#### Option 1: Docker (Recommended)

```bash
# Start test database using docker-compose
docker-compose -f docker-compose.test.yml up -d

# Run tests
make test-integration

# Cleanup after tests
docker-compose -f docker-compose.test.yml down
```

#### Option 2: Local PostgreSQL

If you have PostgreSQL installed locally:

```bash
# Create test database
createdb nodepulse_test

# Set environment variable (optional, defaults to shown below)
export TEST_DATABASE_URL="postgres://nodepulse:testpass@localhost:5432/nodepulse_test?sslmode=disable"

# Run tests
go test ./tests/api/...
```

## Running Tests

### Run All Integration Tests

```bash
go test ./tests/api/...
```

### Run Specific Test

```bash
# Run beacon heartbeat integration tests
go test ./tests/api/ -run TestBeaconHeartbeatIntegration

# Run with verbose output
go test -v ./tests/api/...
```

### Skip Integration Tests

If no test database is available, integration tests will be automatically skipped.

```bash
# Run with -short flag to skip performance tests
go test -short ./tests/api/...
```

## Test Coverage

Current integration tests:

- `beacon_heartbeat_integration_test.go` - Tests for beacon heartbeat API endpoint
  - Valid request handling
  - Invalid node ID validation
  - Metric range validation
  - Performance testing

## Adding New Tests

When adding new integration tests:

1. Follow the naming convention: `<feature>_integration_test.go`
2. Use the `testDBPool(t)` helper to get a database connection
3. Clean up test data after each test (use `defer cleanup...`)
4. Use `t.Skip()` if test database is not available
5. Add documentation to this README

## Troubleshooting

### Connection Refused

Ensure PostgreSQL is running:
```bash
# Docker
docker ps

# Local
pg_ctl status
```

### Authentication Failed

Check database credentials in test files or environment variables.

### Timeout Errors

Increase context timeout in the `testDBPool` function if needed.
