# Story 5.5: Alert Engine Implementation

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Pulse 系统,
I need 实时评估告警规则并触发告警,
So that 及时发现网络异常。

## Acceptance Criteria

**Given** 告警规则已配置（enabled = true）
**When** Beacon 心跳数据到达（POST /api/v1/beacon/heartbeat）
**Then** 检查每个指标是否超过配置的阈值
**And** 超过阈值时创建告警事件
**And** 告警事件包含：node_id、metric、threshold、current_value、level、timestamp
**And** 告警引擎异步评估，不阻塞心跳响应
**And** 评估延迟 < 100ms（NFR-OTHER-002: 心跳数据 5 秒内接收并开始处理）

**Given** 告警规则已禁用（enabled = false）
**When** Beacon 心跳数据到达
**Then** 跳过该规则的评估
**And** 不创建告警事件

**Given** 存在全局告警规则（node_id = NULL）
**When** 任意节点的心跳数据到达
**Then** 评估该全局规则
**And** 超过阈值时创建告警事件

**Given** 存在节点特定告警规则（node_id 设置）
**When** 特定节点的心跳数据到达
**Then** 仅评估该节点的规则
**And** 超过阈值时创建告警事件

**Given** 同一节点同时有全局规则和节点特定规则
**When** 心跳数据到达
**Then** 评估所有适用的规则
**And** 每个规则独立创建告警事件

**覆盖需求:** FR5（告警规则配置）

## Tasks / Subtasks

- [x] Task 1: Design alert engine architecture (AC: Given - 告警规则配置)
  - [x] Subtask 1.1: Design async evaluation pipeline to avoid blocking heartbeat
  - [x] Subtask 1.2: Design metric-to-rule matching algorithm
  - [x] Subtask 1.3: Design alert event structure and storage model
  - [x] Subtask 1.4: Define alert engine interfaces and responsibilities
  - [x] Subtask 1.5: Document evaluation flow and performance requirements

- [x] Task 2: Implement alert evaluation engine (AC: When - 心跳数据到达)
  - [x] Subtask 2.1: Create AlertEngine struct with alert querier dependency
  - [x] Subtask 2.2: Implement EvaluateMetrics function
  - [x] Subtask 2.3: Implement metric value comparison logic (latency > threshold, packet_loss_rate > threshold, jitter > threshold)
  - [x] Subtask 2.4: Implement rule filtering (enabled only, node_id matching)
  - [x] Subtask 2.5: Implement async evaluation with goroutines and channels
  - [x] Subtask 2.6: Add context cancellation support for graceful shutdown

- [x] Task 3: Integrate alert engine with heartbeat handler (AC: Then - 异步评估)
  - [x] Subtask 3.1: Modify HandleHeartbeat to trigger async alert evaluation
  - [x] Subtask 3.2: Pass metric data to alert engine via channel
  - [x] Subtask 3.3: Ensure heartbeat response is not blocked by evaluation
  - [x] Subtask 3.4: Add buffering to prevent backpressure on heartbeat endpoint

- [x] Task 4: Implement alert event creation (AC: Then - 创建告警事件)
  - [x] Subtask 4.1: Define AlertEvent model (alert_id, node_id, metric, threshold, current_value, level, timestamp)
  - [x] Subtask 4.2: Create alert_events table migration
  - [x] Subtask 4.3: Implement CreateAlertEvent database function
  - [x] Subtask 4.4: Store alert events in database for historical tracking
  - [x] Subtask 4.5: Add logging for alert event creation

- [x] Task 5: Implement metric comparison logic (AC: Then - 检查阈值)
  - [x] Subtask 5.1: Implement latency threshold comparison (milliseconds)
  - [x] Subtask 5.2: Implement packet_loss_rate threshold comparison (percentage 0-100)
  - [x] Subtask 5.3: Implement jitter threshold comparison (milliseconds)
  - [x] Subtask 5.4: Handle nil/missing metric values gracefully
  - [x] Subtask 5.5: Support all three metric types in single evaluation pass

- [x] Task 6: Implement rule filtering and matching (AC: Given - 全局/节点规则)
  - [x] Subtask 6.1: Query all enabled alert rules from database
  - [x] Subtask 6.2: Filter rules by metric type (latency, packet_loss_rate, jitter)
  - [x] Subtask 6.3: Match node-specific rules (node_id = current node)
  - [x] Subtask 6.4: Include global rules (node_id IS NULL) for all nodes
  - [x] Subtask 6.5: Cache enabled rules to reduce database queries

- [x] Task 7: Add performance monitoring (AC: Then - 评估延迟 < 100ms)
  - [x] Subtask 7.1: Add timing metrics for evaluation duration
  - [x] Subtask 7.2: Monitor goroutine pool utilization
  - [x] Subtask 7.3: Track channel buffer depth
  - [x] Subtask 7.4: Log slow evaluations (>100ms)
  - [x] Subtask 7.5: Add Prometheus metrics for alert engine health

- [x] Task 8: Write comprehensive tests (AC: 完整功能验证)
  - [x] Subtask 8.1: Unit tests for metric comparison logic
  - [x] Subtask 8.2: Unit tests for rule filtering and matching
  - [x] Subtask 8.3: Unit tests for alert event creation
  - [x] Subtask 8.4: Integration tests for alert evaluation pipeline
  - [x] Subtask 8.5: Test async evaluation does not block heartbeat
  - [x] Subtask 8.6: Test global vs node-specific rules
  - [x] Subtask 8.7: Test multiple metrics evaluation
  - [x] Subtask 8.8: Performance tests for evaluation latency

- [x] Task 9: Update documentation and examples (AC: 文档完整性)
  - [x] Subtask 9.1: Document alert engine architecture and flow
  - [x] Subtask 9.2: Add usage examples for alert evaluation
  - [x] Subtask 9.3: Document alert event schema and storage
  - [x] Subtask 9.4: Document performance characteristics and tuning

## Dev Notes

### Epic Analysis

**Epic 5: 告警规则配置与通知** - 系统可以自动检测异常并通过 Webhook 推送告警

**Story Context in Epic:**
- Story 5.1-5.4: **已完成** - 告警规则 API、Webhook 配置 API、前端页面
- Story 5.5: **告警引擎实现** (本故事) - **核心评估引擎**
- Story 5.6: 告警抑制机制（依赖本故事）
- Story 5.7: Webhook 推送实现（依赖本故事）
- Story 5.8: 健康检查扩展（依赖本故事）

**Critical Prerequisites:**
- **Story 5.1 已完成**: 告警规则 API 已实现（alerts 表、AlertQuerier）
- **Story 3.1 已完成**: 心跳数据接收 API 已实现（POST /api/v1/beacon/heartbeat）
- **Story 3.2 已完成**: 内存缓存已实现（MemoryCache、BatchWriter）
- **数据库已配置**: PostgreSQL + pgx 连接池已就绪

### Architecture Alignment

**Alert Engine Architecture** [Source: Architecture.md#API & Communication Patterns]:
```
告警评估流程：
1. Beacon 上报心跳数据 → HandleHeartbeat 接收
2. HandleHeartbeat 存储数据到内存缓存（已有）
3. HandleHeartbeat 异步触发告警评估（新增）
4. AlertEngine 查询启用的告警规则
5. 评估每个指标是否超过阈值
6. 创建告警事件（alert_events 表）
```

**Database Schema** [Source: Architecture.md#Data Models]:
- `alert_events` 表：id (UUID), node_id (UUID), metric (VARCHAR), threshold (DECIMAL), current_value (DECIMAL), level (VARCHAR), created_at (TIMESTAMP)
- 外键：node_id REFERENCES nodes(id)
- 索引：idx_alert_events_node_id, idx_alert_events_metric, idx_alert_events_created_at
- 注意：alert_events 用于存储告警历史，为 Story 5.6/5.7 提供数据源

**Performance Requirements** [Source: NFR-OTHER-002]:
- 心跳数据 5 秒内接收并开始处理
- 评估延迟 < 100ms（不阻塞心跳响应）
- 异步处理：使用 goroutines + channels

**Alert Rule Types** [Source: FR5 - 告警规则配置]:
- **全局规则**（node_id = NULL）：应用于所有节点
- **节点特定规则**（node_id 设置）：仅应用于指定节点
- **规则启用状态**（enabled = false）：跳过评估

### Project Structure Notes

**Alignment with unified project structure** [Source: Architecture.md#Project Structure & Boundaries]:
```
pulse-api/
├── internal/
│   ├── models/
│   │   └── alert_event.go         # NEW - AlertEvent model
│   ├── db/
│   │   ├── alert_events.go         # NEW - AlertEvent database operations
│   │   └── migrations.go           # UPDATE - Add alert_events table
│   ├── alert/
│   │   └── engine.go               # NEW - Alert evaluation engine
│   └── api/
│       └── beacon_handler.go       # UPDATE - Integrate alert engine
├── tests/
│   └── integration/
│       └── alert_engine_integration_test.go  # NEW - Alert engine tests
└── cmd/server/
    └── main.go                     # UPDATE - Initialize alert engine
```

**Detected conflicts or variances:**
- **No conflicts detected**: This is a new feature addition

### Implementation Strategy

**Phase 1: Core Engine (Tasks 1-2)**
- Design async evaluation pipeline
- Implement AlertEngine with evaluation logic
- Use worker pool pattern for concurrent evaluation

**Phase 2: Integration (Tasks 3-4)**
- Integrate with HandleHeartbeat
- Create alert_events table
- Implement alert event storage

**Phase 3: Logic & Filtering (Tasks 5-6)**
- Implement metric comparison
- Implement rule filtering (global vs node-specific)
- Add rule caching for performance

**Phase 4: Monitoring & Testing (Tasks 7-9)**
- Add performance monitoring
- Write comprehensive tests
- Document architecture and usage

### Key Design Decisions

**1. Async Evaluation Pattern**
```go
// Heartbeat handler sends metric data to channel
metricChannel := make(chan *MetricData, 1000)
go alertEngine.RunEvaluation(metricChannel)

// In HandleHeartbeat (non-blocking)
select {
case metricChannel <- metricData:
    // Queued for evaluation
default:
    // Channel full, log warning but don't block
}
```

**2. Rule Filtering Logic**
```go
// Query enabled rules once, cache in memory
enabledRules := alertQuerier.GetEnabledRules(ctx)

// Filter for current node
for _, rule := range enabledRules {
    if rule.NodeID == nil || *rule.NodeID == nodeID {
        // Rule applies to this node
        evaluateRule(rule, metricData)
    }
}
```

**3. Metric Comparison**
```go
func evaluateRule(rule *Alert, metric *MetricData) *AlertEvent {
    var currentValue float64
    var exceedsThreshold bool

    switch rule.Metric {
    case "latency":
        currentValue = metric.LatencyMs
        exceedsThreshold = currentValue > rule.Threshold
    case "packet_loss_rate":
        currentValue = metric.PacketLossRate
        exceedsThreshold = currentValue > rule.Threshold
    case "jitter":
        currentValue = metric.JitterMs
        exceedsThreshold = currentValue > rule.Threshold
    }

    if exceedsThreshold {
        return &AlertEvent{
            NodeID:        metric.NodeID,
            Metric:        rule.Metric,
            Threshold:     rule.Threshold,
            CurrentValue:  currentValue,
            Level:         rule.Level,
        }
    }
    return nil
}
```

**4. Performance Targets**
- Evaluation latency: < 100ms (measured from metric queued to event created)
- Channel buffer: 1000 metrics (prevent backpressure)
- Worker pool: 10 workers (configurable)
- Rule cache refresh: 60 seconds (balance freshness vs performance)

### Testing Strategy

**Unit Tests:**
- Test metric comparison logic with various thresholds
- Test rule filtering (global vs node-specific)
- Test alert event creation and storage
- Test edge cases (nil values, missing metrics, disabled rules)

**Integration Tests:**
- Test full evaluation pipeline (heartbeat → evaluation → event creation)
- Test async evaluation does not block heartbeat response
- Test multiple rules evaluation
- Test concurrent evaluations

**Performance Tests:**
- Measure evaluation latency with 100 rules
- Measure evaluation latency with 1000 concurrent metrics
- Test channel buffer overflow handling
- Verify < 100ms evaluation target

### Dependencies on Other Stories

**Depends On:**
- **Story 5.1** (Alert Rule API): Provides alerts table and AlertQuerier
- **Story 3.1** (Pulse Data Receiving API): Provides heartbeat endpoint
- **Story 3.2** (Pulse Memory Cache): Provides metric data storage

**Required For:**
- **Story 5.6** (Alert Suppression): Needs alert events for suppression logic
- **Story 5.7** (Webhook Push): Needs alert events for webhook notifications
- **Story 5.8** (Health Check): Needs alert engine status for health monitoring
- **Story 6.1** (Alert Record Storage): Will build upon alert_events table

### Non-Functional Requirements

**Performance:**
- NFR-OTHER-002: 心跳数据 5 秒内接收并开始处理
- Evaluation latency < 100ms (internal target)

**Reliability:**
- Alert evaluation must not block heartbeat endpoint
- Graceful degradation if evaluation is slow
- Logging for all alert events

**Scalability:**
- Support 100+ concurrent rules
- Support 10+ nodes reporting simultaneously
- Async processing to handle spikes in metrics

**Observability:**
- Log all alert events with context
- Metrics for evaluation latency and throughput
- Health check integration (Story 5.8)

### Open Questions

1. **Alert Event Retention**: How long to keep alert_events? (30 days per FR7)
2. **Event Deduplication**: Should we deduplicate identical alerts? (Story 5.6 will handle suppression)
3. **Alert Aggregation**: Should we aggregate multiple alerts? (MVP: create separate event per rule)
4. **Rule Cache Invalidation**: When to refresh rule cache? (60 seconds is reasonable default)

### Success Metrics

- ✅ Alert evaluation completes within 100ms
- ✅ Heartbeat endpoint response time < 50ms (not blocked by evaluation)
- ✅ All enabled rules are evaluated for each metric
- ✅ Global and node-specific rules work correctly
- ✅ Alert events are created and stored successfully
- ✅ No memory leaks or goroutine leaks in evaluation pipeline
- ✅ Comprehensive test coverage (>80%)
