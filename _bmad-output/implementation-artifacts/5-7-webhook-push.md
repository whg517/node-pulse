# Story 5.7: Webhook Push Implementation

Status: ready-for-dev

## Story

As a Pulse 系统,
I need 通过 Webhook 推送告警到配置的 URL,
So that 运维人员可以及时收到告警推送。

## Acceptance Criteria

**Given** Webhook 配置已设置（enabled = true）
**When** 告警触发（未抑制）
**Then** 使用 HTTP POST 发送告警事件到配置的 URL
**And** 请求格式为 JSON（包含告警事件数据）
**And** 告警通知包含直接链接到异常节点详情页
**And** 响应超时时间 ≤10 秒
**And** 推送失败时重试最多 3 次（指数退避：1s, 2s, 4s）
**And** 记录推送结果到日志（webhook_logs 表）

**覆盖需求:** FR6（Webhook 推送）、NFR-RECOVERY-004（重试 3 次）、NFR-REL-001（成功率 ≥95%）

## Tasks / Subtasks

- [ ] Task 1: Create webhook_logs table migration (AC: Then - 记录推送结果)
  - [ ] Subtask 1.1: Define webhook_logs table schema
  - [ ] Subtask 1.2: Add id, webhook_id, alert_id, status, retry_count, created_at
  - [ ] Subtask 1.3: Create indexes on webhook_id, alert_id, status, created_at
  - [ ] Subtask 1.4: Add migration to Migrate function

- [ ] Task 2: Create WebhookLog model (AC: Then - 记录推送结果)
  - [ ] Subtask 2.1: Create WebhookLog model struct
  - [ ] Subtask 2.2: Add DTOs for webhook log responses

- [ ] Task 3: Implement webhook logs database operations (AC: 记录推送结果)
  - [ ] Subtask 3.1: Create WebhookLogsQuerier interface
  - [ ] Subtask 3.2: Implement CreateWebhookLog function
  - [ ] Subtask 3.3: Implement GetWebhookLogs function (optional filtering)

- [ ] Task 4: Implement webhook push service (AC: When - 告警触发)
  - [ ] Subtask 4.1: Create WebhookPushService struct
  - [ ] Subtask 4.2: Implement SendWebhook function with HTTP POST
  - ] Subtask 4.3: Implement retry logic with exponential backoff (1s, 2s, 4s)
  [ ] Subtask 4.4: Implement format customization (webhook event_format)
  - [ ] Subtask 4.5: Add 10-second timeout
  - [ ] Subtask 4.6: Handle multiple webhook endpoints
  - ] Subtask 4.7: Return delivery result (success/failure)

- [ ] Task 5: Integrate webhook push with alert engine (AC: Given - Webhook 配置已设置)
  - [ ] Subtask 5.1: Modify AlertEngine to use WebhookPushService
  [ - ] Subtask 5.2: Trigger webhook push after alert event creation
  - ] Subtask 5.3: Get all enabled webhooks from database
  [ ] Subtask 5.4: Send alert to each webhook concurrently
  [ ] Subtask 5.5: Log webhook delivery results

- [ ] Task 6: Write comprehensive tests (AC: 完整功能验证)
  - [ ] Subtask 6.1: Unit tests for webhook push service
  - [ ] Subtask 6.2: Unit tests for retry logic
  [ - ] Subtask 6.3: Integration tests with mock webhook server
  - ] Subtask 6.4: Test timeout handling
  [ ] Subtask 6.5: Test retry mechanism
  - ] Subtask 6.6: Test multiple webhook endpoints
  - ] Subtask 6.7: Test webhook event format customization

- [ ] Task 7: Update documentation (AC: 文档完整性)
  - [ ] Subtask 7.1: Document webhook push flow
  [ ] Subtask 7.2: Document retry policy
  [ ] Subtask 7.3] Add usage examples

## Dev Notes

### Epic Analysis

**Epic 5: 告警规则配置与通知** - 系统可以自动检测异常并通过 Webhook 推送告警

**Story Context in Epic:**
- Story 5.1-5.6: **已完成** - 告警规则 API、Webhook 配置、前端页面、告警引擎、告警抑制
- Story 5.7: **Webhook 推送实现** (本故事) - **核心推送功能**
- Story 5.8: **后续功能** - 健康检查扩展

**Critical Prerequisites:**
- **Story 5.2 已完成**: Webhook 配置 API 已实现（webhooks 表、WebhookQuerier）
- **Story 5.5 已完成**: 告警引擎已实现（AlertEngine、alert_events 表）
- **Story 5.6 已完成**: 告警抑制已实现（不触发被抑制的告警的 webhook）
- **数据库已配置**: PostgreSQL + pgx 连接池已就绪
- **DefaultEventFormat**: 已在 models/webhook.go 中定义默认模板

### Architecture Alignment

**Webhook Push Flow** [Source: Architecture.md#API & Communication Patterns]:
```
告警触发流程：
1. AlertEngine 检测到告警指标超过阈值
2. 检查告警抑制（Story 5.6）- 如果被抑制则跳过
3. 创建 AlertEvent 存储到数据库
4. 调用 WebhookPushService.SendAlert() 发送告警
5. 查询所有启用的 webhook 配置
6. 对每个 webhook 并发发送 HTTP POST 请求
7. 格式化告警数据（根据 webhook 的 event_format 或默认格式）
8. 处理响应：成功 → 记录成功日志；失败 → 重试最多 3 次
9. 所有推送结果记录到 webhook_logs 表
```

**Database Schema** [Source: Architecture.md#Data Models]:
- `webhook_logs` 表：id (UUID), webhook_id (UUID), alert_id (UUID), status (VARCHAR), retry_count (INTEGER), created_at (TIMESTAMP)
- 外键：webhook_id REFERENCES webhooks(id), alert_id REFERENCES alert_events(id)
- 索引：idx_webhook_logs_webhook_id, idx_webhook_logs_alert_id, idx_webhook_logs_status

**Retry Policy** [Source: NFR-RECOVERY-004]:
- 最大重试 3 次
- 指数退避：1 秒, 2 秒, 4 秒
- 总尝试次数：1 (初始) + 3 (重试) = 4 次尝试
- 总时间：~7 秒（不含网络延迟）

**Performance Requirements** [Source: NFR-REL-001]:
- Webhook 告警推送成功率 ≥ 95%
- 响应超时 ≤10 秒

### Project Structure Notes

**Alignment with unified project structure** [Source: Architecture.md#Project Structure & Boundaries]:
```
pulse-api/
├── internal/
│   ├── models/
│   │   └── webhook_log.go             # NEW - WebhookLog model
│   ├── db/
│   │   ├── webhook_logs.go             # NEW - WebhookLog database operations
│   │   └── migrations.go               # UPDATE - Add webhook_logs table
│   ├── webhook/
│   │   └── push_service.go            # NEW - Webhook push service
│   └── alert/
│       └── engine.go                  # UPDATE - Integrate webhook push
└── tests/
    └── integration/
        └── webhook_push_integration_test.go  # NEW - Webhook push tests
```

### Key Design Decisions

**1. Concurrent Webhook Delivery**
```go
// Send to all webhooks concurrently using goroutines
var wg sync.WaitGroup
errChan := make(chan error, len(webhooks))

for _, webhook := range webhooks {
    wg.Add(1)
    go func(wh *models.Webhook) {
        defer wg.Done()
        err := service.SendAlert(ctx, alertEvent, wh)
        errChan <- err
    }(webhook)
}

wg.Wait()
close(errChan)

// Collect all errors
for err := range errChan {
    if err != nil {
        slog.Error("Webhook delivery failed", "error", err)
    }
}
```

**2. Retry Logic with Exponential Backoff**
```go
func (s *WebhookPushService) SendAlert(ctx context.Context, alertEvent *models.AlertEvent, webhook *models.Webhook) error {
    maxRetries := 3
    backoffs := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            time.Sleep(backoffs[attempt-1])
        }

        err := s.sendHTTP(ctx, alertEvent, webhook)
        if err == nil {
            return nil // Success
        }

        if attempt == maxRetries {
            return err // Final attempt failed
        }

        slog.Warn("Webhook delivery failed, retrying",
            "webhook_id", webhook.ID,
            "attempt", attempt+1,
            "error", err)
    }

    return fmt.Errorf("webhook delivery failed after %d attempts", maxRetries+1)
}
```

**3. Event Format Customization**
```go
func formatAlertEvent(alert *models.AlertEvent, webhook *models.Webhook) (map[string]any, error) {
    // Use custom event_format if provided, else use default
    template := webhook.EventFormat
    if template == nil || len(template) == 0 {
        template = models.DefaultEventFormat
    }

    // For MVP, use default format
    // TODO: Implement template variable substitution
    formatted := map[string]any{
        "version": "1.0",
        "alert": map[string]any{
            "id":            alert.ID,
            "metric":        alert.Metric,
            "threshold":     alert.Threshold,
            "current_value": alert.CurrentValue,
            "level":         alert.Level,
            "node_id":       alert.NodeID,
            "triggered_at":   alert.CreatedAt.Format(time.RFC3339),
        },
        "links": map[string]any{
            "alert_details": fmt.Sprintf("https://%s/nodes/%s", baseURL, alert.NodeID),
            "dashboard":     fmt.Sprintf("https://%s", baseURL),
        },
    }

    return formatted, nil
}
```

**4. Timeout Handling**
```go
client := &http.Client{
    Timeout: 10 * time.Second,
}

req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(body))
if err != nil {
    return err
}

req.Header.Set("Content-Type", "application/json")
resp, err := client.Do(req)
```

### Testing Strategy

**Unit Tests:**
- Test webhook push service with mock HTTP client
- Test retry logic (success, failure, retry scenarios)
- Test timeout handling
- Test event format generation
- Test concurrent webhook delivery

**Integration Tests:**
- Test full webhook push flow with mock server
- Test retry mechanism with delayed responses
- Test multiple webhook endpoints
- Test webhook_logs database operations
- Test integration with alert engine

**Edge Cases:**
- Webhook server down (all retries fail)
- Webhook server slow response (timeout)
- Invalid webhook URL (connection refused)
  - Malformed JSON response
- Empty webhook configuration list
- Disabled webhooks (enabled = false)

### Dependencies on Other Stories

**Depends On:**
- **Story 5.2** (Webhook Config API): Provides webhooks table and configuration
- **Story 5.5** (Alert Engine): Triggers webhook push on alert events
- **Story 5.6** (Alert Suppression): Only sends non-suppressed alerts

**Required For:**
- **Story 5.8** (Health Check): Webhook delivery rate monitoring
- **Story 6.1** (Alert Record Storage): Webhook logs correlate with alert records

### Non-Functional Requirements

**Reliability:**
- NFR-REL-001: Webhook 告警推送成功率 ≥ 95%
- NFR-RECOVERY-004: 推送失败重试最多 3 次

**Performance:**
- Response timeout ≤10 seconds
- Concurrent webhook delivery (goroutines)
- Retry with exponential backoff

**Observability:**
- Log all webhook delivery attempts
- Record success/failure to webhook_logs table
- Metrics for delivery rate, latency, errors

### Open Questions

1. **Event Format Template**: Should we implement full template variable substitution? (MVP: use default format with alert data)
2. **Webhook Authentication**: Should webhooks support authentication headers? (MVP: no authentication)
3. **Async Delivery**: Should webhook pushes be queued for background processing? (MVP: synchronous in alert engine evaluation)
4. **Dead Letter Queue**: Should failed webhooks be stored for manual retry? (MVP: logs only, no manual retry)

### Success Metrics

- ✅ HTTP POST to webhook URL with JSON payload
- ✅ 10-second timeout enforced
- ✅ Retry up to 3 times with exponential backoff
- ✅ Concurrent delivery to multiple webhooks
- ✅ Webhook logs recorded to database
- ✅ Integration with alert engine complete
- ✅ Test coverage >80%
