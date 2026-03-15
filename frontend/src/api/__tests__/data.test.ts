import { describe, it, expect, vi, beforeEach } from 'vitest'
import { fetchMetrics, fetchHistory, exportData, getComparisonData } from '../data'
import { apiClient } from '../client'

vi.mock('../client', () => ({
  apiClient: vi.fn(),
}))

describe('Data API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('fetchMetrics', () => {
    it('fetches metrics for multiple nodes', async () => {
      const mockMetrics = [
        { node_id: 'node-1', latency_ms: 50, packet_loss_rate: 0 },
        { node_id: 'node-2', latency_ms: 80, packet_loss_rate: 1 },
      ]
      vi.mocked(apiClient).mockResolvedValueOnce({ data: mockMetrics })

      const result = await fetchMetrics(['node-1', 'node-2'])

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('node_id=node-1')
      expect(call).toContain('node_id=node-2')
      expect(result.data).toEqual(mockMetrics)
    })

    it('fetches metrics for single node', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchMetrics(['node-1'])

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('/api/v1/data/metrics')
      expect(call).toContain('node_id=node-1')
    })

    it('fetches all metrics when no node IDs provided', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchMetrics([])

      expect(apiClient).toHaveBeenCalledWith('/api/v1/data/metrics')
    })
  })

  describe('fetchHistory', () => {
    const baseQuery = {
      node_ids: ['node-1'],
      start_time: '2024-01-01T00:00:00Z',
      end_time: '2024-01-02T00:00:00Z',
      metrics: ['latency'],
    }

    it('fetches historical data', async () => {
      const mockHistory = [
        {
          node_id: 'node-1',
          metric: 'latency',
          data_points: [{ timestamp: '2024-01-01T00:00:00Z', value: 50 }],
        },
      ]
      vi.mocked(apiClient).mockResolvedValueOnce({ data: mockHistory })

      const result = await fetchHistory(baseQuery)

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('/api/v1/data/history')
      expect(call).toContain('node_id=node-1')
      expect(call).toContain('start_time=')
      expect(call).toContain('end_time=')
      expect(call).toContain('metric=latency')
      expect(result.data).toEqual(mockHistory)
    })

    it('includes aggregation parameter when provided', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchHistory({ ...baseQuery, aggregation: '5m' })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('aggregation=5m')
    })

    it('does not include aggregation when not provided', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchHistory(baseQuery)

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).not.toContain('aggregation=')
    })

    it('handles multiple node IDs and metrics', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchHistory({
        node_ids: ['node-1', 'node-2'],
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-02T00:00:00Z',
        metrics: ['latency', 'packet_loss_rate'],
      })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('node_id=node-1')
      expect(call).toContain('node_id=node-2')
      expect(call).toContain('metric=latency')
      expect(call).toContain('metric=packet_loss_rate')
    })
  })

  describe('exportData', () => {
    it('exports data as CSV', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: { download_url: '/export/file.csv' } })

      const result = await exportData({
        node_ids: ['node-1'],
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-07T23:59:59Z',
        format: 'csv',
      })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('/api/v1/data/export')
      expect(call).toContain('node_id=node-1')
      expect(call).toContain('format=csv')
      expect(result.data.download_url).toBe('/export/file.csv')
    })

    it('exports data with multiple nodes', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: { download_url: '/export/file.csv' } })

      await exportData({
        node_ids: ['node-1', 'node-2'],
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-07T23:59:59Z',
        format: 'excel',
      })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('node_id=node-1')
      expect(call).toContain('node_id=node-2')
      expect(call).toContain('format=excel')
    })
  })

  describe('getComparisonData', () => {
    it('fetches comparison data for nodes', async () => {
      const mockResponse = {
        data: {
          time_range: { start: '2024-01-01T00:00:00Z', end: '2024-01-02T00:00:00Z' },
          nodes: [],
          statistics: {},
        },
        message: 'Success',
        timestamp: '2024-01-01T00:00:00Z',
      }
      vi.mocked(apiClient).mockResolvedValueOnce(mockResponse)

      const result = await getComparisonData({
        node_ids: ['node-1', 'node-2'],
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-02T00:00:00Z',
        metrics: ['latency_ms', 'packet_loss_rate'],
      })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('/api/v1/data/comparison')
      expect(call).toContain('node_ids=node-1%2Cnode-2')
      expect(call).toContain('metrics=latency_ms%2Cpacket_loss_rate')
      expect(result.data).toEqual(mockResponse.data)
    })

    it('includes time range in comparison request', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: { time_range: {}, nodes: [], statistics: {} },
        message: '',
        timestamp: '',
      })

      await getComparisonData({
        node_ids: ['node-1'],
        start_time: '2024-06-01T00:00:00Z',
        end_time: '2024-06-02T00:00:00Z',
        metrics: ['latency_ms'],
      })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('start_time=')
      expect(call).toContain('end_time=')
    })
  })
})
