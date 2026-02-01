# Story 4.4: Dashboard Homepage with Node List

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a 运维主管,
I want 在仪表盘首页查看所有节点的全局视图和健康状态,
So that 可以快速掌握整体网络质量状况并识别异常节点。

## Acceptance Criteria

**Given** 用户已登录并访问仪表盘首页路径 `/` 或 `/dashboard`
**When** 仪表盘页面加载完成
**Then** 显示全局节点列表表格（包含所有注册的 Beacon 节点）
**And** 每个节点显示以下信息：
  - 节点名称（如 "Beacon-CN-East-1"）
  - IP 地址（如 "192.168.1.100"）
  - 地区标签（如 "华东", "华北", "华南"）
  - 健康状态指示器（红/黄/绿三色状态）
    - 绿色：健康（所有指标正常）
    - 黄色：预警（部分指标接近阈值）
    - 红色：异常（指标超过阈值或节点离线）
**And** 显示异常 TOP5 列表（按严重程度排序，优先显示红色状态节点）
**And** 显示核心指标均值卡片（全局汇总统计）：
  - 平均时延（所有在线节点的平均 RTT）
  - 平均丢包率（所有在线节点的平均丢包率）
  - 平均抖动（所有在线节点的平均抖动）
**And** 仪表盘初始加载时间 ≤5 秒（P95 ≤3 秒）
**And** 节点状态自动刷新，刷新周期 ≤5 秒
**And** 支持点击节点行跳转到节点详情页（`/nodes/:id`）

**Given** 用户在仪表盘页面停留
**When** 5 秒刷新周期到达
**Then** 自动从后端 API 拉取最新节点状态
**And** 更新表格中的健康状态指示器（不刷新整个页面，局部更新）
**And** 更新核心指标均值卡片
**And** 更新异常 TOP5 列表

**Given** Beacon 节点离线或心跳超时（超过 2 个周期未收到心跳，即 120 秒）
**When** 节点状态更新
**Then** 健康状态显示为红色
**And** 节点状态显示为 "离线"
**And** 该节点出现在异常 TOP5 列表中

## Tasks / Subtasks

- [ ] Task 1: Create Dashboard page component structure (AC: Given, When - 仪表盘页面)
  - [ ] Create `src/pages/Dashboard.tsx` as main dashboard page component
  - [ ] Create `src/components/dashboard/NodeListTable.tsx` for node list table
  - [ ] Create `src/components/dashboard/HealthStatusBadge.tsx` for status indicator (红/黄/绿)
  - [ ] Create `src/components/dashboard/TopAnomaliesList.tsx` for TOP5 异常列表
  - [ ] Create `src/components/dashboard/MetricsSummaryCards.tsx` for core metrics display
  - [ ] Add Dashboard route to React Router (`/` or `/dashboard`)
  - [ ] Ensure route is protected by authentication guard (from Story 4.1)

- [ ] Task 2: Implement NodeListTable component with real-time data (AC: Then - 节点列表表格)
  - [ ] Define TypeScript interface for NodeListData (extends NodeDTO from API layer)
  - [ ] Implement table layout with columns: 节点名称, IP, 地区, 健康状态
  - [ ] Use nodesStore from Zustand to fetch and cache node list
  - [ ] Call fetchNodes() API from api/nodes.ts on component mount
  - [ ] Implement loading state (skeleton or spinner) while fetching data
  - [ ] Implement error state with retry button if API call fails
  - [ ] Implement empty state if no nodes exist
  - [ ] Add click handler to node rows to navigate to `/nodes/:id`
  - [ ] Use Tailwind CSS for table styling (border, padding, hover effects)
  - [ ] Ensure responsive design (scrollable table on small screens)

- [ ] Task 3: Implement HealthStatusBadge component (AC: And - 健康状态指示器)
  - [ ] Create component accepting `status` prop: 'healthy' | 'warning' | 'critical' | 'offline'
  - [ ] Display status badge with appropriate color:
    - Green badge for 'healthy' (Tailwind: bg-green-100 text-green-800)
    - Yellow badge for 'warning' (Tailwind: bg-yellow-100 text-yellow-800)
    - Red badge for 'critical' (Tailwind: bg-red-100 text-red-800)
    - Gray badge for 'offline' (Tailwind: bg-gray-100 text-gray-800)
  - [ ] Display status text label: "健康", "预警", "异常", "离线"
  - [ ] Add visual indicator (colored dot or icon) for quick scanning
  - [ ] Ensure accessibility (proper ARIA labels for color-blind users)
  - [ ] Use rounded corners and proper padding for polished UI

- [ ] Task 4: Implement health status determination logic (AC: And - 红/黄/绿状态判断)
  - [ ] Create utility function `determineHealthStatus(metrics: NodeMetrics)` in `src/utils/healthStatus.ts`
  - [ ] Define health status rules based on alert thresholds:
    - Green: latency < threshold, packet_loss < threshold, jitter < threshold, and node is online
    - Yellow: any metric within 80-100% of threshold OR node hasn't reported heartbeat in 60-120 seconds
    - Red: any metric exceeds threshold OR node offline (>120 seconds without heartbeat)
  - [ ] Use default thresholds if no alert rules configured:
    - Latency threshold: 200ms
    - Packet loss threshold: 5%
    - Jitter threshold: 50ms
  - [ ] Handle edge cases: missing metrics, null values, division by zero
  - [ ] Export function for use in components and tests
  - [ ] Add unit tests for status determination logic

- [ ] Task 5: Implement TopAnomaliesList component (AC: And - 异常 TOP5 列表)
  - [ ] Create component accepting `nodes` array prop
  - [ ] Sort nodes by severity (critical > warning > healthy)
  - [ ] Filter to show top 5 most critical nodes (prefer critical status, then warning)
  - [ ] Display list with node name, health status badge, and key metric (e.g., latency)
  - [ ] Show severity indicator (icon or color)
  - [ ] Add click handler to navigate to node detail page
  - [ ] Handle empty state (no anomalies)
  - [ ] Use Tailwind CSS for card styling

- [ ] Task 6: Implement MetricsSummaryCards component (AC: And - 核心指标均值卡片)
  - [ ] Create component with three metric cards:
    - Average Latency (mean of all nodes' latency_ms)
    - Average Packet Loss Rate (mean of all nodes' packet_loss_rate)
    - Average Jitter (mean of all nodes' jitter_ms)
  - [ ] Fetch metrics data using fetchMetrics() API from api/data.ts
  - [ ] Calculate mean values for online nodes only
  - [ ] Display metrics with proper units: ms, %, ms
  - [ ] Use card layout with icon, label, and value
  - [ ] Apply color coding: green for good, yellow for warning, red for critical
  - [ ] Handle loading and error states
  - [ ] Use Tailwind CSS for responsive card grid (1 col mobile, 3 cols desktop)

- [ ] Task 7: Implement real-time data polling with useDashboardData hook (AC: When - 5秒刷新)
  - [ ] Create custom hook `src/hooks/useDashboardData.ts`
  - [ ] Use useEffect to set up polling interval (5000ms = 5 seconds)
  - [ ] Call fetchNodes() and fetchMetrics() APIs on each interval
  - [ ] Update nodesStore with fresh data from API
  - [ ] Implement cleanup function to clear interval on component unmount
  - [ ] Add polling state (isPolling) to show refresh indicator
  - [ ] Implement pause/resume polling mechanism (optional, for UX)
  - [ ] Handle API errors gracefully (don't break polling on single failure)
  - [ ] Avoid excessive API calls with proper dependency array
  - [ ] Test polling doesn't cause memory leaks

- [ ] Task 8: Implement offline node detection (AC: When - 节点离线检测)
  - [ ] Track last heartbeat timestamp for each node
  - [ ] Define offline threshold: 120 seconds (2 heartbeat cycles)
  - [ ] Create utility function `isNodeOffline(lastHeartbeat: string): boolean`
  - [ ] Update health status to 'offline' if threshold exceeded
  - [ ] Display "离线" status badge for offline nodes
  - [ ] Include offline nodes in TOP5 anomalies list
  - [ ] Test offline detection with mocked timestamps

- [ ] Task 9: Implement performance optimizations (AC: And - 加载时间 ≤5 秒)
  - [ ] Ensure initial API calls are parallel (fetchNodes() and fetchMetrics() simultaneously)
  - [ ] Use React.memo() for expensive child components
  - [ ] Implement virtual scrolling for node list if node count > 50 (future-proofing)
  - [ ] Lazy load non-critical components (MetricsSummaryCards can load after table)
  - [ ] Add loading skeleton for better perceived performance
  - [ ] Measure dashboard load time with performance API
  - [ ] Optimize re-renders by preventing unnecessary state updates
  - [ ] Use React DevTools Profiler to identify performance bottlenecks

- [ ] Task 10: Create comprehensive tests for Dashboard components (AC: 测试覆盖)
  - [ ] Unit tests for health status determination logic (determineHealthStatus)
  - [ ] Unit tests for offline detection logic (isNodeOffline)
  - [ ] Component tests for NodeListTable (render, loading, error, empty states)
  - [ ] Component tests for HealthStatusBadge (all status types)
  - [ ] Component tests for TopAnomaliesList (sorting, filtering)
  - [ ] Component tests for MetricsSummaryCards (calculation, display)
  - [ ] Integration test for useDashboardData hook (mock polling)
  - [ ] Integration test for Dashboard page (end-to-end flow)
  - [ ] Test navigation to node detail page on click
  - [ ] Test polling mechanism doesn't cause memory leaks

## Dev Notes

### Epic Analysis

**Epic 4: 实时监控仪表盘** - 运维主管可以在仪表盘上查看所有节点的实时状态和核心指标

**Story Context in Epic:**
- Story 4.1: Frontend route auth guard (已完成) - 提供路由保护
- Story 4.2: Zustand state management (已完成) - 创建了 nodesStore, dashboardStore
- Story 4.3: API Layer Encapsulation (已完成) - **提供了完整的 API 调用层**
  - ✅ api/nodes.ts: fetchNodes(), createNode(), updateNode(), deleteNode(), fetchNodeStatus()
  - ✅ api/data.ts: fetchMetrics(), fetchHistory(), exportData()
  - ✅ api/types.ts: NodeDTO, MetricsDTO 等类型定义
  - ✅ 统一错误处理和 Session Cookie 认证
- **Story 4.4: 仪表盘首页与节点列表** (本故事) - **创建仪表盘 UI 组件**
- Story 4.5-4.9: More dashboard features (节点详情页, 趋势图, 通知组件)

**Critical Dependencies:**
- ✅ **Story 4.3 已完成**: API 层完整实现，可以直接调用 fetchNodes(), fetchMetrics()
- ✅ **Story 4.2 已完成**: Zustand stores 已创建，nodesStore 和 dashboardStore 可以复用
- ✅ **Story 4.1 已完成**: 路由守卫已实现，仪表盘路由受保护
- ⚠️ **后端 API 必须实现**: 确保后端已实现节点管理和数据查询 API（Epic 2, Epic 3）

### Architecture Alignment

**Frontend Architecture** [Source: Architecture.md#Frontend Architecture]:
```
pulse-frontend/src/
├── pages/
│   └── Dashboard.tsx          # NEW - 主仪表盘页面
├── components/
│   └── dashboard/             # NEW - 仪表盘组件目录
│       ├── NodeListTable.tsx       # 节点列表表格
│       ├── HealthStatusBadge.tsx   # 健康状态徽章
│       ├── TopAnomaliesList.tsx    # TOP5 异常列表
│       └── MetricsSummaryCards.tsx # 核心指标卡片
├── hooks/
│   └── useDashboardData.ts    # NEW - 数据轮询 Hook
├── utils/
│   └── healthStatus.ts        # NEW - 健康状态判断逻辑
├── stores/
│   ├── nodesStore.ts          # EXISTS from Story 4.2
│   ├── dashboardStore.ts      # EXISTS from Story 4.2
│   └── authStore.ts           # EXISTS from Story 4.2
├── api/
│   ├── nodes.ts               # EXISTS from Story 4.3 - fetchNodes()
│   ├── data.ts                # EXISTS from Story 4.3 - fetchMetrics()
│   └── types.ts               # EXISTS from Story 4.3 - NodeDTO, MetricsDTO
└── types/
    └── auth.ts                # EXISTS from Story 4.1
```

**UI/UX Design Principles** [Source: UX Design Specification]:
- **侧边栏导航**: 左侧固定侧边栏用于主要模块切换（仪表盘、节点管理、告警配置）
- **卡片展开详情**: 点击节点卡片无需翻页直接展开详情，减少点击次数
- **状态刷新**: 实时数据自动刷新，刷新周期 ≤5 秒
- **异常 TOP5 列表**: 按严重程度排序，优先显示红色状态节点
- **健康状态指示**: 红/黄/绿三色状态，快速识别问题节点

**Performance Requirements** [Source: PRD.md#NFR-PERF-002]:
- 仪表盘加载时间 ≤5 秒（P99 ≤3 秒，P95 ≤2 秒）
- 状态刷新周期 ≤5 秒
- API 响应时间 ≤500ms（P99），≤200ms（P95）

### Project Structure Notes

**Alignment with unified project structure** [Source: Architecture.md#Project Structure]:
- ✅ 使用 `src/pages/` 存放页面组件
- ✅ 使用 `src/components/dashboard/` 存放仪表盘相关组件
- ✅ 使用 `src/hooks/` 存放自定义 Hooks
- ✅ 使用 `src/utils/` 存放工具函数
- ✅ 复用 Zustand stores（nodesStore, dashboardStore）
- ✅ 复用 API layer（fetchNodes, fetchMetrics）
- ✅ 使用 Tailwind CSS 进行样式开发
- ✅ 遵循 TypeScript 类型安全

**New Files to Create:**
- `src/pages/Dashboard.tsx` - 主仪表盘页面
- `src/components/dashboard/NodeListTable.tsx` - 节点列表表格
- `src/components/dashboard/HealthStatusBadge.tsx` - 健康状态徽章
- `src/components/dashboard/TopAnomaliesList.tsx` - TOP5 异常列表
- `src/components/dashboard/MetricsSummaryCards.tsx` - 核心指标卡片
- `src/hooks/useDashboardData.ts` - 数据轮询 Hook
- `src/utils/healthStatus.ts` - 健康状态判断逻辑

### Previous Story Intelligence

**From Story 4.3 (API Layer Encapsulation)** [Source: Story 4.3 Implementation]:
- ✅ **Complete API layer implemented**:
  - `api/nodes.ts`: fetchNodes(), createNode(), updateNode(), deleteNode(), fetchNodeStatus()
  - `api/data.ts`: fetchMetrics(), fetchHistory(), exportData()
  - `api/types.ts`: NodeDTO, MetricsDTO, HistoryQueryDTO, ExportQueryDTO
  - Unified error handling with custom error classes (ApiError, AuthenticationError, ValidationError)
  - All API functions use Session Cookie authentication
- **Learnings**:
  - API layer is centralized and type-safe
  - All DTO types are exported and reusable
  - Error handling is consistent across all API calls
  - API client automatically includes credentials for Session Cookie

**From Story 4.2 (Zustand State Management)** [Source: Story 4.2 Implementation]:
- ✅ **Zustand stores created**:
  - `nodesStore`: nodes array, loading state, error state, fetchNodes action
  - `dashboardStore`: dashboard filters, time range, refresh settings
  - `authStore`: user info, auth state, login/logout actions
- **Learnings**:
  - Zustand stores are lightweight and don't need Provider
  - Stores can be used directly in any component
  - TypeScript types are enforced in stores
  - Actions and state are clearly separated

**From Story 4.1 (Frontend Route Auth Guard)** [Source: Story 4.1 Implementation]:
- ✅ **React Router v6 configured** with protected routes
- ✅ **Auth guard implemented**: redirects unauthenticated users to `/login`
- ✅ **Session Cookie authentication**: automatic redirect to original page after login
- **Learnings**:
  - Dashboard route is protected and requires authentication
  - Use `<Outlet />` for nested routes
  - Auth state is managed in authStore

**Git History Analysis** (last 5 commits):
- `4a73254 feat: Implement API Layer Encapsulation (Story 4.3)` - **Latest commit**
- `15852da feat: add Auto-Sprint` - Auto-sprint workflow added
- `dd7aa52 feat: Implement Zustand State Management (Story 4.2)` - Zustand stores created
- `9b4b107 feat: Implement Frontend Route Auth Guard (Story 4.1)` - Route guard implemented
- **Pattern**: Each story builds on previous work
- **Code Quality**: Comprehensive testing, type safety, clean architecture

**Key Learnings from Previous Stories:**
- API layer is complete and ready to use
- Zustand stores provide state management foundation
- Routes are protected and authentication flow works
- Testing infrastructure is set up (Vitest + React Testing Library)
- All stories include comprehensive test coverage

### Technical Requirements

**Data Fetching Strategy**:
- Use `useDashboardData` custom hook to encapsulate polling logic
- Call `fetchNodes()` from api/nodes.ts to get node list
- Call `fetchMetrics()` from api/data.ts to get real-time metrics
- Parallel fetch both APIs on mount (Promise.all or independent useEffects)
- Update Zustand stores with fetched data
- Poll every 5 seconds (5000ms interval)

**Health Status Determination Logic** [Source: PRD.md#FR4]:
```typescript
// src/utils/healthStatus.ts
export type HealthStatus = 'healthy' | 'warning' | 'critical' | 'offline'

export interface HealthThresholds {
  latency: number      // default: 200ms
  packetLoss: number   // default: 5%
  jitter: number       // default: 50ms
}

export function determineHealthStatus(
  metrics: NodeMetrics,
  thresholds: HealthThresholds
): HealthStatus {
  // Offline check
  if (isNodeOffline(metrics.last_heartbeat)) {
    return 'offline'
  }

  // Critical check
  if (
    metrics.latency_ms > thresholds.latency ||
    metrics.packet_loss_rate > thresholds.packetLoss ||
    metrics.jitter_ms > thresholds.jitter
  ) {
    return 'critical'
  }

  // Warning check (80-100% of threshold)
  if (
    metrics.latency_ms > thresholds.latency * 0.8 ||
    metrics.packet_loss_rate > thresholds.packetLoss * 0.8 ||
    metrics.jitter_ms > thresholds.jitter * 0.8
  ) {
    return 'warning'
  }

  return 'healthy'
}

export function isNodeOffline(lastHeartbeat: string): boolean {
  const heartbeatTime = new Date(lastHeartbeat).getTime()
  const now = Date.now()
  const offlineThreshold = 120 * 1000 // 120 seconds
  return (now - heartbeatTime) > offlineThreshold
}
```

**TypeScript Type Definitions** [Reused from api/types.ts]:
```typescript
import type { NodeDTO, MetricsDTO } from '@/api/types'

export interface NodeListData extends NodeDTO {
  healthStatus: 'healthy' | 'warning' | 'critical' | 'offline'
  metrics?: MetricsDTO
}

export interface DashboardMetrics {
  averageLatency: number
  averagePacketLoss: number
  averageJitter: number
  totalNodes: number
  onlineNodes: number
  offlineNodes: number
}
```

**useDashboardData Hook Implementation**:
```typescript
// src/hooks/useDashboardData.ts
import { useEffect, useState } from 'react'
import { useNodesStore } from '@/stores/nodesStore'
import { fetchNodes, fetchMetrics } from '@/api'
import type { MetricsDTO } from '@/api/types'

export interface DashboardData {
  nodes: NodeListData[]
  metrics: MetricsDTO[]
  isLoading: boolean
  error: Error | null
  isPolling: boolean
}

export function useDashboardData(pollingInterval = 5000) {
  const [data, setData] = useState<DashboardData>({
    nodes: [],
    metrics: [],
    isLoading: true,
    error: null,
    isPolling: false,
  })

  const fetchData = async () => {
    try {
      // Parallel fetch for better performance
      const [nodesResponse, metricsResponse] = await Promise.all([
        fetchNodes(),
        fetchMetrics([]), // empty array = fetch all nodes
      ])

      // Transform data with health status
      const nodesWithHealth = nodesResponse.data.map(node => ({
        ...node,
        healthStatus: determineHealthStatus(node.metrics),
      }))

      setData({
        nodes: nodesWithHealth,
        metrics: metricsResponse.data,
        isLoading: false,
        error: null,
        isPolling: true,
      })
    } catch (error) {
      setData(prev => ({
        ...prev,
        isLoading: false,
        error: error as Error,
        isPolling: false,
      }))
    }
  }

  useEffect(() => {
    // Initial fetch
    fetchData()

    // Set up polling
    const interval = setInterval(fetchData, pollingInterval)

    // Cleanup on unmount
    return () => clearInterval(interval)
  }, [pollingInterval])

  return {
    ...data,
    refetch: fetchData,
  }
}
```

**Component Structure - Dashboard Page**:
```typescript
// src/pages/Dashboard.tsx
import { useDashboardData } from '@/hooks/useDashboardData'
import { NodeListTable } from '@/components/dashboard/NodeListTable'
import { TopAnomaliesList } from '@/components/dashboard/TopAnomaliesList'
import { MetricsSummaryCards } from '@/components/dashboard/MetricsSummaryCards'
import { HealthStatusBadge } from '@/components/dashboard/HealthStatusBadge'

export default function Dashboard() {
  const { nodes, metrics, isLoading, error } = useDashboardData()

  if (isLoading) {
    return <DashboardSkeleton />
  }

  if (error) {
    return <ErrorState error={error} onRetry={refetch} />
  }

  return (
    <div className="dashboard-page">
      <header>
        <h1>Dashboard</h1>
        <p>Real-time network monitoring</p>
      </header>

      {/* Metrics Summary Cards */}
      <MetricsSummaryCards metrics={metrics} />

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Node List Table (2/3 width) */}
        <div className="lg:col-span-2">
          <NodeListTable nodes={nodes} />
        </div>

        {/* Top Anomalies List (1/3 width) */}
        <div>
          <TopAnomaliesList nodes={nodes} />
        </div>
      </div>
    </div>
  )
}
```

**Component Structure - Node List Table**:
```typescript
// src/components/dashboard/NodeListTable.tsx
import { useNavigate } from 'react-router-dom'
import { HealthStatusBadge } from './HealthStatusBadge'
import type { NodeListData } from '@/types/dashboard'

interface NodeListTableProps {
  nodes: NodeListData[]
}

export function NodeListTable({ nodes }: NodeListTableProps) {
  const navigate = useNavigate()

  const handleRowClick = (nodeId: string) => {
    navigate(`/nodes/${nodeId}`)
  }

  return (
    <div className="node-list-table">
      <table className="min-w-full divide-y divide-gray-200">
        <thead>
          <tr>
            <th>Node Name</th>
            <th>IP Address</th>
            <th>Region</th>
            <th>Health Status</th>
          </tr>
        </thead>
        <tbody>
          {nodes.map(node => (
            <tr
              key={node.id}
              onClick={() => handleRowClick(node.id)}
              className="cursor-pointer hover:bg-gray-50"
            >
              <td>{node.name}</td>
              <td>{node.ip}</td>
              <td>{node.region}</td>
              <td>
                <HealthStatusBadge status={node.healthStatus} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
```

**Component Structure - Health Status Badge**:
```typescript
// src/components/dashboard/HealthStatusBadge.tsx
import type { HealthStatus } from '@/utils/healthStatus'

interface HealthStatusBadgeProps {
  status: HealthStatus
}

const statusConfig = {
  healthy: {
    label: '健康',
    bgColor: 'bg-green-100',
    textColor: 'text-green-800',
    dotColor: 'bg-green-500',
  },
  warning: {
    label: '预警',
    bgColor: 'bg-yellow-100',
    textColor: 'text-yellow-800',
    dotColor: 'bg-yellow-500',
  },
  critical: {
    label: '异常',
    bgColor: 'bg-red-100',
    textColor: 'text-red-800',
    dotColor: 'bg-red-500',
  },
  offline: {
    label: '离线',
    bgColor: 'bg-gray-100',
    textColor: 'text-gray-800',
    dotColor: 'bg-gray-500',
  },
}

export function HealthStatusBadge({ status }: HealthStatusBadgeProps) {
  const config = statusConfig[status]

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium ${config.bgColor} ${config.textColor}`}
      role="status"
      aria-label={`Health status: ${config.label}`}
    >
      <span className={`w-2 h-2 rounded-full ${config.dotColor}`} />
      {config.label}
    </span>
  )
}
```

**Tailwind CSS Styling Guidelines**:
- Use responsive design classes (`grid-cols-1 lg:grid-cols-3`)
- Use spacing utilities (`gap-6`, `p-4`, `m-2`)
- Use color utilities for status badges (`bg-green-100`, `text-green-800`)
- Use typography utilities (`text-xs`, `font-medium`, `text-lg`)
- Use interactive states (`hover:bg-gray-50`, `cursor-pointer`)
- Use layout utilities (`flex`, `grid`, `min-w-full`)
- Follow consistent spacing scale (2, 4, 6, 8)

### Implementation Guidelines

**Step-by-Step Implementation Order:**

1. **Create utility functions first** (no dependencies):
   - `src/utils/healthStatus.ts` - determineHealthStatus(), isNodeOffline()
   - Write unit tests for utility functions

2. **Create custom hook** (depends on utils and API layer):
   - `src/hooks/useDashboardData.ts` - polling logic, data fetching
   - Write integration tests with mocked API

3. **Create small UI components** (depends on utils):
   - `src/components/dashboard/HealthStatusBadge.tsx`
   - Write component tests for all status types

4. **Create medium UI components** (depends on small components):
   - `src/components/dashboard/NodeListTable.tsx`
   - `src/components/dashboard/TopAnomaliesList.tsx`
   - `src/components/dashboard/MetricsSummaryCards.tsx`
   - Write component tests for each

5. **Create main Dashboard page** (depends on all above):
   - `src/pages/Dashboard.tsx`
   - Add route to App.tsx
   - Write integration test for full page

6. **Performance optimization** (after basic implementation):
   - Add React.memo() to expensive components
   - Optimize re-renders with useMemo, useCallback
   - Add loading skeletons for better UX
   - Measure performance with DevTools Profiler

**Critical Implementation Notes:**
- ✅ **Reuse existing API layer**: Don't create new API functions, use api/nodes.ts and api/data.ts
- ✅ **Reuse Zustand stores**: Use nodesStore and dashboardStore for state management
- ✅ **Follow naming conventions**: Components use PascalCase, functions use camelCase
- ✅ **Maintain type safety**: All props and state must have TypeScript types
- ✅ **Handle all error states**: Loading, error, empty states for all components
- ✅ **Accessibility**: Add ARIA labels, keyboard navigation, semantic HTML
- ✅ **Responsive design**: Mobile-first approach, test on different screen sizes
- ✅ **Performance**: Parallel API calls, polling cleanup, React optimization

### Testing Requirements

**Unit Tests** (Vitest + Testing Library):
- `src/utils/__tests__/healthStatus.test.ts`:
  - Test determineHealthStatus() for all status types
  - Test isNodeOffline() with various timestamps
  - Test edge cases (null values, missing metrics)
- `src/hooks/__tests__/useDashboardData.test.ts`:
  - Test polling mechanism (setup, cleanup, interval)
  - Test data fetching and transformation
  - Test loading, error, and success states
  - Test polling doesn't cause memory leaks

**Component Tests** (React Testing Library):
- `src/components/dashboard/__tests__/HealthStatusBadge.test.tsx`:
  - Render all status types with correct colors
  - Test ARIA labels for accessibility
- `src/components/dashboard/__tests__/NodeListTable.test.tsx`:
  - Render table with nodes data
  - Test navigation on row click
  - Test loading, error, empty states
- `src/components/dashboard/__tests__/TopAnomaliesList.test.tsx`:
  - Test sorting by severity (critical > warning)
  - Test top 5 filtering
- `src/components/dashboard/__tests__/MetricsSummaryCards.test.tsx`:
  - Test metric calculations (mean values)
  - Test loading and error states

**Integration Tests**:
- `src/pages/__tests__/Dashboard.test.tsx`:
  - Full page render with mocked API
  - Test navigation to node detail page
  - Test polling mechanism integration
  - Test error handling and retry

**Test Coverage Requirements**:
- Utility functions: 100% coverage
- Custom hooks: 100% coverage
- Components: ≥90% coverage
- Integration tests for critical user flows

**Mocking Strategy**:
- Mock fetch API for api/nodes.ts and api/data.ts
- Mock Date.now() for offline detection tests
- Mock timers (setInterval, clearInterval) for polling tests
- Use vi.clearAllTimers() in afterEach cleanup

### References

- [Source: Architecture.md#Frontend Architecture] - Project structure, component organization
- [Source: Architecture.md#UI/UX Design Principles] - Dashboard design patterns
- [Source: Architecture.md#Performance Requirements] - Loading time, refresh rate requirements
- [Source: PRD.md#FR4] - Real-time dashboard functional requirements
- [Source: PRD.md#NFR-PERF-002] - Dashboard loading time ≤5 seconds
- [Source: Epics.md > Epic 4 > Story 4.4] - Story requirements and acceptance criteria
- [Source: Story 4.3 Implementation] - API layer to reuse (fetchNodes, fetchMetrics)
- [Source: Story 4.2 Implementation] - Zustand stores to reuse (nodesStore, dashboardStore)
- [Source: Story 4.1 Implementation] - Route auth guard integration
- [Source: UX Design Specification] - UI/UX patterns and interactions

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-5-20250929

### Debug Log References

### Completion Notes List

**Ultimate Context Engine Analysis Summary:**

✅ **Epic Context**: Story 4.4 is the fourth story in Epic 4 (Dashboard UI), following Stories 4.1-4.3 which completed the foundational infrastructure (routes, state management, API layer).

✅ **Story Foundation**: Dashboard homepage with global node list, health status indicators (红/黄/绿), TOP5 anomalies list, and core metrics summary. Real-time polling every 5 seconds.

✅ **Critical Dependencies Completed**:
- Story 4.3: Complete API layer (fetchNodes, fetchMetrics, DTO types)
- Story 4.2: Zustand stores (nodesStore, dashboardStore)
- Story 4.1: Route auth guard and React Router setup
- All previous stories passed comprehensive testing (114 tests passing)

✅ **Architecture Compliance**:
- Frontend architecture: src/pages/, src/components/dashboard/, src/hooks/, src/utils/
- UI/UX principles: Side nav, card expansion, real-time refresh, TOP5 anomalies
- Performance: Dashboard load ≤5s, refresh ≤5s, API response ≤500ms
- Styling: Tailwind CSS with responsive design
- Type safety: Full TypeScript with strict mode

✅ **Previous Story Intelligence**:
- API layer is complete, type-safe, and unified
- Zustand stores are lightweight and don't need Provider
- Route protection and authentication flow works
- Testing infrastructure is mature (Vitest + React Testing Library)
- Git history shows consistent quality and comprehensive commits

✅ **Technical Stack**:
- React 18 with TypeScript
- React Router v6 for routing
- Zustand for state management
- Tailwind CSS for styling
- Vitest + React Testing Library for testing
- Apache ECharts (for future trend charts in Story 4.6-4.7)

✅ **Implementation Readiness**:
- All prerequisite stories completed and committed
- API layer fully functional and tested
- State management infrastructure ready
- Routing and authentication working
- Testing patterns established
- No blockers detected

**Ready for Development: Story 4.4 is fully prepared for implementation by the dev-story agent.**

### File List

**New Files to Create (Implementation Phase):**
- pulse-frontend/src/pages/Dashboard.tsx - Main dashboard page component
- pulse-frontend/src/components/dashboard/NodeListTable.tsx - Node list table with health status
- pulse-frontend/src/components/dashboard/HealthStatusBadge.tsx - Status badge component (红/黄/绿)
- pulse-frontend/src/components/dashboard/TopAnomaliesList.tsx - TOP5 anomalies list
- pulse-frontend/src/components/dashboard/MetricsSummaryCards.tsx - Core metrics summary cards
- pulse-frontend/src/hooks/useDashboardData.ts - Real-time data polling hook
- pulse-frontend/src/utils/healthStatus.ts - Health status determination logic
- pulse-frontend/src/utils/__tests__/healthStatus.test.ts - Unit tests for health status
- pulse-frontend/src/hooks/__tests__/useDashboardData.test.ts - Hook tests
- pulse-frontend/src/components/dashboard/__tests__/HealthStatusBadge.test.tsx - Component tests
- pulse-frontend/src/components/dashboard/__tests__/NodeListTable.test.tsx - Component tests
- pulse-frontend/src/components/dashboard/__tests__/TopAnomaliesList.test.tsx - Component tests
- pulse-frontend/src/components/dashboard/__tests__/MetricsSummaryCards.test.tsx - Component tests
- pulse-frontend/src/pages/__tests__/Dashboard.test.tsx - Integration tests

**Files to Modify:**
- pulse-frontend/src/App.tsx - Add Dashboard route (`/` or `/dashboard`)
- pulse-frontend/src/types/dashboard.ts - Create dashboard-specific types (if needed)

**Existing Files to Reuse:**
- pulse-frontend/src/api/nodes.ts - fetchNodes(), fetchNodeStatus() (from Story 4.3)
- pulse-frontend/src/api/data.ts - fetchMetrics() (from Story 4.3)
- pulse-frontend/src/api/types.ts - NodeDTO, MetricsDTO (from Story 4.3)
- pulse-frontend/src/stores/nodesStore.ts - Node state management (from Story 4.2)
- pulse-frontend/src/stores/dashboardStore.ts - Dashboard state (from Story 4.2)
- pulse-frontend/src/stores/authStore.ts - Authentication state (from Story 4.2)
- pulse-frontend/src/router/index.tsx - Routing configuration (from Story 4.1)
