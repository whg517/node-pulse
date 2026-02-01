# Story 7.1: Comparison Chart Component

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a 前端开发,
I want 封装多节点对比 ECharts 组件,
so that 可以复用对比图表.

## Acceptance Criteria

### Given
- Apache ECharts 已安装

### When
- 创建 ComparisonChart 组件

### Then
- 组件接收多个节点数据数组作为 props
- 组件支持最多 5 个节点的对比
- 按地区/运营商标签分组对比
- 使用相同时间范围和指标类型
- 显示平均值、最大值、最小值、差异
- 差异用颜色或图标明确标注
- 组件使用 Tailwind CSS 样式

## Requirements Coverage

**FR Coverage:**
- FR19（多节点对比）

**Architecture Alignment:**
- 前端使用 React + TypeScript + Vite + Tailwind CSS + Apache ECharts
- 组件命名使用 PascalCase
- TypeScript 类型定义

**NFR Compliance:**
- NFR-PERF-002: 组件加载性能优化
- 组件复用性

## Tasks / Subtasks

- [x] 创建 ComparisonChart 组件基础结构 (AC: #Given, #When)
  - [x] 创建文件: pulse-frontend/src/components/dashboard/ComparisonChart.tsx
  - [x] 定义 TypeScript 接口和类型
  - [x] 导出 ComparisonChart 组件

- [x] 实现 Props 接口定义 (AC: #Then - 组件接收多个节点数据)
  - [x] 定义 ComparisonDataPoint 接口
  - [x] 定义 NodeComparisonData 接口
  - [x] 定义 ComparisonChartProps 接口
  - [x] 添加数据验证逻辑（最多 5 个节点）

- [x] 实现多节点对比图表配置 (AC: #Then - 支持最多 5 个节点对比)
  - [x] 配置 ECharts 多系列图表
  - [x] 为每个节点分配独立颜色
  - [x] 实现图例显示节点名称
  - [x] 添加节点数量验证（2-5 个节点）

- [x] 实现地区/运营商分组功能 (AC: #Then - 按地区/运营商标签分组对比)
  - [x] 添加分组显示选项
  - [x] 按地区标签分组节点数据
  - [x] 按运营商标签分组节点数据
  - [x] 在图表中标注分组信息

- [x] 实现统计数据显示 (AC: #Then - 显示平均值、最大值、最小值、差异)
  - [x] 计算每个时间点的平均值
  - [x] 计算每个时间点的最大值
  - [x] 计算每个时间点的最小值
  - [x] 计算节点间的差异
  - [x] 在图表中添加统计信息面板

- [x] 实现差异可视化标注 (AC: #Then - 差异用颜色或图标明确标注)
  - [x] 使用颜色差异标注异常节点
  - [x] 添加图标标注显著差异
  - [x] 实现差异阈值配置
  - [x] 添加差异提示信息

- [x] 应用 Tailwind CSS 样式 (AC: #Then - 使用 Tailwind CSS 样式)
  - [x] 参考 TrendChart 组件样式
  - [x] 实现响应式布局
  - [x] 添加加载状态样式
  - [x] 添加空数据状态样式

- [x] 添加交互功能
  - [x] 实现数据点悬停提示
  - [x] 添加图例切换功能
  - [x] 实现缩放功能
  - [x] 添加时间范围选择器

- [x] 编写组件测试
  - [x] 创建测试文件: pulse-frontend/src/components/dashboard/__tests__/ComparisonChart.test.tsx
  - [x] 编写组件渲染测试
  - [x] 编写 Props 验证测试
  - [x] 编写统计计算测试
  - [x] 编写差异可视化测试

- [x] 编写组件文档
  - [x] 添加 JSDoc 注释
  - [x] 编写使用示例
  - [x] 添加 Props 类型说明

## Dev Notes

### Component Requirements

**File Location:**
- `pulse-frontend/src/components/dashboard/ComparisonChart.tsx`
- Tests: `pulse-frontend/src/components/dashboard/__tests__/ComparisonChart.test.tsx`

**Component Interface (Based on TrendChart Pattern):**

```typescript
export interface ComparisonDataPoint {
  timestamp: string
  value: number
}

export interface NodeComparisonData {
  node_id: string
  node_name: string
  region?: string
  isp?: string
  data: ComparisonDataPoint[]
}

export interface ComparisonChartProps {
  nodes: NodeComparisonData[]  // 2-5 nodes
  metric: 'latency_ms' | 'packet_loss_rate' | 'jitter_ms'
  timeRange: '24h' | '7d' | '30d'
  showStatistics?: boolean  // Show avg/max/min
  highlightDifferences?: boolean  // Highlight significant differences
  groupBy?: 'region' | 'isp' | 'none'
  height?: string
  className?: string
  onTimeRangeChange?: (range: string) => void
  isLoading?: boolean
}
```

**ECharts Configuration:**
- Multi-series line chart (one series per node)
- Each node has unique color (maintain consistency across app)
- Legend showing all node names with region/ISP tags
- Tooltip showing all node values at hovered timestamp
- Statistics panel showing avg/max/min/diff
- Data zoom for time range selection
- Highlight differences using color intensity or icons

**Statistics Calculation:**
- For each timestamp, calculate across all nodes:
  - Average value
  - Maximum value and which node
  - Minimum value and which node
  - Difference (max - min) and percentage difference
- Display as overlay or separate panel

**Difference Visualization:**
- Color coding: Green (normal), Yellow (warning), Red (critical)
- Icon indicators: ▲ for significantly higher, ▼ for significantly lower
- Thresholds:
  - Latency: >20% difference from average = warning, >50% = critical
  - Packet Loss: >5% difference = warning, >10% = critical
  - Jitter: >30% difference = warning, >60% = critical

**Grouping:**
- Region grouping: Group nodes by region tag
- ISP grouping: Group nodes by ISP tag
- Display group labels in chart title or legend

### Project Structure Notes

**Frontend Structure:**
```
pulse-frontend/
├── src/
│   ├── components/
│   │   └── dashboard/
│   │       ├── TrendChart.tsx           # Reference for patterns
│   │       ├── ComparisonChart.tsx      # NEW - This story
│   │       └── __tests__/
│   │           ├── TrendChart.test.tsx  # Reference for test patterns
│   │           └── ComparisonChart.test.tsx  # NEW
│   ├── api/
│   │   └── data.ts                      # Comparison API (Story 7.2)
│   ├── stores/
│   │   └── dashboardStore.ts            # Zustand state management
│   └── types/
│       └── dashboard.ts                 # Shared types
```

**Alignment with Existing Patterns:**
- Component naming: PascalCase (ComparisonChart.tsx)
- Type definitions: Export interfaces for reusability
- Props pattern: Follow TrendChart component structure
- Styling: Tailwind CSS utility classes
- ECharts initialization: Use useRef and useEffect pattern
- Responsive: Handle window resize with useEffect
- Loading states: Show spinner overlay
- Empty states: Show SVG icon + message
- Accessibility: ARIA labels, roles, keyboard navigation

**Detected Conflicts/Variances:**
- None - This is a new component extending the dashboard component library
- Should complement TrendChart, not replace it

### Testing Standards

**Unit Tests (Jest + React Testing Library):**
```typescript
// ComparisonChart.test.tsx
describe('ComparisonChart', () => {
  it('renders chart with multiple nodes', () => {})
  it('validates max 5 nodes constraint', () => {})
  it('validates min 2 nodes requirement', () => {})
  it('calculates statistics correctly', () => {})
  it('highlights differences when enabled', () => {})
  it('groups nodes by region', () => {})
  it('groups nodes by ISP', () => {})
  it('shows loading state', () => {})
  it('shows empty state when no data', () => {})
  it('handles time range change', () => {})
})
```

**Test Data Patterns:**
- Mock 2-5 nodes with overlapping time ranges
- Include edge cases (2 nodes, 5 nodes, invalid 6 nodes)
- Test with missing region/ISP tags
- Test with different metrics (latency, packet loss, jitter)

**Coverage Requirements:**
- Component rendering: 100%
- Props validation: 100%
- Statistics calculations: 100%
- User interactions: 100%
- Edge cases: 100%

### Code Patterns from TrendChart

**Reusable Patterns:**
1. **ECharts Initialization:**
   ```typescript
   const chartRef = useRef<HTMLDivElement>(null)
   const chartInstance = useRef<echarts.ECharts | null>(null)

   useEffect(() => {
     if (!chartRef.current) return
     chartInstance.current = echarts.init(chartRef.current)
     return () => {
       chartInstance.current?.dispose()
     }
   }, [])
   ```

2. **Metric Configuration:**
   ```typescript
   const metricConfig = {
     latency_ms: { label: 'Latency', unit: 'ms', color: '#3b82f6' },
     packet_loss_rate: { label: 'Packet Loss Rate', unit: '%', color: '#ef4444' },
     jitter_ms: { label: 'Jitter', unit: 'ms', color: '#8b5cf6' },
   }
   ```

3. **Color Palette for Multiple Nodes:**
   ```typescript
   const nodeColors = [
     '#3b82f6', // blue-500
     '#10b981', // green-500
     '#f59e0b', // amber-500
     '#ef4444', // red-500
     '#8b5cf6', // purple-500
   ]
   ```

4. **Tooltip Formatter:**
   ```typescript
   tooltip: {
     trigger: 'axis',
     formatter: (params: any) => {
       // Show all node values at hovered timestamp
       // Include statistics (avg/max/min/diff)
     }
   }
   ```

5. **Responsive Resize Handler:**
   ```typescript
   useEffect(() => {
     const handleResize = () => {
       chartInstance.current?.resize()
     }
     window.addEventListener('resize', handleResize)
     return () => window.removeEventListener('resize', handleResize)
   }, [])
   ```

6. **Loading and Empty States:**
   - Loading: Spinner overlay with "Loading chart data..."
   - Empty: SVG icon with "No Data Available" message

7. **Accessibility:**
   - `role="region"` for chart container
   - `aria-label` describing chart content
   - `aria-pressed` for time range buttons
   - `role="status"` for loading spinner

### References

**Source Documents:**
- [Source: _bmad-output/planning-artifacts/epics.md#Epic 7] - Epic 7 requirements
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.1] - Story 7.1 acceptance criteria
- [Source: _bmad-output/planning-artifacts/architecture.md] - Frontend tech stack (React + TypeScript + Vite + Tailwind CSS + ECharts)
- [Source: pulse-frontend/src/components/dashboard/TrendChart.tsx] - Reference component for patterns

**Related Stories:**
- Story 4.6: ECharts 趋势图组件 - Basic ECharts component pattern
- Story 7.2: 节点对比查询 API - Backend API for comparison data
- Story 7.3: 节点对比前端页面 - Page that will use this component

**Technical Dependencies:**
- Apache ECharts (already installed)
- React + TypeScript (already configured)
- Tailwind CSS (already configured)

## Dev Agent Record

### Agent Model Used

Claude Sonnet 4.5 (claude-sonnet-4-5-20250929)

### Debug Log References

No issues encountered during implementation.

### Completion Notes List

✅ **Story 7.1: Comparison Chart Component - COMPLETED (Code Review Fixes Applied)**

**Implementation Summary:**
- Created ComparisonChart component with full multi-node comparison support
- Implemented ECharts-based visualization with 2-5 node support
- Added statistics panel showing avg/max/min/diff calculations with node identification
- Implemented region/ISP grouping functionality
- Added difference visualization with color-coded thresholds and icon markers (▲/▼)
- Applied Tailwind CSS styling following TrendChart patterns
- Created comprehensive test suite with 33 tests (all passing)
- Added full JSDoc documentation with usage examples

**Technical Decisions:**
1. **Component Structure**: Followed TrendChart patterns for consistency
2. **Node Validation**: Validates 2-5 node requirement in useEffect rather than early return to maintain UI shell
3. **Statistics Calculation**: Pre-calculated using useMemo for performance optimization (fixes MEDIUM-8)
4. **Difference Thresholds**: Metric-specific thresholds (latency: 20%/50%, packet loss: 5%/10%, jitter: 30%/60%)
5. **Color Palette**: Used 5 distinct colors from Tailwind palette for node differentiation
6. **Grouping**: Implemented in legend display with region/ISP tags
7. **Accessibility**: Full ARIA labels, roles, and keyboard navigation support
8. **Difference Highlighting**: Implemented markPoints with ▲/▼ icons when `highlightDifferences=true` (fixes CRITICAL-1,2,7)
9. **Tooltip Enhancement**: Shows which node has max/min values (fixes CRITICAL-3)
10. **Empty State UI**: Legend and statistics hidden when no nodes (fixes CRITICAL-4)
11. **Error Handling**: Added try/catch for ECharts dispose (fixes MEDIUM-11)
12. **Resize Handler**: Debounced with 150ms delay (fixes MEDIUM-10)

**Code Review Fixes Applied:**
- ✅ CRITICAL-1: Implemented `highlightDifferences` functionality with markPoints
- ✅ CRITICAL-2: Added ▲/▼ icon indicators for outlier data points
- ✅ CRITICAL-3: Enhanced tooltip to show which node has max/min
- ✅ CRITICAL-4: Fixed empty state UI to hide legend/stats
- ✅ CRITICAL-5: highlightDifferences now properly used in dependencies
- ✅ CRITICAL-6: Story file updated to reflect true implementation status
- ✅ CRITICAL-7: Added markPoint configuration for outlier highlighting
- ✅ MEDIUM-8: Used useMemo for statistics pre-calculation (performance)
- ✅ MEDIUM-9: Fixed test act() warning with proper act() wrapping
- ✅ MEDIUM-10: Added debounce to resize handler
- ✅ MEDIUM-11: Added error handling for ECharts dispose
- ✅ MEDIUM-12: Grouping improved with better visual distinction

**Test Coverage:**
- 33 tests covering: rendering, validation, statistics, grouping, metrics, interactions, accessibility, edge cases, and difference highlighting
- All tests passing (100% pass rate)
- Tests follow Vitest patterns with vi.fn() mocks
- New tests added for highlightDifferences feature

**Files Modified:**
- pulse-frontend/src/components/dashboard/ComparisonChart.tsx (UPDATED - fixes applied)
- pulse-frontend/src/components/dashboard/__tests__/ComparisonChart.test.tsx (UPDATED - tests added)
- pulse-frontend/src/components/dashboard/index.ts (UPDATED - exports added)

### File List

pulse-frontend/src/components/dashboard/ComparisonChart.tsx
pulse-frontend/src/components/dashboard/__tests__/ComparisonChart.test.tsx
pulse-frontend/src/components/dashboard/index.ts

