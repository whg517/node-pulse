# NodePulse UI/UX Design Document

**Version:** 2.0
**Date:** 2026-02-22
**Author:** Design Team
**Status:** Approved

---

## Design Direction

**Aesthetic:** Technical-Industrial with Data-Centric Precision
**Tech Stack:** React 19 + TypeScript + Tailwind CSS + ECharts + Zustand

---

## 1. Information Architecture

### 1.1 Navigation Structure

Based on PRD user journeys and functional requirements:

```
NodePulse
├── 📊 Dashboard                    /dashboard
│   ├── World Map (节点分布)
│   ├── Core Metrics Panel
│   ├── Alert Stream
│   └── Node Quick List
│
├── 🖥️ Nodes                       /nodes
│   ├── Node List / Management     /nodes
│   ├── Node Detail                /nodes/:id
│   │   ├── Real-time Metrics
│   │   ├── Trend Charts (24h/7d/30d)
│   │   ├── MTR Path Visualization
│   │   └── Diagnostic Report Export
│   └── Node Comparison            /nodes/comparison
│
├── 🚨 Alerts                      /alerts
│   ├── Alert Rules                /alerts/rules
│   ├── Active Alerts / Records    /alerts/records
│   └── Alert History              /alerts/history
│
├── 📈 Reports                     /reports
│   ├── Report Generator           /reports
│   └── Export History             /reports/history
│
├── 🔗 Integrations                /integrations
│   ├── Webhooks                   /integrations/webhooks
│   └── System Health              /integrations/health
│
└── ⚙️ Settings                    /settings
    ├── Preferences                /settings/preferences
    │   ├── Timezone
    │   ├── Language
    │   └── Theme
    ├── Sessions                   /settings/sessions
    └── Users (Admin only)         /settings/users
```

### 1.2 Key User Flows

| Flow | Entry Point | Key Actions | Exit Point |
|------|-------------|-------------|------------|
| **Journey 1: Manager Overview** | Dashboard Map | Click region -> View anomaly -> Export report | PDF Download |
| **Journey 2: Node Deployment** | Node Management | Add Node -> Copy script -> Deploy | Beacon Running |
| **Journey 3: Daily Inspection** | Dashboard | View health map -> Generate report | Email/PDF |
| **Journey 4: Optimization Analysis** | Reports | Select time ranges -> Compare -> Export | PDF with charts |
| **Journey 5: Emergency Response** | Mobile Push | View alert -> MTR path -> Add notes | Alert Resolved |

---

## 2. Layout Architecture

### 2.1 Shared Layout Component

All authenticated pages use a shared `AppLayout` component:

```
+============================================================================+
|                         [Header - 64px]                                    |
|  [Hamburger] NodePulse        [Search...]    [TZ] [EN] [🌙] [User ▼]      |
+============================================================================+
|        |                                                                   |
|   S    |                                                                   |
|   I    |                     Main Content                                  |
|   D    |                                                                   |
|   E    |                   (max-w-7xl, scrollable)                         |
|   B    |                                                                   |
|   A    |                                                                   |
|   R    |                                                                   |
|        |                                                                   |
| 256px |                                                                   |
| (64px |                                                                   |
|collap)|                                                                   |
|        |                                                                   |
+============================================================================+
```

### 2.2 Sidebar Design

**Desktop (≥768px):**
- Width: 256px (expanded) / 64px (collapsed)
- Sections with collapsible groups
- Active state highlighting
- Badge counts for alerts

**Mobile (<768px):**
- Hidden by default
- Overlay when opened (hamburger menu)
- Full-height with backdrop
- Touch-friendly tap targets (48px min)

```typescript
interface SidebarItem {
  icon: React.ReactNode
  label: string
  path: string
  badge?: number        // Alert count, etc.
  children?: SidebarItem[]
  requiredRole?: string // RBAC support
}
```

### 2.3 Header Design

```
+============================================================================+
| [≡] NodePulse    [🔍 Search...]     [UTC+8 ▼] [EN/中文 ▼] [🌙] [Admin ▼]  |
+============================================================================+
```

**Components:**
- **Logo/Brand:** "NodePulse" (consistent)
- **Search:** Global search (future feature)
- **Timezone Selector:** User's preferred timezone
- **Language Switcher:** EN/中文
- **Theme Toggle:** Dark/Light mode
- **User Menu:** Profile, Settings, Logout

---

## 3. Visual Design System

### 3.1 Color Palette - Health States

```css
:root {
  /* Primary Brand */
  --color-primary-500: #2563eb;    /* Blue - Primary actions */
  --color-primary-600: #1d4ed8;    /* Blue - Hover state */

  /* Health State Colors */
  --color-healthy-500: #059669;    /* Emerald - Good health */
  --color-healthy-100: #d1fae5;    /* Light emerald bg */

  --color-warning-500: #d97706;    /* Amber - Warning */
  --color-warning-100: #fef3c7;    /* Light amber bg */

  --color-critical-500: #dc2626;   /* Red - Critical */
  --color-critical-100: #fee2e2;   /* Light red bg */

  --color-offline-400: #9ca3af;    /* Gray - Offline */
  --color-offline-100: #f3f4f6;    /* Light gray bg */

  /* Chart Colors - ECharts compatible */
  --chart-latency: #3b82f6;        /* Blue-500 */
  --chart-packet-loss: #ef4444;    /* Red-500 */
  --chart-jitter: #8b5cf6;         /* Purple-500 */
  --chart-baseline: #10b981;       /* Green-500 */

  /* Semantic - Alert Levels */
  --alert-p0: #dc2626;             /* Red - Critical */
  --alert-p1: #f59e0b;             /* Amber - Warning */
  --alert-p2: #3b82f6;             /* Blue - Notice */

  /* Sidebar */
  --sidebar-bg: #1e293b;           /* Slate-800 */
  --sidebar-hover: #334155;        /* Slate-700 */
  --sidebar-active: #3b82f6;       /* Blue-500 */
}
```

| Status | Color | Hex | Usage |
|--------|-------|-----|-------|
| Healthy | Emerald | `#059669` | Normal nodes |
| Warning | Amber | `#d97706` | Mild anomaly |
| Critical | Red | `#dc2626` | Severe fault |
| Offline | Gray | `#9ca3af` | No response |

### 3.2 Typography

```css
:root {
  /* Primary: JetBrains Mono for data/technical feel */
  --font-display: 'JetBrains Mono', 'SF Mono', 'Fira Code', monospace;

  /* Body: Source Sans 3 for readability */
  --font-body: 'Source Sans 3', 'Inter', system-ui, sans-serif;

  /* Fallback for Chinese content */
  --font-chinese: 'Noto Sans SC', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

/* Typography Scale */
.text-display { font: 600 2rem/1.2 var(--font-display); }
.text-title { font: 600 1.5rem/1.3 var(--font-body); }
.text-body { font: 400 1rem/1.5 var(--font-body); }
.text-mono { font: 400 0.875rem/1.4 var(--font-display); }
.text-label { font: 600 0.75rem/1 var(--font-body); text-transform: uppercase; letter-spacing: 0.05em; }
```

### 3.3 Dark/Light Mode

```css
/* Dark Mode (Default for monitoring systems) */
@media (prefers-color-scheme: dark) {
  :root {
    --bg-primary: #0f172a;      /* Slate-900 */
    --bg-secondary: #1e293b;    /* Slate-800 */
    --bg-tertiary: #334155;     /* Slate-700 */
    --text-primary: #f1f5f9;    /* Slate-100 */
    --text-secondary: #94a3b8;  /* Slate-400 */
    --border-color: #334155;    /* Slate-700 */
  }
}

@media (prefers-color-scheme: light) {
  :root {
    --bg-primary: #ffffff;
    --bg-secondary: #f8fafc;    /* Slate-50 */
    --bg-tertiary: #f1f5f9;     /* Slate-100 */
    --text-primary: #0f172a;
    --text-secondary: #64748b;  /* Slate-500 */
    --border-color: #e2e8f0;    /* Slate-200 */
  }
}
```

### 3.4 Chart Theme (ECharts)

```javascript
const nodePulseTheme = {
  color: ['#3b82f6', '#ef4444', '#8b5cf6', '#10b981', '#f59e0b'],
  backgroundColor: 'transparent',
  textStyle: {
    fontFamily: 'Source Sans 3, sans-serif',
  },
  title: {
    textStyle: { fontWeight: 600, fontSize: 16 },
  },
  line: {
    smooth: 0.3,
    symbol: 'circle',
    symbolSize: 4,
  },
  grid: {
    left: 60, right: 40, top: 60, bottom: 60,
  },
  tooltip: {
    backgroundColor: 'rgba(17, 24, 39, 0.95)',
    borderColor: '#374151',
    textStyle: { color: '#f9fafb' },
  },
}
```

---

## 4. Key UI Patterns

### 4.1 Dashboard Layout

```
+============================================================================+
|  DASHBOARD                           [Refresh] [Export] [24h ▼]            |
+============================================================================+
|                                                                            |
|  +--------------------------------------------------------------------+   |
|  |                    WORLD MAP - ECharts Geo                         |   |
|  |                                                                    |   |
|  |   [Regional clusters with health color coding]                    |   |
|  |   [Pulsing markers for critical alerts]                           |   |
|  |   [Click to navigate to node detail]                              |   |
|  |                                                                    |   |
|  +--------------------------------------------------------------------+   |
|                                                                            |
|  +------------+  +------------+  +------------+  +------------+           |
|  | ONLINE     |  | ANOMALY    |  | 24H ALERTS |  | PROBE      |           |
|  | 94.2%      |  | 5.8%       |  | 23         |  | SUCCESS    |           |
|  | ^ 2.1%     |  | v 1.2%     |  | [sparkline]|  | 98.5%      |           |
|  +------------+  +------------+  +------------+  +------------+           |
|                                                                            |
|  +---------------------------------------+  +-----------------------+     |
|  | LATENCY TREND                         |  | PACKET LOSS TREND     |     |
|  | ------------------------------------- |  | --------------------- |     |
|  | [ECharts line chart with baseline]    |  | [ECharts area chart]  |     |
|  | [24h/7d/30d toggle]                   |  | [Threshold markers]   |     |
|  +---------------------------------------+  +-----------------------+     |
|                                                                            |
|  +---------------------------------------+  +-----------------------+     |
|  | NODE HEALTH CARDS (Top 6)             |  | ALERT STREAM          |     |
|  | ------------------------------------- |  | --------------------- |     |
|  | [Card] [Card]                         |  | [P0] Packet Loss 85%  |     |
|  | [Card] [Card]                         |  | [P1] Latency 320ms    |     |
|  | [Card] [Card]                         |  | [P2] Jitter 45ms      |     |
|  +---------------------------------------+  +-----------------------+     |
|                                                                            |
|  +--------------------------------------------------------------------+   |
|  | NODE LIST TABLE                                                    |   |
|  | ------------------------------------------------------------------ |   |
|  | Name | Status | Region | Latency | Packet Loss | Jitter | Actions |   |
|  | ------------------------------------------------------------------ |   |
|  | ...  | ...    | ...    | ...     | ...         | ...    | ...     |   |
|  +--------------------------------------------------------------------+   |
|                                                                            |
+============================================================================+
```

### 4.2 Health Distribution Map (FR-4.3.1)

**Implementation:**
- ECharts `geo` component with world map
- Custom SVG markers for nodes with health-state colors
- `effectScatter` for pulsing animation on alerts
- Tooltip with node details
- Data zoom for regional drill-down

**Marker Size:**
| Node Count | Marker Size |
|------------|-------------|
| 1-10 | 8px |
| 11-50 | 12px |
| 50+ | 18px |

**Interaction:**
- **Hover:** Tooltip with node count, avg latency, worst node
- **Click:** Navigate to node detail page
- **Long-press (mobile):** Context menu for quick actions

### 4.3 Alert Stream Component

**Display:**
- Latest 10 active alerts
- Auto-scrolling (new alerts insert at top)
- Each alert shows: Level (P0/P1/P2), Node name, Type, Time, Status

**Status Indicators:**
- Unacknowledged: Red background
- In Progress: Yellow background
- Resolved: Gray background

```typescript
interface AlertStreamItem {
  id: string
  level: 'P0' | 'P1' | 'P2'
  nodeName: string
  type: 'latency' | 'packet_loss' | 'jitter'
  value: number
  timestamp: string
  status: 'new' | 'acknowledged' | 'resolved'
}
```

### 4.4 MTR Path Visualization (FR-4.3.5)

**Visual Design:**
```
Source                                              Destination
   |                                                    |
   v                                                    ^
+-------+      +-------+      +-------+      +-------+
| Hop 1 |----->| Hop 2 |----->| Hop 3 |----->| Hop 4 |
|  2ms  |      |  5ms  |      | 12ms  |      | 45ms  |
|  0%   |      |  0%   |      |  0%   |      |  2%   |
+-------+      +-------+      +-------+      +-------+
                                                  |
+===================================================+  <-- Red border (critical)
|  Hop 5: 8.8.8.8                                   |
|  Latency: 85ms | Loss: 15% | AS: 15169           |
|  Location: United States, California             |
+===================================================+
```

**Risk Indicators:**
| Condition | Visual Treatment |
|-----------|------------------|
| Loss >= 10% | Red border, pulsing animation |
| Latency >= 200ms | Red border, warning icon |
| Jitter >= 50ms | Yellow border |
| Timeout | Gray box with question mark |
| Normal | Default blue/gray styling |

### 4.5 Multi-Metric Time Series (FR-4.3.4)

**Features:**
- Metric selector (latency/loss/jitter toggle)
- Overlay multiple metrics on same chart
- Baseline comparison (shaded area, not just line)
- Export chart as PNG (ECharts toolbox)
- Annotation markers for alert events

---

## 5. Wireframes

### 5.1 Global Dashboard (Desktop) - With Sidebar

```
+============================================================================+
|                         HEADER (64px)                                      |
|  [≡] NodePulse Dashboard      [UTC+8 ▼] [EN ▼] [🌙] [Admin ▼]            |
+============================================================================+
|        |                                                                   |
|  S     |  DASHBOARD                                                        |
|  I     |  Real-time Network Overview        [Refresh] [Export PDF]         |
|  D     |                                                                   |
|  E     |  +------------------------------------------------------------+   |
|  B     |  |                                                            |   |
|  A     |  |              [WORLD MAP - ECharts Geo]                     |   |
|  R     |  |                                                            |   |
|        |  |   [APAC: 12 nodes]  [EMEA: 8 nodes]  [AMER: 5 nodes]      |   |
|  ───   |  |   [●] healthy       [●] warning      [●] critical          |   |
|  📊    |  |                                                            |   |
|  Dash  |  +------------------------------------------------------------+   |
|  ───   |                                                                   |
|  🖥️    |  +----------+  +----------+  +----------+  +----------+          |
|  Nodes |  | ONLINE   |  | ANOMALY  |  | ALERTS   |  | LATENCY  |          |
|  ───   |  | 94.2%    |  | 5.8%     |  | 23       |  | 45ms     |          |
|  🚨    |  +----------+  +----------+  +----------+  +----------+          |
|  Alerts|                                                                   |
|  ───   |  +-----------------------------------+  +-------------------+    |
|  📈    |  | LATENCY TREND (24h)               |  | PACKET LOSS       |    |
|  Rep   |  | [Chart with baseline overlay]     |  | [Area chart]      |    |
|  ───   |  +-----------------------------------+  +-------------------+    |
|  🔗    |                                                                   |
|  Integ |  +-----------------------------------+  +-------------------+    |
|  ───   |  | NODE HEALTH CARDS (2x3 grid)     |  | ALERT STREAM      |    |
|  ⚙️    |  | [SG-1] [TK-1] [LD-1]              |  | [P0] SG-3 Loss    |    |
|  Set   |  | [NY-1] [FR-1] [SY-1]              |  | [P1] TK-1 Latency |    |
|        |  +-----------------------------------+  +-------------------+    |
|        |                                                                   |
|        |  +------------------------------------------------------------+   |
|        |  | NODE LIST TABLE                          [View All →]     |   |
|        |  | Name | Status | Region | Latency | Loss | Jitter | Actions|   |
|        |  +------------------------------------------------------------+   |
|        |                                                                   |
+============================================================================+
```

### 5.2 Node Detail Page (Desktop)

```
+============================================================================+
|                         HEADER                                             |
+============================================================================+
|        |                                                                   |
|  S     |  [<] Nodes / Singapore-Primary                                    |
|  I     |                                                                   |
|  D     |  +------------------------------------------------------------+   |
|  E     |  | NODE INFO                                                  |   |
|  B     |  | IP: 203.0.113.45 | Region: APAC-SG | ISP: AWS | Uptime: 45d|   |
|  A     |  | Tags: [prod] [edge] [primary] | Last Heartbeat: 30s ago    |   |
|  R     |  +------------------------------------------------------------+   |
|        |                                                                   |
|  ───   |  +------------------+  +------------------+  +------------------+  |
|  📊    |  | LATENCY          |  | PACKET LOSS      |  | JITTER           |  |
|  Dash  |  |      45 ms       |  |      0.2 %       |  |      12 ms       |  |
|  ───   |  |   [Good]         |  |   [Good]         |  |   [Good]         |  |
|  🖥️    |  +------------------+  +------------------+  +------------------+  |
|  Nodes |                                                                   |
|  ●     |  TREND CHARTS                                     [24h] [7d] [30d] |
|  ───   |  +------------------------------------------------------------+   |
|  🚨    |  |                                                            |   |
|  Alerts|  |  [ECharts - Latency with baseline overlay]                 |   |
|  ───   |  |                                                            |   |
|  📈    |  +------------------------------------------------------------+   |
|  Rep   |                                                                   |
|  ───   |  MTR PATH (8 hops)                                                 |
|  🔗    |  +------------------------------------------------------------+   |
|  Integ |  | [1] 10.0.0.1 -----> [2] 203.0.113.1 -----> [3] ... -----> [8]|   |
|  ───   |  |  2ms, 0%            5ms, 0%             12ms, 0%       92ms  |   |
|  ⚙️    |  |                                                            |   |
|  Set   |  |  [!] Hop 5: 8.8.8.8 - 85ms, 15% loss [CRITICAL]            |   |
|        |  +------------------------------------------------------------+   |
|        |                                                                   |
|        |  DIAGNOSIS                                                 [Expand]|
|        |  +------------------------------------------------------------+   |
|        |  | Root Cause: Cross-border link degradation (SG -> US)       |   |
|        |  | Confidence: High (85%)                                     |   |
|        |  | Recommendation: Contact carrier, consider route optimization|   |
|        |  +------------------------------------------------------------+   |
|        |                                                                   |
+============================================================================+
```

### 5.3 Mobile Layout (<768px)

```
+============================+
|  [≡] NodePulse      [🌙]   |
+============================+
|                            |
|  DASHBOARD                 |
|  [Refresh] [Export]        |
|                            |
|  +------------------------+|
|  |      WORLD MAP         ||
|  |   (simplified view)    ||
|  +------------------------+|
|                            |
|  +--------+  +--------+    |
|  | ONLINE |  |ANOMALY |    |
|  | 94.2%  |  | 5.8%   |    |
|  +--------+  +--------+    |
|                            |
|  ALERT STREAM              |
|  +------------------------+|
|  | [P0] SG-3: Loss 85%   ||
|  | [P1] TK-1: 320ms      ||
|  +------------------------+|
|                            |
|  NODE LIST                 |
|  +------------------------+|
|  | SG-1 | ● | 45ms       ||
|  | TK-1 | ⚠ | 180ms      ||
|  | LD-1 | ● | 32ms       ||
|  +------------------------+|
|                            |
+============================+

Sidebar (overlay when opened):
+============================+
|  NodePulse            [✕]  |
+============================+
|  📊 Dashboard              |
|  🖥️ Nodes                  |
|  🚨 Alerts          [5]    |
|  📈 Reports                |
|  🔗 Integrations           |
|  ─────────────────────     |
|  ⚙️ Settings               |
|  [Logout]                  |
+============================+
```

---

## 6. Component Architecture

### 6.1 Layout Components

```
frontend/src/components/layout/
├── AppLayout.tsx           # Main layout wrapper with sidebar + header
├── Sidebar.tsx             # Navigation sidebar (collapsible)
├── SidebarItem.tsx         # Individual navigation item
├── SidebarGroup.tsx        # Grouped navigation items
├── Header.tsx              # Top header with user actions
├── Breadcrumb.tsx          # Navigation breadcrumbs
├── PageHeader.tsx          # Standardized page header
└── index.ts
```

### 6.2 Dashboard Components

```
frontend/src/components/dashboard/
├── WorldMap.tsx            # ✓ Health distribution map
├── AlertStream.tsx         # NEW: Real-time alert feed
├── MetricsSummaryCards.tsx # ✓ Dashboard summary stats
├── NodeSummaryCard.tsx     # ✓ Individual node card
├── NodeListTable.tsx       # ✓ Node list table
├── TopAnomaliesList.tsx    # ✓ Top anomalies list
├── TrendChart.tsx          # ✓ Time series chart
├── MetricCard.tsx          # ✓ Single metric display
├── HealthStatusBadge.tsx   # ✓ Status badge
├── ProblemDiagnosis.tsx    # ✓ Root cause analysis
└── index.ts
```

### 6.3 Common Components

```
frontend/src/components/common/
├── ThemeToggle.tsx         # ✓ Dark/light mode toggle
├── LanguageSwitcher.tsx    # ✓ EN/中文 switcher
├── TimezoneSelector.tsx    # ✓ Timezone picker
├── ProtectedRoute.tsx      # ✓ Auth route guard
├── LoadingSpinner.tsx      # Loading indicator
├── ErrorBoundary.tsx       # Error handling
├── EmptyState.tsx          # Empty data display
└── ConfirmDialog.tsx       # Confirmation modal
```

---

## 7. Route Structure

### 7.1 App.tsx Routes

```typescript
<Routes>
  {/* Public */}
  <Route path="/login" element={<LoginPage />} />

  {/* Protected - Wrapped in AppLayout */}
  <Route element={<ProtectedRoute><AppLayout /></ProtectedRoute>}>
    {/* Dashboard */}
    <Route path="/dashboard" element={<DashboardPage />} />
    <Route path="/" element={<Navigate to="/dashboard" replace />} />

    {/* Nodes */}
    <Route path="/nodes" element={<NodeManagementPage />} />
    <Route path="/nodes/:id" element={<NodeDetailPage />} />
    <Route path="/nodes/comparison" element={<NodeComparisonPage />} />

    {/* Alerts */}
    <Route path="/alerts" element={<Navigate to="rules" replace />} />
    <Route path="/alerts/rules" element={<AlertRulesPage />} />
    <Route path="/alerts/records" element={<AlertRecordsPage />} />
    <Route path="/alerts/history" element={<AlertHistoryPage />} />

    {/* Reports */}
    <Route path="/reports" element={<ReportsPage />} />
    <Route path="/reports/history" element={<ExportHistoryPage />} />

    {/* Integrations */}
    <Route path="/integrations" element={<Navigate to="webhooks" replace />} />
    <Route path="/integrations/webhooks" element={<WebhooksPage />} />
    <Route path="/integrations/health" element={<SystemHealthPage />} />

    {/* Settings */}
    <Route path="/settings" element={<Navigate to="preferences" replace />} />
    <Route path="/settings/preferences" element={<PreferencesPage />} />
    <Route path="/settings/sessions" element={<SessionsPage />} />
    <Route
      path="/settings/users"
      element={
        <ProtectedRoute requiredRole="admin">
          <UsersPage />
        </ProtectedRoute>
      }
    />
  </Route>

  {/* 404 */}
  <Route path="*" element={<NotFoundPage />} />
</Routes>
```

---

## 8. Accessibility & i18n

### 8.1 WCAG 2.1 AA Compliance

| Requirement | Implementation |
|-------------|----------------|
| **Keyboard Navigation** | Tab order follows logical flow; focus trap in modals |
| **Focus Indicators** | 2px solid outline with 2px offset, using primary color |
| **Color Contrast** | 4.5:1 minimum; health states use icons + color |
| **Screen Reader** | ARIA labels on all charts; live regions for alerts |
| **Chart Alternatives** | Collapsible data tables for all visualizations |
| **Motion** | Respect `prefers-reduced-motion`; disable animations |

### 8.2 Multi-Timezone Display (FR-4.3.14)

**Display Pattern:**
```
+-------------------------------------------------------+
| Timestamp: 2026-02-22 10:30:00 SGT (UTC+8)           |
|           = 2026-02-22 02:30:00 UTC                  |
|           = 2026-02-21 21:30:00 EST (UTC-5)          |
+-------------------------------------------------------+
```

**Settings:**
```typescript
type TimezoneDisplay = 'utc' | 'local' | 'node_local' | 'multi'

interface TimezoneConfig {
  display: TimezoneDisplay
  primaryTimezone: string  // e.g., 'Asia/Singapore'
  showMultiZone: boolean
}
```

### 8.3 Chinese/English Language Switching

**i18n Structure:**
```
frontend/src/
├── locales/
│   ├── en.json
│   └── zh-CN.json
├── i18n.ts
└── ...
```

**Key Translations:**
| English | Chinese |
|---------|---------|
| Dashboard | 仪表盘 |
| Nodes | 节点 |
| Alerts | 告警 |
| Reports | 报告 |
| Settings | 设置 |
| Healthy | 健康 |
| Warning | 预警 |
| Critical | 异常 |
| Offline | 离线 |

---

## 9. Implementation Priority

### Phase 1: Foundation (Week 1)
| Priority | Task | Effort |
|----------|------|--------|
| **P0** | Create AppLayout component | High |
| **P0** | Create Sidebar component | High |
| **P0** | Create Header component | Medium |
| **P0** | Update App.tsx routes | Medium |

### Phase 2: Migration (Week 2)
| Priority | Task | Effort |
|----------|------|--------|
| **P0** | Migrate DashboardPage to AppLayout | Medium |
| **P0** | Integrate WorldMap into Dashboard | Medium |
| **P0** | Migrate all pages to AppLayout | High |
| **P1** | Create AlertStream component | Medium |

### Phase 3: Enhancement (Week 3)
| Priority | Task | Effort |
|----------|------|--------|
| **P1** | Create PreferencesPage | Medium |
| **P1** | Create UsersPage | Medium |
| **P1** | Create SystemHealthPage | Medium |
| **P2** | Mobile responsive optimization | High |

### Phase 4: Polish (Week 4)
| Priority | Task | Effort |
|----------|------|--------|
| **P2** | E2E test updates | High |
| **P2** | Accessibility audit | Medium |
| **P2** | Performance optimization | Medium |
| **P3** | Animation polish | Low |

---

## 10. Summary

This UI/UX design document provides a comprehensive framework for NodePulse's frontend with a focus on:

### Key Design Highlights

1. **Shared Layout Architecture**
   - Single AppLayout component for all authenticated pages
   - Collapsible sidebar navigation for scalability
   - Consistent header with global actions

2. **Dashboard Enhancement**
   - WorldMap integration for geographic node visualization
   - AlertStream for real-time monitoring
   - Improved visual hierarchy with card-based layout

3. **Color System**
   - Health-state colors (emerald/amber/red/gray) with clear semantic meaning
   - WCAG-compliant contrast ratios
   - Dark mode as default (appropriate for monitoring systems)

4. **Key Differentiators**
   - Interactive world map with real-time health visualization
   - MTR path diagram with risk highlighting
   - Multi-timezone collaboration support
   - Bilingual (Chinese/English) interface

5. **Mobile-First for Alerts**
   - Emergency response flow optimized for mobile
   - Overlay sidebar for navigation
   - Touch-friendly interactions

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-02-17 | Design Team | Initial UI/UX design document |
| 2.0 | 2026-02-22 | Design Team | Comprehensive restructure: added shared layout, sidebar navigation, integrated WorldMap, reorganized routes |
