# Story 5.8: Health Check Extension

Status: ready-for-dev

## Story

As a 运维人员,
I need 查看告警系统的健康状态,
So that 可以确保告警功能正常运行。

## Acceptance Criteria

**Given** 告警引擎已启动并运行
**When** 访问健康检查端点 GET /health
**Then** 返回整体状态为 "healthy" 或 "unhealthy"
**And** 包含告警引擎状态检查 (alert_engine)
**And** 告警引擎状态包含: cached_rules, rule_cache_last_refresh, metric_channel_depth
**And** 包含 Webhook 推送成功率检查 (webhook_delivery)
**And** Webhook 推送成功率基于最近 100 条推送记录
**And** 成功率计算: success_count / total_count * 100
**And** 包含告警抑制服务状态 (alert_suppression)
**And** 告警抑制状态显示活跃抑制记录数量
**And** 健康检查响应时间 < 1 秒
**And** 所有检查项返回 "ok" 或具体错误信息

**覆盖需求:** NFR-AVAIL-001 (系统可用性 ≥99.5%)、NFR-REL-002 (错误率 <5%)

## Tasks / Subtasks

- [ ] Task 1: Extend health check response to include alert system status (AC: Then - 返回整体状态)
  - [ ] Subtask 1.1: Add AlertSystemStatus struct to health response
  - [ ] Subtask 1.2: Update HealthResponse to include AlertSystem field
  - [ ] Subtask 1.3: Define AlertSystemStatus schema with alert_engine, webhook_delivery, alert_suppression

- [ ] Task 2: Implement alert engine health check (AC: And - 包含告警引擎状态检查)
  - [ ] Subtask 2.1: Add AlertEngine to HealthChecker struct
  - [ ] Subtask 2.2: Implement checkAlertEngine function
  - [ ] Subtask 2.3: Call engine.GetStats() to retrieve metrics
  - [ ] Subtask 2.4: Validate rule cache freshness (last refresh < 5 minutes)
  - [ ] Subtask 2.5: Check metric channel depth (should not be at capacity)
  - [ ] Subtask 2.6: Return status: "ok", "stale", or "full"

- [ ] Task 3: Implement webhook delivery health check (AC: And - 包含 Webhook 推送成功率检查)
  - [ ] Subtask 3.1: Add WebhookLogsQuerier to HealthChecker struct
  - [ ] Subtask 3.2: Implement checkWebhookDelivery function
  - [ ] Subtask 3.3: Query last 100 webhook logs from database
  - [ ] Subtask 3.4: Calculate success rate: success_count / total_count
  - [ ] Subtask 3.5: Determine health status: "healthy" if ≥95%, "degraded" if ≥80%, "unhealthy" if <80%
  - [ ] Subtask 3.6: Return status with success_rate percentage

- [ ] Task 4: Implement alert suppression health check (AC: And - 包含告警抑制服务状态)
  - [ ] Subtask 4.1: Add AlertSuppressionsQuerier to HealthChecker struct
  - [ ] Subtask 4.2: Implement checkAlertSuppression function
  - [ ] Subtask 4.3: Query count of active suppression records
  - [ ] Subtask 4.4: Return status: "ok" with active_suppression_count
  - [ ] Subtask 4.5: Handle database errors gracefully (fail-open)

- [ ] Task 5: Update health check handler to include alert system checks (AC: Given - 告警引擎已启动)
  - [ ] Subtask 5.1: Modify Handler() to call alert system checks
  - [ ] Subtask 5.2: Set overall status to "unhealthy" if any critical check fails
  - [ ] Subtask 5.3: Set overall status to "degraded" if webhook delivery <95%
  - [ ] Subtask 5.4: Ensure health check completes within 1 second
  - [ ] Subtask 5.5: Add timeout context for individual checks

- [ ] Task 6: Write comprehensive tests (AC: 完整功能验证)
  - [ ] Subtask 6.1: Unit tests for checkAlertEngine (ok, stale, full scenarios)
  - [ ] Subtask 6.2: Unit tests for checkWebhookDelivery (healthy, degraded, unhealthy)
  - [ ] Subtask 6.3: Unit tests for checkAlertSuppression (various counts)
  - [ ] Subtask 6.4: Integration test with full health check endpoint
  - [ ] Subtask 6.5: Test health check response time (< 1s)
  - [ ] Subtask 6.6: Test database failure handling (graceful degradation)

- [ ] Task 7: Update main.go to wire alert system dependencies (AC: 集成到现有系统)
  - [ ] Subtask 7.1: Pass AlertEngine to health checker
  - [ ] Subtask 7.2: Pass WebhookLogsQuerier to health checker
  - [ ] Subtask 7.3: Pass AlertSuppressionsQuerier to health checker
  - [ ] Subtask 7.4: Update health.New() call with new parameters

- [ ] Task 8: Update documentation (AC: 文档完整性)
  - [ ] Subtask 8.1: Document health check endpoint response schema
  - [ ] Subtask 8.2: Document health status thresholds and meanings
  - [ ] Subtask 8.3: Add API documentation for /health endpoint
  - [ ] Subtask 8.4: Document monitoring and alerting recommendations

## Dev Notes

### Epic Analysis

**Epic 5: 告警规则配置与通知** - 系统可以自动检测异常并通过 Webhook 推送告警

**Story Context in Epic:**
- Story 5.1-5.7: **已完成** - 告警规则 API、Webhook 配置、前端页面、告警引擎、告警抑制、Webhook 推送
- Story 5.8: **健康检查扩展** (本故事) - **Epic 5 最终故事**

### Technical Context

**Existing Infrastructure:**
- Health check system at `/internal/health/health.go`
- Current checks: database, scheduler
- AlertEngine with `GetStats()` method returning metrics
- WebhookLogs table for delivery status tracking
- AlertSuppressions table for active suppression records

**Health Check Extension Design:**

```go
type AlertSystemStatus struct {
    AlertEngine         *AlertEngineStatus         `json:"alert_engine,omitempty"`
    WebhookDelivery     *WebhookDeliveryStatus     `json:"webhook_delivery,omitempty"`
    AlertSuppression    *AlertSuppressionStatus    `json:"alert_suppression,omitempty"`
}

type AlertEngineStatus struct {
    Status               string  `json:"status"` // ok, stale, full
    CachedRules          int     `json:"cached_rules"`
    RuleCacheLastRefresh string  `json:"rule_cache_last_refresh"`
    MetricChannelDepth   int     `json:"metric_channel_depth"`
    MetricChannelCapacity int    `json:"metric_channel_capacity"`
}

type WebhookDeliveryStatus struct {
    Status      string  `json:"status"` // healthy, degraded, unhealthy
    SuccessRate float64 `json:"success_rate"` // 0-100
    TotalCount  int     `json:"total_count"`
    SuccessCount int    `json:"success_count"`
}

type AlertSuppressionStatus struct {
    Status                string `json:"status"` // ok
    ActiveSuppressionCount int64  `json:"active_suppression_count"`
}
```

**Health Status Logic:**

1. **Alert Engine:**
   - "ok": Rule cache fresh (<5 min) AND channel not full
   - "stale": Rule cache stale (>5 min)
   - "full": Metric channel at capacity
   - "error": Database/other error

2. **Webhook Delivery:**
   - "healthy": Success rate ≥95%
   - "degraded": Success rate ≥80% AND <95%
   - "unhealthy": Success rate <80%
   - "nodata": No webhook logs in last 100 records

3. **Alert Suppression:**
   - "ok": Query succeeded (return count)
   - "error": Database error (fail-open)

**Overall Health Determination:**
- "healthy": All checks "ok"/"healthy"
- "degraded": Webhook delivery "degraded" OR alert engine "stale"
- "unhealthy": Any check "unhealthy"/"full"/"error"

### Database Queries

**Webhook Delivery Rate Calculation:**
```sql
SELECT
    COUNT(*) as total_count,
    SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_count
FROM webhook_logs
WHERE id IN (
    SELECT id FROM webhook_logs
    ORDER BY created_at DESC
    LIMIT 100
)
```

**Active Suppression Count:**
```sql
SELECT COUNT(*) as active_count
FROM alert_suppressions
WHERE suppressed_until > NOW()
```

### Performance Requirements

- Health check must complete in < 1 second
- Database queries should use indexes:
  - `idx_webhook_logs_created_at` for webhook delivery
  - `idx_alert_suppressions_suppressed_until` for suppression count
- Use context with 500ms timeout per individual check

### Error Handling

**Fail-Open Design:**
- Webhook delivery check returns "nodata" if no logs (not error)
- Alert suppression check returns "ok" on database error (fail-open)
- Individual check failures don't crash entire health check

**Graceful Degradation:**
- If alert engine check fails, still show webhook and suppression status
- If database timeout, return partial status with degraded overall health
- Always return HTTP 200 with status field for monitoring systems

### Dependencies

**Internal Dependencies:**
- AlertEngine (from Story 5.5) - for alert engine status
- WebhookLogsQuerier (from Story 5.7) - for webhook delivery metrics
- AlertSuppressionsQuerier (from Story 5.6) - for suppression status
- Health checker infrastructure (existing) - for health check framework

**External Dependencies:**
- PostgreSQL database - for querying webhook logs and suppressions
- Gin framework - for HTTP handler

### Testing Strategy

**Unit Tests:**
- Mock AlertEngine.GetStats() with various scenarios
- Mock WebhookLogsQuerier for delivery rate calculation
- Mock AlertSuppressionsQuerier for suppression count
- Test timeout handling with context cancellation

**Integration Tests:**
- Full health check endpoint with real database
- Test with active webhook logs (100 records)
- Test with stale rule cache
- Test with full metric channel
- Test database connection failures

**Performance Tests:**
- Measure health check response time
- Verify < 1 second requirement
- Test with 1000+ webhook logs

### Monitoring Recommendations

**Prometheus Metrics:**
- `pulse_health_check_status{component="alert_engine"}` (0=unhealthy, 1=healthy)
- `pulse_health_check_status{component="webhook_delivery"}` (0=unhealthy, 1=degraded, 2=healthy)
- `pulse_webhook_delivery_success_rate` (gauge, percentage)
- `pulse_alert_suppression_active_count` (gauge)

**Alerting Rules:**
- Alert if overall health = "unhealthy"
- Warn if webhook delivery success rate < 95%
- Warn if alert engine rule cache stale > 10 minutes

### API Documentation

**GET /health**

Response:
```json
{
  "status": "healthy",
  "checks": {
    "database": "ok",
    "alert_engine": "ok",
    "webhook_delivery": "healthy",
    "alert_suppression": "ok"
  },
  "scheduler": {
    "running": true,
    "tasks": {
      "metrics-cleanup": {
        "is_running": false,
        "last_run": "2025-02-01T10:30:00Z",
        "run_count": 1234,
        "last_error": ""
      }
    }
  },
  "alert_system": {
    "alert_engine": {
      "status": "ok",
      "cached_rules": 15,
      "rule_cache_last_refresh": "2025-02-01T10:29:50Z",
      "metric_channel_depth": 23,
      "metric_channel_capacity": 1000
    },
    "webhook_delivery": {
      "status": "healthy",
      "success_rate": 98.5,
      "total_count": 100,
      "success_count": 98
    },
    "alert_suppression": {
      "status": "ok",
      "active_suppression_count": 3
    }
  },
  "timestamp": "2025-02-01T10:30:05Z"
}
```

### Success Metrics

**Functional:**
- ✅ Health check returns alert system status for all 3 components
- ✅ Alert engine status includes rule cache and channel metrics
- ✅ Webhook delivery success rate calculated from last 100 logs
- ✅ Alert suppression count shows active suppressions
- ✅ Overall health determined correctly (healthy/degraded/unhealthy)

**Non-Functional:**
- ✅ Health check response time < 1 second (95th percentile)
- ✅ Database queries use proper indexes
- ✅ Graceful degradation on component failures
- ✅ Fail-open design for non-critical checks

### Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] Unit tests written and passing (≥90% coverage)
- [ ] Integration tests written and passing
- [ ] Health check endpoint returns alert system status
- [ ] Response time < 1 second verified
- [ ] API documentation updated
- [ ] Code reviewed and approved
- [ ] Committed to git with descriptive message
- [ ] Epic 5 retrospective ready (final story)

---

**Story Priority:** High (Epic 5 completion)
**Story Points:** 5
**Target Release:** Sprint 5 (Epic 5 Final Story)
**Dependencies:** Stories 5.5, 5.6, 5.7 must be complete
**Blockers:** None
**Risk:** Low (extension of existing health check system)
