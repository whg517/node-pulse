# NodePulse UI/UX Design Document

**Version:** 1.0
**Date:** 2026-02-17
**Author:** Design Team
**Status:** Draft

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
├── Dashboard (Global Overview)
│   ├── Health Distribution Map (FR-4.3.1)
│   ├── Core Metrics Panel (FR-4.3.2)
│   ├── Alert Stream (FR-4.3.3)
│   └── Node Quick List
│
├── Nodes
│   ├── Node List / Management (FR-4.3.7)
│   ├── Node Detail (FR-4.3.4, FR-4.3.5)
│   │   ├── Real-time Metrics
│   │   ├── Trend Charts (24h/7d/30d)
│   │   ├── MTR Path Visualization
│   │   └── Diagnostic Report Export (FR-4.3.6)
│   └── Node Comparison (FR-4.3.12)
│
├── Alerts
│   ├── Alert Rules (FR-4.3.8)
│   ├── Active Alerts / Records (FR-4.3.3)
│   └── Alert History
│
├── Reports
│   ├── Health Report Generator (FR-4.3.11)
│   ├── Performance Comparison (FR-4.3.12)
│   └── Export History
│
├── Integrations
│   ├── Webhooks (FR-4.3.9)
│   └── System Health (FR-4.3.10)
│
└── Settings
    ├── User Preferences (timezone, language - FR-4.3.14)
    └── Session Management
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

## 2. Visual Design System

### 2.1 Color Palette - Health States (FR-4.3.1)

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
}
```

| Status | Color | Hex | Usage |
|--------|-------|-----|-------|
| Healthy | Emerald | `#059669` | Normal nodes |
| Warning | Amber | `#d97706` | Mild anomaly |
| Critical | Red | `#dc2626` | Severe fault |
| Offline | Gray | `#9ca3af` | No response |

### 2.2 Typography

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

### 2.3 Dark/Light Mode

```css
@media (prefers-color-scheme: dark) {
  :root {
    --bg-primary: #0f172a;      /* Slate-900 */
    --bg-secondary: #1e293b;    /* Slate-800 */
    --text-primary: #f1f5f9;    /* Slate-100 */
    --text-secondary: #94a3b8;  /* Slate-400 */
    --border-color: #334155;    /* Slate-700 */
  }
}

@media (prefers-color-scheme: light) {
  :root {
    --bg-primary: #ffffff;
    --bg-secondary: #f8fafc;    /* Slate-50 */
    --text-primary: #0f172a;
    --text-secondary: #64748b;  /* Slate-500 */
    --border-color: #e2e8f0;    /* Slate-200 */
  }
}
```

### 2.4 Chart Theme (ECharts)

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

## 3. Key UI Patterns

### 3.1 Health Distribution Map (FR-4.3.1)

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
- **Click:** Navigate to filtered node list for that region
- **Long-press (mobile):** Context menu for quick actions

### 3.2 MTR Path Visualization (FR-4.3.5)

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

### 3.3 Multi-Metric Time Series (FR-4.3.4)

**Features:**
- Metric selector (latency/loss/jitter toggle)
- Overlay multiple metrics on same chart
- Baseline comparison (shaded area, not just line)
- Export chart as PNG (ECharts toolbox)
- Annotation markers for alert events

### 3.4 Alert Stream (FR-4.3.3)

**Display:**
- Latest 10 active alerts
- Auto-scrolling (new alerts insert at top)
- Each alert shows: Level (P0/P1/P2), Node name, Type, Time, Status

**Status Indicators:**
- Unacknowledged: Red
- In Progress: Yellow
- Resolved: Gray

---

## 4. Wireframes

### 4.1 Global Dashboard (Desktop)

```
+============================================================================+
|  NODEPULSE                    [Dashboard] [Nodes] [Alerts] [Reports] [Set] |
|  Network Monitoring                       [EN/中文] [UTC+8 v] [Refresh 5s] |
+============================================================================+
|                                                                            |
|  +--------------------------------------------------------------------+   |
|  |                                                                    |   |
|  |                    [WORLD MAP - ECharts Geo]                       |   |
|  |                                                                    |   |
|  |   [Regional clusters with health color coding]                    |   |
|  |   [Pulsing markers for alerts]                                    |   |
|  |                                                                    |   |
|  +--------------------------------------------------------------------+   |
|                                                                            |
|  +------------+  +------------+  +------------+  +------------+           |
|  | ONLINE     |  | ANOMALY    |  | 24H ALERTS |  | AVG LATENCY|           |
|  | 94.2%      |  | 5.8%       |  | 23         |  | 45ms       |           |
|  | ^ 2.1%     |  | v 1.2%     |  | [mini]     |  | v 3ms      |           |
|  +------------+  +------------+  +------------+  +------------+           |
|                                                                            |
|  +---------------------------------------+  +-----------------------+     |
|  | NODE LIST (Top 10 by anomaly)         |  | ALERT STREAM          |     |
|  | ------------------------------------- |  | --------------------- |     |
|  | [!] Singapore-3  | 85% loss | 2m ago  |  | [P0] Packet Loss 85%  |     |
|  | [!] Tokyo-1      | 320ms    | 5m ago  |  | [P1] Latency 320ms    |     |
|  | [w] London-Edge  | 45ms jit | 12m ago |  | [P2] Jitter 45ms      |     |
|  | ...                                   |  | ...                   |     |
|  +---------------------------------------+  +-----------------------+     |
|                                                                            |
+============================================================================+
```

### 4.2 Node Detail Page (Desktop)

```
+============================================================================+
|  [<] Node: Singapore-Primary              [ONLINE] [Live] [Export PDF]    |
+============================================================================+
|                                                                            |
|  +--------------------------------------------------------------------+   |
|  | NODE INFO                                                          |   |
|  | IP: 203.0.113.45 | Region: APAC-SG | ISP: AWS | Uptime: 45d       |   |
|  | Tags: [prod] [edge] [primary] | Last Heartbeat: 30s ago            |   |
|  +--------------------------------------------------------------------+   |
|                                                                            |
|  +------------------+  +------------------+  +------------------+          |
|  | LATENCY          |  | PACKET LOSS      |  | JITTER           |          |
|  |      45 ms       |  |      0.2 %       |  |      12 ms       |          |
|  |   [Good]         |  |   [Good]         |  |   [Good]         |          |
|  +------------------+  +------------------+  +------------------+          |
|                                                                            |
|  TREND CHARTS                                     [24h] [7d] [30d]         |
|  +--------------------------------------------------------------------+   |
|  |                                                                    |   |
|  |  [ECharts - Latency with baseline overlay]                         |   |
|  |                                                                    |   |
|  +--------------------------------------------------------------------+   |
|                                                                            |
|  MTR PATH (8 hops)                                                         |
|  +--------------------------------------------------------------------+   |
|  | [1] 10.0.0.1 -----> [2] 203.0.113.1 -----> [3] ... -----> [8]      |   |
|  |  2ms, 0%            5ms, 0%             12ms, 0%       92ms, 0%    |   |
|  |                                                                    |   |
|  |  [!] Hop 5: 8.8.8.8 - 85ms, 15% loss [CRITICAL]                    |   |
|  +--------------------------------------------------------------------+   |
|                                                                            |
|  DIAGNOSIS                                                         [Expand]|
|  +--------------------------------------------------------------------+   |
|  | Root Cause: Cross-border link degradation (Singapore -> US)        |   |
|  | Confidence: High (85%)                                             |   |
|  | Recommendation: Contact carrier, consider route optimization       |   |
|  +--------------------------------------------------------------------+   |
|                                                                            |
+============================================================================+
```

### 4.3 Alert Management Page

```
+============================================================================+
|  ALERTS                         [Rules] [Records] [History]               |
+============================================================================+
|                                                                            |
|  Filter: [All Levels v] [All Nodes v] [All Status v]    [Search...]       |
|                                                                            |
|  +--------------------------------------------------------------------+   |
|  | LEVEL | NODE           | TYPE        | VALUE    | TIME    | STATUS |   |
|  |-------|----------------|-------------|----------|---------|--------|   |
|  | [P0]  | Singapore-3    | Packet Loss | 85%      | 2m ago  | New    |   |
|  | [P0]  | Tokyo-Primary  | Latency     | 500ms    | 5m ago  | Progress|  |
|  | [P1]  | London-Edge    | Jitter      | 45ms     | 12m ago | Done   |   |
|  | [P1]  | Sydney-Backup  | Latency     | 280ms    | 1h ago  | Done   |   |
|  | [P2]  | Frankfurt-1    | Packet Loss | 3%       | 2h ago  | Done   |   |
|  +--------------------------------------------------------------------+   |
|                                                                            |
|  Showing 5 of 23 active alerts                           [< 1 2 3 ... >] |
|                                                                            |
+============================================================================+
```

### 4.4 Mobile Emergency View (FR-4.3.13)

```
+============================+
|  NODEPULSE          [Menu] |
+============================+
|                            |
|  ALERT: Singapore-3        |
|  Packet Loss: 85%          |
|  Started: 2 min ago        |
|                            |
|  [View Details]            |
|                            |
+----------------------------+
|                            |
|  MTR PATH                  |
|  +------------------------+|
|  | 1. Gateway    2ms     ||
|  | 2. ISP        5ms     ||
|  | 3. Regional   12ms    ||
|  | ...                    ||
|  | 6. 8.8.8.8   85ms     ||
|  |    [CRITICAL - 15%]   ||
|  +------------------------+|
|                            |
|  Add Note:                 |
|  +------------------------+|
|  | Contacted carrier...  ||
|  +------------------------+|
|                            |
|  [Mark In Progress]        |
|  [Resolve]                 |
|                            |
+============================+
```

---

## 5. Accessibility & i18n

### 5.1 WCAG 2.1 AA Compliance

| Requirement | Implementation |
|-------------|----------------|
| **Keyboard Navigation** | Tab order follows logical flow; focus trap in modals |
| **Focus Indicators** | 2px solid outline with 2px offset, using primary color |
| **Color Contrast** | 4.5:1 minimum; health states use icons + color |
| **Screen Reader** | ARIA labels on all charts; live regions for alerts |
| **Chart Alternatives** | Collapsible data tables for all visualizations |
| **Motion** | Respect `prefers-reduced-motion`; disable animations |

### 5.2 Multi-Timezone Display (FR-4.3.14)

**Display Pattern:**
```
+-------------------------------------------------------+
| Timestamp: 2026-02-17 10:30:00 SGT (UTC+8)           |
|           = 2026-02-17 02:30:00 UTC                  |
|           = 2026-02-16 21:30:00 EST (UTC-5)          |
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

### 5.3 Chinese/English Language Switching

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
| Healthy | 健康 |
| Warning | 预警 |
| Critical | 异常 |
| Offline | 离线 |

---

## 6. Component Architecture

### 6.1 New Components Required

```
frontend/src/components/
├── dashboard/
│   ├── WorldMap.tsx           # Health distribution map
│   ├── AlertStream.tsx        # Real-time alert feed
│   ├── TopNodesList.tsx       # Top anomalies/delays
│   └── MetricsPanel.tsx       # Dashboard summary
├── nodes/
│   ├── MTRVisualization.tsx   # Route path display
│   ├── NodeStatusHeader.tsx   # Status + live indicator
│   └── DiagnosticSummary.tsx  # Root cause analysis
├── alerts/
│   ├── AlertRuleEditor.tsx    # Policy configuration
│   └── AlertTimeline.tsx      # Event history
├── reports/
│   ├── ReportGenerator.tsx    # PDF export UI
│   └── ComparisonView.tsx     # Before/after analysis
└── common/
    ├── LanguageSwitcher.tsx
    ├── TimezoneSelector.tsx
    └── AccessibilityToggle.tsx
```

### 6.2 Components to Enhance

| Component | Enhancement |
|-----------|-------------|
| `TrendChart.tsx` | Add multi-metric overlay, improved baseline visualization |
| `MetricCard.tsx` | Add sparkline mini-charts, improved accessibility |
| `HealthStatusBadge.tsx` | Add animation for state changes |
| `DashboardPage.tsx` | Add world map, reorganize layout |

---

## 7. Implementation Priority

| Priority | Component | FR Reference | Effort |
|----------|-----------|--------------|--------|
| **P0** | Health Distribution Map | FR-4.3.1 | High |
| **P0** | Alert Stream Component | FR-4.3.3 | Medium |
| **P0** | Enhanced Metric Cards | FR-4.3.2 | Low |
| **P1** | MTR Path Visualization | FR-4.3.5 | High |
| **P1** | Node Detail Page Enhancements | FR-4.3.4, FR-4.3.6 | Medium |
| **P1** | Alert Rules Editor | FR-4.3.8 | Medium |
| **P2** | PDF Report Export | FR-4.3.6, FR-4.3.11 | High |
| **P2** | Mobile Responsive Layout | NFR-5.4.1 | Medium |
| **P2** | Dark Mode | - | Medium |
| **P3** | Performance Comparison | FR-4.3.12 | Medium |
| **P3** | Timezone/Language Switcher | FR-4.3.14 | Low |

---

## 8. Summary

This UI/UX design document provides a comprehensive framework for implementing NodePulse's frontend based on the PRD requirements.

### Key Design Highlights

1. **Aesthetic Direction:** Technical-industrial with JetBrains Mono for data display and Source Sans 3 for body text

2. **Color System:** Health-state colors (emerald/amber/red/gray) with clear semantic meaning and WCAG-compliant contrast

3. **Component Architecture:** Build on existing React/TypeScript/Tailwind/ECharts foundation with enhanced accessibility

4. **Key Differentiators:**
   - Interactive world map with real-time health visualization
   - MTR path diagram with risk highlighting
   - Multi-timezone collaboration support
   - Bilingual (Chinese/English) interface

5. **Mobile-First for Alerts:** Emergency response flow optimized for mobile with essential MTR and action capabilities

---

## Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-02-17 | Design Team | Initial UI/UX design document |
