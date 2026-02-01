feat: Implement Alert Engine (Story 5.5)

Implement real-time alert evaluation engine that checks metrics against
configured alert rules and triggers alert events when thresholds are
exceeded.

## Features Implemented

### Core Components
- **Alert Engine** (`internal/alert/engine.go`):
  - Async evaluation with worker pool (10 workers)
  - Buffered channel for metrics (1000 capacity)
  - Non-blocking evaluation to prevent heartbeat backpressure
  - Rule caching with 60-second refresh interval
  - Support for global and node-specific rules
  - Performance monitoring (logs slow evaluations >100ms)
  - Graceful shutdown with context cancellation

- **Alert Events** (`internal/models/alert_event.go`):
  - AlertEvent model with node_id, metric, threshold, current_value, level
  - Database storage for historical tracking

### Database Schema
- **alert_events table**:
  - id (UUID primary key)
  - node_id (foreign key to nodes)
  - metric (latency/packet_loss_rate/jitter)
  - threshold, current_value (DECIMAL)
  - level (P0/P1/P2)
  - created_at (TIMESTAMPTZ)
  - Indexes on node_id, metric, created_at

### API Integration
- **Beacon Heartbeat Handler** (`internal/api/beacon_handler.go`):
  - Trigger async alert evaluation on heartbeat
  - Pass metrics to alert engine via non-blocking channel
  - Maintain heartbeat response performance

### Evaluation Logic
- Compares latency_ms, packet_loss_rate, jitter_ms against thresholds
- Creates alert events when thresholds exceeded
- Evaluates all applicable rules (global + node-specific)
- Skips disabled rules

### Performance
- Target: <100ms evaluation latency
- Worker pool: 10 concurrent workers (configurable)
- Channel buffer: 1000 metrics (configurable)
- Rule cache refresh: 60 seconds (configurable)

## Testing
- Unit tests for alert engine structure
- Integration tests for:
  - Threshold exceeded scenarios
  - Rule scoping (global vs node-specific)
  - Disabled rules
  - Performance (100 metrics)
- Database tests for alert event creation
- Updated beacon_handler_test.go for new parameter

## Files Created
- pulse-api/internal/models/alert_event.go
- pulse-api/internal/db/alert_events.go
- pulse-api/internal/alert/engine.go
- pulse-api/internal/alert/engine_test.go
- pulse-api/tests/integration/alert_engine_integration_test.go
- pulse-api/internal/db/alert_events_test.go

## Files Modified
- pulse-api/internal/db/migrations.go (added alert_events table)
- pulse-api/internal/api/beacon_handler.go (integrated alert engine)
- pulse-api/internal/api/routes.go (initialize alert engine)
- pulse-api/cmd/server/main.go (added alert engine shutdown)
- pulse-api/internal/api/beacon_handler_test.go (updated constructor call)

## Dependencies
- Story 5.1 (Alert Rule API) - provides alert rules and querier
- Story 3.1 (Pulse Data Receiving API) - provides heartbeat endpoint
- Story 3.2 (Pulse Memory Cache) - provides metric storage

## Required For
- Story 5.6 (Alert Suppression) - needs alert events
- Story 5.7 (Webhook Push) - needs alert events
- Story 5.8 (Health Check) - needs alert engine status
- Story 6.1 (Alert Record Storage) - builds on alert_events table

Co-authored-by: BMAD-Auto-Sprint <bmad@node-pulse.dev>
