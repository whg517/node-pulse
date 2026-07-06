import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook } from '@testing-library/react'
import type { ReactNode } from 'react'

// Capture the message handler so tests can invoke it directly.
let capturedHandler: ((msg: { type: string; payload: unknown; timestamp: string }) => void) | null = null

// Hoist the mock fns so vi.mock factories (which are hoisted) can reference them.
const { mockSetNodeStatus, mockUpsertAlertRecord, mockShowAlertNotification } = vi.hoisted(() => ({
  mockSetNodeStatus: vi.fn(),
  mockUpsertAlertRecord: vi.fn(),
  mockShowAlertNotification: vi.fn(),
}))

vi.mock('@/stores/nodesStore', () => ({
  useNodesStore: (selector: (s: unknown) => unknown) =>
    selector({ setNodeStatus: mockSetNodeStatus }),
}))

vi.mock('@/stores/alertsStore', () => ({
  useAlertsStore: (selector: (s: unknown) => unknown) =>
    selector({ upsertAlertRecord: mockUpsertAlertRecord }),
}))

vi.mock('@/services/NotificationService', () => ({
  initialize: vi.fn(),
  destroy: vi.fn(),
  showAlertNotification: mockShowAlertNotification,
  // F4: setNotificationPrefsSource is called on mount to wire the settings
  // store into the service; the test doesn't exercise filtering so a noop is fine.
  setNotificationPrefsSource: vi.fn(),
}))

vi.mock('@/services/WebSocketService', () => ({
  initialize: vi.fn((handler: (msg: { type: string; payload: unknown; timestamp: string }) => void) => {
    capturedHandler = handler
  }),
  connect: vi.fn(),
  disconnect: vi.fn(),
}))

vi.mock('react-router-dom', () => ({
  useNavigate: () => vi.fn(),
}))

import { useGlobalRealtime } from '../useGlobalRealtime'

describe('useGlobalRealtime', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    capturedHandler = null
  })

  function renderIt() {
    return renderHook(() => useGlobalRealtime(), {
      wrapper: ({ children }: { children: ReactNode }) => <>{children}</>,
    })
  }

  it('updates nodesStore on node:online event', () => {
    renderIt()
    expect(capturedHandler).not.toBeNull()
    capturedHandler!({
      type: 'node:online',
      payload: { node_id: 'node-abc', status: 'online' },
      timestamp: new Date().toISOString(),
    })
    expect(mockSetNodeStatus).toHaveBeenCalledWith('node-abc', 'online')
  })

  it('updates nodesStore on node:offline event', () => {
    renderIt()
    capturedHandler!({
      type: 'node:offline',
      payload: { node_id: 'node-xyz', status: 'offline' },
      timestamp: new Date().toISOString(),
    })
    expect(mockSetNodeStatus).toHaveBeenCalledWith('node-xyz', 'offline')
  })

  it('ignores unrelated event types', () => {
    renderIt()
    capturedHandler!({
      type: 'system:heartbeat',
      payload: {},
      timestamp: new Date().toISOString(),
    })
    expect(mockSetNodeStatus).not.toHaveBeenCalled()
  })
})
