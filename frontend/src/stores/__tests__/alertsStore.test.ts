import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useAlertsStore } from '../alertsStore'
import * as alertsApi from '../../api/alerts'

// Mock the alerts API
vi.mock('../../api/alerts', () => ({
  fetchAlertRules: vi.fn(),
  fetchAlertRecords: vi.fn(),
}))

describe('useAlertsStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAlertsStore.setState({
      alertRules: [],
      alertRecords: [],
      filter: {
        level: 'all',
        status: 'all',
        nodeId: null,
        searchQuery: '',
      },
    })
    vi.clearAllMocks()
  })

  it('should have initial state', () => {
    const { result } = renderHook(() => useAlertsStore())

    expect(result.current.alertRules).toEqual([])
    expect(result.current.alertRecords).toEqual([])
    expect(result.current.filter).toEqual({
      level: 'all',
      status: 'all',
      nodeId: null,
      searchQuery: '',
    })
  })

  it('should set alert rules', () => {
    const { result } = renderHook(() => useAlertsStore())

    const mockRules = [
      {
        id: 'rule-1',
        metric: 'latency' as const,
        threshold: 100,
        level: 'P1' as const,
        nodeId: 'node-1',
        enabled: true,
      },
    ]

    act(() => {
      result.current.setAlertRules(mockRules)
    })

    expect(result.current.alertRules).toEqual(mockRules)
  })

  it('should set alert records', () => {
    const { result } = renderHook(() => useAlertsStore())

    const mockRecords = [
      {
        id: 'record-1',
        nodeId: 'node-1',
        metric: 'latency',
        level: 'P1',
        status: 'pending' as const,
        timestamp: '2024-01-01T00:00:00Z',
      },
    ]

    act(() => {
      result.current.setAlertRecords(mockRecords)
    })

    expect(result.current.alertRecords).toEqual(mockRecords)
  })

  it('should add alert rule', () => {
    const { result } = renderHook(() => useAlertsStore())

    const mockRule = {
      id: 'rule-1',
      metric: 'latency' as const,
      threshold: 100,
      level: 'P1' as const,
      nodeId: 'node-1',
      enabled: true,
    }

    act(() => {
      result.current.addAlertRule(mockRule)
    })

    expect(result.current.alertRules).toHaveLength(1)
    expect(result.current.alertRules[0]).toEqual(mockRule)
  })

  it('should update alert rule', () => {
    const { result } = renderHook(() => useAlertsStore())

    const initialRules = [
      {
        id: 'rule-1',
        metric: 'latency' as const,
        threshold: 100,
        level: 'P1' as const,
        nodeId: 'node-1',
        enabled: true,
      },
    ]

    act(() => {
      result.current.setAlertRules(initialRules)
    })

    act(() => {
      result.current.updateAlertRule('rule-1', { threshold: 200, enabled: false })
    })

    expect(result.current.alertRules[0].threshold).toBe(200)
    expect(result.current.alertRules[0].enabled).toBe(false)
    expect(result.current.alertRules[0].metric).toBe('latency') // Other fields unchanged
  })

  it('should remove alert rule', () => {
    const { result } = renderHook(() => useAlertsStore())

    const initialRules = [
      {
        id: 'rule-1',
        metric: 'latency' as const,
        threshold: 100,
        level: 'P1' as const,
        nodeId: 'node-1',
        enabled: true,
      },
      {
        id: 'rule-2',
        metric: 'jitter' as const,
        threshold: 50,
        level: 'P2' as const,
        nodeId: 'node-2',
        enabled: true,
      },
    ]

    act(() => {
      result.current.setAlertRules(initialRules)
    })

    act(() => {
      result.current.removeAlertRule('rule-1')
    })

    expect(result.current.alertRules).toHaveLength(1)
    expect(result.current.alertRules[0].id).toBe('rule-2')
  })

  it('should set filter', () => {
    const { result } = renderHook(() => useAlertsStore())

    const newFilter = {
      level: 'P1' as const,
      status: 'pending' as const,
      nodeId: 'node-1',
      searchQuery: 'test',
    }

    act(() => {
      result.current.setFilter(newFilter)
    })

    expect(result.current.filter).toEqual(newFilter)
  })

  it('should fetch alert rules from API', async () => {
    const mockRulesData = [
      {
        id: 'rule-1',
        metric: 'latency' as const,
        threshold: 100,
        level: 'P1' as const,
        node_id: 'node-1',
        enabled: true,
      },
    ]

    vi.mocked(alertsApi.fetchAlertRules).mockResolvedValueOnce({ data: mockRulesData })

    const { result } = renderHook(() => useAlertsStore())

    await act(async () => {
      await result.current.fetchAlertRules()
    })

    expect(result.current.alertRules).toHaveLength(1)
    expect(result.current.alertRules[0].id).toBe('rule-1')
    expect(result.current.alertRules[0].nodeId).toBe('node-1')
    expect(alertsApi.fetchAlertRules).toHaveBeenCalled()
  })

  it('should fetch alert records from API', async () => {
    const mockRecordsData = [
      {
        id: 'record-1',
        node_id: 'node-1',
        metric: 'latency',
        level: 'P1',
        status: 'pending' as const,
        timestamp: '2024-01-01T00:00:00Z',
      },
    ]

    vi.mocked(alertsApi.fetchAlertRecords).mockResolvedValueOnce({ data: mockRecordsData })

    const { result } = renderHook(() => useAlertsStore())

    await act(async () => {
      await result.current.fetchAlertRecords()
    })

    expect(result.current.alertRecords).toHaveLength(1)
    expect(result.current.alertRecords[0].id).toBe('record-1')
    expect(result.current.alertRecords[0].nodeId).toBe('node-1')
    expect(alertsApi.fetchAlertRecords).toHaveBeenCalled()
  })

  it('should handle fetch alert rules error', async () => {
    vi.mocked(alertsApi.fetchAlertRules).mockRejectedValueOnce(new Error('Network error'))

    const { result } = renderHook(() => useAlertsStore())

    await expect(async () => {
      await act(async () => {
        await result.current.fetchAlertRules()
      })
    }).rejects.toThrow('Network error')
  })

  it('should handle fetch alert records error', async () => {
    vi.mocked(alertsApi.fetchAlertRecords).mockRejectedValueOnce(new Error('Network error'))

    const { result } = renderHook(() => useAlertsStore())

    await expect(async () => {
      await act(async () => {
        await result.current.fetchAlertRecords()
      })
    }).rejects.toThrow('Network error')
  })
})
