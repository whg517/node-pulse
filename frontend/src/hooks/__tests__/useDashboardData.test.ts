import { renderHook, waitFor, act } from '@testing-library/react'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useDashboardData } from '../useDashboardData'
import * as api from '../../api'

// Mock API modules
vi.mock('../../api', () => ({
  fetchNodes: vi.fn(),
  fetchMetrics: vi.fn(),
}))

describe('useDashboardData', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch initial data on mount', async () => {
    const mockNodes = {
      data: [
        { id: '1', name: 'Node-1', ip: '192.168.1.1', region: '华东' },
      ],
    }
    const mockMetrics = { data: [] }

    vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
    vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

    const { result } = renderHook(() => useDashboardData(5000))

    expect(result.current.isLoading).toBe(true)

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(api.fetchNodes).toHaveBeenCalledTimes(1)
    expect(api.fetchMetrics).toHaveBeenCalledTimes(1)
    expect(result.current.nodes).toEqual(mockNodes.data)
  })

  it('should set polling state after successful fetch', async () => {
    const mockNodes = { data: [] }
    const mockMetrics = { data: [] }

    vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
    vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

    const { result } = renderHook(() => useDashboardData(5000))

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.isPolling).toBe(true)
  })

  it('should handle API errors gracefully', async () => {
    const error = new Error('Network error')
    vi.mocked(api.fetchNodes).mockRejectedValue(error)
    vi.mocked(api.fetchMetrics).mockRejectedValue(error)

    const { result } = renderHook(() => useDashboardData(5000))

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.error).toEqual(error)
    expect(result.current.isPolling).toBe(false)
  })

  it('should poll data at specified interval', async () => {
    vi.useFakeTimers()

    try {
      const mockNodes = { data: [] }
      const mockMetrics = { data: [] }

      vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
      vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

      renderHook(() => useDashboardData(5000))

      // Initial fetch happens synchronously with fake timers
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)

      // Fast-forward 5 seconds to trigger polling
      act(() => {
        vi.advanceTimersByTime(5000)
      })

      // Should have called fetchNodes again
      expect(api.fetchNodes).toHaveBeenCalledTimes(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('should cleanup interval on unmount', async () => {
    vi.useFakeTimers()

    try {
      const mockNodes = { data: [] }
      const mockMetrics = { data: [] }

      vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
      vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

      const { unmount } = renderHook(() => useDashboardData(5000))

      // Initial fetch
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)

      unmount()

      // Fast-forward time (should not trigger another fetch after unmount)
      act(() => {
        vi.advanceTimersByTime(5000)
      })

      // Should not call fetchNodes again after unmount
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)
    } finally {
      vi.useRealTimers()
    }
  })

  it('should provide refetch function', async () => {
    const mockNodes = { data: [] }
    const mockMetrics = { data: [] }

    vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
    vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

    const { result } = renderHook(() => useDashboardData(5000))

    await waitFor(() => {
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)
    })

    act(() => {
      result.current.refetch()
    })

    await waitFor(() => {
      expect(api.fetchNodes).toHaveBeenCalledTimes(2)
    })
  })

  it('should use default polling interval of 5000ms', async () => {
    vi.useFakeTimers()

    try {
      const mockNodes = { data: [] }
      const mockMetrics = { data: [] }

      vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
      vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

      renderHook(() => useDashboardData())

      // Initial fetch
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)

      act(() => {
        vi.advanceTimersByTime(5000)
      })

      // Should have called fetchNodes again after 5000ms
      expect(api.fetchNodes).toHaveBeenCalledTimes(2)
    } finally {
      vi.useRealTimers()
    }
  })
})
