# QA Journey Audit — user-journey.md v3.0 断言验证

**Date:** 2026-07-05
**Scope:** 对 `docs/user-journey.md` v3.0 的 17 条旅程（D1/D2/O1/O2 + J1–J13）逐条核对实际代码，识别「文档与实现不一致」的全部问题。
**Method:** 三组并行 QA 探查（部署/运维、J1–J7、J8–J13）+ 人工复核 P0 断言。所有判定均附 `file:line` 证据。

---

## 1. 总览

| 类别 | 数量 |
|------|------|
| 核对断言总数 | ~70 条 |
| ✅ 完全一致 | ~60 条 |
| ⚠️ 部分一致 / ❌ 不一致 | 11 条 |
| **新发现问题** | **8 条**（文档未记录的真实 bug / 文档过时未回写） |

**关键结论**：v3.0 文档的 P0/P1 缺口标注（O-G1、D-G1~G5、F1 等）**经代码核实全部属实**，可信度显著高于 v2.x。但 QA 同时发现了文档自身的若干**过时/不严谨**描述，以及代码层的若干**新 bug**。

---

## 2. 问题清单（按严重度排序）

### 2.1 P0 — 数据安全 / 真实 bug

| # | 旅程 | 问题 | 实际情况 | 证据 |
|---|------|------|----------|------|
| **Q-P0-1** | O2 / O-G1 | **审计清理【文档谎称】+ 函数死代码**（文档已诚实标注为缺口，但 `authentication.md` 仍谎称） | `auth.NewCleanupJob`（`cleanup_job.go:21`）实现了 refresh_tokens/token_blacklist/auth_audit_logs/api_keys 的批量清理（`RunAll:199-234`），但**全仓库无任何调用**（死代码）。已注册的 `cleanup.CleanupTask`（`registry.go:108`）**仅删 metrics 表**（`cleanup.go:77`）。`authentication.md:517-525` 仍谎称"90 天自动清理，每天 02:00 UTC"。4 张表无限增长。 | `pulse/internal/auth/cleanup_job.go:21`、`pulse/internal/server/registry.go:108`、`pulse/internal/cleanup/cleanup.go:77`、`docs/authentication.md:517-525` |
| **Q-P0-2** | J8 / F1 | **`Reports.tsx` 计划按钮无角色关卡**（文档已标注，确认属实） | `Reports.tsx` 全文（311 行）无 `isAdmin`/`role` 检查；按钮 `Reports.tsx:186` 对所有人可见，Viewer/Operator 点击 → 后端 admin-only 返回 403。 | `frontend/src/pages/Reports.tsx:186`、`pulse/internal/api/routes.go:471-473` |

### 2.2 P1 — 生产可用性 / 安全防护缺失

| # | 旅程 | 问题 | 实际情况 | 证据 |
|---|------|------|----------|------|
| **Q-P1-1** | J10 | **后端 `UpdateUser` 缺少 self-role-change 防护**（文档未记录的新发现） | 前端 `UsersPage.tsx:169,182` 对「改自己角色 / 删自己」做了 `disabled` 守卫；但后端 `UpdateUser`（`admin_user_handler.go:361-525`）**无 self-check**（对比 `DeleteUser:564` 有 `ErrCannotDeleteSelf`）。绕过前端可 PUT 自身 role，admin 可能把自己降级失去 admin 权限。 | `pulse/internal/api/admin_user_handler.go:361-525` vs `:564` |
| **Q-P1-2** | O1 | **`/api/v1/health` 的 scheduler 探测不参与降级判定**（文档未记录的新发现） | `health.go:122-143` scheduler 分支**只填充状态结构体**，不修改 `isHealthy`/`isDegraded`。scheduler 停摆/任务异常时健康端点仍报 `healthy(200)`，违背监控承诺。对比 alert_engine 分支（:159-162）会设置降级。 | `pulse/internal/health/health.go:122-143` vs `:159-162` |
| **Q-P1-3** | J4 | **`routes.go:406` 误写 `nodes.Use(csrf)` 应为 `probes.Use`**（文档未记录的新发现） | probes 段（:401-416）的 CSRF 中间件写成了 `nodes.Use(middleware.RBACMiddleware...)` 上一行的 `nodes.Use(csrf.CSRFMiddleware())`。由于 nodes/probes 都需 CSRF，功能上无害，但是真实的代码 bug。 | `pulse/internal/api/routes.go:406` |

### 2.2b P2 — 完善性

| # | 旅程 | 问题 | 实际情况 | 证据 |
|---|------|------|----------|------|
| **Q-P2-1** | J8 | **导出历史删除按钮是 no-op**（文档未记录的新发现） | 文档 §17.1 称「导出历史（过滤/分页/下载/删除）✅」，但 `Reports.tsx:245` 与 `DataExportPage.tsx:84` 都传 `onDelete={() => {}}`，后端 `routes.go:451-463` 无 DELETE 路由。删除功能实际不存在。 | `frontend/src/pages/Reports.tsx:245`、`frontend/src/pages/DataExportPage.tsx:84`、`pulse/internal/api/routes.go:451-463` |

### 2.3 P3 — 文档勘误（断言过时 / 不严谨）

| # | 旅程 | 问题 | 实际情况 | 证据 |
|---|------|------|----------|------|
| **Q-P3-1** | J3 | **§12.1 RBAC 不一致断言已过时**（已修复未回写） | 文档称「`NodeManagementPage.tsx:34` 把 canEdit 限定为 admin…Operator 看不到创建按钮」。实际第 34-36 行已改为 `canEdit = admin \|\| operator`，附修复注释，前后端 RBAC 已对齐，Operator 能看到按钮。 | `frontend/src/pages/NodeManagementPage.tsx:34-36`、`pulse/internal/api/routes.go:379,382,385,388` |
| **Q-P3-2** | J3 | **§12.1 行号引用 `routes.go:347-353` 错误** | 347-353 实际是 `configGroup`（admin-only 配置端点），不是 nodes。nodes 的 RBAC + POST/PUT/DELETE 在 363-388。 | `pulse/internal/api/routes.go:347-356` vs `:363-388` |
| **Q-P3-3** | J4 | **§13.1/§13.2「前端无显式角色关卡」断言已过时**（已修复未回写） | `ProbeManagementPage.tsx:48-50` 现有显式 UI 关卡 `canEdit = admin \|\| operator`，Viewer 点不到按钮，不会触发 403。 | `frontend/src/pages/ProbeManagementPage.tsx:48-50` |
| **Q-P3-4** | J9 | **§18.1「自定义 headers / 严重级别过滤 ❌ 计划中」过严** | **严重级别过滤已实现**（`router.go:66-86` `rule.Severities` 数组匹配 `event.Level`，ADR-002 Tier 1）。仅「自定义 headers」未实现（`models/webhook.go:6-12` 无 Headers 字段）。应拆分标注。 | `pulse/internal/webhook/router.go:66-86`、`pulse/internal/models/webhook.go:6-12` |
| **Q-P3-5** | J2 | **§11.1/§11.2 符号名 `isValidStatusTransition` 实为前端函数** | 该符号仅是前端纯函数（`api/alertRecords.ts:106`），后端方法实为 `CanTransitionTo`（`alert_record.go:64-80`，行号正确）。文档把它当后端校验器引用，措辞误导。 | `pulse/internal/models/alert_record.go:64-80`、`frontend/src/api/alertRecords.ts:106` |
| **Q-P3-6** | J12 | **§21「API Key + 审计日志查询」措辞误导** | 审计日志查询是独立的 `AuditLogsPage`，`ApiKeysPage` 内未集成该 key 的操作历史。文档把两者并列易让读者以为在 API Key 详情里能看操作历史。 | `frontend/src/pages/ApiKeysPage.tsx`、`frontend/src/pages/AuditLogsPage.tsx` |
| **Q-P3-7** | O2 | **§9.4 优雅关闭顺序描述略简** | 实际顺序为「刷批处理 → **停 nodeSweeper** → 停调度 → HTTP drain」，文档漏了 nodeSweeper.Stop 一步。语义无影响。 | `pulse/internal/server/server.go:82-86` |
| **Q-P3-8** | J1 | **§10.3「告警流含 alert:note_created 4 事件」前端未消费** | 后端确实广播 `alert:note_created`（`hub.go:149-158`），但前端 `useGlobalRealtime` 与 `AlertStream` 都不消费该事件（仅订阅 new/updated/resolved）。 | 后端发：`pulse/internal/realtime/hub.go:149-158`；前端未处理：`frontend/src/hooks/useGlobalRealtime.ts:48-89` |

---

## 3. 已确认属实的文档缺口（无需文档修正，列入实现规划）

以下文档标注的缺口经核实**完全属实**，是真正的实现 backlog（不是文档错误）：

| 缺口 | 核实结果 |
|------|----------|
| **D-G1** 无 TLS 终止 | ✅ Pulse 监听明文 HTTP，无反代文档 |
| **D-G2** 无数据库备份 | ✅ 无 pg_dump/cron，仅 postgres_data 卷 |
| **D-G3** 无升级文档 | ✅ 无 upgrade.md / runbook |
| **D-G4** Beacon 无 systemd | ✅ `make install` 仅拷贝二进制 |
| **D-G5** 无版本/发布系统 | ✅ 无 git tags / release workflow / -ldflags version |
| **D-G6** `.env.example` "Frontend (nginx)" 死引用 | ✅ `.env.example:37-38` `FRONTEND_PORT=80` 无消费者 |
| **O-G1** 审计清理未接线 | ✅（见 Q-P0-1）|
| **O-G2** 无管理员解锁用户 UI | ✅ UsersPage 无解锁按钮，后端无 endpoint |
| **O-G3** JWT 无轮换窗口 | ✅ `VerifyKeyID` 仅匹配单 keyID |
| **O-G4** Pulse 无热重载 | ✅ Beacon 有 SIGHUP，Pulse 无 |
| **F1** Reports 计划按钮无角色关卡 | ✅（见 Q-P0-2）|

---

## 4. 修复策略分组

按「问题性质」归类，便于分批实现：

### Group A — 真实 bug 修复（代码改动）
- **Q-P0-1**（O-G1）：注册 auth 清理任务到 scheduler + 修正 `authentication.md`
- **Q-P0-2**（F1）：`Reports.tsx` 加角色守卫
- **Q-P1-1**：后端 `UpdateUser` 加 self-role-change 防护
- **Q-P1-2**：`/health` scheduler 探测参与降级判定
- **Q-P1-3**：`routes.go:406` 修正 `nodes.Use` → `probes.Use`
- **Q-P2-1**：导出删除按钮 — 实现 DELETE 端点 + 接线前端（或移除假按钮）

### Group B — 文档勘误（仅 user-journey.md）
- Q-P3-1 / Q-P3-2 / Q-P3-3 / Q-P3-4 / Q-P3-5 / Q-P3-6 / Q-P3-7 / Q-P3-8

### Group C — 重大缺口（实现工作量大，本轮不做，仅记录）
- D-G1（TLS）、D-G2（备份）、D-G3（升级文档）、D-G4（systemd）、D-G5（版本系统）、O-G2（解锁用户）、O-G3（JWT 轮换）、O-G4（热重载）

---

## 5. 下一步

见 `docs/iteration-plan-v3.1.md`（基于本报告的迭代实现规划）。
