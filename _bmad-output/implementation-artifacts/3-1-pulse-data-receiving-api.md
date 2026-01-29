# Story 3.1: Pulse 数据接收 API

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Pulse 系统,
I want 接收 Beacon 心跳数据,
So that 可以存储和处理网络质量指标。

## Acceptance Criteria

**Given** Pulse API 服务已运行
**When** Beacon 发送 `POST /api/v1/beacon/heartbeat` 请求
**Then** 验证节点 ID 是否有效（存在于 nodes 表）
**And** 验证指标值在合理范围（时延 0-60000ms，丢包率 0-100%，抖动 0-50000ms）
**And** 数据在 5 秒内开始处理
**And** 处理失败时返回 400 错误码

**覆盖需求:** FR14（心跳上报）、NFR-OTHER-001（心跳 5 秒处理）

**创建表:** 无（使用 nodes 表）

## Tasks / Subtasks

- [x] 实现 Beacon 心跳数据接收 API 端点 (AC: #1, #2, #4)
  - [x] 创建 `POST /api/v1/beacon/heartbeat` 路由
  - [x] 定义心跳数据请求结构体（JSON）
  - [x] 实现节点 ID 有效性验证
  - [x] 实现指标值范围验证
  - [x] 实现 5 秒处理时间保证
  - [x] 返回合适的 HTTP 状态码（200/400）
- [x] 编写单元测试和集成测试 (AC: #1, #2, #4)
  - [x] 测试有效节点 ID 和有效指标值
  - [x] 测试无效节点 ID 返回 400
  - [x] 测试超出范围指标值返回 400
  - [x] 测试 API 响应时间 ≤5 秒

## Dev Notes

### Architecture Compliance

**核心架构要求:**
- **API 框架**: 使用 Gin Web 框架（最新稳定版）[Source: architecture.md#API & Communication Patterns]
- **路由设计**: REST API 风格，端点 `/api/v1/beacon/heartbeat` [Source: architecture.md#API & Communication Patterns]
- **数据格式**: JSON 格式（MVP 阶段不压缩）[Source: architecture.md#Data Architecture]
- **错误响应**: 统一错误格式 `{code: "ERR_XXX", message: "...", details: {...}}` [Source: architecture.md#Format Patterns]
- **速率限制**: Beacon 心跳每个节点 60 秒最多 1 次上报 [Source: architecture.md#API & Communication Patterns]

**命名约定:**
- API 端点: 使用复数形式（虽然 beacon 是单数，但保持与 /api/v1/nodes 一致的风格）
- JSON 字段: 使用 snake_case（与 PostgreSQL 一致）[Source: architecture.md#Naming Patterns]
- HTTP 状态码: 200（成功）、400（验证失败）、429（速率限制）[Source: architecture.md#API & Communication Patterns]

**请求/响应格式:**

```go
// 请求格式
type HeartbeatRequest struct {
    NodeID          string  `json:"node_id" binding:"required"`
    ProbeID         string  `json:"probe_id" binding:"required"`
    LatencyMs       float64 `json:"latency_ms" binding:"required"`
    PacketLossRate  float64 `json:"packet_loss_rate" binding:"required"`
    JitterMs        float64 `json:"jitter_ms" binding:"required"`
    Timestamp       string  `json:"timestamp" binding:"required"` // ISO 8601
}

// 成功响应格式
type HeartbeatSuccessResponse struct {
    Data      interface{} `json:"data"`
    Message   string      `json:"message"`
    Timestamp string      `json:"timestamp"`
}

// 错误响应格式
type ErrorResponse struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details"`
}
```

**验证规则:**
- `node_id`: 必须存在于 `nodes` 表中（UUID 格式）
- `latency_ms`: 0-60000 范围
- `packet_loss_rate`: 0-100 范围（百分比，转换为 0.0-1.0 存储）
- `jitter_ms`: 0-50000 范围
- `timestamp`: ISO 8601 格式，可解析为时间戳

**性能要求:**
- 数据在 5 秒内开始处理 [Source: epics.md#Story 3.1]
- API 响应时间 P99 ≤ 500ms, P95 ≤ 200ms [Source: architecture.md#NonFunctional Requirements]

**安全要求:**
- Beacon 与 Pulse 之间采用 TLS 1.2 或更高版本加密传输 [Source: architecture.md#Security Requirements]
- 无需 Session 认证（MVP 阶段使用简化 token 认证或 IP 白名单）[Source: architecture.md#Authentication & Security]

**代码位置:**
- 路由定义: `pulse-api/internal/api/beacon_handler.go`
- 数据模型: `pulse-api/internal/models/beacon.go`
- 验证逻辑: `pulse-api/internal/api/middleware/validation.go`

### Technical Requirements

**依赖项:**
1. **PostgreSQL 数据库连接** (Story 1.2 已实现)
   - 使用 pgx 驱动和 pgxpool 连接池 [Source: architecture.md#Data Architecture]
   - 验证节点 ID 时查询 `nodes` 表

2. **Gin Web 框架** (Story 1.2 已实现)
   - 路由注册: `router.POST("/api/v1/beacon/heartbeat", handlers.HandleHeartbeat)`

3. **速率限制中间件** (Story 1.2 可能已实现)
   - 使用 Gin 中间件实现
   - 限制: 每个 node_id 60 秒最多 1 次请求

**实现步骤:**
1. 在 `internal/api/` 创建 `beacon_handler.go`
2. 定义 `HeartbeatRequest` 结构体（使用 Go struct tags 验证）
3. 实现验证逻辑：
   - 节点 ID 存在性检查（查询 nodes 表）
   - 指标值范围检查（时延、丢包率、抖动）
4. 实现 Gin 路由处理函数
5. 添加速率限制中间件
6. 编写单元测试和集成测试

**数据库查询:**
```sql
-- 验证节点 ID 存在性
SELECT id FROM nodes WHERE id = $1;
```

**错误处理:**
- `ERR_NODE_NOT_FOUND`: 节点 ID 不存在
- `ERR_INVALID_LATENCY`: 时延超出范围
- `ERR_INVALID_PACKET_LOSS`: 丢包率超出范围
- `ERR_INVALID_JITTER`: 抖动超出范围
- `ERR_RATE_LIMIT_EXCEEDED`: 超过速率限制

### Integration with Subsequent Stories

**依赖关系:**
- **被 Story 3.2 依赖**: 本故事实现的数据接收 API 将被 Story 3.2（内存缓存与异步批量写入）使用
- **被 Story 3.7 依赖**: 本故事实现的心跳端点是 Beacon 数据上报的基础

**数据流转:**
1. Beacon 发送心跳 → 本故事 API 接收
2. 验证通过 → Story 3.2 写入内存缓存
3. Story 3.2 异步批量写入 → PostgreSQL `metrics` 表

**接口设计:**
- 本故事仅实现数据接收和验证
- 不实现数据持久化（由 Story 3.2 完成）
- 不实现告警检测（由 Story 3.2 + Story 5.5 完成）

### Previous Story Intelligence

**从 Epic 2 Stories 学到的经验:**

**Story 2.1 (节点管理 API):**
- ✅ 使用 Gin 框架成功实现 REST API
- ✅ 统一错误响应格式工作良好
- ✅ PostgreSQL 查询使用 pgx 驱动正常
- ⚠️ 注意: 节点 ID 使用 UUID 格式，验证时需要检查格式和存在性

**Story 2.2 (节点状态查询 API):**
- ✅ GET 端点实现模式可参考
- ✅ 数据库连接池（pgxpool）配置正确
- ⚠️ 注意: 状态查询需要考虑缓存（本故事不需要，但后续需要）

**Story 2.6 (Beacon 进程管理):**
- ✅ Beacon 已实现与 Pulse 的通信基础
- ✅ Beacon 配置文件包含 `pulse_server` 地址
- 📌 Beacon 在启动时注册节点（Story 2.5），本故事验证 `node_id` 应该能找到注册的节点

**代码模式参考:**
```go
// 从 Story 2.1 学到的模式
// 路由定义
nodes := v1.Group("/nodes")
{
    nodes.POST("", handlers.CreateNode)
    nodes.GET("", handlers.ListNodes)
    nodes.GET("/:id", handlers.GetNode)
}

// 错误响应格式
c.JSON(http.StatusBadRequest, gin.H{
    "code": "ERR_INVALID_INPUT",
    "message": "Invalid input parameters",
    "details": gin.H{
        "field": "node_id",
        "reason": "Node ID not found",
    },
})
```

**Git 智能分析:**
- 最新提交: `3246220 fix: 修复 Story 2.6 代码审查发现的问题`
- Epic 2 已完成所有 6 个故事，节点管理和 Beacon 基础功能已实现
- Epic 3 是第一次涉及时序数据处理的 Epic，需要特别注意数据验证和性能

### Testing Requirements

**单元测试:**
- 测试有效节点 ID 和有效指标值 → 返回 200
- 测试无效节点 ID（UUID 格式正确但不存在）→ 返回 400
- 测试无效节点 ID（UUID 格式错误）→ 返回 400
- 测试时延超出范围（-1, 60001）→ 返回 400
- 测试丢包率超出范围（-1, 101）→ 返回 400
- 测试抖动超出范围（-1, 50001）→ 返回 400
- 测试缺少必填字段 → 返回 400

**集成测试:**
- 测试完整的请求-响应流程
- 测试数据库连接和查询
- 测试速率限制中间件
- 测试性能（响应时间 ≤5 秒）

**测试文件位置:**
- 单元测试: `pulse-api/internal/api/beacon_handler_test.go`
- 集成测试: `pulse-api/tests/api/beacon_heartbeat_integration_test.go`

**测试数据准备:**
- 在测试数据库中预先插入测试节点数据
- 测试完成后清理数据

### Project Structure Notes

**文件组织:**
```
pulse-api/
├── internal/
│   ├── api/
│   │   ├── beacon_handler.go        # 本故事新增
│   │   ├── beacon_handler_test.go   # 本故事新增
│   │   ├── middleware/
│   │   │   └── rate_limit.go        # 速率限制中间件（可能需要新增）
│   ├── models/
│   │   └── beacon.go                # 本故事新增（数据模型）
│   └── db/
│       └── nodes.go                 # 节点数据库操作（Story 2.1 已实现）
├── tests/
│   └── api/
│       └── beacon_heartbeat_integration_test.go  # 本故事新增
```

**与统一项目结构对齐:**
- ✅ 遵循 `internal/` 目录组织
- ✅ 遵循测试文件与源代码并行组织
- ✅ 使用 Go 标准项目布局

**无冲突检测:**
- 本故事新增文件，不修改 Epic 1 和 Epic 2 的现有代码
- 路由端点 `/api/v1/beacon/heartbeat` 不与现有端点冲突

### References

**Architecture 文档引用:**
- [Source: architecture.md#Data Architecture] - PostgreSQL + pgx 驱动配置
- [Source: architecture.md#API & Communication Patterns] - Gin 框架和 REST API 设计
- [Source: architecture.md#Format Patterns] - API 响应格式和错误处理
- [Source: architecture.md#Naming Patterns] - 数据库表命名和 JSON 字段命名

**Epics 文档引用:**
- [Source: epics.md#Epic 3] - Epic 3 技术基础和包含的 NFR
- [Source: epics.md#Story 3.1] - 完整的验收标准和需求覆盖

**Previous Stories:**
- Story 1.2: 后端项目初始化与数据库设置（Gin 框架、PostgreSQL 连接）
- Story 2.1: 节点管理 API 实现（nodes 表、UUID 验证）
- Story 2.2: 节点状态查询 API（GET 端点模式）
- Story 2.5: Beacon 节点注册功能（节点注册流程）

**NFR 引用:**
- NFR-OTHER-001: 心跳数据 5 秒内接收并开始处理 [Source: epics.md#NonFunctional Requirements]

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### Debug Log References

No critical issues encountered during implementation. All tests passed on first run.

### Completion Notes List

**Implementation Summary:**
- ✅ Created `POST /api/v1/beacon/heartbeat` endpoint at `/api/v1/beacon/heartbeat`
- ✅ Defined `HeartbeatRequest` model with validation tags (node_id, probe_id, latency_ms, packet_loss_rate, jitter_ms, timestamp)
- ✅ Implemented node ID validation (UUID format + existence check in nodes table)
- ✅ Implemented probe_id validation (max length 255 characters)
- ✅ Implemented metric range validation:
  - latency_ms: 0-60000ms
  - packet_loss_rate: 0-100%
  - jitter_ms: 0-50000ms
- ✅ Implemented timestamp validation (ISO 8601 format)
- ✅ Returns 200 on success, 400 on validation failures
- ✅ API response time well under 5-second requirement (typically < 1ms in tests)

**Testing Coverage:**
- ✅ Unit tests: 10 test cases covering all validation scenarios
  - Valid request handling
  - Invalid node ID format
  - Node not found
  - Metric range violations (latency, packet loss, jitter)
  - Missing required fields
  - Invalid timestamp format
  - Invalid probe_id (too long)
- ✅ Integration tests: 4 test cases for end-to-end validation
  - Valid request with performance measurement
  - Invalid node ID
  - Metric validation with database
  - Performance test (10 requests)

**Files Created/Modified:**
- `pulse-api/internal/models/beacon.go` - Beacon data models
- `pulse-api/internal/api/beacon_handler.go` - Heartbeat handler with probe_id validation
- `pulse-api/internal/api/beacon_handler_test.go` - Unit tests including probe_id validation
- `pulse-api/internal/api/routes.go` - Added beacon routes
- `pulse-api/tests/api/beacon_heartbeat_integration_test.go` - Integration tests
- `pulse-api/tests/api/README.md` - Integration test documentation

**Technical Decisions:**
1. No authentication required for beacon endpoint (MVP simplification per architecture)
2. Uses existing NodesQuerier interface for database operations
3. Consistent error response format with other API endpoints
4. TODO comment added for Story 3.2 (memory cache + async write)
5. Beacon endpoint is public (no auth middleware) as per MVP requirements
6. Rate limiting uses IP-based middleware (per-node rate limiting deferred)

**Code Review Fixes Applied (2026-01-29):**
- ✅ Added probe_id length validation (max 255 characters)
- ✅ Added unit test for probe_id validation
- ✅ Added integration test documentation (README.md)
- ✅ Committed all files to git (tests were previously untracked)
- ✅ Updated sprint status to 'done'

**Performance Validation:**
- API response time: < 1ms (far below 5-second NFR requirement)
- All validation checks are O(1) complexity
- Database query uses indexed UUID lookup

**Acceptance Criteria Status:**
- ✅ AC #1: Validates node ID exists in nodes table
- ✅ AC #2: Validates metric ranges (latency 0-60000ms, loss 0-100%, jitter 0-50000ms)
- ✅ AC #3: Data processing begins immediately (well under 5 seconds)
- ✅ AC #4: Returns 400 on validation failures

### File List

pulse-api/internal/models/beacon.go
pulse-api/internal/api/beacon_handler.go
pulse-api/internal/api/beacon_handler_test.go
pulse-api/internal/api/routes.go
pulse-api/tests/api/beacon_heartbeat_integration_test.go

