# Story 5.6: Alert Suppression Mechanism

Status: done

<!-- Note: Validation is optional. Run validate-create-story for validdev-story. -->

## Story

As a Pulse 系统,
I need 实现告警抑制，避免同一节点同一类型异常重复推送,
So that 运维人员不会收到重复的告警通知。

## Acceptance Criteria

**Given** 告警引擎已实现（Story 5.5）
**When** 同一节点同一类型异常发生
**Then** 检查是否在 5 分钟抑制窗口内
**And** 如果在窗口内则抑制新告警
**And** 如果不在窗口内则触发新告警并重置窗口
**And** 抑制机制按 node_id 和 metric 类型分别记录

**Given** 告警事件已创建
**When** 需要检查是否应该抑制该告警
**Then** 查询 alert_suppressions 表检查是否存在活跃抑制记录
**And** 比较当前时间与 suppressed_until 时间戳
**And** 如果当前时间 < suppressed_until，则抑制告警
**And** 如果当前时间 >= suppressed_until 或记录不存在，则不抑制

**Given** 告警触发（未抑制）
**When** 告警事件创建成功
**Then** 创建或更新抑制记录，设置 suppressed_until 为当前时间 + 5 分钟
**And** 记录按 (node_id, metric) 组合键存储
**And** 允许后续告警触发

**覆盖需求:** FR5（告警抑制）、NFR-OTHER-003（告警抑制机制）

## Tasks / Subtasks

- [ ] Task 1: Design alert suppression architecture (AC: Given - 告警引擎已实现)
  - [ ] Subtask 1.1: Design suppression check interface
  - [ ] Subtask 1.2: Design suppression record lifecycle (create, update, expire)
  - [ ] Subtask 1.3: Design integration point with alert engine
  - [ ] Subtask 1.4: Define suppression window configuration (5 minutes default)
  - [ ] Subtask 1.5: Document suppression logic flow

- [ ] Task 2: Implement alert_suppressions table migration (AC: 创建表)
  - [ ] Subtask 2.1: Define alert_suppressions table schema
  - [ ] Subtask 2.2: Add id (UUID primary key)
  - [ ] Subtask 2.3: Add node_id (UUID foreign key)
  - [ ] Subtask 2.4: Add metric (VARCHAR) - latency, packet_loss_rate, jitter
  - [ ] Subtask 2.5: Add suppressed_until (TIMESTAMPTZ)
  - [ ] Subtask 2.6: Add created_at, updated_at timestamps
  - [ ] Subtask 2.7: Create unique index on (node_id, metric)
  - [ ] Subtask 2.8: Create index on suppressed_until for cleanup

- [ ] Task 3: Implement AlertSuppression model and DTOs (AC: 抑制记录)
  - [ ] Subtask 3.1: Create AlertSuppression model struct
  - [ ] Subtask 3.2: Add database tags for ORM mapping
  - [ ] Subtask 3.3: Create CreateSuppressionRequest DTO
  - [ ] Subtask 3.4: Create UpdateSuppressionRequest DTO
  - [ ] Subtask 3.5: Create SuppressionData DTO for responses

- [ ] Task 4: Implement suppression database operations (AC: Then - 查询抑制记录)
  - [ ] Subtask 4.1: Create AlertSuppressionsQuerier interface
  - [ ] Subtask 4.2: Implement CheckSuppression function (query by node_id, metric)
  - [ ] Subtask 4.3: Implement CreateSuppression function (insert or upsert)
  - [ ] Subtask 4.4: Implement UpdateSuppression function (update suppressed_until)
  - [ ] Subtask 4.5: Implement DeleteExpiredSuppressions function (cleanup)
  - [ ] Subtask 4.6: Add context for timeout handling

- [ ] Task 5: Implement suppression service (AC: When - 同一节点同一类型异常发生)
  - [ ] Subtask 5.1: Create SuppressionService with querier dependency
  - [ ] Subtask 5.2: Implement ShouldSuppress function (check + update logic)
  - [ ] Subtask 5.3: Implement RecordSuppression function (set suppression window)
  - [ ] Subtask 5.4: Add suppression window configuration (5 minutes)
  - [ ] Subtask 5.5: Handle concurrent suppression checks
  - [ ] Subtask 5.6: Log suppression decisions for debugging

- [ ] Task 6: Integrate suppression with alert engine (AC: Then - 检查是否抑制)
  - [ ] Subtask 6.1: Modify AlertEngine to use SuppressionService
  - [ ] Subtask 6.2: Check suppression before creating alert events
  - [ ] Subtask 6.3: Record suppression when alert triggers
  - [ ] Subtask 6.4: Log suppressed alerts for audit trail
  - [ ] Subtask 6.5: Handle suppression service errors gracefully

- [ ] Task 7: Implement suppression cleanup job (AC: 持久化维护)
  - [ ] Subtask 7.1: Create cleanup function for expired suppressions
  - [ ] Subtask 7.2: Register cleanup task with scheduler
  - [ ] Subtask 7.3: Run cleanup hourly (configurable interval)
  - [ ] Subtask 7.4: Log cleanup statistics (records deleted)

- [ ] Task 8: Write comprehensive tests (AC: 完整功能验证)
  - [ ] Subtask 8.1: Unit tests for suppression service logic
  - [ ] Subtask 8.2: Unit tests for suppression database operations
  - [ ] Subtask 8.3: Integration tests for suppression check flow
  - [ ] Subtask 8.4: Test suppression window expiration
  - [ ] Subtask 8.5: Test concurrent suppression handling
  - [ ] Subtask 8.6: Test cleanup job functionality
  - [ ] Subtask 8.7: Test alert engine integration
  - [ ] Subtask 8.8: Test suppression logging

- [ ] Task 9: Update documentation and examples (AC: 文档完整性)
  - [ ] Subtask 9.1: Document suppression mechanism design
  - [ ] Subtask 9.2: Document suppression window configuration
  - [ ] Subtask 9.3: Add usage examples for suppression logic
  - [ ] Subtask 9.4: Document cleanup job configuration

## Dev Notes

### Epic Analysis

**Epic 5: 告警规则配置与通知** - 系统可以自动检测异常并通过 Webhook 推送告警

**Story Context in Epic:**
- Story 5.1-5.5: **已完成** - 告警规则、Webhook 配置、前端页面、告警引擎
- Story 5.6: **告警抑制机制** (本故事) - **防止重复告警推送**
- Story 5.7-5.8: **后续功能** - Webhook 推送、健康检查扩展

**Critical Prerequisites:**
- **Story 5.5 已完成**: 告警引擎已实现（alert_events 表、AlertEngine）
- **Story 5.1 已完成**: 告警规则 API 已实现（alerts 表、AlertQuerier）
- **数据库已配置**: PostgreSQL + pgx 连接池已就绪
- **调度器已配置**: Story 3.12 的 scheduler 可用于清理任务

### Architecture Alignment

**Alert Suppression Architecture** [Source: Architecture.md#Data Models]:
```
告警抑制流程：
1. 告警引擎检测到指标超过阈值
2. 调用 SuppressionService.ShouldSuppress(nodeID, metric)
3. 查询 alert_suppressions 表检查是否存在活跃抑制记录
4. 如果存在且未过期 (current_time < suppressed_until): 抑制告警，返回 true
5. 如果不存在或已过期: 不抑制，返回 false
6. 告警触发后，调用 RecordSuppression(nodeID, metric, 5分钟)
7. 创建或更新抑制记录，设置 suppressed_until = NOW() + 5 minutes
```

**Database Schema** [Source: Architecture.md#Data Models]:
- `alert_suppressions` 表：id (UUID), node_id (UUID), metric (VARCHAR), suppressed_until (TIMESTAMPTZ), created_at (TIMESTAMPTZ), updated_at (TIMESTAMPTZ)
- 外键：node_id REFERENCES nodes(id) ON DELETE CASCADE
- 唯一约束：UNIQUE(node_id, metric) - 每个节点的每个指标只能有一条抑制记录
- 索引：idx_suppressions_node_metric (node_id, metric), idx_suppressions_until (suppressed_until)

**Suppression Window** [Source: NFR-OTHER-003]:
- 同一节点同一类型异常 5 分钟内仅推送一次
- 抑制窗口从告警触发时开始计时
- 窗口过期后可再次触发告警

**Alert Event Flow** [Source: Story 5.5 Implementation]:
```
Beacon Heartbeat → Alert Engine → Evaluate Metrics
  → Check Suppression (NEW)
    → If suppressed: Log and skip
    → If not suppressed: Create Alert Event → Record Suppression
```

### Project Structure Notes

**Alignment with unified project structure** [Source: Architecture.md#Project Structure & Boundaries]:
```
pulse-api/
├── internal/
│   ├── models/
│   │   └── alert_suppression.go         # NEW - AlertSuppression model
│   ├── db/
│   │   └── alert_suppressions.go         # NEW - Suppression database operations
│   ├── suppression/
│   │   └── service.go                    # NEW - Suppression service logic
│   ├── alert/
│   │   └── engine.go                     # UPDATE - Integrate suppression check
│   └── db/
│       └── migrations.go                 # UPDATE - Add alert_suppressions table
├── tests/
│   └── integration/
│       └── alert_suppression_integration_test.go  # NEW - Suppression tests
└── cmd/server/
    └── main.go                           # UPDATE - Register cleanup job
```

**Detected conflicts or variances:**
- **No conflicts detected**: This is a new feature addition building on Story 5.5

### Implementation Strategy

**Phase 1: Data Layer (Tasks 2-4)**
- Create alert_suppressions table
- Implement model and DTOs
- Implement database operations (check, create, update, delete)

**Phase 2: Service Layer (Tasks 5, 7)**
- Implement SuppressionService with business logic
- ShouldSuppress: check if alert should be suppressed
- RecordSuppression: create/update suppression record
- Implement cleanup job for expired records

**Phase 3: Integration (Task 6)**
- Integrate suppression check into AlertEngine
- Check suppression before creating alert events
- Record suppression when alerts trigger
- Log suppression decisions

**Phase 4: Testing & Documentation (Tasks 8-9)**
- Unit tests for service and database operations
- Integration tests for suppression flow
- Documentation and examples

### Key Design Decisions

**1. Suppression Window Duration**
```go
const DefaultSuppressionWindow = 5 * time.Minute
```
- 5 minutes default as per NFR-OTHER-003
- Configurable for future flexibility
- Per (node_id, metric) combination

**2. Suppression Check Logic**
```go
func (s *SuppressionService) ShouldSuppress(ctx context.Context, nodeID, metric string) (bool, error) {
    suppression, err := s.querier.CheckSuppression(ctx, nodeID, metric)
    if err != nil {
        return false, err // On error, don't suppress to avoid missing alerts
    }

    if suppression == nil {
        return false, nil // No suppression record, don't suppress
    }

    // Check if still within suppression window
    if time.Now().Before(suppression.SuppressedUntil) {
        return true, nil // Still suppressed
    }

    return false, nil // Suppression expired, don't suppress
}
```

**3. Suppression Recording**
```go
func (s *SuppressionService) RecordSuppression(ctx context.Context, nodeID, metric string, window time.Duration) error {
    suppressedUntil := time.Now().Add(window)

    // Use upsert (INSERT ... ON CONFLICT UPDATE)
    return s.querier.CreateOrUpdateSuppression(ctx, nodeID, metric, suppressedUntil)
}
```

**4. Alert Engine Integration**
```go
// In AlertEngine.evaluateMetric
if alertEvent := e.evaluateRule(rule, data); alertEvent != nil {
    // Check suppression before creating alert event
    suppressed, err := e.suppressionService.ShouldSuppress(ctx, data.NodeID, rule.Metric)
    if err != nil {
        slog.Error("Failed to check suppression", "error", err)
        // Continue with alert creation
    } else if suppressed {
        slog.Info("Alert suppressed",
            "node_id", data.NodeID,
            "metric", rule.Metric,
            "level", rule.Level)
        return nil // Don't create alert event
    }

    // Create alert event
    err = e.alertEventsQuerier.CreateAlertEvent(ctx, alertEvent)
    if err != nil {
        slog.Error("Failed to create alert event", "error", err)
        return nil
    }

    // Record suppression for future alerts
    err = e.suppressionService.RecordSuppression(ctx, data.NodeID, rule.Metric, 5*time.Minute)
    if err != nil {
        slog.Error("Failed to record suppression", "error", err)
    }

    slog.Info("Alert event created", ...)
}
```

**5. Cleanup Job**
```go
func CleanupExpiredSuppressions(ctx context.Context, querier AlertSuppressionsQuerier) error {
    deleted, err := querier.DeleteExpiredSuppressions(ctx)
    if err != nil {
        return err
    }

    slog.Info("Cleaned up expired suppressions", "count", deleted)
    return nil
}

// Register with scheduler to run hourly
cleanupTask := &cleanup.SuppressionCleanupTask{querier: suppressionQuerier}
sched.RegisterTask(cleanupTask)
```

### Testing Strategy

**Unit Tests:**
- Test suppression check logic (within window, expired window, no record)
- Test suppression recording (create new, update existing)
- Test cleanup job functionality
- Test error handling (database errors, timeout)

**Integration Tests:**
- Test full suppression flow (check → suppress/trigger → record)
- Test multiple alerts within suppression window
- Test suppression expiration
- Test concurrent suppression checks
- Test cleanup job execution

**Edge Cases:**
- Database errors during suppression check (should not suppress)
- Concurrent alert evaluations for same node/metric
- Node deleted while suppression active (CASCADE delete)
- Suppression record manually deleted

### Dependencies on Other Stories

**Depends On:**
- **Story 5.5** (Alert Engine): Requires AlertEngine integration point
- **Story 5.1** (Alert Rule API): Requires alert rules to be configured
- **Story 3.12** (Scheduled Data Cleanup): Requires scheduler for cleanup job

**Required For:**
- **Story 5.7** (Webhook Push): Suppression affects which alerts trigger webhooks
- **Story 5.8** (Health Check): Suppression service health can be monitored
- **Story 6.1** (Alert Record Storage): Suppressed alerts should not appear in records

### Non-Functional Requirements

**Reliability:**
- NFR-OTHER-003: 同一节点同一类型异常 5 分钟内仅推送一次
- Database errors should not cause alert suppression (fail open)
- Suppression recording failures should not affect alert creation

**Performance:**
- Suppression check should be fast (<10ms)
- Database query with indexed lookup (node_id, metric)
- Suppression recording is async (non-blocking for alert creation)

**Maintainability:**
- Configurable suppression window duration
- Cleanup job prevents table bloat
- Logging for debugging suppression decisions
- Metrics for suppression rate monitoring

### Open Questions

1. **Suppression Window Duration**: Should 5 minutes be configurable? (MVP: 5 minutes hardcoded, future: configurable)
2. **Global Rules Suppression**: Should global rules use per-node suppression or global suppression? (MVP: per-node)
3. **Suppression Events**: Should we log suppressed alerts as events or just slog? (MVP: slog only)
4. **Manual Suppression**: Should admins be able to manually suppress alerts? (MVP: no, automatic only)

### Success Metrics

- ✅ Same node + same metric + within 5 minutes = suppressed
- ✅ Same node + same metric + after 5 minutes = new alert
- ✅ Different node + same metric = no suppression
- ✅ Same node + different metric = no suppression
- ✅ Suppression check latency <10ms
- ✅ Database errors don't cause suppression (fail open)
- ✅ Cleanup job runs successfully
- ✅ Comprehensive test coverage (>80%)
