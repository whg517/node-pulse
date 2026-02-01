# Story 4.5: Node Detail Page

**Epic:** Epic 4 - 实时监控仪表盘
**Status:** ready-for-dev
**Assignee:** TBD
**Priority:** High
**Story Points:** 5
**Estimated Days:** 2

## User Story

As a 运维主管,
I can 在节点详情页查看单个节点的详细网络指标,
So that 可以深入分析单个节点的性能状况。

## Acceptance Criteria

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

## Requirements Coverage

- **FR4:** 实时仪表盘（单节点详情页显示时延/丢包率/抖动）
- **FR22:** 问题类型自动判断
- **UX Design:** 卡片展开详情交互模式

## Technical Implementation Notes

### Backend APIs Required
- `GET /api/v1/nodes/{id}` - Get node details
- `GET /api/v1/nodes/{id}/status` - Get node status
- `GET /api/v1/data/metrics?node_id={id}` - Get node metrics

### Frontend Components
- Create `/pulse-frontend/src/pages/NodeDetailPage.tsx`
- Create `/pulse-frontend/src/components/dashboard/MetricCard.tsx` (if not exists)
- Create `/pulse-frontend/src/components/dashboard/ProblemDiagnosis.tsx` (for problem type detection)
- Add route `/nodes/:id` to React Router

### Data Flow
1. Page loads with node_id from URL params
2. Fetch node basic info from `/api/v1/nodes/{id}`
3. Fetch node status from `/api/v1/nodes/{id}/status`
4. Fetch real-time metrics from `/api/v1/data/metrics?node_id={id}`
5. Display metric cards with real-time values
6. Implement expandable card interaction for problem diagnosis
7. Poll metrics every 5 seconds for real-time updates

### State Management (Zustand)
- Use `nodesStore` for node information
- Store current node details in component state or nodesStore

### UI Design
- Use Tailwind CSS for styling
- Metric cards should be visually distinct with color coding
- Green/Yellow/Red status indicators for health
- Expandable section for problem diagnosis details
- Responsive design for mobile and desktop

### Problem Type Detection Display
- Node local fault (节点本地故障)
- Cross-border link issue (跨境链路问题)
- Carrier routing issue (运营商路由问题)
- Show confidence level: High/Medium/Low

## Dependencies

- Story 4.1 (Frontend Route Auth Guard) - must be completed
- Story 4.2 (Zustand State Management) - must be completed
- Story 4.3 (API Layer Encapsulation) - must be completed
- Story 4.4 (Dashboard Homepage Node List) - should be completed

## Definition of Done

- [ ] Story file created and reviewed
- [ ] Backend API endpoints implemented (if not already)
- [ ] NodeDetailPage component created
- [ ] MetricCard component created
- [ ] ProblemDiagnosis component created
- [ ] Route `/nodes/:id` configured in React Router
- [ ] Real-time data polling implemented (5-second interval)
- [ ] Expandable card interaction implemented
- [ ] Problem type detection display implemented
- [ ] Unit tests written for components
- [ ] Integration tests written for page flow
- [ ] UI/UX reviewed against design spec
- [ ] Accessibility validated (ARIA labels, keyboard navigation)
- [ ] Performance tested (page load time < 5 seconds)
- [ ] Code reviewed and approved
- [ ] Documentation updated

## Tasks

1. **Backend Verification**
   - Verify `/api/v1/nodes/{id}` endpoint exists
   - Verify `/api/v1/nodes/{id}/status` endpoint exists
   - Verify `/api/v1/data/metrics` endpoint supports node_id filter

2. **Frontend Setup**
   - Create NodeDetailPage component structure
   - Add route to React Router configuration
   - Create navigation from dashboard to node detail

3. **Component Development**
   - Create MetricCard component for displaying individual metrics
   - Create ProblemDiagnosis component for problem type display
   - Implement expandable card interaction pattern

4. **Data Integration**
   - Integrate API calls using api/nodes.ts and api/data.ts
   - Implement real-time data polling with useDashboardData hook
   - Handle loading and error states

5. **UI Implementation**
   - Design metric card layout with Tailwind CSS
   - Implement color-coded health indicators
   - Add last heartbeat timestamp display
   - Implement expandable details section

6. **Problem Diagnosis Display**
   - Display problem type classification
   - Show confidence level
   - Include diagnostic information

7. **Testing**
   - Write unit tests for MetricCard component
   - Write unit tests for ProblemDiagnosis component
   - Write integration tests for NodeDetailPage
   - Test navigation from dashboard to detail page
   - Test real-time data updates
   - Test expandable card interaction

8. **Documentation**
   - Document component props and usage
   - Update API documentation if needed
   - Add troubleshooting guide

## Testing Strategy

### Unit Tests
- MetricCard component rendering with different metric values
- ProblemDiagnosis component with different problem types
- API call functions with mock data

### Integration Tests
- Complete page flow from navigation to data display
- Real-time data polling behavior
- Expandable card interaction
- Error handling for API failures

### E2E Tests
- Navigate to node detail page from dashboard
- Verify all metrics display correctly
- Verify real-time updates work
- Verify problem diagnosis display

## Performance Requirements

- Page load time: ≤ 5 seconds
- Real-time data refresh: ≤ 5 seconds
- Metric card expansion: < 100ms response time

## Accessibility Requirements

- Semantic HTML elements
- ARIA labels for interactive elements
- Keyboard navigation support
- Color contrast ratios meet WCAG standards
- Screen reader compatibility

## Notes

- This story implements the node detail view that allows drilling down from the dashboard
- Expandable card interaction pattern should be consistent with dashboard
- Problem type detection logic may be a placeholder until Story 7.4 (Problem Diagnosis Engine) is complete
- Consider adding a "Back to Dashboard" button for easy navigation
