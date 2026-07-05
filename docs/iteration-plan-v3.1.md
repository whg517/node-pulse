# 迭代实现规划 v3.1 — 基于 QA Journey Audit

**Date:** 2026-07-05
**Basis:** `docs/qa-journey-audit.md`（17 条旅程审计结果）
**Scope:** 本轮迭代修复全部「Group A 真实 bug + Group B 文档勘误」，共 **6 个代码问题 + 8 个文档问题**。
**Group C 重大缺口**（TLS/备份/升级/systemd/版本系统/解锁用户/JWT 轮换/热重载）工作量过大，不在本轮范围，留作独立迭代。

---

## 修复优先级与方式

### Group A — 真实 bug 修复（代码）

| # | 问题 | 修复方式 | 复杂度 |
|---|------|----------|--------|
| **A1**（Q-P0-1）| O-G1 审计清理未接线 | 新建 `auth.CleanupAdapter` 实现 `scheduler.Task`，包装 `CleanupJob.RunAll`；在 `registry.go` 注册。修正 `authentication.md:517-525` 把"已实现"改为"计划/澄清"或直接接线后保留为真 | 中 |
| **A2**（Q-P0-2）| F1 Reports 按钮无角色守卫 | `Reports.tsx` 引入 `useAuth`，按钮加 `disabled={!isAdmin}` 或隐藏；schedule dialog 触发前做角色判断 | 小 |
| **A3**（Q-P1-1）| UpdateUser 缺 self-role 防护 | 后端 `UpdateUser` 增加：若 `idParam == requestingUserID` 且 `role` 字段被改动，返回 400 `ErrCannotChangeOwnRole`。参照 `DeleteUser:564` 的 self-check 模式 | 小 |
| **A4**（Q-P1-2）| /health scheduler 不参与降级 | `health.go` scheduler 分支：scheduler 停摆 → `isHealthy=false`；任务 lastError 非空且 lastRun 超 3 倍 interval → `isDegraded=true` | 小 |
| **A5**（Q-P1-3）| routes.go:406 误写 | `nodes.Use(csrf.CSRFMiddleware())` → `probes.Use(csrf.CSRFMiddleware())`。一行字面修正 | 极小 |
| **A6**（Q-P2-1）| 导出删除按钮 no-op | 选项 a：实现后端 `DELETE /data/export/:id` + 前端接线；选项 b：移除假按钮。**选 a**（更完整），删除文件 + DB 记录 | 中 |

### Group B — 文档勘误（user-journey.md）

| # | 问题 | 修复方式 |
|---|------|----------|
| **B1**（Q-P3-1）| §12.1 J3 RBAC 不一致已过时 | 改为"前后端 RBAC 已对齐为 admin+operator，Operator 能看到创建/编辑/删除按钮"，删除"与 RBAC 不一致"措辞 |
| **B2**（Q-P3-2）| §12.1 行号 routes.go:347-353 错误 | 改为 `routes.go:379,382,385,388`（nodes 的 RBAC + 写路由） |
| **B3**（Q-P3-3）| §13 J4 角色关卡断言已过时 | 改为"`ProbeManagementPage.tsx:48-50` 现有 `canEdit = admin\|\|operator` UI 关卡，Viewer 点不到按钮" |
| **B4**（Q-P3-4）| §18.1 严重级别过滤已实现 | 拆为「自定义 headers：未实现；严重级别过滤：已实现（ADR-002 Tier 1）」 |
| **B5**（Q-P3-5）| §11 状态机符号名误导 | 后端方法名改为 `CanTransitionTo`（`alert_record.go:64-80`），注明 `isValidStatusTransition` 是前端纯函数 |
| **B6**（Q-P3-6）| §21 API Key 审计日志措辞 | 改为"审计日志查询（独立 `/settings/audit-logs` 页，非 API Key 详情内）" |
| **B7**（Q-P3-7）| §9.4 优雅关闭顺序 | 补 nodeSweeper.Stop 一步：刷批处理 → 停 nodeSweeper → 停调度 → HTTP drain |
| **B8**（Q-P3-8）| §10.3 alert:note_created | 标注"后端广播但前端不消费（new/updated/resolved 三个被消费）"或直接移除该事件名 |

### 同步项

- **A1 完成后**：更新 `user-journey.md` §9.2 数据保留表、§23.1 P0 O-G1 状态（从【文档谎称】→【已修复】）
- **A2 完成后**：更新 `user-journey.md` §23.1 P0 F1 状态（从 真实 bug →【已修复】）
- **A4 完成后**：更新 §8.2 「调度器 — metrics-cleanup 任务状态」描述（注明现在参与降级判定）
- **A6 完成后**：确认 §17.1「导出历史...删除」断言变为真实

---

## 实施批次

由于工作量适中，**全部在一个 worktree `fix-qa-journey-v3.1` 内完成**，分多个原子提交：

```
fix(qa): wire auth cleanup job into scheduler (O-G1)
fix(reports): add admin-only guard to schedule button (F1)
fix(admin): reject self role change in UpdateUser
fix(health): degrade status on scheduler failure
fix(routes): correct probes CSRF middleware typo
feat(export): implement delete endpoint and wire frontend
docs(journey): correct J3/J4 RBAC, line refs, and stale claims
docs(journey): fix J9 severity filter, J2 symbol, J12 wording
docs(journey): refine shutdown order and note_created event
docs(auth): correct audit log retention claim (was false)
docs(journey): mark O-G1 and F1 as resolved in §23
```

每个提交独立可编译，整体通过 4 个 gate（pulse lint+build、beacon lint+build、frontend lint+build）。

---

## 范围之外（Group C，未来迭代）

- D-G1 TLS 反代 / D-G2 备份 / D-G3 升级文档 / D-G4 systemd / D-G5 版本系统
- O-G2 解锁用户 UI / O-G3 JWT 轮换 / O-G4 Pulse 热重载
- F2 集中式 RBAC 路由守卫（根因 F1）

这些需要独立设计与 PRD 同步，本轮不涉及。
