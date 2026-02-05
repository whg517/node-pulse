import { renderHook, act, waitFor } from '@testing-library/react'
import { vi } from 'vitest'
import { useNodeDetail } from '../useNodeDetail'
import { fetchNode, fetchNodeStatus, fetchMetrics } from '../../api'

// Mock the API modules
vi.mock('../../api/nodes')
vi.mock('../../api/data')

const mockFetchNode = fetchNode as ReturnType<typeof vi.mocked<typeof fetchNode>>
const mockFetchNodeStatus = fetchNodeStatus as ReturnType<typeof vi.mocked<typeof fetchNodeStatus>>
const mockFetchMetrics = fetchMetrics as ReturnType<typeof vi.mocked<typeof fetchMetrics>>

describe('useNodeDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches node data on mount', async () => {
    const mockNode = {
      data: {
        id: 'node-1',
        name: 'Test Node',
        ip: '192.168.1.1',
        region: 'us-east',
        tags: ['production'],
        status: 'online',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    }

    const mockStatus = {
      data: {
        status: 'online',
        last_heartbeat: '2024-01-01T12:00:00Z',
      },
    }

    const mockMetrics = {
      data: [
        {
          node_id: 'node-1',
          latency_ms: 45,
          packet_loss_rate: 0,
          jitter_ms: 5,
          timestamp: '2024-01-01T12:00:00Z',
        },
      ],
    }

    mockFetchNode.mockResolvedValue(mockNode as any)
    mockFetchNodeStatus.mockResolvedValue(mockStatus as any)
    mockFetchMetrics.mockResolvedValue(mockMetrics as any)

    const { result } = renderHook(() => useNodeDetail('node-1', 0))

    expect(result.current.isLoading).toBe(true)

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(mockFetchNode).toHaveBeenCalledWith('node-1')
    expect(mockFetchNodeStatus).toHaveBeenCalledWith('node-1')
    expect(mockFetchMetrics).toHaveBeenCalledWith(['node-1'])

    expect(result.current.node).toEqual(mockNode.data)
    expect(result.current.nodeStatus).toEqual(mockStatus.data)
    expect(result.current.metrics).toEqual(mockMetrics.data[0])
    expect(result.current.error).toBeNull()
  })

  it('handles fetch errors', async () => {
    const error = new Error('Failed to fetch')
    mockFetchNode.mockRejectedValue(error)
    mockFetchNodeStatus.mockRejectedValue(error)
    mockFetchMetrics.mockRejectedValue(error)

    const { result } = renderHook(() => useNodeDetail('node-1', 0))

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(result.current.error).toEqual(error)
    expect(result.current.node).toBeNull()
    expect(result.current.nodeStatus).toBeNull()
    expect(result.current.metrics).toBeNull()
  })

  it('polls data at specified interval', async () => {
    vi.useFakeTimers()

    try {
      const mockNode = {
        data: {
          id: 'node-1',
          name: 'Test Node',
          ip: '192.168.1.1',
          region: 'us-east',
          tags: ['production'],
          status: 'online',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      }

      const mockStatus = {
        data: {
          status: 'online',
          last_heartbeat: '2024-01-01T12:00:00Z',
        },
      }

      const mockMetrics = {
        data: [
          {
            node_id: 'node-1',
            latency_ms: 45,
            packet_loss_rate: 0,
            jitter_ms: 5,
            timestamp: '2024-01-01T12:00:00Z',
          },
        ],
      }

      mockFetchNode.mockResolvedValue(mockNode as any)
      mockFetchNodeStatus.mockResolvedValue(mockStatus as any)
      mockFetchMetrics.mockResolvedValue(mockMetrics as any)

      renderHook(() => useNodeDetail('node-1', 5000))

      // Initial fetch happens synchronously with fake timers
      expect(mockFetchNode).toHaveBeenCalledTimes(1)

      // Fast-forward time
      act(() => {
        vi.advanceTimersByTime(5000)
      })

      // Should have called fetchNode again
      expect(mockFetchNode).toHaveBeenCalledTimes(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('does not poll when interval is 0', async () => {
    vi.useFakeTimers()

    try {
      const mockNode = {
        data: {
          id: 'node-1',
          name: 'Test Node',
          ip: '192.168.1.1',
          region: 'us-east',
          tags: ['production'],
          status: 'online',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      }

      const mockStatus = {
        data: {
          status: 'online',
          last_heartbeat: '2024-01-01T12:00:00Z',
        },
      }

      const mockMetrics = {
        data: [
          {
            node_id: 'node-1',
            latency_ms: 45,
            packet_loss_rate: 0,
            jitter_ms: 5,
            timestamp: '2024-01-01T12:00:00Z',
          },
        ],
      }

      mockFetchNode.mockResolvedValue(mockNode as any)
      mockFetchNodeStatus.mockResolvedValue(mockStatus as any)
      mockFetchMetrics.mockResolvedValue(mockMetrics as any)

      renderHook(() => useNodeDetail('node-1', 0))

      // Initial fetch
      expect(mockFetchNode).toHaveBeenCalledTimes(1)

      // Fast-forward time (should not trigger another fetch)
      act(() => {
        vi.advanceTimersByTime(10000)
      })

      // Should still be 1 (no polling when interval is 0)
      expect(mockFetchNode).toHaveBeenCalledTimes(1)
    } finally {
      vi.useRealTimers()
    }
  })

  it('provides refetch function', async () => {
    const mockNode = {
      data: {
        id: 'node-1',
        name: 'Test Node',
        ip: '192.168.1.1',
        region: 'us-east',
        tags: ['production'],
        status: 'online',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    }

    const mockStatus = {
      data: {
        status: 'online',
        last_heartbeat: '2024-01-01T12:00:00Z',
      },
    }

    const mockMetrics = {
      data: [
        {
          node_id: 'node-1',
          latency_ms: 45,
          packet_loss_rate: 0,
          jitter_ms: 5,
          timestamp: '2024-01-01T12:00:00Z',
        },
      ],
    }

    mockFetchNode.mockResolvedValue(mockNode as any)
    mockFetchNodeStatus.mockResolvedValue(mockStatus as any)
    mockFetchMetrics.mockResolvedValue(mockMetrics as any)

    const { result } = renderHook(() => useNodeDetail('node-1', 0))

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    expect(mockFetchNode).toHaveBeenCalledTimes(1)

    await act(async () => {
      await result.current.refetch()
    })

    expect(mockFetchNode).toHaveBeenCalledTimes(2)
  })

  it('cleans up interval on unmount', async () => {
    vi.useFakeTimers()

    try {
      const mockNode = {
        data: {
          id: 'node-1',
          name: 'Test Node',
          ip: '192.168.1.1',
          region: 'us-east',
          tags: ['production'],
          status: 'online',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      }

      const mockStatus = {
        data: {
          status: 'online',
          last_heartbeat: '2024-01-01T12:00:00Z',
        },
      }

      const mockMetrics = {
        data: [
          {
            node_id: 'node-1',
            latency_ms: 45,
            packet_loss_rate: 0,
            jitter_ms: 5,
            timestamp: '2024-01-01T12:00:00Z',
          },
        ],
      }

      mockFetchNode.mockResolvedValue(mockNode as any)
      mockFetchNodeStatus.mockResolvedValue(mockStatus as any)
      mockFetchMetrics.mockResolvedValue(mockMetrics as any)

      const { unmount } = renderHook(() => useNodeDetail('node-1', 5000))

      // Initial fetch
      expect(mockFetchNode).toHaveBeenCalledTimes(1)

      unmount()

      // Fast-forward time (should not trigger another fetch after unmount)
      act(() => {
        vi.advanceTimersByTime(5000)
      })

      // Should still be 1 (no additional fetch after unmount)
      expect(mockFetchNode).toHaveBeenCalledTimes(1)
    } finally {
      vi.useRealTimers()
    }
  })
})
