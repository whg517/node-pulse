import { renderHook, waitFor, act } from '@testing-library/react'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useNodesQuery } from '../useNodesQuery'
import { createQueryWrapper } from '@/test/utils'

// Mock the API barrel; useNodesQuery imports { fetchNodes, fetchMetrics } from '@/api'
vi.mock('@/api', () => ({
  fetchNodes: vi.fn(),
  fetchMetrics: vi.fn(),
}))

// Import after mock so vi.mocked() can target the stubbed module.
import * as api from '@/api'

describe('useNodesQuery', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches nodes and metrics in parallel and merges them', async () => {
    const nodes = [{ id: '1', name: 'Node-1', ip: '10.0.0.1' }]
    const metrics = [{ node_id: '1', latency: 5 }]
    vi.mocked(api.fetchNodes).mockResolvedValue({ data: { nodes } } as any)
    vi.mocked(api.fetchMetrics).mockResolvedValue({ data: metrics } as any)

    const { wrapper } = createQueryWrapper()
    const { result } = renderHook(() => useNodesQuery(), { wrapper })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })

    expect(api.fetchNodes).toHaveBeenCalledTimes(1)
    expect(api.fetchMetrics).toHaveBeenCalledTimes(1)
    expect(result.current.data).toEqual({ nodes, metrics })
  })

  it('defaults to empty arrays when API returns nullish data', async () => {
    vi.mocked(api.fetchNodes).mockResolvedValue({ data: {} } as any) // nodes undefined
    vi.mocked(api.fetchMetrics).mockResolvedValue({ data: undefined } as any)

    const { wrapper } = createQueryWrapper()
    const { result } = renderHook(() => useNodesQuery(), { wrapper })

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })
    expect(result.current.data).toEqual({ nodes: [], metrics: [] })
  })

  it('accepts a custom polling interval', async () => {
    vi.useFakeTimers()
    try {
      const nodes = [{ id: '1', name: 'Node-1' }]
      vi.mocked(api.fetchNodes).mockResolvedValue({ data: { nodes } } as any)
      vi.mocked(api.fetchMetrics).mockResolvedValue({ data: [] } as any)

      const { wrapper } = createQueryWrapper()
      renderHook(() => useNodesQuery(5_000), { wrapper })

      await act(async () => {
        await Promise.resolve()
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(1)

      // custom 5s interval -> refetch after 5s
      await act(async () => {
        await vi.advanceTimersByTimeAsync(5_000)
      })
      expect(api.fetchNodes).toHaveBeenCalledTimes(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('surfaces errors from fetchNodes', async () => {
    const error = new Error('network error')
    vi.mocked(api.fetchNodes).mockRejectedValue(error)
    vi.mocked(api.fetchMetrics).mockResolvedValue({ data: [] } as any)

    const { wrapper } = createQueryWrapper()
    const { result } = renderHook(() => useNodesQuery(), { wrapper })

    await waitFor(() => {
      expect(result.current.isError).toBe(true)
    })
    expect(result.current.error).toEqual(error)
  })
})
