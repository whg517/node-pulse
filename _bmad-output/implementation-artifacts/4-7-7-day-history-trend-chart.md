# Story 4.7: 7-Day History Trend Chart

**Epic:** Epic 4 - 实时监控仪表盘
**Status:** ready-for-dev
**Assignee:** TBD
**Priority:** High
**Story Points:** 5
**Estimated Days:** 2

## User Story

As a 运维主管,
I can 在节点详情页查看 7 天历史趋势图,
So that 可以分析节点的长期性能趋势。

## Acceptance Criteria

**Given** 用户已登录并访问节点详情页
**When** 趋势图数据加载完成
**Then** 显示最近 24 小时、7 天、30 天时间范围选择
**And** 显示时延、丢包率、抖动指标曲线
**And** 实时数据（< 1 小时）从内存缓存加载，历史数据（1 小时 - 7 天）从 PostgreSQL `metrics` 表查询
**And** 数据按 1 分钟或 5 分钟聚合
**And** 包含 7 天基线参考线（绿色虚线）
**And** 支持鼠标悬停显示具体时间点的数值

## Requirements Coverage

- **FR18:** 7 天历史趋势图
- **Architecture:** 数据分层查询策略

## Technical Implementation Notes

### Backend API Endpoints Required

- `GET /api/v1/data/history` - Get historical metrics for a node
  - Query params: `node_id`, `start_time`, `end_time`, `metric`, `aggregation`
  - Response: Array of time-series data points
  - Real-time data (< 1 hour): from memory cache
  - Historical data (1 hour - 7 days): from PostgreSQL metrics table

### Frontend Components to Update

- **Update `/pulse-frontend/src/pages/NodeDetailPage.tsx`**
  - Import TrendChart component
  - Add trend chart section below metric cards
  - Fetch historical data based on selected time range
  - Display three trend charts (one for each metric: latency, packet loss, jitter)
  - Calculate and display 7-day baseline

- **Create `/pulse-frontend/src/api/data.ts` (update if needed)**
  - Add `fetchHistoryData()` function
  - Support time range and metric parameters
  - Handle data aggregation

### Data Flow

1. User opens NodeDetailPage for a specific node
2. Page loads node details and current metrics (existing functionality)
3. Page fetches historical data for default time range (24h)
4. TrendChart components display data for each metric
5. User can switch time ranges (24h/7d/30d)
6. On time range change, refetch data and update charts
7. Charts show baseline reference line (7-day average)

### Time Range Implementation

- **24h**: Fetch data from `now - 24h` to `now`, 1-minute aggregation
- **7d**: Fetch data from `now - 7d` to `now`, 5-minute aggregation
- **30d**: Fetch data from `now - 30d` to `now`, 1-hour aggregation

### Data Layer Strategy

**Real-time Data (< 1 hour):**
- Source: Memory cache (Pulse in-memory cache)
- Aggregation: 1-minute
- Query latency target: < 50ms

**Historical Data (1 hour - 7 days):**
- Source: PostgreSQL `metrics` table
- Aggregation: 1-minute or 5-minute (based on time range)
- Query latency target: < 500ms

### Baseline Calculation

- Calculate 7-day rolling average from historical data
- Display as green dashed line on charts
- Only show when time range is 7d or 30d
- Label with baseline value and unit

### API Request Format

```
GET /api/v1/data/history?node_id={id}&start_time={iso}&end_time={iso}&metric={metric}&aggregation={agg}
```

**Response:**
```json
{
  "data": [
    {
      "timestamp": "2024-01-01T00:00:00Z",
      "value": 45.2
    },
    ...
  ],
  "baseline": 48.5,
  "aggregation": "1m"
}
```

### UI Layout

In NodeDetailPage, add trend charts section below the existing metrics cards:

```
[Node Information Card]
[Metric Cards - 3 cards]
[Trend Charts Section - NEW]
  ├── Time Range Selector (shared across all charts)
  ├── Latency Trend Chart
  ├── Packet Loss Rate Trend Chart
  └── Jitter Trend Chart
[Problem Diagnosis Card]
```

### Component Structure

```typescript
// In NodeDetailPage.tsx
const [timeRange, setTimeRange] = useState<TimeRange>('24h')
const [historyData, setHistoryData] = useState<HistoryData>({
  latency_ms: [],
  packet_loss_rate: [],
  jitter_ms: []
})
const [isLoadingHistory, setIsLoadingHistory] = useState(false)

// Fetch history data
useEffect(() => {
  const fetchHistory = async () => {
    setIsLoadingHistory(true)
    const data = await fetchHistoryData(nodeId, timeRange)
    setHistoryData(data)
    setIsLoadingHistory(false)
  }
  fetchHistory()
}, [nodeId, timeRange])
```

### State Management

- Time range state managed at page level
- Historical data fetched and stored in component state
- Loading state for history data
- Error handling for API failures

### Error Handling

- Show friendly error message if history data fetch fails
- Provide retry button
- Show empty state if no historical data available
- Handle partial data (some metrics available, others not)

### Performance Considerations

- Fetch all three metrics in parallel (Promise.all)
- Implement data caching to avoid redundant fetches
- Use proper aggregation to limit data points
- Limit to ~1000 data points per chart for performance
- Implement lazy loading if needed

## Dependencies

- Story 4.5 (Node Detail Page) - must be completed
- Story 4.6 (ECharts Trend Chart Component) - must be completed
- Backend API endpoint `/api/v1/data/history` must be implemented

## Definition of Done

- [ ] Story file created and reviewed
- [ ] Backend API endpoint implemented (or verified existing)
- [ ] NodeDetailPage updated with trend charts section
- [ ] Three trend charts displayed (latency, packet loss, jitter)
- [ ] Time range selector implemented and functional
- [ ] Historical data fetching implemented
- [ ] Data layer strategy working (cache vs PostgreSQL)
- [ ] Baseline reference line calculated and displayed
- [ ] Loading states implemented
- [ ] Error handling implemented
- [ ] Unit tests written for new functionality
- [ ] Integration tests written for history data fetching
- [ ] UI/UX reviewed against design spec
- [ ] Performance tested (data loading time < 1s)
- [ ] Code reviewed and approved
- [ ] Documentation updated

## Tasks

1. **Backend Verification**
   - Verify `/api/v1/data/history` endpoint exists
   - Test endpoint with different time ranges
   - Verify data aggregation works correctly
   - Verify baseline calculation

2. **API Layer Update**
   - Add `fetchHistoryData()` function to `/pulse-frontend/src/api/data.ts`
   - Handle time range parameters
   - Handle metric type parameters
   - Type the response properly

3. **NodeDetailPage Update**
   - Import TrendChart component
   - Add state for time range and history data
   - Implement history data fetching
   - Add trend charts section to UI
   - Implement time range change handler
   - Calculate and display baseline

4. **Data Management**
   - Fetch historical data on mount and time range change
   - Fetch all three metrics in parallel
   - Implement data caching if needed
   - Handle loading and error states

5. **UI Implementation**
   - Add time range selector above charts
   - Render three TrendChart components
   - Style with Tailwind CSS
   - Ensure responsive design
   - Add loading overlay
   - Add error message display

6. **Baseline Calculation**
   - Calculate 7-day rolling average
   - Pass baseline to TrendChart
   - Display baseline in legend

7. **Testing**
   - Write unit tests for history data fetching
   - Write integration tests for trend chart updates
   - Test time range switching
   - Test loading and error states
   - Test baseline calculation

8. **Documentation**
   - Document API usage
   - Update component documentation
   - Add usage examples

## Testing Strategy

### Unit Tests

- `fetchHistoryData()` function with mock data
- Time range state management
- Baseline calculation logic

### Integration Tests

- Complete history data flow from API to chart
- Time range switching triggers refetch
- Error handling for API failures
- Loading states during fetch

### E2E Tests

- Navigate to node detail page
- Verify trend charts load with 24h data
- Switch to 7d time range and verify data updates
- Verify baseline appears on 7d and 30d ranges
- Verify hover tooltips show correct values

## Performance Requirements

- Initial history data load: < 1 second
- Time range switch: < 500ms
- Chart rendering: < 200ms
- API response time: < 500ms for PostgreSQL queries

## Accessibility Requirements

- Time range selector keyboard accessible
- Charts properly labeled with ARIA
- Loading states announced to screen readers
- Error messages accessible

## Notes

- This story integrates the TrendChart component from Story 4.6
- Backend may need to be updated to support history API if not already implemented
- Data aggregation strategy is critical for performance
- Consider implementing data pagination if datasets are large
- Baseline should be calculated on backend for consistency

## Related Stories

- Story 4.5: Node Detail Page (integration target)
- Story 4.6: ECharts Trend Chart Component (dependency)
- Story 3.3: Probe Config API Timeseries Table (metrics table creation)
