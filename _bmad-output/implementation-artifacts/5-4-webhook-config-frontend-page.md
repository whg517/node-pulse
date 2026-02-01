# Story 5.4: Webhook Config Frontend Page

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a 运维主管,
I can 在前端页面配置 Webhook,
So that 可以可视化地管理告警推送地址。

## Acceptance Criteria

**Given** 用户已登录并访问 Webhook 配置页面
**When** 页面加载完成
**Then** 显示所有 Webhook 配置列表
**And** 提供"添加 Webhook"按钮
**And** 支持编辑和删除 Webhook
**And** 表单包含：URL 输入（验证 HTTPS）、事件格式编辑

**覆盖需求:** FR6（Webhook 配置）

**创建表:** 无（后端已完成 Story 5.2）

## Tasks / Subtasks

- [ ] Task 1: Create WebhooksPage component (AC: Then - 显示 Webhook 配置列表)
  - [ ] Subtask 1.1: Create page component file structure
  - [ ] Subtask 1.2: Add routing configuration for /webhooks route
  - [ ] Subtask 1.3: Implement page layout with header and container
  - [ ] Subtask 1.4: Add loading state handling
  - [ ] Subtask 1.5: Add error state handling with retry option

- [ ] Task 2: Implement webhooks list table (AC: Then - 显示所有 Webhook 配置列表)
  - [ ] Subtask 2.1: Create WebhooksTable component
  - [ ] Subtask 2.2: Display columns: ID, URL, Event Format, Enabled, Status, Actions
  - [ ] Subtask 2.3: Add color-coded status badges (enabled/disabled)
  - [ ] Subtask 2.4: Truncate long URLs with tooltip
  - [ ] Subtask 2.5: Implement empty state when no webhooks
  - [ ] Subtask 2.6: Add responsive table design

- [ ] Task 3: Implement CreateWebhook modal/form (AC: Then - 提供"添加 Webhook"按钮)
  - [ ] Subtask 3.1: Create WebhookForm component (reusable for create/edit)
  - [ ] Subtask 3.2: Add "Add Webhook" button in page header
  - [ ] Subtask 3.3: Implement modal/dialog for form
  - [ ] Subtask 3.4: Add URL input with HTTPS validation
  - [ ] Subtask 3.5: Add event format JSON editor (or textarea)
  - [ ] Subtask 3.6: Add enabled/disabled toggle switch
  - [ ] Subtask 3.7: Implement form validation (HTTPS required)
  - [ ] Subtask 3.8: Add submit and cancel buttons

- [ ] Task 4: Implement Edit functionality (AC: Then - 支持编辑 Webhook)
  - [ ] Subtask 4.1: Add edit button to each table row
  - [ ] Subtask 4.2: Pre-populate form with existing webhook data
  - [ ] Subtask 4.3: Implement update API call
  - [ ] Subtask 4.4: Show success/error toast notifications
  - [ ] Subtask 4.5: Refresh webhooks list after successful update

- [ ] Task 5: Implement Delete functionality (AC: Then - 支持删除 Webhook)
  - [ ] Subtask 5.1: Add delete button to each table row
  - [ ] Subtask 5.2: Implement confirmation dialog
  - [ ] Subtask 5.3: Implement delete API call
  - [ ] Subtask 5.4: Show success/error toast notifications
  - [ ] Subtask 5.5: Refresh webhooks list after successful deletion

- [ ] Task 6: Add HTTPS URL validation (AC: Then - 验证 HTTPS URL)
  - [ ] Subtask 6.1: Client-side URL validation
  - [ ] Subtask 6.2: Show inline error messages for invalid URLs
  - [ ] Subtask 6.3: Prevent form submission with invalid URL
  - [ ] Subtask 6.4: Display clear validation requirements

- [ ] Task 7: Implement event format editor (AC: Then - 事件格式编辑)
  - [ ] Subtask 7.1: Add JSON editor/textarea for event format
  - [ ] Subtask 7.2: Validate JSON format
  - [ ] Subtask 7.3: Provide default event format template
  - [ ] Subtask 7.4: Add syntax validation feedback

- [ ] Task 8: Write comprehensive tests (AC: 完整功能验证)
  - [ ] Subtask 8.1: Unit tests for WebhooksPage component
  - [ ] Subtask 8.2: Unit tests for WebhooksTable component
  - [ ] Subtask 8.3: Unit tests for WebhookForm component
  - [ ] Subtask 8.4: Integration tests for HTTPS validation
  - [ ] Subtask 8.5: Integration tests for CRUD operations
  - [ ] Subtask 8.6: Test JSON validation in event format

- [ ] Task 9: Update API integration (AC: API 集成)
  - [ ] Subtask 9.1: Create webhooks API functions (if not exists)
  - [ ] Subtask 9.2: Create webhooks store (if not exists)
  - [ ] Subtask 9.3: Handle API errors gracefully
  - [ ] Subtask 9.4: Implement proper loading states

## Dev Notes

### Epic Analysis

**Epic 5: 告警规则配置与通知** - 系统可以自动检测异常并通过 Webhook 推送告警

**Story Context in Epic:**
- Story 5.1: 告警规则 API (已完成 - 后端 API)
- Story 5.2: Webhook 配置 API (已完成 - 后端 API)
- Story 5.3: 告警规则前端页面 (已完成 - 前端 UI)
- Story 5.4: **Webhook 配置前端页面** (本故事) - **前端 UI 实现**
- Story 5.5-5.8: 后续告警功能

**Critical Prerequisites:**
- **Story 5.2 已完成**: Webhook Config API 后端已实现
- **Epic 4 已完成**: 前端基础设施已完成（React Router, Zustand, API layer）
- **Story 5.3 已完成**: Alert Rules Frontend Page 提供了参考模式
- **需要创建**: Webhooks API functions (类似 alerts.ts)
- **需要创建**: Webhooks store (类似 alertsStore.ts)

### Architecture Alignment

**Frontend Architecture** [Source: Story 5.3 Implementation]:
```
pulse-frontend/src/
├── pages/
│   └── WebhooksPage.tsx            # NEW - Main page component
├── components/
│   └── webhooks/
│       ├── index.ts                # NEW - Export barrel
│       ├── WebhooksTable.tsx       # NEW - Webhooks list table
│       ├── WebhookForm.tsx         # NEW - Create/edit form
│       └── WebhookDialog.tsx       # NEW - Modal wrapper
├── stores/
│   └── webhooksStore.ts            # NEW - Webhooks state management
├── api/
│   └── webhooks.ts                 # NEW - Webhooks API functions
└── routes/
    └── App.tsx                     # UPDATE - Add /webhooks route
```

**Component Patterns** [Source: Story 5.3]:
- Follow AlertRulesPage structure exactly
- Similar table design with status badges
- Similar modal form with validation
- Similar delete confirmation dialog
- Same RBAC pattern (admin-only access)

**Key Differences from Alert Rules:**
- **Admin-only access**: Webhook config is admin-only (stricter than alert rules)
- **HTTPS validation**: Must enforce HTTPS URLs
- **Event format**: JSON editor instead of dropdowns
- **No node selection**: Webhooks are global (no node scope)

### Previous Story Intelligence

**From Story 5.3 (Alert Rules Frontend)** [Source: Story 5.3 Implementation]:
- **Page structure**: Header with actions, main content area, error/loading states
- **Table pattern**: Responsive tables with action buttons, status badges
- **Form pattern**: Modal dialogs with validation, submit/cancel buttons
- **State management**: Zustand stores with optimistic updates
- **Error handling**: Toast notifications via ToastContext
- **Access control**: RBAC from authStore

**Key Learnings from Story 5.3**:
- Use existing ToastContext for notifications
- Follow consistent table design patterns
- Implement proper loading and error states
- Use responsive design with Tailwind CSS
- Leverage existing stores and API functions
- Test components with Vitest and React Testing Library
- Fix TypeScript type issues in mocks

### Technical Requirements

**API Layer** (src/api/webhooks.ts):
```typescript
import { apiClient } from './client'

export interface WebhookDTO {
  id: string
  url: string
  event_format: Record<string, any>
  enabled: boolean
  created_at: string
}

export interface CreateWebhookRequest {
  url: string
  event_format?: Record<string, any>
  enabled?: boolean
}

export interface UpdateWebhookRequest {
  url?: string
  event_format?: Record<string, any>
  enabled?: boolean
}

export async function fetchWebhooks(): Promise<{ data: WebhookDTO[] }> {
  return apiClient<{ data: WebhookDTO[] }>('/api/v1/webhooks')
}

export async function createWebhook(
  request: CreateWebhookRequest
): Promise<{ data: WebhookDTO }> {
  return apiClient<{ data: WebhookDTO }>('/api/v1/webhooks', {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

export async function updateWebhook(
  id: string,
  request: UpdateWebhookRequest
): Promise<{ data: WebhookDTO }> {
  return apiClient<{ data: WebhookDTO }>(`/api/v1/webhooks/${id}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  })
}

export async function deleteWebhook(
  id: string
): Promise<{ message: string }> {
  return apiClient<{ message: string }>(`/api/v1/webhooks/${id}`, {
    method: 'DELETE',
  })
}
```

**Store** (src/stores/webhooksStore.ts):
```typescript
import { create } from 'zustand'
import * as webhooksAPI from '../api/webhooks'

export interface Webhook {
  id: string
  url: string
  eventFormat: Record<string, any>
  enabled: boolean
}

export interface WebhooksState {
  webhooks: Webhook[]
}

export interface WebhooksActions {
  setWebhooks: (webhooks: Webhook[]) => void
  addWebhook: (webhook: Webhook) => void
  updateWebhook: (id: string, updates: Partial<Webhook>) => void
  removeWebhook: (id: string) => void
  fetchWebhooks: () => Promise<void>
}

type WebhooksStore = WebhooksState & WebhooksActions

export const useWebhooksStore = create<WebhooksStore>((set, get) => ({
  // State
  webhooks: [],

  // Actions
  setWebhooks: (webhooks) => set({ webhooks }),

  addWebhook: (webhook) => {
    set((state) => ({ webhooks: [...state.webhooks, webhook] }))
  },

  updateWebhook: (id, updates) => {
    set((state) => ({
      webhooks: state.webhooks.map((w) =>
        w.id === id ? { ...w, ...updates } : w
      ),
    }))
  },

  removeWebhook: (id) => {
    set((state) => ({
      webhooks: state.webhooks.filter((w) => w.id !== id),
    }))
  },

  fetchWebhooks: async () => {
    try {
      const response = await webhooksAPI.fetchWebhooks()
      const webhooks: Webhook[] = response.data.map((webhook) => ({
        id: webhook.id,
        url: webhook.url,
        eventFormat: webhook.event_format,
        enabled: webhook.enabled,
      }))
      set({ webhooks })
    } catch (error) {
      console.error('Failed to fetch webhooks:', error)
      throw error
    }
  },
}))
```

**Page Component** (src/pages/WebhooksPage.tsx):
```typescript
// Similar structure to AlertRulesPage.tsx
// Key differences:
// - Admin-only access control
// - HTTPS URL validation in form
// - Event format JSON editor
```

**Form Component** (src/components/webhooks/WebhookForm.tsx):
```typescript
interface WebhookFormProps {
  mode: 'create' | 'edit'
  initialData?: Webhook
  onSubmit: (data: any) => Promise<void>
  onCancel: () => void
}

export function WebhookForm({ mode, initialData, onSubmit, onCancel }: WebhookFormProps) {
  const [url, setUrl] = useState(initialData?.url || '')
  const [eventFormat, setEventFormat] = useState(
    JSON.stringify(initialData?.eventFormat || {}, null, 2)
  )
  const [enabled, setEnabled] = useState(initialData?.enabled ?? true)

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  const validate = () => {
    const newErrors: Record<string, string> = {}

    // Validate HTTPS URL
    if (!url) {
      newErrors.url = 'URL is required'
    } else if (!url.startsWith('https://')) {
      newErrors.url = 'URL must use HTTPS protocol for security'
    }

    // Validate JSON
    try {
      JSON.parse(eventFormat)
    } catch (e) {
      newErrors.eventFormat = 'Invalid JSON format'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) return

    setIsSubmitting(true)
    try {
      const data = {
        url,
        event_format: JSON.parse(eventFormat),
        enabled,
      }
      await onSubmit(data)
    } catch (error) {
      console.error('Failed to submit webhook:', error)
      throw error
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      {/* URL Input */}
      <div>
        <label htmlFor="url">Webhook URL</label>
        <input
          type="url"
          id="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://example.com/webhook"
          className={errors.url ? 'border-red-300' : 'border-gray-300'}
        />
        {errors.url && <span className="text-red-600">{errors.url}</span>}
        <p className="text-sm text-gray-500">
          Must be a valid HTTPS URL for secure notification delivery
        </p>
      </div>

      {/* Event Format JSON Editor */}
      <div>
        <label htmlFor="eventFormat">Event Format (JSON)</label>
        <textarea
          id="eventFormat"
          value={eventFormat}
          onChange={(e) => setEventFormat(e.target.value)}
          rows={10}
          className="font-mono text-sm"
          placeholder='{\n  "version": "1.0",\n  "alert": {...}\n}'
        />
        {errors.eventFormat && <span className="text-red-600">{errors.eventFormat}</span>}
        <p className="text-sm text-gray-500">
          Customize the JSON payload sent to this webhook endpoint
        </p>
      </div>

      {/* Enabled Toggle */}
      <div>
        <input
          type="checkbox"
          id="enabled"
          checked={enabled}
          onChange={(e) => setEnabled(e.target.checked)}
        />
        <label htmlFor="enabled">Enabled</label>
      </div>

      {/* Buttons */}
      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Saving...' : mode === 'create' ? 'Add Webhook' : 'Update Webhook'}
      </button>
      <button type="button" onClick={onCancel}>
        Cancel
      </button>
    </form>
  )
}
```

### Testing Requirements

**Component Tests:**
- Test page renders correctly
- Test loading state displays
- Test error state displays with retry button
- Test empty state displays when no webhooks
- Test create button opens dialog
- Test webhooks are fetched on mount
- Test admin-only access (viewer cannot see buttons)

**Form Tests:**
- Test HTTPS URL validation
- Test HTTP URL rejected
- Test JSON validation
- Test invalid JSON shows error
- Test form submission with valid data

**Integration Tests:**
- Test create webhook flow
- Test edit webhook flow
- Test delete webhook flow with confirmation
- Test HTTPS URL enforcement
- Test JSON event format handling

### Implementation Guidelines

**UX Best Practices:**
- Use clear labels for form fields
- Provide inline validation feedback
- Show loading states during API calls
- Display success/error toast notifications
- Confirm destructive actions (delete)
- Use color coding for visual hierarchy

**HTTPS Validation:**
- Client-side validation before API calls
- Clear error messages for non-HTTPS URLs
- Prevent form submission with invalid URLs
- Display validation requirements

**JSON Editor:**
- Use textarea with monospace font
- Validate JSON format on submit
- Provide default event format template
- Show clear error messages for invalid JSON

**Access Control:**
- Admin-only access (stricter than alert rules)
- Hide all buttons for non-admin users
- Show read-only message for non-admin users
- Use authStore to check user role

**State Management:**
- Create webhooksStore following alertsStore pattern
- Implement optimistic updates
- Refresh list after mutations
- Handle API errors gracefully

### References

- [Source: Story 5.2 Implementation] - Webhook Config API backend
- [Source: Story 5.3 Implementation] - Alert Rules Frontend Page pattern
- [Source: Epic 4 Implementation] - Frontend infrastructure and patterns
- [Source: webhooks.go Backend] - API endpoints and validation
- [Source: webhook.go Model] - Webhook data structure
- [Source: Architecture.md#Frontend] - Frontend architecture guidelines
- [Source: Epics.md > Epic 5 > Story 5.4] - Story requirements and acceptance criteria

## Dev Agent Record

### Agent Model Used

claude-sonnet-4.5-20250929

### Debug Log References

### Completion Notes List

**To be completed after implementation...**
