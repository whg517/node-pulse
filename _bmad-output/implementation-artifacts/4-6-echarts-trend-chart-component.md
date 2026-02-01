# Story 4.6: ECharts Trend Chart Component

**Epic:** Epic 4 - 实时监控仪表盘
**Status:** ready-for-dev
**Assignee:** TBD
**Priority:** High
**Story Points:** 5
**Estimated Days:** 2

## User Story

As a 前端开发,
I can 封装 ECharts 趋势图组件,
So that 可以在多处复用趋势图功能。

## Acceptance Criteria

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

## Requirements Coverage

- **FR18:** 7 天历史趋势图
- **UX Design:** 卡片展开详情交互模式

## Technical Implementation Notes

### Component Props

```typescript
interface TrendChartProps {
  data: DataPoint[]
  metric: 'latency_ms' | 'packet_loss_rate' | 'jitter_ms'
  timeRange: '24h' | '7d' | '30d'
  showBaseline?: boolean
  height?: string
  className?: string
  onTimeRangeChange?: (range: '24h' | '7d' | '30d') => void
}

interface DataPoint {
  timestamp: string
  value: number
}
```

### ECharts Configuration

- Chart Type: Line chart
- X-Axis: Time (auto-formatted based on time range)
- Y-Axis: Metric value with proper scale
- Tooltip: Hover to show exact values
- Zoom: Mouse wheel zoom support
- Toolbox: Zoom reset, data zoom controls
- Grid: Responsive sizing
- Animation: Smooth transitions

### Time Range Handling

- **24h**: Show data points at 1-minute intervals, format X-axis as HH:mm
- **7d**: Show data points at 5-minute intervals, format X-axis as MM-DD HH:mm
- **30d**: Show data points at 1-hour intervals, format X-axis as MM-DD

### Multi-Metric Support

- Component should be able to display different metrics
- Color coding:
  - Latency: Blue (#3b82f6)
  - Packet Loss Rate: Red (#ef4444)
  - Jitter: Purple (#8b5cf6)
- Y-axis scale automatically adjusts based on metric type and values

### Baseline Reference Line

- 7-day baseline (average of past 7 days)
- Displayed as green dashed line
- Only shown when time range is 7d or 30d
- Calculated from data or provided as prop

### Data Aggregation

- Data may come pre-aggregated from backend
- Support for 1-minute, 5-minute, and 1-hour aggregation
- Component should handle data points efficiently (limit to ~1000 points for performance)

### Responsive Design

- Chart container should be responsive
- Use ResizeObserver to detect container size changes
- Call chart.resize() on container resize
- Support different screen sizes (mobile, tablet, desktop)

### Accessibility

- Provide ARIA labels for chart
- Support keyboard navigation for time range selector
- Ensure color contrast meets WCAG standards
- Provide alternative text for screen readers

### Error Handling

- Display friendly message when no data available
- Show loading state while fetching data
- Handle malformed data gracefully

## Component Structure

Create `/pulse-frontend/src/components/dashboard/TrendChart.tsx`:

```typescript
import { useEffect, useRef } from 'react'
import * as echarts from 'echarts'

interface TrendChartProps {
  data: DataPoint[]
  metric: 'latency_ms' | 'packet_loss_rate' | 'jitter_ms'
  timeRange: '24h' | '7d' | '30d'
  showBaseline?: boolean
  height?: string
  className?: string
  onTimeRangeChange?: (range: '24h' | '7d' | '30d') => void
}

export default function TrendChart({
  data,
  metric,
  timeRange,
  showBaseline = false,
  height = '400px',
  className = '',
  onTimeRangeChange
}: TrendChartProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<echarts.ECharts | null>(null)

  // Implementation...
}
```

## Styling (Tailwind CSS)

- Container: `bg-white rounded-lg shadow-sm p-4`
- Time range selector: Button group with active state styling
- Chart container: Responsive with configurable height
- Loading overlay: Semi-transparent overlay with spinner
- Empty state: Centered message with icon

## Dependencies

- Story 4.1 (Frontend Route Auth Guard) - must be completed
- Story 4.2 (Zustand State Management) - must be completed
- Story 4.3 (API Layer Encapsulation) - must be completed
- Apache ECharts must be installed (from Story 4.1)

## API Integration

This component will receive data through props. The parent component (NodeDetailPage in Story 4.7) will be responsible for:
- Fetching historical data from `/api/v1/data/history`
- Passing aggregated data to TrendChart
- Handling loading and error states

## Definition of Done

- [ ] Story file created and reviewed
- [ ] TrendChart component created with all required props
- [ ] ECharts integration working correctly
- [ ] Time range selector implemented and functional
- [ ] Multi-metric support implemented
- [ ] Tooltip hover display working
- [ ] Zoom functionality working (mouse wheel)
- [ ] Baseline reference line implemented
- [ ] Responsive design working
- [ ] Loading state implemented
- [ ] Empty state implemented
- [ ] Error handling implemented
- [ ] Component styled with Tailwind CSS
- [ ] Unit tests written for TrendChart component
- [ ] Integration tests written for chart interactions
- [ ] Accessibility validated (ARIA labels, keyboard navigation)
- [ ] Performance tested (handles 1000+ data points smoothly)
- [ ] Code reviewed and approved
- [ ] Documentation updated

## Tasks

1. **Component Setup**
   - Create TrendChart component file
   - Define TypeScript interfaces for props
   - Set up ECharts instance with ref

2. **Chart Configuration**
   - Implement basic line chart
   - Configure X-axis (time-based)
   - Configure Y-axis (metric-specific)
   - Add tooltip configuration
   - Add zoom and toolbox configuration

3. **Time Range Selector**
   - Create time range buttons (24h/7d/30d)
   - Implement state management for selected range
   - Format X-axis labels based on time range
   - Call onTimeRangeChange callback

4. **Multi-Metric Support**
   - Implement color coding for different metrics
   - Adjust Y-axis scale based on metric type
   - Update chart title based on metric

5. **Baseline Reference Line**
   - Calculate baseline from data (if not provided)
   - Add green dashed line to chart
   - Show/hide based on showBaseline prop

6. **Data Handling**
   - Transform data points to ECharts format
   - Handle empty data arrays
   - Implement data point limiting for performance

7. **Responsive Design**
   - Use ResizeObserver to detect size changes
   - Call chart.resize() on container resize
   - Test on different screen sizes

8. **Loading and Error States**
   - Implement loading overlay with spinner
   - Show empty state message when no data
   - Handle chart initialization errors

9. **Styling**
   - Apply Tailwind CSS classes
   - Style time range selector
   - Style container and layout
   - Ensure consistent design with other components

10. **Testing**
    - Write unit tests for component rendering
    - Test chart configuration and options
    - Test time range switching
    - Test multi-metric display
    - Test zoom functionality
    - Test responsive behavior
    - Test loading and error states

11. **Documentation**
    - Document component props and usage
    - Add usage examples
    - Document ECharts configuration options
    - Add troubleshooting guide

## Testing Strategy

### Unit Tests

- Component rendering with different props
- ECharts instance creation and cleanup
- Chart configuration with different metrics
- Time range switching
- Baseline line display
- Loading and empty states

### Integration Tests

- Chart interactions (zoom, hover)
- Time range selector functionality
- Responsive behavior
- Data updates and re-renders

### E2E Tests

- Chart displays correctly in NodeDetailPage
- Time range switching triggers data fetch
- User can zoom and pan the chart
- Tooltip shows correct values on hover

## Performance Requirements

- Initial render time: < 500ms
- Chart update time: < 200ms
- Zoom response time: < 100ms
- Handle up to 1000 data points smoothly

## Accessibility Requirements

- ARIA labels for chart container
- Keyboard navigation for time range selector
- Color contrast meets WCAG AA standards
- Screen reader announces chart title and current selection
- Focus indicators visible

## Technical Constraints

- ECharts version: ^5.4.0 (installed in Story 4.1)
- React: ^18.2.0
- TypeScript: ^5.0.0
- Tailwind CSS: ^3.3.0

## Future Enhancements

- Support for multiple metrics on same chart (multi-line)
- Export chart as image
- Custom date range picker
- Advanced statistical overlays (min/max/avg)
- Comparison with other nodes

## Notes

- This component provides the foundation for Story 4.7 (7-day history trend chart)
- Component should be reusable across different pages
- Focus on clean component API for easy integration
- ECharts provides extensive configuration options - expose commonly used ones as props
- Consider creating a custom hook for chart data fetching if needed in Story 4.7

## Related Stories

- Story 4.5: Node Detail Page (will use TrendChart in Story 4.7)
- Story 4.7: 7-Day History Trend Chart (integrates TrendChart into NodeDetailPage)
