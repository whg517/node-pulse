# Story 5.3: Alert Rule Frontend Page

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a 运维主管,
I can 在前端页面配置告警规则,
So that 可以可视化地管理告警规则。

## Acceptance Criteria

**Given** 用户已登录并访问告警规则页面
**When** 页面加载完成
**Then** 显示所有现有告警规则列表
**And** 提供"创建规则"按钮
**And** 支持编辑和删除规则
**And** 表单包含：指标类型选择、阈值输入、告警级别选择、节点选择、启用/禁用开关

**覆盖需求:** FR5（告警规则配置）

**创建表:** 无（后端已完成 Story 5.1）

## Tasks / Subtasks

- [x] Task 1: Create AlertRulesPage component (AC: Then - 显示告警规则列表)
  - [x] Subtask 1.1: Create page component file structure
  - [x] Subtask 1.2: Add routing configuration for /alerts/rules route
  - [x] Subtask 1.3: Implement page layout with header and container
  - [x] Subtask 1.4: Add loading state handling
  - [x] Subtask 1.5: Add error state handling with retry option

- [x] Task 2: Implement alert rules list table (AC: Then - 显示所有现有告警规则列表)
  - [x] Subtask 2.1: Create AlertRulesTable component
  - [x] Subtask 2.2: Display columns: ID, Metric, Threshold, Level, Node, Enabled, Status, Actions
  - [x] Subtask 2.3: Add color-coded status badges (enabled/disabled)
  - [x] Subtask 2.4: Add color-coded level badges (P0=red, P1=orange, P2=yellow)
  - [x] Subtask 2.5: Implement empty state when no rules exist
  - [x] Subtask 2.6: Add responsive table design

- [x] Task 3: Implement CreateAlertRule modal/form (AC: Then - 提供"创建规则"按钮)
  - [x] Subtask 3.1: Create AlertRuleForm component (reusable for create/edit)
  - [x] Subtask 3.2: Add "Create Rule" button in page header
  - [x] Subtask 3.3: Implement modal/dialog for form
  - [x] Subtask 3.4: Add metric type dropdown (latency, packet_loss_rate, jitter)
  - [x] Subtask 3.5: Add threshold input with validation (numeric > 0)
  - [x] Subtask 3.6: Add alert level dropdown (P0, P1, P2)
  - [x] Subtask 3.7: Add node selection dropdown (with option for global rules)
  - [x] Subtask 3.8: Add enabled/disabled toggle switch
  - [x] Subtask 3.9: Implement form validation
  - [x] Subtask 3.10: Add submit and cancel buttons

- [x] Task 4: Implement Edit functionality (AC: Then - 支持编辑规则)
  - [x] Subtask 4.1: Add edit button to each table row
  - [x] Subtask 4.2: Pre-populate form with existing rule data
  - [x] Subtask 4.3: Implement update API call
  - [x] Subtask 4.4: Show success/error toast notifications
  - [x] Subtask 4.5: Refresh rules list after successful update

- [x] Task 5: Implement Delete functionality (AC: Then - 支持删除规则)
  - [x] Subtask 5.1: Add delete button to each table row
  - [x] Subtask 5.2: Implement confirmation dialog
  - [x] Subtask 5.3: Implement delete API call
  - [x] Subtask 5.4: Show success/error toast notifications
  - [x] Subtask 5.5: Refresh rules list after successful deletion

- [x] Task 6: Add real-time status display (AC: Then - 实时状态显示)
  - [x] Subtask 6.1: Display rule enabled status with visual indicator
  - [x] Subtask 6.2: Show last updated timestamp
  - [x] Subtask 6.3: Add toggle switch for quick enable/disable
  - [x] Subtask 6.4: Implement optimistic UI updates

- [x] Task 7: Enhance UX with filtering and search (Optional - for better UX)
  - [x] Subtask 7.1: Add search input for filtering rules
  - [x] Subtask 7.2: Add metric type filter
  - [x] Subtask 7.3: Add alert level filter
  - [x] Subtask 7.4: Add enabled/disabled filter

- [x] Task 8: Write comprehensive tests (AC: 完整功能验证)
  - [x] Subtask 8.1: Unit tests for AlertRulesPage component
  - [x] Subtask 8.2: Unit tests for AlertRulesTable component
  - [x] Subtask 8.3: Unit tests for AlertRuleForm component
  - [x] Subtask 8.4: Integration tests for form validation
  - [x] Subtask 8.5: Integration tests for CRUD operations
  - [x] Subtask 8.6: Test error handling scenarios

- [x] Task 9: Update API integration (AC: API 集成)
  - [x] Subtask 9.1: Integrate with existing alerts API (already implemented)
  - [x] Subtask 9.2: Use alertsStore for state management
  - [x] Subtask 9.3: Handle API errors gracefully
  - [x] Subtask 9.4: Implement proper loading states

## Dev Notes

### Epic Analysis

**Epic 5: 告警规则配置与通知** - 系统可以自动检测异常并通过 Webhook 推送告警

**Story Context in Epic:**
- Story 5.1: 告警规则 API (已完成 - 后端 API)
- Story 5.2: Webhook 配置 API (已完成 - 后端 API)
- Story 5.3: **告警规则前端页面** (本故事) - **前端 UI 实现**
- Story 5.4: Webhook 配置前端页面 (下一个故事)
- Story 5.5-5.8: 后续告警功能

**Critical Prerequisites:**
- **Story 5.1 已完成**: Alert Rule API 后端已实现
- **Epic 4 已完成**: 前端基础设施已完成（React Router, Zustand, API layer, Toast components）
- **API Layer**: `pulse-frontend/src/api/alerts.ts` 已实现
- **Store**: `pulse-frontend/src/stores/alertsStore.ts` 已实现
- **Types**: `pulse-frontend/src/api/types.ts` 已定义 AlertRuleDTO

### Architecture Alignment

**Frontend Architecture** [Source: Epic 4 Implementation]:
```
pulse-frontend/src/
├── pages/
│   └── AlertRulesPage.tsx         # NEW - Main page component
├── components/
│   └── alerts/
│       ├── index.ts               # NEW - Export barrel
│       ├── AlertRulesTable.tsx    # NEW - Rules list table
│       ├── AlertRuleForm.tsx      # NEW - Create/edit form
│       └── AlertRuleDialog.tsx    # NEW - Modal wrapper
├── stores/
│   └── alertsStore.ts             # EXISTING - Already implemented
├── api/
│   └── alerts.ts                  # EXISTING - Already implemented
└── routes/
    └── App.tsx                    # UPDATE - Add /alerts/rules route
```

**API Integration** [Source: Story 5.1 & Epic 4]:
- Backend API: `/api/v1/alerts/rules` (GET, POST, PUT, DELETE)
- Frontend API functions: `fetchAlertRules()`, `createAlertRule()`, `updateAlertRule()`, `deleteAlertRule()`
- Zustand store: `useAlertsStore` with state and actions

**Component Patterns** [Source: Epic 4 (DashboardPage, NodeDetailPage)]:
- Page structure: Header with actions, main content area, error/loading states
- Table pattern: Responsive tables with action buttons, status badges
- Form pattern: Modal dialogs with validation, submit/cancel buttons
- State management: Zustand stores with optimistic updates
- Error handling: Toast notifications via ToastContext

**RBAC Integration** [Source: Epic 1 & Story 5.1]:
- Admin/Operator: Can create, edit, delete alert rules
- Viewer: Read-only access (hide create/edit/delete buttons)
- Use `useAuthStore` to check user role

### Previous Story Intelligence

**From Epic 4 (Frontend Infrastructure)** [Source: Stories 4.1-4.9]:
- **React Router v6**: Route setup with protected routes
- **Zustand**: State management with stores pattern
- **API Layer**: Centralized API client with error handling
- **Toast Notifications**: Via ToastContext for success/error messages
- **Responsive Design**: Tailwind CSS utility classes
- **Form Validation**: Client-side validation with error messages
- **Loading States**: Spinners and skeleton screens
- **Error States**: Error messages with retry buttons

**From Story 4.4 (Dashboard Homepage)**:
- **Table Pattern**: NodeListTable component structure
- **Status Badges**: Color-coded badges for status display
- **Action Buttons**: Edit/delete buttons in table rows
- **Empty States**: Friendly messages when no data exists

**From Story 4.5 (Node Detail Page)**:
- **Modal/Dialog Pattern**: For forms and confirmations
- **Form Validation**: Real-time validation with error messages
- **Success/Error Feedback**: Toast notifications after actions

**Key Learnings from Previous Stories**:
- Use existing ToastContext for notifications
- Follow consistent table design patterns
- Implement proper loading and error states
- Use responsive design with Tailwind CSS
- Leverage existing stores and API functions
- Test components with Vitest and React Testing Library

### Technical Requirements

**Page Component Structure** (AlertRulesPage.tsx):
```tsx
import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAlertsStore } from '../stores/alertsStore'
import { useAuthStore } from '../stores/authStore'
import { useToast } from '../contexts/ToastContext'
import { AlertRulesTable } from '../components/alerts/AlertRulesTable'
import { AlertRuleDialog } from '../components/alerts/AlertRuleDialog'

export default function AlertRulesPage() {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const { alertRules, fetchAlertRules, isLoading, error } = useAlertsStore()
  const { showSuccess, showError } = useToast()

  useEffect(() => {
    fetchAlertRules()
  }, [])

  const handleCreate = () => {
    // Open create dialog
  }

  const handleEdit = (id: string) => {
    // Open edit dialog with rule data
  }

  const handleDelete = (id: string) => {
    // Show confirmation and delete
  }

  const handleToggleEnabled = (id: string, enabled: boolean) => {
    // Quick toggle enable/disable
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header with create button */}
      {/* AlertRulesTable component */}
      {/* AlertRuleDialog for create/edit */}
      {/* Loading and error states */}
    </div>
  )
}
```

**AlertRulesTable Component**:
```tsx
interface AlertRulesTableProps {
  rules: AlertRule[]
  onEdit: (id: string) => void
  onDelete: (id: string) => void
  onToggleEnabled: (id: string, enabled: boolean) => void
  canEdit: boolean  // Based on user role
}

export function AlertRulesTable({ rules, onEdit, onDelete, onToggleEnabled, canEdit }: AlertRulesTableProps) {
  if (rules.length === 0) {
    return <EmptyState />
  }

  return (
    <table className="min-w-full divide-y divide-gray-200">
      <thead>
        <tr>
          <th>Metric</th>
          <th>Threshold</th>
          <th>Level</th>
          <th>Node</th>
          <th>Status</th>
          {canEdit && <th>Actions</th>}
        </tr>
      </thead>
      <tbody>
        {rules.map((rule) => (
          <tr key={rule.id}>
            <td>{rule.metric}</td>
            <td>{rule.threshold}</td>
            <td><LevelBadge level={rule.level} /></td>
            <td>{rule.nodeId || 'Global'}</td>
            <td><EnabledBadge enabled={rule.enabled} /></td>
            {canEdit && (
              <td>
                <EditButton onClick={() => onEdit(rule.id)} />
                <DeleteButton onClick={() => onDelete(rule.id)} />
              </td>
            )}
          </tr>
        ))}
      </tbody>
    </table>
  )
}
```

**AlertRuleForm Component**:
```tsx
interface AlertRuleFormProps {
  mode: 'create' | 'edit'
  initialData?: AlertRule
  onSubmit: (data: CreateAlertRuleRequest | UpdateAlertRuleRequest) => Promise<void>
  onCancel: () => void
}

export function AlertRuleForm({ mode, initialData, onSubmit, onCancel }: AlertRuleFormProps) {
  const [metric, setMetric] = useState<string>(initialData?.metric || 'latency')
  const [threshold, setThreshold] = useState<number>(initialData?.threshold || 0)
  const [level, setLevel] = useState<string>(initialData?.level || 'P1')
  const [nodeId, setNodeId] = useState<string | null>(initialData?.nodeId || null)
  const [enabled, setEnabled] = useState<boolean>(initialData?.enabled ?? true)

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  const validate = () => {
    const newErrors: Record<string, string> = {}

    if (!threshold || threshold <= 0) {
      newErrors.threshold = 'Threshold must be greater than 0'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) return

    setIsSubmitting(true)
    try {
      await onSubmit({
        metric,
        threshold,
        level,
        node_id: nodeId,
        enabled,
      })
    } catch (error) {
      console.error('Failed to submit form:', error)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      {/* Metric Type Select */}
      <select value={metric} onChange={(e) => setMetric(e.target.value)}>
        <option value="latency">Latency</option>
        <option value="packet_loss_rate">Packet Loss Rate</option>
        <option value="jitter">Jitter</option>
      </select>

      {/* Threshold Input */}
      <input
        type="number"
        value={threshold}
        onChange={(e) => setThreshold(Number(e.target.value))}
        min="0"
        step="0.01"
      />
      {errors.threshold && <span className="error">{errors.threshold}</span>}

      {/* Alert Level Select */}
      <select value={level} onChange={(e) => setLevel(e.target.value)}>
        <option value="P0">P0 - Critical</option>
        <option value="P1">P1 - Warning</option>
        <option value="P2">P2 - Info</option>
      </select>

      {/* Node Selection */}
      <select value={nodeId || ''} onChange={(e) => setNodeId(e.target.value || null)}>
        <option value="">Global Rule</option>
        {/* Load nodes from API */}
      </select>

      {/* Enabled Toggle */}
      <label>
        <input
          type="checkbox"
          checked={enabled}
          onChange={(e) => setEnabled(e.target.checked)}
        />
        Enabled
      </label>

      {/* Submit and Cancel Buttons */}
      <button type="submit" disabled={isSubmitting}>
        {mode === 'create' ? 'Create Rule' : 'Update Rule'}
      </button>
      <button type="button" onClick={onCancel}>
        Cancel
      </button>
    </form>
  )
}
```

**Routing Configuration** (App.tsx):
```tsx
import { ProtectedRoute } from './components/auth/ProtectedRoute'
import AlertRulesPage from './pages/AlertRulesPage'

// Add to routes
<Route
  path="/alerts/rules"
  element={
    <ProtectedRoute>
      <AlertRulesPage />
    </ProtectedRoute>
  }
/>
```

### Testing Requirements

**Component Tests** (AlertRulesPage.test.tsx):
- Test page renders correctly
- Test loading state displays
- Test error state displays with retry button
- Test empty state displays when no rules
- Test create button opens dialog
- Test rules are fetched on mount

**Table Component Tests** (AlertRulesTable.test.tsx):
- Test table renders with rules
- Test empty state renders
- Test status badges display correctly
- Test level badges display correctly
- Test edit/delete buttons show for admin/operator
- Test edit/delete buttons hide for viewer
- Test edit callback fires
- Test delete callback fires

**Form Component Tests** (AlertRuleForm.test.tsx):
- Test form renders with initial data
- Test form validation works
- Test submit creates rule in create mode
- Test submit updates rule in edit mode
- Test cancel callback fires
- Test validation errors display
- Test submit button disabled while submitting

**Integration Tests**:
- Test create rule flow
- Test edit rule flow
- Test delete rule flow with confirmation
- Test API error handling
- Test toast notifications display
- Test list refreshes after CRUD operations

### Implementation Guidelines

**UX Best Practices:**
- Use clear, descriptive labels for form fields
- Provide inline validation feedback
- Show loading states during API calls
- Display success/error toast notifications
- Confirm destructive actions (delete)
- Use color coding for visual hierarchy
- Implement responsive design for mobile

**Form Validation:**
- Client-side validation before API calls
- Clear error messages
- Disable submit button while submitting
- Highlight invalid fields
- Validate threshold > 0
- Validate required fields

**State Management:**
- Use existing alertsStore
- Implement optimistic updates for better UX
- Refresh list after mutations
- Handle API errors gracefully

**Access Control:**
- Check user role from useAuthStore
- Hide create/edit/delete buttons for viewers
- Show read-only message for viewers
- Disable form fields in view mode

**Error Handling:**
- Display user-friendly error messages
- Provide retry mechanism for failed API calls
- Log errors to console for debugging
- Show toast notifications for success/error

**Responsive Design:**
- Use Tailwind CSS responsive utilities
- Table should scroll horizontally on small screens
- Stack form fields on mobile
- Touch-friendly button sizes

### References

- [Source: Story 5.1 Implementation] - Alert Rule API backend
- [Source: Epic 4 Implementation] - Frontend infrastructure and patterns
- [Source: DashboardPage.tsx] - Page layout pattern
- [Source: NodeDetailPage.tsx] - Modal/dialog pattern
- [Source: alertsStore.ts] - State management pattern
- [Source: alerts.ts API] - API integration pattern
- [Source: Architecture.md#Frontend] - Frontend architecture guidelines
- [Source: Epics.md > Epic 5 > Story 5.3] - Story requirements and acceptance criteria
