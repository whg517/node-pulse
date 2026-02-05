import { describe, it, expect, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useDashboardStore } from '../dashboardStore'

describe('useDashboardStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useDashboardStore.setState({
      filters: {
        region: null,
        status: 'all',
        searchQuery: '',
      },
      timeRange: '24h',
      refreshInterval: 5,
      autoRefresh: true,
      top5AbnormalNodes: [],
    })
  })

  it('should have initial state', () => {
    const { result } = renderHook(() => useDashboardStore())

    expect(result.current.filters).toEqual({
      region: null,
      status: 'all',
      searchQuery: '',
    })
    expect(result.current.timeRange).toBe('24h')
    expect(result.current.refreshInterval).toBe(5)
    expect(result.current.autoRefresh).toBe(true)
    expect(result.current.top5AbnormalNodes).toEqual([])
  })

  it('should set filters', () => {
    const { result } = renderHook(() => useDashboardStore())

    const newFilters = {
      region: 'us-east',
      status: 'online' as const,
      searchQuery: 'test',
    }

    act(() => {
      result.current.setFilters(newFilters)
    })

    expect(result.current.filters).toEqual(newFilters)
  })

  it('should set time range', () => {
    const { result } = renderHook(() => useDashboardStore())

    act(() => {
      result.current.setTimeRange('7d')
    })

    expect(result.current.timeRange).toBe('7d')
  })

  it('should set refresh interval', () => {
    const { result } = renderHook(() => useDashboardStore())

    act(() => {
      result.current.setRefreshInterval(10)
    })

    expect(result.current.refreshInterval).toBe(10)
  })

  it('should toggle auto refresh', () => {
    const { result } = renderHook(() => useDashboardStore())

    expect(result.current.autoRefresh).toBe(true)

    act(() => {
      result.current.toggleAutoRefresh()
    })

    expect(result.current.autoRefresh).toBe(false)

    act(() => {
      result.current.toggleAutoRefresh()
    })

    expect(result.current.autoRefresh).toBe(true)
  })

  it('should set top 5 abnormal nodes', () => {
    const { result } = renderHook(() => useDashboardStore())

    const mockNodes = [
      {
        id: 'node-1',
        name: 'Node 1',
        ip: '192.168.1.1',
        region: 'us-east',
        tags: ['tag1'],
        status: 'offline' as const,
      },
      {
        id: 'node-2',
        name: 'Node 2',
        ip: '192.168.1.2',
        region: 'us-west',
        tags: ['tag2'],
        status: 'online' as const,
      },
    ]

    act(() => {
      result.current.setTop5AbnormalNodes(mockNodes)
    })

    expect(result.current.top5AbnormalNodes).toEqual(mockNodes)
  })

  it('should support all time range options', () => {
    const { result } = renderHook(() => useDashboardStore())

    act(() => {
      result.current.setTimeRange('24h')
    })
    expect(result.current.timeRange).toBe('24h')

    act(() => {
      result.current.setTimeRange('7d')
    })
    expect(result.current.timeRange).toBe('7d')

    act(() => {
      result.current.setTimeRange('30d')
    })
    expect(result.current.timeRange).toBe('30d')
  })

  it('should support all status filter options', () => {
    const { result } = renderHook(() => useDashboardStore())

    const statuses: Array<'online' | 'offline' | 'all'> = ['online', 'offline', 'all']

    statuses.forEach((status) => {
      act(() => {
        result.current.setFilters({
          region: null,
          status,
          searchQuery: '',
        })
      })

      expect(result.current.filters.status).toBe(status)
    })
  })
})
