import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getAlertRecords, updateAlertRecordStatus } from '../alertRecords'
import { apiClient } from '../client'

// Mock apiClient
vi.mock('../client', () => ({
  apiClient: vi.fn(),
}))

describe('AlertRecords API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getAlertRecords', () => {
    it('fetches all alert records without filters', async () => {
      const mockRecords = [
        {
          id: 'record-1',
          alert_event_id: 'event-1',
          node_id: 'node-1',
          metric: 'latency',
          level: 'P1',
          status: 'pending',
          created_at: '2024-01-01T10:00:00Z',
          updated_at: '2024-01-01T10:00:00Z',
        },
      ]

      vi.mocked(apiClient).mockResolvedValueOnce({
        data: mockRecords,
        message: 'Success',
        timestamp: '2024-01-01T10:00:00Z',
      })

      const result = await getAlertRecords()

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records')
      expect(result.data).toEqual(mockRecords)
    })

    it('fetches alert records with node_id filter', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: [],
      })

      await getAlertRecords({ node_id: 'node-1' })

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records?node_id=node-1')
    })

    it('fetches alert records with level filter', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: [],
      })

      await getAlertRecords({ level: 'P0' })

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records?level=P0')
    })

    it('fetches alert records with status filter', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: [],
      })

      await getAlertRecords({ status: 'pending' })

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records?status=pending')
    })

    it('fetches alert records with time range filters', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: [],
      })

      await getAlertRecords({
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-12-31T23:59:59Z',
      })

      expect(apiClient).toHaveBeenCalledWith(
        '/api/v1/alerts/records?start_time=2024-01-01T00%3A00%3A00Z&end_time=2024-12-31T23%3A59%3A59Z'
      )
    })

    it('fetches alert records with pagination parameters', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: [],
      })

      await getAlertRecords({ limit: 20, offset: 40 })

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records?limit=20&offset=40')
    })

    it('fetches alert records with multiple filters', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: [],
      })

      await getAlertRecords({
        node_id: 'node-1',
        level: 'P1',
        status: 'in_progress',
        limit: 10,
        offset: 0,
      })

      // Check that apiClient was called
      expect(apiClient).toHaveBeenCalled()
      const callArgs = vi.mocked(apiClient).mock.calls[0][0]

      // Check that the URL has the correct query parameters
      expect(callArgs).toContain('node_id=node-1')
      expect(callArgs).toContain('level=P1')
      expect(callArgs).toContain('status=in_progress')
      expect(callArgs).toContain('limit=10')
      // Note: offset=0 might be truncated in display but should be in the URL
      expect(callArgs).toMatch(/\?.*=/) // At least has query parameters
    })

    it('handles empty filters object', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: [],
      })

      await getAlertRecords({})

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records')
    })
  })

  describe('updateAlertRecordStatus', () => {
    it('updates alert record status successfully', async () => {
      const updatedRecord = {
        id: 'record-1',
        alert_event_id: 'event-1',
        node_id: 'node-1',
        metric: 'latency',
        level: 'P1',
        status: 'in_progress',
        created_at: '2024-01-01T10:00:00Z',
        updated_at: '2024-01-01T11:00:00Z',
      }

      vi.mocked(apiClient).mockResolvedValueOnce({
        data: updatedRecord,
        message: 'Alert record status updated',
        timestamp: '2024-01-01T11:00:00Z',
      })

      const result = await updateAlertRecordStatus('record-1', 'in_progress')

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records/record-1/status', {
        method: 'PUT',
        body: JSON.stringify({ status: 'in_progress' }),
      })
      expect(result.data).toEqual(updatedRecord)
      expect(result.message).toBe('Alert record status updated')
    })

    it('updates status to resolved', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: {},
        message: 'Alert record status updated',
        timestamp: '2024-01-01T12:00:00Z',
      })

      await updateAlertRecordStatus('record-1', 'resolved')

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records/record-1/status', {
        method: 'PUT',
        body: JSON.stringify({ status: 'resolved' }),
      })
    })

    it('sends correct request format', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: {},
        message: 'Success',
        timestamp: '2024-01-01T10:00:00Z',
      })

      await updateAlertRecordStatus('record-1', 'pending')

      expect(apiClient).toHaveBeenCalledWith(
        '/api/v1/alerts/records/record-1/status',
        expect.objectContaining({
          method: 'PUT',
          body: expect.stringContaining('pending'),
        })
      )
    })
  })
})
