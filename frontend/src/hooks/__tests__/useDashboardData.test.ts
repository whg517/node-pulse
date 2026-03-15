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
    const mockNodesList = [
      { id: '1', name: 'Node-1', ip: '192.168.1.1', region: '华东' },
    ]
    const mockNodes = { data: { nodes: mockNodesList } }
    const mockMetrics = { data: [] }

    vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
    vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

    const { result } = renderHook(() => useDashboardData())

    expect(result.current.isLoading).toBe(true)

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(api.fetchNodes).toHaveBeenCalledTimes(1)
    expect(api.fetchMetrics).toHaveBeenCalledTimes(1)
    expect(result.current.nodes).toEqual(mockNodesList)
  })

  it('should set polling state after successful fetch', async () => {
    const mockNodes = { data: { nodes: [] } }
    const mockMetrics = { data: [] }

    vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
    vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

    const { result } = renderHook(() => useDashboardData())

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.isPolling).toBe(true)
  })

  it('should handle API errors gracefully', async () => {
    const error = new Error('Network error')
    vi.mocked(api.fetchNodes).mockRejectedValue(error)
    vi.mocked(api.fetchMetrics).mockRejectedValue(error)

    const { result } = renderHook(() => useDashboardData())

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.error).toEqual(error)
    expect(result.current.isPolling).toBe(false)
  })

  it('should auto-refetch at polling interval', async () => {
    vi.useFakeTimers()

    try {
      const mockNodes = { data: { nodes: [] } }
      const mockMetrics = { data: [] }

      vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
      vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

      renderHook(() => useDashboardData())
      await act(async () => {
        await Promise.resolve()
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)

      await act(async () => {
        await vi.advanceTimersByTimeAsync(5000)
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('should pause polling when page is hidden and resume when visible', async () => {
    vi.useFakeTimers()

    const originalVisibility = document.visibilityState
    const setVisibility = (state: DocumentVisibilityState) => {
      Object.defineProperty(document, 'visibilityState', {
        configurable: true,
        value: state,
      })
      document.dispatchEvent(new Event('visibilitychange'))
    }

    try {
      const mockNodes = { data: { nodes: [] } }
      const mockMetrics = { data: [] }

      vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
      vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

      renderHook(() => useDashboardData())
      await act(async () => {
        await Promise.resolve()
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)

      await act(async () => {
        setVisibility('hidden')
      })
      await act(async () => {
        await vi.advanceTimersByTimeAsync(10000)
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)

      await act(async () => {
        setVisibility('visible')
        await Promise.resolve()
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(2)
    } finally {
      Object.defineProperty(document, 'visibilityState', {
        configurable: true,
        value: originalVisibility,
      })
      vi.useRealTimers()
    }
  })

  it('should use exponential backoff after failure', async () => {
    vi.useFakeTimers()

    try {
      const error = new Error('Network error')
      const mockNodes = { data: { nodes: [] } }
      const mockMetrics = { data: [] }
      vi.mocked(api.fetchNodes)
        .mockRejectedValueOnce(error)
        .mockResolvedValue(mockNodes as any)
      vi.mocked(api.fetchMetrics)
        .mockRejectedValueOnce(error)
        .mockResolvedValue(mockMetrics as any)

      renderHook(() => useDashboardData())
      await act(async () => {
        await Promise.resolve()
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)

      // first failure => next interval should be 10s
      await act(async () => {
        await vi.advanceTimersByTimeAsync(5000)
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)

      await act(async () => {
        await vi.advanceTimersByTimeAsync(5000)
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('should provide refetch function', async () => {
    const mockNodes = { data: { nodes: [] } }
    const mockMetrics = { data: [] }

    vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
    vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

    const { result } = renderHook(() => useDashboardData())

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

  it('should start polling by default', async () => {
    const mockNodes = { data: { nodes: [] } }
    const mockMetrics = { data: [] }

    vi.mocked(api.fetchNodes).mockResolvedValue(mockNodes as any)
    vi.mocked(api.fetchMetrics).mockResolvedValue(mockMetrics as any)

    renderHook(() => useDashboardData())

    await waitFor(() => {
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)
      expect(api.fetchMetrics).toHaveBeenCalledTimes(1)
    })
  })
})
