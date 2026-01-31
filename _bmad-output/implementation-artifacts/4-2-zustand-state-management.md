# Story 4.2: Zustand State Management Setup

Status: done

## Story

As a 前端开发人员,
I need 初始化 Zustand 状态管理并创建所有必要的 stores,
So that 前端应用可以高效管理全局状态（认证、节点、告警、仪表盘）。

## Acceptance Criteria

**Given** Pulse 前端项目已初始化（Story 1.1）并使用 React + TypeScript + Vite
**When** 开发人员安装 Zustand 并初始化状态管理
**Then** 创建 4 个核心 stores（authStore, nodesStore, alertsStore, dashboardStore）
**And** 所有 stores 支持 TypeScript 类型定义
**And** stores 可以被任意组件直接使用（无需 Provider 包装）
**And** stores 遵循 Zustand 最佳实践和架构模式
**And** 所有 stores 包含必要的 actions 和 selectors
**And** 创建示例组件验证 store 功能

## Tasks / Subtasks

- [x] Task 1: Install Zustand and setup project structure (AC: Given)
  - [x] Install Zustand: `npm install zustand`
  - [x] Verify Zustand version is latest stable
  - [x] Create `src/stores/` directory for store files
- [x] Task 2: Create authStore with authentication state (AC: Then - authStore)
  - [x] Define TypeScript interfaces for AuthState
  - [x] Implement user authentication state (user, isAuthenticated, role)
  - [x] Add login/logout actions
  - [x] Add session management actions
- [x] Task 3: Create nodesStore with node management state (AC: Then - nodesStore)
  - [x] Define TypeScript interfaces for NodesState
  - [x] Implement nodes list state
  - [x] Implement node detail state
  - [x] Add node status actions (online/offline)
  - [x] Add CRUD actions for nodes
- [x] Task 4: Create alertsStore with alert management state (AC: Then - alertsStore)
  - [x] Define TypeScript interfaces for AlertsState
  - [x] Implement alert rules state
  - [x] Implement alert records state
  - [x] Add alert CRUD actions
  - [x] Add alert filtering actions
- [x] Task 5: Create dashboardStore with dashboard settings (AC: Then - dashboardStore)
  - [x] Define TypeScript interfaces for DashboardState
  - [x] Implement filter settings (region, status)
  - [x] Implement time range settings
  - [x] Implement refresh settings
  - [x] Add TOP5 abnormal nodes state
- [x] Task 6: Add TypeScript type definitions for all stores (AC: And - TS types)
  - [x] Create shared types file for common interfaces
  - [x] Export all store types for component use
- [x] Task 7: Create example component to verify stores (AC: And - 示例组件)
  - [x] Create test component using all stores
  - [x] Verify state updates work correctly
  - [x] Verify TypeScript types are correct

## Dev Notes

### Architecture Alignment

**State Management Decision** [Source: Architecture.md#Frontend Architecture > 状态管理]:
- Use Zustand for global state management
- Modular store organization (按功能分 store)
- No Provider wrapper needed - any component can directly use stores
- Built-in TypeScript support
- Optional DevTools integration

**Store Organization** [Source: Architecture.md#Frontend Architecture > 状态管理]:
- `authStore`：用户认证状态、角色信息
- `nodesStore`：节点列表、节点详情、在线/离线状态
- `alertsStore`：告警规则、告警记录
- `dashboardStore`：仪表盘筛选、时间范围、刷新设置、异常 TOP5 列表

**State Management Patterns** [Source: Architecture.md#Implementation Patterns & Consistency Rules > State Management Patterns]:
- 状态更新：使用不可变更新（Zustand 内置支持）
- 动作命名：`setField`, `addField`, `removeItem` 格式
- 状态组织：按功能分 store，每个 store 保持聚焦（单一职责）

### Implementation Context

**Dependencies from Previous Stories:**
- Story 1.1: Frontend project initialized with Vite + React + TypeScript
- Story 4.1: Frontend route auth guard implemented (references authStore but used placeholder)
  - **Critical**: This story (4.2) implements the actual authStore that 4.1 referenced
  - Story 4.1 created `src/hooks/useAuth.ts` which will integrate with authStore
  - After implementing this story, Story 4.1's useAuth hook should be updated to use authStore

**Zustand Benefits:**
- 极简主义 API - minimal boilerplate
- 无需 Provider - direct import and use
- 内置 TypeScript support - type safety out of the box
- 小包体积 - ~1KB minizipped
- Easy testing - no wrapper components needed

### Project Structure Notes

**Alignment with unified project structure** [Source: Architecture.md#Project Structure & Boundaries]:
```
pulse-frontend/
├── src/
│   ├── stores/
│   │   ├── authStore.ts        # NEW - User authentication state
│   │   ├── nodesStore.ts       # NEW - Node management state
│   │   ├── alertsStore.ts      # NEW - Alert management state
│   │   ├── dashboardStore.ts   # NEW - Dashboard settings state
│   │   └── types.ts            # NEW - Shared TypeScript interfaces
│   ├── components/
│   │   └── common/
│   │       └── StoreTest.tsx   # NEW - Example component to verify stores
│   ├── hooks/
│   │   └── useAuth.ts          # EXISTS from Story 4.1 - will integrate with authStore
│   └── App.tsx
├── package.json
└── vite.config.ts
```

**Detected conflicts or variances:**
- None identified. This is foundational state management work.
- **Note**: Story 4.1 referenced authStore but didn't implement it - this story provides that implementation.

### Technical Requirements

**State Update Patterns** [Source: Architecture.md#State Management Patterns]:
- Use immutable updates (Zustand built-in support with immer)
- Action naming: `setField`, `addField`, `removeItem`
- Example actions:
  - `setNodeId(nodeId: string)`
  - `setAlerts(alerts: Alert[])`
  - `addNode(node: Node)`
  - `removeAlert(alertId: string)`

**TypeScript Type Definitions** [Source: Architecture.md#Implementation Patterns > Naming Patterns]:
- Component naming: PascalCase (MetricCard.tsx, NodeDetail.tsx)
- Function naming: camelCase (getUserData, createNode, validateAlert)
- Variable naming: camelCase (userId, nodeId, probeConfig)
- Interface naming: PascalCase with descriptive names (AuthState, NodesState, AlertsState)

**API Integration Context** [Source: Architecture.md#API & Communication Patterns]:
- API response format: `{data: ..., message: "...", timestamp: "..."}` or `{code: "ERR_XXX", message: "...", details: {...}}`
- API endpoints that stores will interact with:
  - `POST /api/v1/auth/login` - authStore login action
  - `POST /api/v1/auth/logout` - authStore logout action
  - `GET /api/v1/nodes` - nodesStore fetch nodes action
  - `GET /api/v1/alerts/rules` - alertsStore fetch alert rules
  - `GET /api/v1/alerts/records` - alertsStore fetch alert records

**Performance Requirements** [Source: Architecture.md#Frontend Architecture > 性能优化]:
- Dashboard real-time updates using `useDashboardData` Hook (每 5 秒轮询)
- State changes should trigger minimal re-renders
- Zustand automatically handles this with selector-based subscriptions

### Store Implementation Guidelines

**authStore Structure** (参考 Architecture.md#Frontend Architecture > 状态管理):
```typescript
interface AuthState {
  // State
  user: User | null;
  isAuthenticated: boolean;
  role: 'admin' | 'operator' | 'viewer' | null;
  sessionId: string | null;

  // Actions
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  setUser: (user: User) => void;
  clearAuth: () => void;
}

interface User {
  id: string;
  username: string;
  role: 'admin' | 'operator' | 'viewer';
}
```

**nodesStore Structure**:
```typescript
interface NodesState {
  // State
  nodes: Node[];
  selectedNode: Node | null;
  nodeStatuses: Record<string, 'online' | 'offline' | 'connecting'>;

  // Actions
  setNodes: (nodes: Node[]) => void;
  setSelectedNode: (node: Node | null) => void;
  addNode: (node: Node) => void;
  updateNode: (id: string, updates: Partial<Node>) => void;
  removeNode: (id: string) => void;
  setNodeStatus: (nodeId: string, status: 'online' | 'offline' | 'connecting') => void;
  fetchNodes: () => Promise<void>;
}

interface Node {
  id: string;
  name: string;
  ip: string;
  region: string;
  tags: string[];
  status: 'online' | 'offline' | 'connecting';
}
```

**alertsStore Structure**:
```typescript
interface AlertsState {
  // State
  alertRules: AlertRule[];
  alertRecords: AlertRecord[];
  filter: AlertFilter;

  // Actions
  setAlertRules: (rules: AlertRule[]) => void;
  setAlertRecords: (records: AlertRecord[]) => void;
  addAlertRule: (rule: AlertRule) => void;
  updateAlertRule: (id: string, updates: Partial<AlertRule>) => void;
  removeAlertRule: (id: string) => void;
  setFilter: (filter: AlertFilter) => void;
  fetchAlertRules: () => Promise<void>;
  fetchAlertRecords: () => Promise<void>;
}

interface AlertRule {
  id: string;
  metric: 'latency' | 'packet_loss_rate' | 'jitter';
  threshold: number;
  level: 'P0' | 'P1' | 'P2';
  nodeId: string | null;
  enabled: boolean;
}

interface AlertRecord {
  id: string;
  nodeId: string;
  metric: string;
  level: string;
  status: 'pending' | 'processing' | 'resolved';
  timestamp: string;
}
```

**dashboardStore Structure**:
```typescript
interface DashboardState {
  // State
  filters: DashboardFilter;
  timeRange: '24h' | '7d' | '30d';
  refreshInterval: number; // seconds
  autoRefresh: boolean;
  top5AbnormalNodes: Node[];

  // Actions
  setFilters: (filters: DashboardFilter) => void;
  setTimeRange: (range: '24h' | '7d' | '30d') => void;
  setRefreshInterval: (interval: number) => void;
  toggleAutoRefresh: () => void;
  setTop5AbnormalNodes: (nodes: Node[]) => void;
}

interface DashboardFilter {
  region: string | null;
  status: 'online' | 'offline' | 'all';
  searchQuery: string;
}
```

### Previous Story Intelligence

**From Story 4.1 (Frontend Route Auth Guard)**:
- Created `src/hooks/useAuth.ts` which references authStore
- useAuth hook currently uses placeholder authentication logic
- After implementing authStore in this story, update useAuth to integrate with authStore
- Story 4.1 completion notes indicate React Router v7.13.0 is installed
- ProtectedRoute component in `src/components/common/ProtectedRoute.tsx` will use authStore

**Git History Analysis** (last 5 commits):
- Most recent: Story 4.1 implemented route auth guard
- Pattern: Stories implement backend functionality before frontend
- All stories include comprehensive testing
- Code follows TypeScript and naming convention best practices

**Files to be Aware Of**:
- `pulse-frontend/src/hooks/useAuth.ts` - Will integrate with new authStore
- `pulse-frontend/src/components/common/ProtectedRoute.tsx` - Uses authStore
- `pulse-frontend/src/pages/LoginPage.tsx` - Will use authStore for login/logout

### Testing Requirements

**Unit Tests** (using Vitest + React Testing Library):
- Test authStore login/logout actions
- Test authStore state updates
- Test nodesStore CRUD operations
- Test alertsStore filtering
- Test dashboardStore settings updates
- Test TypeScript type safety for all stores

**Test File Location** [Source: Architecture.md#Implementation Patterns]:
- Place test files in `src/stores/__tests__/` or parallel to store files
- Use naming pattern: `authStore.test.ts`, `nodesStore.test.ts`

**Integration Tests**:
- Test store integration with React components
- Test state persistence across component updates
- Test async actions (API calls in stores)

### References

- [Source: Architecture.md#Frontend Architecture > 状态管理] - Zustand store design and organization
- [Source: Architecture.md#Implementation Patterns & Consistency Rules > State Management Patterns] - Action naming and update patterns
- [Source: Architecture.md#Project Structure & Boundaries] - Frontend project structure
- [Source: Architecture.md#API & Communication Patterns] - API response formats and endpoints
- [Source: PRD.md#Non-Functional Requirements] - Performance requirements (NFR-PERF-002)
- [Source: Epics.md] - Story 4.2 requirements and acceptance criteria
- [Source: Story 4.1 Implementation] - Previous story context and authStore placeholder

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-5-20250929

### Debug Log References

### Completion Notes List

**Implementation Summary:**
- ✅ All 4 core stores implemented (authStore, nodesStore, alertsStore, dashboardStore)
- ✅ TypeScript types defined for all stores with proper interfaces
- ✅ All stores follow Zustand best practices and architecture patterns
- ✅ Stores integrate with backend API endpoints
- ✅ Updated existing components to use new authStore interface
- ✅ Created comprehensive test suite for all stores (37 tests total)
- ✅ All 78 project tests passing
- ✅ Example component created to verify store functionality

**Key Technical Decisions:**
- Used Zustand v5.0.10 (already installed)
- Implemented separate interface for State and Actions for better type safety
- Used async actions for API calls (login, logout, fetchNodes, etc.)
- Integrated with existing API layer from Story 4.1
- Updated useAuth hook to integrate with new authStore
- Updated all existing test files to use new store interfaces

**Test Coverage:**
- authStore: 9 tests (login, logout, session validation, state management)
- nodesStore: 9 tests (CRUD operations, fetch API, status management)
- alertsStore: 11 tests (CRUD operations, fetch APIs, filtering)
- dashboardStore: 8 tests (filters, time range, refresh settings)
- All store tests cover state updates, actions, and edge cases

**Integration with Existing Code:**
- Updated useAuth.ts hook to work with new authStore structure
- Updated ProtectedRoute component and tests
- Updated DashboardPage component and tests
- Updated LoginPage component to use new store actions
- All existing tests updated and passing

**Code Review Fixes Applied (2026-01-31):**
- Created `config/constants.ts` for magic numbers (SESSION_EXPIRY_HOURS, API_BASE_URL)
- Updated authStore to import constants instead of defining inline
- Fixed StoreTest.tsx to use `AlertRecord` type instead of `any`
- Created unified API layer (`api/nodes.ts`, `api/alerts.ts`) for consistent API calls
- Updated nodesStore and alertsStore to use API layer instead of raw fetch
- Updated store tests to mock API layer properly
- Fixed story File List to accurately document all changes

### File List

**New Files Created:**
- `pulse-frontend/src/config/constants.ts` - Application constants (session expiry, API base URL, dashboard defaults)
- `pulse-frontend/src/api/nodes.ts` - Nodes API endpoints with typed DTOs
- `pulse-frontend/src/api/alerts.ts` - Alerts API endpoints with typed DTOs
- `pulse-frontend/src/stores/types.ts` - Shared TypeScript interfaces
- `pulse-frontend/src/stores/nodesStore.ts` - Node management state
- `pulse-frontend/src/stores/alertsStore.ts` - Alert management state
- `pulse-frontend/src/stores/dashboardStore.ts` - Dashboard settings state
- `pulse-frontend/src/stores/index.ts` - Centralized exports
- `pulse-frontend/src/components/common/StoreTest.tsx` - Example component for verification
- `pulse-frontend/src/stores/__tests__/` - Test directory for all store tests
  - `authStore.test.ts` - Auth store tests (9 tests)
  - `nodesStore.test.ts` - Nodes store tests (9 tests)
  - `alertsStore.test.ts` - Alerts store tests (11 tests)
  - `dashboardStore.test.ts` - Dashboard store tests (8 tests)

**Modified Files:**
- `pulse-frontend/src/stores/authStore.ts` - Enhanced with login/logout API integration, now imports constants
- `pulse-frontend/src/hooks/useAuth.ts` - Updated to use new authStore interface
- `pulse-frontend/src/hooks/useAuth.test.ts` - Updated for new authStore structure
- `pulse-frontend/src/pages/LoginPage.tsx` - Updated to use new authStore actions
- `pulse-frontend/src/pages/DashboardPage.tsx` - Updated to use new authStore
- `pulse-frontend/src/pages/DashboardPage.test.tsx` - Updated for new authStore
- `pulse-frontend/src/components/common/ProtectedRoute.test.tsx` - Updated for new authStore
- `pulse-frontend/src/components/common/ProtectedRoute.integration.test.tsx` - Updated for new authStore
- `pulse-frontend/src/components/common/StoreTest.tsx` - Fixed TypeScript type (removed `any`)
- `pulse-frontend/src/stores/nodesStore.ts` - Now uses API layer instead of raw fetch
- `pulse-frontend/src/stores/alertsStore.ts` - Now uses API layer instead of raw fetch
- `pulse-frontend/src/stores/__tests__/nodesStore.test.ts` - Updated to mock API layer
- `pulse-frontend/src/stores/__tests__/alertsStore.test.ts` - Updated to mock API layer
