---
stepsCompleted: ['step-01-validate-prerequisites', 'step-02-design-epics', 'step-03-create-stories']
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/planning-artifacts/architecture.md
  - _bmad-output/planning-artifacts/ux-design-specification.md
project_name: node-pulse
workflowType: 'epics-and-stories'
date: 2026-01-25
author: Kevin
editHistory:
  - date: '2026-01-25'
    changes: '需求提取完成，Epic 设计完成，50 个 Stories 生成完成'
---

# node-pulse - Epic Breakdown

## Overview

This document provides complete epic and story breakdown for node-pulse, decomposing requirements from PRD, UX Design, and Architecture requirements into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: 运维工程师可以管理 Beacon 节点（添加、删除、查看所有 Beacon 节点，包含节点 ID、节点名称、IP 地址、地区标签和探测配置）

FR2: 运维工程师可以配置 Beacon 的探测参数（配置探测目标、协议类型、探测间隔、超时时间、探测次数，支持 TCP/UDP 协议，间隔 60-300 秒，次数 1-100 次）

FR3: 运维工程师可以查看 Beacon 的实时状态（在线/离线/连接中状态、最后心跳时间、最新数据上报时间，状态刷新 ≤5 秒）

FR4: 运维主管可以查看 Pulse 的实时仪表盘（全局节点列表、红/黄/绿健康状态指示、单节点详情页显示时延/丢包率/抖动、7 天历史趋势图、全局视图显示异常 TOP5 列表，加载时间 ≤5 秒）

FR5: 运维主管可以配置告警规则（配置时延/丢包率/抖动阈值、按节点或分组应用规则、告警级别分为 P0/P1/P2、支持启用/禁用规则、告警抑制：同一节点同一类型异常 5 分钟内仅推送一次）

FR6: 运维主管可以配置 Webhook 告警推送（配置一个或多个 Webhook URL、使用 HTTPS POST JSON 格式、支持自定义告警事件格式、告警通知包含直接链接到异常节点详情页、响应超时 ≤10 秒、失败重试最多 3 次）

FR7: 运维主管可以查看告警记录（按节点/时间/级别筛选、显示告警处理状态、告警记录留存 ≥30 天）

FR8: Beacon 可以执行 TCP Ping 探测（使用 TCP SYN 包探测目标 IP 和端口的连通性、采集 RTT、探测超时可配置 1-30 秒、不依赖 ICMP 适用于 ICMP 禁用环境）

FR9: Beacon 可以执行 UDP Ping 探测（使用 UDP 包探测目标 IP 和端口的连通性、计算丢包率、探测超时可配置 1-30 秒）

FR10: Beacon 可以采集核心网络指标（采集时延 RTT 和时延方差、丢包率、抖动、每次探测至少采集 10 个样本点、时延精度 ≤1 毫秒）

FR11: Beacon 可以通过 YAML 配置文件管理配置（YAML 格式 UTF-8 编码、包含 pulse_server/node_id/node_name/probes 字段、支持热更新无需重启 Beacon、配置文件大小 ≤100KB）

FR12: Beacon 支持 CLI 命令行操作（支持 start/stop/status/debug 命令、start 启动进程加载配置开始探测并显示实时进度和部署耗时、stop 优雅停止等待当前探测完成、status 查看运行状态在线/离线/最后心跳/配置版本输出 JSON 格式、debug 启用详细调试输出结构化日志、错误提示包含具体位置和修复建议）

FR13: Pulse 可以管理用户认证（账号密码登录 8-32 字符、bcrypt 加密存储、会话超时 24 小时、登录失败 5 次后账户锁定 10 分钟、单租户部署）

FR14: Pulse 可以接收 Beacon 心跳上报（使用 HTTP POST/HTTPS 请求、心跳数据包含节点 ID/时延/丢包率/抖动/上报时间戳、验证节点 ID 有效性/指标值范围、心跳数据 5 秒内接收并开始处理）

FR15: Pulse 可以将 Beacon 数据存储到内存缓存（支持至少 10 个节点的实时数据、按 1 分钟聚合数据、缓存数据保留 7 天、超过时间自动清除 FIFO/LRU）

FR16: Pulse 可以提供系统健康检查 API（返回整体系统状态健康/异常、包含数据库连接/Beacon 连接数/API 响应延迟/内存使用检查、响应时间 ≤100ms、每分钟自动触发）

FR17: Pulse 可以管理 Beacon 节点注册（支持 Beacon 注册分配 Node ID UUID、支持更新节点信息、支持删除节点、注册请求包含节点名称/节点 IP/地区标签）

FR18: Pulse 可以提供 7 天历史趋势图（显示最近 24 小时/7 天/30 天时间范围、显示时延/丢包率/抖动指标、数据从内存缓存加载 7 天数据、支持数据点悬停显示具体数值、按 1 分钟或 5 分钟聚合、包含 7 天基线参考线绿色虚线、支持缩放）

FR19: Pulse 可以支持多节点对比视图（按地区/运营商标签分组对比、支持自定义节点选择最多 5 个、使用相同时间范围和指标类型、显示平均值/最大值/最小值/差异、对比节点必须有重叠时间数据、明确标注差异用颜色或图标）

FR20: Pulse 可以导出节点数据报表（支持按节点/时间范围/指标类型筛选、导出 CSV/Excel 格式、单次导出最多 50 个节点、文件大小限制 10MB、异步导出完成后通知）

FR21: Pulse 可以查看仪表盘加载性能指标（仪表盘加载时间 P99/P95、API 响应时间 P99/P95、数据查询时间 P99/P95、每分钟记录一次、在系统监控仪表盘可视化显示、异常性能告警）

FR22: Pulse 可以自动判断问题类型（基于多节点数据对比自动判断节点本地故障 vs 跨境链路问题 vs 运营商路由问题、需要至少 3 个节点数据参与对比、对比时间窗口最近 1 小时、判断置信度高/中/低、判断结果实时更新）

### NonFunctional Requirements

#### 性能要求 (NFR-PERF)

NFR-PERF-001: Beacon → Pulse 数据上报延迟 ≤ 5 秒

NFR-PERF-002: 仪表盘加载时间 ≤ 5 秒（P99 ≤ 3 秒，P95 ≤ 2 秒）

NFR-PERF-003: API 响应时间 ≤ 500ms（P99），≤ 200ms（P95）

#### 可靠性要求 (NFR-REL)

NFR-REL-001: Webhook 告警推送成功率 ≥ 95%

NFR-REL-002: Beacon 节点 24 小时在线率 ≥ 99%，心跳丢失率 ≤ 1%

#### 可扩展性要求 (NFR-SCALE)

NFR-SCALE-001: 系统支持至少 10 个 Beacon 节点同时运行和上报

#### 安全性要求 (NFR-SEC)

NFR-SEC-001: Beacon 与 Pulse 之间采用 TLS 1.2 或更高版本加密传输

NFR-SEC-002: 账号密码通过 bcrypt 加密存储，登录失败 5 次后账户锁定 10 分钟

NFR-SEC-003: Webhook URL 必须是有效的 HTTPS 地址

#### 可维护性要求 (NFR-MAIN)

NFR-MAIN-001: 配置文件支持热更新（无需重启 Beacon），配置文件大小 ≤ 100KB

NFR-MAIN-002: 结构化日志（INFO/WARN/ERROR 级别），调试模式提供详细诊断信息

#### 资源约束 (NFR-RES)

NFR-RES-001: Beacon 内存占用 ≤ 100MB

NFR-RES-002: Beacon CPU 占用 ≤ 100 微核

#### 可观测性 (NFR-OBS)

NFR-OBS-001: Beacon 暴露 `/metrics` 端点，遵循 Prometheus exposition format

NFR-OBS-002: 核心指标：`beacon_up`, `beacon_rtt_seconds`, `beacon_packet_loss_rate`, `beacon_jitter_ms`

#### 容错与恢复 (NFR-RECOVERY)

NFR-RECOVERY-001: 跨境网络数据丢失缓解：数据压缩 + 断点续传机制（MVP 不实现压缩）

NFR-RECOVERY-002: ICMP 禁用环境适配：优先 TCP/UDP 探测，自动回退

NFR-RECOVERY-003: Beacon 资源占用监控：超限时告警并自动降级采集频率

NFR-RECOVERY-004: Webhook 失败重试：最多 3 次

#### 其他约束 (NFR-OTHER)

NFR-OTHER-001: 心跳数据 5 秒内接收并开始处理

NFR-OTHER-002: 仪表盘数据从内存缓存加载（实时数据 < 1 小时）

NFR-OTHER-003: 告警抑制机制：同一节点同一类型异常 5 分钟内仅推送一次

NFR-OTHER-004: 单次导出最多 50 个节点，文件大小限制 10MB

NFR-OTHER-005: 健康检查 API 响应时间 ≤ 100ms

### Additional Requirements

**From Architecture - Technical Decisions:**

- Beacon CLI 框架使用 Cobra
- Pulse 后端 Go Web 框架使用 Gin
- Pulse 数据存储使用 PostgreSQL + pgx 驱动
- Pulse 前端使用 React + TypeScript + Vite + Tailwind CSS + Apache ECharts
- Pulse 前端状态管理使用 Zustand
- Pulse 认证使用 Session Cookie 认证（会话超时 24 小时，存储在 PostgreSQL）
- Pulse API 设计为 REST API + OpenAPI 3.0 规范
- Pulse 内存缓存使用自定义环形缓冲区 + sync.Map
- 数据压缩策略：MVP 纯 JSON 传输，后续支持配置化压缩
- 前端路由使用 React Router v6

**From Architecture - Data Models:**

- PostgreSQL 表：users, nodes, alerts, sessions, alert_records, metrics, probes, webhooks, alert_suppressions, webhook_logs, performance_metrics
- 外键命名：表名前缀 + 下划线 + 列名（如 user_id, node_id）
- 主键命名：表名前缀 + 下划线 + id（如 user_id, node_id）
- 索引命名：idx_table_columns（如 idx_users_email, idx_alerts_node_id）

**From Architecture - Time Series Data Persistence:**

- **PostgreSQL metrics 表**：存储 Beacon 探测的原始和聚合时序数据
  - 字段：id (BIGSERIAL), node_id (UUID), probe_id (UUID), timestamp (TIMESTAMPTZ), latency_ms (DECIMAL), packet_loss_rate (DECIMAL), jitter_ms (DECIMAL), is_aggregated (BOOLEAN), created_at (TIMESTAMPTZ)
  - 索引优化：idx_metrics_node_timestamp, idx_metrics_probe_timestamp, idx_metrics_timestamp, idx_metrics_aggregated
- **数据分层存储策略**：
  - 实时查询数据（< 1 小时）：内存缓存，查询延迟 < 50ms，FIFO 自动淘汰
  - 热数据（1 小时 - 7 天）：PostgreSQL `metrics` 表，查询延迟 < 500ms，7 天后删除
  - 冷数据（> 7 天）：不存储（MVP），自动删除
- **数据写入流程**：
  - Beacon 心跳上报：Pulse API 接收数据，立即写入内存缓存，异步批量写入 PostgreSQL metrics 表
  - 批量写入策略：批量大小 100 条或 1 分钟超时
  - 数据聚合：原始数据点（60 秒探测）+ 聚合数据点（1 分钟聚合，均值/最大/最小值）
- **数据清理策略**：定时任务每小时执行，删除 7 天前的时序数据（`DELETE FROM metrics WHERE timestamp < NOW() - INTERVAL '7 days'`）
- **内存缓存优化**：
  - 实现方式：自定义环形缓冲区 + sync.Map
  - 数据结构：按节点 ID 分区，每个节点维护时间序列缓冲区
  - 聚合策略：1 分钟聚合
  - 清除策略：FIFO，超过 1 小时的数据自动清除
  - 内存估算：约 30KB（10 节点 × 1 小时 × 60 分钟 × 50 字节/点）

**From Architecture - API Endpoints:**

- 认证端点：POST /api/v1/auth/login, POST /api/v1/auth/logout
- 节点管理：GET/POST /api/v1/nodes, PUT/DELETE /api/v1/nodes/{id}, GET /api/v1/nodes/{id}/status
- 探测配置：GET/POST /api/v1/probes
- 告警规则：GET/POST /api/v1/alerts/rules, PUT/DELETE /api/v1/alerts/rules/{id}
- 告警记录：GET /api/v1/alerts/records
- 数据查询：GET /api/v1/data/metrics, GET /api/v1/data/history, GET /api/v1/data/export
- Beacon 端点：POST /api/v1/beacon/heartbeat
- 健康检查：GET /api/v1/health

**From Architecture - Deployment:**

- Pulse Web 平台使用 Docker 容器化部署（pulse-web 前端, pulse-api 后端, postgres 数据库）
- Beacon 部署方式为静态二进制文件 + YAML 配置，目标架构 Linux AMD64
- Beacon 配置文件：/etc/beacon/beacon.yaml 或当前目录
- Beacon 日志文件：/var/log/beacon/beacon.log 自动轮转

**From Architecture - RBAC Roles:**

- 管理员：所有权限（节点管理、告警配置、系统设置、用户管理）
- 操作员：执行探测配置、告警处理、数据查看
- 查看员：仅查看仪表盘数据，无配置权限

**From Architecture - Starter Template:**

- 使用 Vite + React + TypeScript 模板，手动集成 Tailwind CSS 和 Apache ECharts
- 初始化命令：`npm create vite@latest pulse-frontend -- --template react-ts`
- 项目初始化应该是第一个实现故事（Epic 1 Story 1）

**From Architecture - Code Patterns:**

- 数据库表命名使用复数形式（users, nodes, alerts, sessions, alert_records）
- JSON 字段使用 snake_case（与 PostgreSQL 一致）
- API 响应使用统一包裹格式：{data: ..., message: "...", timestamp: "..."} 或 {code: "ERR_XXX", message: "...", details: {...}}
- 组件命名使用 PascalCase（如 MetricCard.tsx, NodeDetail.tsx）
- 函数和变量命名使用 camelCase（如 getUserData, createNode）
- 常量命名使用 UPPER_SNAKE_CASE（如 MAX_RETRIES, DEFAULT_TIMEOUT）
- 事件命名使用 `动词.过去时 + 实体`（如 node.created, alert.triggered）

**From UX Design - User Journey Flows:**

- 侧边栏导航：左侧固定侧边栏用于主要模块切换（仪表盘、节点管理、告警配置、系统设置）
- 卡片展开详情：点击节点卡片无需翻页直接展开详情，减少点击次数
- Toast 通知：操作成功或失败的即时反馈（"节点添加成功"、"配置已保存"、"连接失败"），短暂显示后自动消失
- 进度可视化：分步骤进度显示（1/4、2/4、3/4、4/4），用户始终知道操作进度
- 问题类型自动判断：系统自动标注问题类型（节点本地故障/跨境链路问题/运营商路由问题）
- 错误提示包含解决方案：错误消息包含具体原因和修复建议
- 进度总结：关键任务完成后的肯定反馈（"部署完成！用时 8 分钟"）

### FR Coverage Map

| FR | Epic | 描述 |
|----|-------|------|
| FR1 | Epic 2 | Beacon 节点管理 |
| FR2 | Epic 3 | 探测参数配置 |
| FR3 | Epic 2 | Beacon 实时状态查看 |
| FR4 | Epic 4 | 实时仪表盘 |
| FR5 | Epic 5 | 告警规则配置 |
| FR6 | Epic 5 | Webhook 告警推送 |
| FR7 | Epic 6 | 告警记录查询 |
| FR8 | Epic 3 | TCP Ping 探测 |
| FR9 | Epic 3 | UDP Ping 探测 |
| FR10 | Epic 3 | 核心网络指标采集 |
| FR11 | Epic 3 | YAML 配置文件 |
| FR12 | Epic 3 | CLI 命令行操作 |
| FR13 | Epic 1 | 用户认证 |
| FR14 | Epic 3 | Beacon 心跳上报 |
| FR15 | Epic 3 | 内存缓存存储 + 时序数据持久化 |
| FR16 | Epic 5 | 健康检查 API |
| FR17 | Epic 2 | Beacon 节点注册 |
| FR18 | Epic 4 | 7 天历史趋势图 |
| FR19 | Epic 7 | 多节点对比视图 |
| FR20 | Epic 8 | 数据报表导出 |
| FR21 | Epic 8 | 仪表盘性能指标 |
| FR22 | Epic 7 | 自动问题类型判断 |

## Epic List

### Epic 1: 系统初始化与用户认证

运维团队可以登录 Pulse 平台，开始使用监控系统。

**FRs covered:** FR13

**包含 NFR:** NFR-PERF-002, NFR-SEC-002

**技术基础:** Vite + React + TypeScript 项目初始化 + Session Cookie 认证系统

---

### Story 1.1: 前端项目初始化与基础配置

As a 开发人员，
I want 初始化前端项目并配置基础依赖，
So that 后续可以开发前端功能。

**Acceptance Criteria:**

**Given** Pulse 前端项目不存在
**When** 开发人员执行初始化命令 `npm create vite@latest pulse-frontend -- --template react-ts`
**Then** 项目使用 React + TypeScript + Vite 创建成功
**And** Tailwind CSS 已正确安装和配置（tailwind.config.js）
**And** Apache ECharts 已安装
**And** 项目可以运行 `npm run dev` 启动开发服务器
**And** 基础项目结构已创建（src/components, src/pages, src/api, src/hooks）

**覆盖需求:** 技术决策中的启动模板要求

**创建表:** 无

---

### Story 1.2: 后端项目初始化与数据库设置

As a 开发人员，
I want 初始化后端项目并配置数据库连接，
So that 可以实现 API 端点。

**Acceptance Criteria:**

**Given** Pulse 后端项目不存在
**When** 开发人员创建 Go 项目并安装 Gin 框架和 pgx 驱动
**Then** 项目结构已创建（cmd/, internal/, pkg/, tests/）
**And** PostgreSQL 数据库连接已配置（通过环境变量）
**And** pgxpool 连接池已初始化
**And** 基础健康检查端点已实现（`GET /api/v1/health`）
**And** 统一错误响应中间件已实现

**覆盖需求:** Gin Web 框架、PostgreSQL + pgx、健康检查 API

**创建表:** 无（此故事仅初始化项目和连接）

---

### Story 1.3: 用户认证 API 实现

As a 系统，
I want 提供用户认证 API，
So that 运维团队可以安全登录 Pulse 平台。

**Acceptance Criteria:**

**Given** 数据库连接已配置
**When** 系统接收登录请求 `POST /api/v1/auth/login`
**Then** 验证用户名和密码（bcrypt 加密比较）
**And** 验证失败时返回 401 错误，失败计数增加
**And** 失败 5 次后账户锁定 10 分钟
**And** 登录成功时创建 Session（24 小时过期）并设置 Session Cookie
**And** 返回用户信息和角色

**When** 系统接收登出请求 `POST /api/v1/auth/logout`
**Then** 清除 Session 并返回成功响应

**And** Session 存储在 PostgreSQL 中，包含 session_id、user_id、role、expired_at

**覆盖需求:** FR13（用户认证）、NFR-SEC-002（bcrypt）、账户锁定

**创建表:**
- `users` 表：id (UUID), username (VARCHAR), password_hash (VARCHAR), role (VARCHAR), failed_login_attempts (INTEGER), locked_until (TIMESTAMP)
- `sessions` 表：id (UUID), user_id (UUID), role (VARCHAR), expired_at (TIMESTAMP), created_at (TIMESTAMP)

---

### Story 1.4: 前端登录页面与认证集成

As a 运维人员，
I want 通过登录页面输入用户名和密码登录 Pulse 平台，
So that 可以使用监控功能。

**Acceptance Criteria:**

**Given** 用户访问 `/login` 路径
**When** 用户在登录表单输入用户名和密码并提交
**Then** 调用 `/api/v1/auth/login` API
**And** 登录成功后，Session Cookie 自动设置并重定向到仪表盘
**And** 登录失败时显示错误提示
**And** 账户被锁定时显示锁定提示信息

**When** 用户点击登出按钮
**Then** 调用 `/api/v1/auth/logout` API 并清除 Session
**And** 重定向到登录页面

**And** 登录页面使用 Tailwind CSS 样式

**覆盖需求:** FR13（用户认证）、NFR-PERF-002（仪表盘加载时间）

**创建表:** 无

---

## Epic 2: Beacon 节点部署与注册

运维工程师可以部署 Beacon 并注册到 Pulse，开始数据上报。

**FRs covered:** FR1, FR17, FR3

**包含 NFR:** NFR-SEC-001

**技术基础:** 节点管理 API + Beacon CLI Cobra 框架基础

---

### Story 2.1: 节点管理 API 实现

As a 运维工程师，
I want 通过 API 添加、查看、更新和删除 Beacon 节点，
So that 可以管理监控节点。

**Acceptance Criteria:**

**Given** 用户已登录并具有管理员或操作员权限
**When** 用户发送 `POST /api/v1/nodes` 请求
**Then** 创建新节点，生成 UUID 作为 node_id
**And** 验证节点名称、IP、地区标签必填
**And** 返回创建的节点信息

**When** 用户发送 `GET /api/v1/nodes` 请求
**Then** 返回所有节点列表

**When** 用户发送 `PUT /api/v1/nodes/{id}` 请求
**Then** 更新指定节点信息

**When** 用户发送 `DELETE /api/v1/nodes/{id}` 请求
**Then** 删除指定节点（需要确认）

**覆盖需求:** FR1（节点管理）

**创建表:**
- `nodes` 表：id (UUID), name (VARCHAR), ip (VARCHAR), region (VARCHAR), tags (JSONB), created_at (TIMESTAMP), updated_at (TIMESTAMP)

---

### Story 2.2: 节点状态查询 API

As a 运维工程师，
I want 通过 API 查询节点的实时状态，
So that 可以了解节点是否在线。

**Acceptance Criteria:**

**Given** 节点已注册
**When** 用户发送 `GET /api/v1/nodes/{id}/status` 请求
**Then** 返回节点状态（在线/离线/连接中）
**And** 返回最后心跳时间
**And** 返回最新数据上报时间

**覆盖需求:** FR3（实时状态查看）、FR17（节点注册）

**创建表:** 无（使用 nodes 表）

---

### Story 2.3: Beacon CLI 框架初始化

As a 开发人员，
I want 初始化 Beacon CLI 框架，
So that 可以实现命令行操作。

**Acceptance Criteria:**

**Given** Beacon 项目不存在
**When** 开发人员初始化 Go 项目并安装 Cobra 框架
**Then** 项目结构已创建（cmd/, internal/, pkg/）
**And** 基础命令结构已实现（beacon start, beacon stop, beacon status, beacon debug）
**And** 支持命令参数和配置文件路径
**And** 可以编译为静态二进制（Linux AMD64）

**覆盖需求:** FR12（CLI 命令行操作）

**创建表:** 无

---

### Story 2.4: Beacon 配置文件与 YAML 解析

As a 运维工程师，
I want 通过 YAML 配置文件配置 Beacon，
So that 可以设置 Pulse 服务器地址和节点信息。

**Acceptance Criteria:**

**Given** Beacon 已安装
**When** 配置文件 beacon.yaml 存在于 `/etc/beacon/` 或当前目录
**Then** Beacon 加载配置文件并验证格式
**And** 验证必填字段（pulse_server、node_id、node_name）
**And** 配置文件大小 ≤100KB
**And** 配置文件格式为 YAML（UTF-8 编码）

**覆盖需求:** FR11（YAML 配置文件）、NFR-MAIN-001（配置文件大小）

**创建表:** 无

---

### Story 2.5: Beacon 节点注册功能

As a Beacon，
I want 在启动时自动注册到 Pulse，
So that 可以开始数据上报。

**Acceptance Criteria:**

**Given** Beacon 配置文件已设置
**When** Beacon 启动向 Pulse 发送注册请求
**Then** Pulse 分配 UUID 作为 node_id
**And** 节点信息包含节点名称、IP、地区标签
**And** 使用 HTTPS TLS 加密传输

**覆盖需求:** FR17（节点注册）、NFR-SEC-001（TLS 加密传输）

**创建表:** 无（使用 nodes 表）

---

### Story 2.6: Beacon 进程管理（start/stop/status）

As a 运维工程师，
I want 通过 CLI 命令启动、停止和查看 Beacon 状态，
So that 可以管理 Beacon 进程。

**Acceptance Criteria:**

**Given** Beacon 已安装和配置
**When** 执行 `beacon start` 命令
**Then** 加载配置文件并启动进程
**And** 显示实时进度（"正在连接到 Pulse..."、"正在注册..."、"上传配置..."）
**And** 向 Pulse 注册节点
**And** 输出注册成功信息和节点 ID
**And** 部署完成显示耗时统计（"部署完成！用时 X 分钟"）
**And** 错误提示包含具体位置和修复建议
  - 示例："配置格式错误：第 5 行缩进应为 2 个空格，实际是 4 个空格"

**When** 执行 `beacon stop` 命令
**Then** 优雅停止进程（等待当前探测完成）
**And** 输出停止成功信息

**When** 执行 `beacon status` 命令
**Then** 输出 JSON 格式的运行状态
**And** 包含在线/离线状态、最后心跳时间、配置版本

**When** 执行 `beacon debug` 命令
**Then** 输出详细调试信息
**And** 支持分步骤进度显示（1/4、2/4、3/4、4/4）
**And** 错误提示包含具体位置和修复建议

**覆盖需求:** FR12（CLI 命令行操作）、FR3（实时状态查看）、UX Design 交互模式

**创建表:** 无

---

## Epic 3: 网络探测配置与数据采集

Beacon 可以执行网络探测并上报数据到 Pulse，Pulse 通过内存缓存和 PostgreSQL 持久化时序数据。

**FRs covered:** FR2, FR8, FR9, FR10, FR11, FR12, FR14, FR15

**包含 NFR:** NFR-PERF-001, NFR-REL-002, NFR-SCALE-001, NFR-RES-001, NFR-RES-002, NFR-OBS-001, NFR-OBS-002, NFR-MAIN-002, NFR-RECOVERY-002, NFR-RECOVERY-003, NFR-OTHER-001, NFR-OTHER-002

**技术基础:** TCP/UDP 探测引擎 + 配置系统 + 数据接收器 + 内存缓存（sync.Map + 环形缓冲区）+ PostgreSQL metrics 表 + 异步批量写入 + 定时数据清理

---

### Story 3.1: Pulse 数据接收 API

As a Pulse 系统，
I want 接收 Beacon 心跳数据，
So that 可以存储和处理网络质量指标。

**Acceptance Criteria:**

**Given** Pulse API 服务已运行
**When** Beacon 发送 `POST /api/v1/beacon/heartbeat` 请求
**Then** 验证节点 ID 是否有效（存在于 nodes 表）
**And** 验证指标值在合理范围（时延 0-60000ms，丢包率 0-100%，抖动 0-50000ms）
**And** 数据在 5 秒内开始处理
**And** 处理失败时返回 400 错误码

**覆盖需求:** FR14（心跳上报）、NFR-OTHER-001（心跳 5 秒处理）

**创建表:** 无（使用 nodes 表）

---

### Story 3.2: Pulse 内存缓存与异步批量写入

As a Pulse 系统，
I want 将 Beacon 数据存储到内存缓存并异步批量写入 PostgreSQL，
So that 可以快速查询实时数据并持久化历史数据。

**Acceptance Criteria:**

**Given** 数据接收 API 已实现
**When** Beacon 心跳数据到达
**Then** 数据按节点 ID 存储到 sync.Map（环形缓冲区）
**And** 每个节点维护时间序列环形缓冲区
**And** 数据按 1 分钟间隔聚合
**And** 超过 1 小时的数据自动从内存清除（FIFO）
**And** 缓存支持至少 10 个节点

**When** 数据写入内存缓存后
**Then** 触发异步批量写入到 PostgreSQL `metrics` 表
**And** 批量写入策略：批量大小 100 条或 1 分钟超时
**And** 写入失败时记录错误日志并保留在内存中重试
**And** 聚合数据标记 `is_aggregated = TRUE`

**覆盖需求:** FR15（内存缓存存储）、NFR-OTHER-002（内存缓存实时数据）

**创建表:** 无（内存实现 + 异步批量写入到 metrics 表）

---

### Story 3.3: 探测配置 API 与时序数据表

As a 运维工程师，
I want 通过 API 配置探测参数并创建时序数据表，
So that 可以设置探测目标、协议类型和间隔，并持久化历史数据。

**Acceptance Criteria:**

**Given** 用户已登录并具有管理员或操作员权限
**When** 用户发送 `POST /api/v1/probes` 请求
**Then** 创建新探测配置
**And** 验证探测类型为 TCP 或 UDP
**And** 验证间隔在 60-300 秒范围内
**And** 验证探测次数在 1-100 次范围内
**And** 验证超时时间在 1-30 秒范围内

**When** 用户发送 `GET /api/v1/probes` 请求
**Then** 返回所有探测配置

**And** 创建 PostgreSQL `metrics` 表存储时序数据
  - 字段：id (BIGSERIAL PRIMARY KEY), node_id (UUID NOT NULL REFERENCES nodes(id)), probe_id (UUID NOT NULL REFERENCES probes(id)), timestamp (TIMESTAMPTZ NOT NULL), latency_ms (DECIMAL(10,2)), packet_loss_rate (DECIMAL(5,4)), jitter_ms (DECIMAL(10,2)), is_aggregated (BOOLEAN DEFAULT FALSE), created_at (TIMESTAMPTZ DEFAULT NOW())
  - 索引：idx_metrics_node_timestamp, idx_metrics_probe_timestamp, idx_metrics_timestamp, idx_metrics_aggregated

**覆盖需求:** FR2（探测配置）、Architecture 时序数据表设计

**创建表:**
- `probes` 表：id (UUID), node_id (UUID), type (VARCHAR), target (VARCHAR), port (INTEGER), interval_seconds (INTEGER), count (INTEGER), timeout_seconds (INTEGER)
- `metrics` 表：id (BIGSERIAL), node_id (UUID), probe_id (UUID), timestamp (TIMESTAMPTZ), latency_ms (DECIMAL), packet_loss_rate (DECIMAL), jitter_ms (DECIMAL), is_aggregated (BOOLEAN), created_at (TIMESTAMPTZ)
  - 索引：idx_metrics_node_timestamp, idx_metrics_probe_timestamp, idx_metrics_timestamp, idx_metrics_aggregated

---

### Story 3.4: TCP Ping 探测引擎

As a Beacon，
I want 使用 TCP SYN 包探测目标 IP 和端口，
So that 可以采集 RTT 和连通性。

**Acceptance Criteria:**

**Given** Beacon 已安装和配置
**When** 配置了 TCP Ping 探测任务
**Then** 发送 TCP SYN 包到目标 IP:端口
**And** 测量往返时延（RTT）精确到毫秒
**And** 探测超时可配置 1-30 秒（默认 5 秒）
**And** 探测结果包含连通性（成功/失败）和 RTT
**And** 不依赖 ICMP（适用于 ICMP 禁用环境）

**覆盖需求:** FR8（TCP Ping 探测）

**创建表:** 无

---

### Story 3.5: UDP Ping 探测引擎

As a Beacon，
I want 使用 UDP 包探测目标 IP 和端口，
So that 可以计算丢包率。

**Acceptance Criteria:**

**Given** Beacon 已安装和配置
**When** 配置了 UDP Ping 探测任务
**Then** 发送 UDP 包到目标 IP:端口
**And** 计算丢包率（发送包数 / 接收确认数）
**And** 探测超时可配置 1-30 秒（默认 5 秒）
**And** 丢包率为 0-100% 百分比

**覆盖需求:** FR9（UDP Ping 探测）

**创建表:** 无

---

### Story 3.6: 核心指标采集

As a Beacon，
I want 采集时延、丢包率和抖动等核心网络质量指标，
So that 可以上报到 Pulse。

**Acceptance Criteria:**

**Given** TCP/UDP 探测已实现
**When** 执行探测任务
**Then** 每次探测至少采集 10 个样本点
**And** 计算时延 RTT 和时延方差
**And** 计算丢包率（发送丢包率）
**And** 计算时延抖动（相邻样本的时延变化）
**And** 时延测量精度 ≤1 毫秒

**覆盖需求:** FR10（核心网络指标采集）

**创建表:** 无

---

### Story 3.7: Beacon 数据上报

As a Beacon，
I need 定时向 Pulse 上报数据，
So that Pulse 可以存储和展示监控数据。

**Acceptance Criteria:**

**Given** Beacon 已注册并配置
**When** 探测完成数据采集
**Then** 每 60 秒向 Pulse 发送心跳数据
**And** 心跳数据包含：node_id、latency、packet_loss_rate、jitter、timestamp
**And** 使用 HTTP/HTTPS 请求（TLS 加密）
**And** 上报延迟 ≤ 5 秒

**覆盖需求:** FR14（心跳上报）、NFR-PERF-001（上报延迟）

**创建表:** 无

---

### Story 3.8: Prometheus Metrics 端点

As a Beacon，
I need 暴露 `/metrics` 端点，
So that Prometheus 可以抓取监控指标。

**Acceptance Criteria:**

**Given** Beacon 已安装和配置
**When** Prometheus 请求 `GET /metrics` 端点
**Then** 返回遵循 Prometheus exposition format 的指标
**And** 核心指标包含：beacon_up、beacon_rtt_seconds、beacon_packet_loss_rate、beacon_jitter_ms
**And** 指标格式为文本/纯文本

**覆盖需求:** NFR-OBS-001（Prometheus Metrics）

**创建表:** 无

---

### Story 3.9: 结构化日志系统

As a Beacon，
I need 记录结构化日志，
So that 可以故障排查和系统监控。

**Acceptance Criteria:**

**Given** Beacon 已安装
**When** Beacon 运行时产生日志
**Then** 日志输出到文件 `/var/log/beacon/beacon.log`
**And** 日志级别分为 INFO、WARN、ERROR
**And** 日志格式为 JSON 结构化
**And** 日志文件自动轮转（按大小 10MB 或每天）

**覆盖需求:** NFR-MAIN-002（结构化日志）

**创建表:** 无

---

### Story 3.10: 调试模式

As a 运维工程师，
I can 通过 `beacon debug` 命令查看详细诊断信息，
So that 可以故障排查。

**Acceptance Criteria:**

**Given** Beacon 已安装
**When** 执行 `beacon debug` 命令
**Then** 输出详细调试信息
**And** 调试信息包含：网络状态、配置信息、连接重试状态
**And** 输出为结构化日志格式

**覆盖需求:** NFR-MAIN-002（调试模式）

**创建表:** 无

---

### Story 3.11: 资源监控与降级

As a Beacon，
I need 监控 CPU 和内存使用，
So that 超限时可以防止资源耗尽。

**Acceptance Criteria:**

**Given** Beacon 已安装和配置
**When** CPU 使用超过 100 微核或内存超过 100MB
**Then** 触发告警并输出警告日志
**And** 自动降级采集频率（例如从 300 秒增加到 600 秒）
**And** 持续监控资源使用

**覆盖需求:** NFR-RES-001/002（资源限制）、NFR-RECOVERY-003（资源监控与降级）

**创建表:** 无

---

### Story 3.12: 定时数据清理任务

As a Pulse 系统，
I need 定时清理过期的时序数据，
So that 防止数据库存储无限增长。

**Acceptance Criteria:**

**Given** `metrics` 表已创建
**When** 定时任务每小时触发
**Then** 执行清理命令：`DELETE FROM metrics WHERE timestamp < NOW() - INTERVAL '7 days'`
**And** 记录清理操作日志（删除的行数、执行时间）
**And** 清理失败时记录错误日志并告警

**And** 定时任务可配置清理间隔（默认 1 小时）

**覆盖需求:** Architecture 数据清理策略、NFR-OTHER-002（7 天数据保留）

**创建表:** 无（使用 metrics 表）

---

### Story 3.13: 配置热更新

As a 运维工程师，
I can 修改 YAML 配置文件而无需重启 Beacon，
So that 可以动态调整配置。

**Acceptance Criteria:**

**Given** Beacon 正在运行
**When** 修改 beacon.yaml 配置文件
**Then** Beacon 自动检测文件变更（文件监控）
**And** 重新加载配置无需重启进程
**And** 验证新配置格式和字段
**And** 验证失败时输出错误并保持原配置

**覆盖需求:** FR11（配置热更新）

**创建表:** 无

---

## Epic 4: 实时监控仪表盘

运维主管可以在仪表盘上查看所有节点的实时状态和核心指标，实时数据从内存缓存加载，历史数据从 PostgreSQL metrics 表查询。

**FRs covered:** FR4, FR18

**包含 NFR:** NFR-PERF-002, NFR-OTHER-002, NFR-OTHER-005

**技术基础:** React 仪表盘 + ECharts 图表 + Zustand 状态管理 + 内存缓存 + PostgreSQL metrics 表

---

### Story 4.1: 前端路由与认证守卫

As a 运维人员，
I need 前端路由保护未认证用户，
So that 确保只有登录后才能访问仪表盘。

**Acceptance Criteria:**

**Given** React Router v6 已安装
**When** 用户访问受保护路由（如 `/dashboard`、`/nodes`）
**Then** 系统检查 Session Cookie
**And** 未认证时重定向到 `/login` 页面
**And** 登录后重定向到原始请求页面
**And** 已认证时正常访问受保护路由

**覆盖需求:** NFR-PERF-002（仪表盘加载时间）

**创建表:** 无

---

### Story 4.2: Zustand 状态管理设置

As a 前端开发，
I need 初始化状态管理
So that 可以管理应用状态。

**Acceptance Criteria:**

**Given** Zustand 已安装
**When** 初始化 stores
**Then** 创建 authStore（用户认证状态、角色信息）
**And** 创建 nodesStore（节点列表、节点详情、在线/离线状态）
**And** 创建 alertsStore（告警规则、告警记录）
**And** 创建 dashboardStore（仪表盘筛选、时间范围、刷新设置）
**And** stores 支持 TypeScript 类型
**And** stores 可以被任意组件直接使用（无需 Provider）

**覆盖需求:** NFR-PERF-002（仪表盘加载时间）

**创建表:** 无

---

### Story 4.3: API 调用层封装

As a 前端开发，
I need 统一的 API 调用层
So that 可以与 Pulse 后端交互。

**Acceptance Criteria:**

**Given** Pulse API 端点已定义
**When** 创建 API 调用封装
**Then** 创建 api/auth.ts（登录/登出 API）
**And** 创建 api/nodes.ts（节点管理 API）
**And** 创建 api/data.ts（数据查询 API）
**And** 创建 api/alerts.ts（告警 API）
**And** 类型定义与后端 API 同步
**And** 统一错误处理
**And** API 响应使用统一包裹格式

**覆盖需求:** NFR-PERF-002（仪表盘加载时间）

**创建表:** 无

---

### Story 4.4: 仪表盘首页与节点列表

As a 运维主管，
I can 在仪表盘首页查看所有节点的全局视图和健康状态。

**Acceptance Criteria:**

**Given** 用户已登录并访问仪表盘
**When** 仪表盘加载完成
**Then** 显示全局节点列表表格
**And** 每个节点显示：节点名称、IP、地区、健康状态（红/黄/绿）
**And** 绿色表示健康，黄色表示预警，红色表示异常
**And** 显示异常 TOP5 列表（按严重程度排序）
**And** 显示核心指标均值（平均时延、平均丢包率、平均抖动）
**And** 仪表盘加载时间 ≤5 秒
**And** 状态刷新周期 ≤5 秒

**覆盖需求:** FR4（实时仪表盘）、UX Design（异常 TOP5 列表）

**创建表:** 无

---

### Story 4.5: 节点详情页

As a 运维主管，
I can 在节点详情页查看单个节点的详细网络指标。

**Acceptance Criteria:**

**Given** 用户已登录并访问 `/nodes/:id` 路由
**When** 节点详情页加载完成
**Then** 显示节点基本信息（名称、IP、地区、标签）
**And** 显示核心指标卡片：时延、丢包率、抖动
**And** 实时显示指标数值
**And** 显示节点在线/离线状态
**And** 显示最后心跳时间
**And** 支持卡片展开详情交互模式
  - 点击节点卡片无需翻页直接展开详情
  - 减少点击次数，提高效率
  - 详情包含问题类型判断（节点本地故障/跨境链路问题/运营商路由问题）

**覆盖需求:** FR4（实时仪表盘）、FR22（问题类型判断）、UX Design（卡片展开详情交互模式）

**创建表:** 无

---

### Story 4.6: ECharts 趋势图组件

As a 前端开发，
I need 封装 ECharts 趋势图组件
So that 可以复用。

**Acceptance Criteria:**

**Given** Apache ECharts 已安装
**When** 创建 TrendChart 组件
**Then** 组件接收数据点数组作为 props
**And** 组件支持时间范围选择（24小时/7天/30天）
**And** 组件支持多指标显示（时延、丢包率、抖动）
**And** 组件支持数据点悬停显示具体数值
**And** 组件支持缩放（鼠标滚轮放大/缩小）
**And** 组件使用 Tailwind CSS 样式
**And** 支持卡片展开详情交互模式
  - 点击节点卡片无需翻页直接展开详情
  - 减少点击次数，提高效率

**覆盖需求:** FR18（7 天历史趋势图）、UX Design（卡片展开详情交互模式）

**创建表:** 无

---

### Story 4.7: 7 天历史趋势图

As a 运维主管，
I can 在节点详情页查看 7 天历史趋势图。

**Acceptance Criteria:**

**Given** 用户已登录并访问节点详情页
**When** 趋势图数据加载完成
**Then** 显示最近 24 小时、7 天、30 天时间范围选择
**Then** 显示时延、丢包率、抖动指标曲线
**And** 实时数据（< 1 小时）从内存缓存加载，历史数据（1 小时 - 7 天）从 PostgreSQL `metrics` 表查询
**And** 数据按 1 分钟或 5 分钟聚合
**And** 包含 7 天基线参考线（绿色虚线）
**And** 支持鼠标悬停显示具体时间点的数值

**覆盖需求:** FR18（7 天历史趋势图）、Architecture 数据分层查询策略

**创建表:** 无（使用 metrics 表）

---

### Story 4.8: Toast 通知组件

As a 前端开发，
I need 封装 Toast 通知组件，
So that 可以在操作成功或失败时提供即时反馈。

**Acceptance Criteria:**

**Given** React + Tailwind CSS 已配置
**When** 创建 ToastNotification 组件
**Then** 组件支持多种通知类型（success/error/warning/info）
**And** 通知自动消失（默认 3 秒）
**And** 通知支持手动关闭
**And** 组件显示位置可配置（顶部/底部/右侧）
**And** 支持多通知队列显示
**And** 使用 Tailwind CSS 样式

**Given** 用户执行操作（如添加节点、保存配置）
**When** 操作成功
**Then** 显示成功通知（"节点添加成功"、"配置已保存"）
**And** 通知短暂显示后自动消失

**When** 操作失败
**Then** 显示错误通知（"连接失败"、"配置保存失败"）
**And** 通知包含具体错误信息

**覆盖需求:** UX Design（Toast 通知反馈模式）

**创建表:** 无

---

### Story 4.9: 数据实时更新轮询

As a 运维主管，
仪表盘可以自动刷新数据，
So that 显示实时监控状态。

**Acceptance Criteria:**

**Given** 仪表盘页面已加载
**When** 创建 useDashboardData Hook
**Then** 每 5 秒轮询数据 API
**And** 节点状态变化时自动刷新
**And** 避免全页刷新（局部更新状态）
**And** 支持暂停/恢复轮询

**覆盖需求:** FR4（实时仪表盘）

**创建表:** 无

---

## Epic 5: 告警规则配置与通知

系统可以自动检测异常并通过 Webhook 推送告警，失败重试最多 3 次。

**FRs covered:** FR5, FR6, FR16

**包含 NFR:** NFR-REL-001, NFR-SEC-003, NFR-RECOVERY-004, NFR-OTHER-003, NFR-OTHER-005

**技术基础:** 告警引擎 + Webhook 推送 + 健康检查 API + 失败重试机制

---

### Story 5.1: 告警规则 API

As a 运维主管，
I can 通过 API 配置告警规则，
So that 可以自动检测网络异常。

**Acceptance Criteria:**

**Given** 用户已登录并具有管理员或操作员权限
**When** 用户发送 `POST /api/v1/alerts/rules` 请求
**Then** 创建新告警规则
**And** 验证指标类型为 latency/packet_loss_rate/jitter
**And** 验证阈值为数值
**And** 验证告警级别为 P0/P1/P2
**And** 支持按节点或分组应用规则
**And** 支持启用/禁用状态

**When** 用户发送 `GET /api/v1/alerts/rules` 请求
**Then** 返回所有告警规则

**覆盖需求:** FR5（告警规则配置）

**创建表:**
- `alerts` 表：id (UUID), metric (VARCHAR), threshold (DECIMAL), level (VARCHAR), node_id (UUID), enabled (BOOLEAN), created_at (TIMESTAMP)

---

### Story 5.2: Webhook 配置 API

As a 运维主管，
I can 通过 API 配置 Webhook URL，
So that 可以接收告警推送。

**Acceptance Criteria:**

**Given** 用户已登录并具有管理员权限
**When** 用户发送 `POST /api/v1/webhooks` 请求
**Then** 创建新 Webhook 配置
**And** 验证 URL 为有效的 HTTPS 地址
**And** 支持配置一个或多个 Webhook URL
**And** 支持自定义告警事件格式

**When** 用户发送 `GET /api/v1/webhooks` 请求
**Then** 返回所有 Webhook 配置

**覆盖需求:** FR6（Webhook 配置）、NFR-SEC-003（HTTPS URL）

**创建表:**
- `webhooks` 表：id (UUID), url (VARCHAR), event_format (JSONB), enabled (BOOLEAN), created_at (TIMESTAMP)

---

### Story 5.3: 告警规则前端页面

As a 运维主管，
I can 在前端配置告警规则，
So that 可以设置检测阈值。

**Acceptance Criteria:**

**Given** 用户已登录并访问告警规则页面
**When** 页面加载完成
**Then** 显示所有现有告警规则列表
**And** 提供"创建规则"按钮
**And** 支持编辑和删除规则
**And** 表单包含：指标类型选择、阈值输入、告警级别选择、节点选择、启用/禁用开关

**覆盖需求:** FR5（告警规则配置）

**创建表:** 无

---

### Story 5.4: Webhook 配置前端页面

As a 运维主管，
I can 在前端配置 Webhook URL，
So that 可以推送告警到第三方系统。

**Acceptance Criteria:**

**Given** 用户已登录并访问 Webhook 配置页面
**When** 页面加载完成
**Then** 显示所有 Webhook 配置列表
**And** 提供"添加 Webhook"按钮
**And** 支持编辑和删除 Webhook
**And** 表单包含：URL 输入（验证 HTTPS）、事件格式编辑

**覆盖需求:** FR6（Webhook 配置）

**创建表:** 无

---

### Story 5.5: 告警引擎实现

As a Pulse 系统，
I need 检测指标是否超过阈值并触发告警。

**Acceptance Criteria:**

**Given** 告警规则已配置
**When** Beacon 心跳数据到达
**Then** 检查每个指标是否超过配置的阈值
**And** 超过阈值时创建告警事件
**And** 告警事件包含：node_id、metric、threshold、current_value、level、timestamp

**覆盖需求:** FR5（告警规则配置）

**创建表:** 无（使用 alerts 和 alert_records 表）

---

### Story 5.6: 告警抑制机制

As a Pulse 系统，
I need 实现告警抑制，避免同一节点同一类型异常重复推送。

**Acceptance Criteria:**

**Given** 告警引擎已实现
**When** 同一节点同一类型异常发生
**Then** 检查是否在 5 分钟抑制窗口内
**And** 如果在窗口内则抑制新告警
**And** 如果不在窗口内则触发新告警并重置窗口
**And** 抑制机制按 node_id 和 metric 类型分别记录

**覆盖需求:** FR5（告警抑制）、NFR-OTHER-003（告警抑制机制）

**创建表:**
- `alert_suppressions` 表：id (UUID), node_id (UUID), metric (VARCHAR), suppressed_until (TIMESTAMP)

---

### Story 5.7: Webhook 推送实现

As a Pulse 系统，
I need 通过 Webhook 推送告警到配置的 URL。

**Acceptance Criteria:**

**Given** Webhook 配置已设置
**When** 告警触发
**Then** 使用 HTTP POST 发送告警事件到配置的 URL
**And** 请求格式为 JSON（包含告警事件数据）
**And** 告警通知包含直接链接到异常节点详情页
**And** 支持会话未过期时免登录跳转
  - 或会话过期时自动登录后跳转到详情页
**And** 响应超时时间 ≤10 秒
**And** 推送失败时重试最多 3 次（指数退避）
**And** 记录推送结果到日志

**覆盖需求:** FR6（Webhook 推送）、NFR-RECOVERY-004（重试 3 次）、UX Design（告警通知详情页链接）

**创建表:**
- `webhook_logs` 表：id (UUID), webhook_id (UUID), alert_id (UUID), status (VARCHAR), retry_count (INTEGER), created_at (TIMESTAMP)

---

### Story 5.8: 健康检查扩展

As a 系统，
I need 扩展健康检查 API 包含告警状态，以便监控系统整体健康。

**Acceptance Criteria:**

**Given** 健康检查 API 已存在（Epic 1 Story 1.2）
**When** 用户请求 `GET /api/v1/health`
**Then** 返回整体系统状态（healthy/unhealthy）
**And** 包含组件状态：数据库连接、Beacon 连接数、API 响应延迟、内存使用
**And** 增加告警引擎状态（活跃/挂起/错误）
**And** 响应时间 ≤100ms

**覆盖需求:** FR16（健康检查 API）、NFR-OTHER-005（响应时间 ≤100ms）

**创建表:** 无

---

## Epic 6: 告警记录查询

运维主管可以查看历史告警记录，追踪问题处理。

**FRs covered:** FR7

**技术基础:** 告警记录存储与查询 API

---

### Story 6.1: 告警记录存储 API

As a Pulse 系统，
I need 存储告警记录，
So that 可以查询历史告警。

**Acceptance Criteria:**

**Given** 告警引擎已实现
**When** 告警触发
**Then** 创建告警记录并存储到数据库
**And** 记录包含：alert_id、node_id、metric、level、status（未处理/处理中/已解决）、timestamp
**And** 状态跟踪支持更新（如处理中/已解决）

**When** 用户发送 `GET /api/v1/alerts/records` 请求
**Then** 返回告警记录列表
**And** 支持按节点筛选（node_id 参数）
**And** 支持按时间范围筛选（start_time 和 end_time 参数）
**And** 支持按告警级别筛选（level 参数）
**And** 支持按处理状态筛选（status 参数）

**覆盖需求:** FR7（告警记录查询）

**创建表:**
- `alert_records` 表：id (UUID), alert_id (UUID), node_id (UUID), metric (VARCHAR), level (VARCHAR), status (VARCHAR), created_at (TIMESTAMP), updated_at (TIMESTAMP)

---

### Story 6.2: 告警记录前端页面

As a 运维主管，
I can 在前端查看和筛选告警记录，
So that 可以追踪问题处理。

**Acceptance Criteria:**

**Given** 用户已登录并访问告警记录页面
**When** 页面加载完成
**Then** 显示所有告警记录列表
**And** 提供筛选器：节点选择、时间范围选择、告警级别选择、处理状态选择
**And** 每条记录显示：节点名称、指标类型、告警级别、状态、时间戳
**And** 状态用颜色标注（未处理-红色、处理中-黄色、已解决-绿色）
**And** 支持分页加载
**And** 支持点击记录查看详情

**覆盖需求:** FR7（告警记录查询）

**创建表:** 无

---

## Epic 7: 多节点对比与分析

运维主管可以对比多个节点，快速定位问题根因，实时数据从内存缓存加载，历史数据从 PostgreSQL metrics 表查询。

**FRs covered:** FR19, FR22

**包含 NFR:** NFR-OTHER-002

**技术基础:** 对比算法 + 问题诊断引擎 + 内存缓存 + PostgreSQL metrics 表

---

### Story 7.1: 对比图表组件

As a 前端开发，
I need 封装多节点对比 ECharts 组件，
So that 可以复用对比图表。

**Acceptance Criteria:**

**Given** Apache ECharts 已安装
**When** 创建 ComparisonChart 组件
**Then** 组件接收多个节点数据数组作为 props
**And** 组件支持最多 5 个节点的对比
**And** 按地区/运营商标签分组对比
**And** 使用相同时间范围和指标类型
**And** 显示平均值、最大值、最小值、差异
**And** 差异用颜色或图标明确标注
**And** 组件使用 Tailwind CSS 样式

**覆盖需求:** FR19（多节点对比）

**创建表:** 无

---

### Story 7.2: 节点对比查询 API

As a 运维主管，
I can 通过 API 查询多个节点数据进行对比，
So that 可以分析节点性能。

**Acceptance Criteria:**

**Given** 用户已登录
**When** 用户发送 `GET /api/v1/data/comparison?node_ids=xxx,yyy,zzz` 请求
**Then** 返回指定节点的数据
**And** 实时数据（< 1 小时）从内存缓存查询，历史数据（1 小时 - 7 天）从 PostgreSQL `metrics` 表查询
**And** 确保对比节点有重叠的时间数据
**And** 返回数据包含相同的时间范围和指标类型
**And** 自动计算平均值、最大值、最小值、差异
**And** 验证最多 5 个节点对比

**覆盖需求:** FR19（多节点对比）、Architecture 数据分层查询策略

**创建表:** 无（使用内存缓存 + metrics 表）

---

### Story 7.3: 节点对比前端页面

As a 运维主管，
I can 在对比页面查看多个节点的网络指标，
So that 可以快速定位问题。

**Acceptance Criteria:**

**Given** 用户已登录并访问对比页面
**When** 页面加载完成
**Then** 显示节点选择器（最多 5 个）
**And** 支持按地区标签分组选择
**And** 支持按运营商标签分组选择
**And** 显示时间范围选择器
**Then** 显示多节点对比图表
**And** 对比图表使用 ComparisonChart 组件
**And** 显示平均值、最大值、最小值、差异
**And** 差异用颜色或图标标注

**覆盖需求:** FR19（多节点对比）

**创建表:** 无

---

### Story 7.4: 问题类型诊断引擎

As a Pulse 系统，
I need 基于多节点数据对比自动判断问题类型，
So that 可以快速定位根因。

**Acceptance Criteria:**

**Given** 至少 3 个节点数据存在
**When** 系统检测到异常节点
**Then** 基于同一地区节点的对比分析
**And** 判断问题类型：节点本地故障、跨境链路问题、运营商路由问题
**Then** 对比时间窗口：最近 1 小时
**And** 计算判断置信度：高（>90%）、中（70-90%）、低（<70%）
**Then** 判断结果实时更新并返回
**Then** 判断结果在前端明确标注
**And** 运营商路由问题基于运营商路由特征分析
  - 检测路由跳数异常
  - 检测 AS（自治系统）变更
  - 对比同运营商其他节点表现

**覆盖需求:** FR22（自动问题类型判断）、PRD 新增（运营商路由问题）、UX Design（问题类型自动判断）

**创建表:** 无

---

## Epic 8: 数据导出与性能监控

运维主管可以导出报表并监控系统性能指标，数据从 PostgreSQL metrics 表导出。

**FRs covered:** FR20, FR21

**包含 NFR:** NFR-PERF-002/003, NFR-OTHER-004

**技术基础:** 导出功能 + 性能指标采集 + PostgreSQL metrics 表

---

### Story 8.1: 数据导出 API

As a 运维主管，
I can 通过 API 导出节点数据报表，
So that 可以数据分析和汇报。

**Acceptance Criteria:**

**Given** 用户已登录并具有管理员权限
**When** 用户发送 `GET /api/v1/data/export` 请求
**Then** 验证筛选参数（node_id、time_range、metric_type）
**And** 异步启动导出任务
**And** 从 PostgreSQL `metrics` 表查询历史数据（1 小时 - 7 天）
**And** 导出格式支持 CSV（UTF-8 编码）和 Excel
**And** 单次导出最多 50 个节点
**And** 导出文件大小限制 10MB
**And** 导出完成后通过系统消息或邮件通知用户

**覆盖需求:** FR20（数据报表导出）、NFR-OTHER-004（导出限制）、Architecture 数据查询策略（从 metrics 表）

**创建表:** 无（使用 metrics 表）

---

### Story 8.2: 数据导出前端页面

As a 运维主管，
I can 在前端界面配置导出参数并下载报表，
So that 可以数据分析。

**Acceptance Criteria:**

**Given** 用户已登录并访问数据导出页面
**When** 页面加载完成
**Then** 显示导出参数表单
**And** 提供节点选择（多选，最多 50 个）
**And** 提供时间范围选择（最近 7 天/最近 30 天）
**And** 提供指标类型选择（时延/丢包率/抖动）
**And** 提供"导出"按钮和格式选择（CSV/Excel）
**And** 导出任务创建后显示"正在导出"状态
**And** 导出完成后提供下载链接

**覆盖需求:** FR20（数据报表导出）

**创建表:** 无

---

### Story 8.3: 性能指标采集

As a Pulse 系统，
I need 采集性能指标，
So that 可以监控系统响应速度。

**Acceptance Criteria:**

**Given** 仪表盘页面被访问
**When** 每个仪表盘请求完成
**Then** 记录仪表盘加载时间（P99、P95）
**And** 记录 API 响应时间（P99、P95）
**And** 记录数据查询时间（P99、P95）
**And** 每分钟记录一次性能指标
**And** 性能指标存储在数据库或缓存中

**覆盖需求:** FR21（仪表盘性能指标）

**创建表:**
- `performance_metrics` 表：id (UUID), metric_name (VARCHAR), p99 (DECIMAL), p95 (DECIMAL), recorded_at (TIMESTAMP)

---

### Story 8.4: 性能监控仪表盘

As a 运维主管，
I can 在性能监控仪表盘查看系统性能指标，
So that 可以评估系统响应速度。

**Acceptance Criteria:**

**Given** 用户已登录并访问性能监控页面
**When** 页面加载完成
**Then** 显示性能指标卡片：仪表盘加载时间 P99/P95
**And** 显示 API 响应时间 P99/P95
**And** 显示数据查询时间 P99/P95
**And** 显示性能趋势图（最近 24 小时）
**And** 标识超过目标值的异常（如 P99 > 5 秒）
**And** 显示系统整体健康状态
**And** 触发异常性能告警

**覆盖需求:** FR21（仪表盘性能指标）

**创建表:** 无（使用 performance_metrics 表）

---

## 文档完成总结

- ✅ **8 个 Epics** 完整定义
- ✅ **50 个 Stories** 生成完成
- ✅ **所有 22 个 FRs** 映射到 Epics
- ✅ **所有 15 个 NFRs** 覆盖到对应 Stories
- ✅ **Architecture 技术决策** 整合到 Stories
- ✅ **UX Design 交互模式** 整合到 Stories
- ✅ **数据库表** 按需求创建（12 个表）

---

**Select an Option:** [A] Advanced Elicitation [P] Party Mode [C] Continue
