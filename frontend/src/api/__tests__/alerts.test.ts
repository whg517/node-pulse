import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  fetchAlertRules,
  createAlertRule,
  updateAlertRule,
  deleteAlertRule,
  fetchAlertRecords,
} from '../alerts'
import { apiClient } from '../client'

vi.mock('../client', () => ({
  apiClient: vi.fn(),
}))

describe('Alerts API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('fetchAlertRules', () => {
    it('fetches all alert rules', async () => {
      const mockRules = [
        { id: 'rule-1', metric: 'latency', threshold: 100, level: 'P1', enabled: true },
      ]
      vi.mocked(apiClient).mockResolvedValueOnce({ data: { alerts: mockRules } })

      const result = await fetchAlertRules()

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/rules')
      expect(result.data.alerts).toEqual(mockRules)
    })

    it('returns empty alerts array when no rules exist', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: { alerts: [] } })

      const result = await fetchAlertRules()

      expect(result.data.alerts).toEqual([])
    })
  })

  describe('createAlertRule', () => {
    it('creates a new alert rule', async () => {
      const request = {
        metric: 'latency' as const,
        threshold: 100,
        level: 'P1' as const,
        node_id: null,
      }
      const mockCreated = { id: 'rule-new', ...request, enabled: true }
      vi.mocked(apiClient).mockResolvedValueOnce({ data: mockCreated })

      const result = await createAlertRule(request)

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/rules', {
        method: 'POST',
        body: JSON.stringify(request),
      })
      expect(result.data).toEqual(mockCreated)
    })

    it('creates alert rule with node_id', async () => {
      const request = {
        metric: 'packet_loss_rate' as const,
        threshold: 5,
        level: 'P0' as const,
        node_id: 'node-1',
      }
      vi.mocked(apiClient).mockResolvedValueOnce({ data: { id: 'rule-2', ...request, enabled: true } })

      await createAlertRule(request)

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/rules', {
        method: 'POST',
        body: JSON.stringify(request),
      })
    })
  })

  describe('updateAlertRule', () => {
    it('updates an existing alert rule', async () => {
      const id = 'rule-1'
      const request = { threshold: 150, enabled: false }
      const mockUpdated = { id, metric: 'latency', threshold: 150, level: 'P1', enabled: false }
      vi.mocked(apiClient).mockResolvedValueOnce({ data: mockUpdated })

      const result = await updateAlertRule(id, request)

      expect(apiClient).toHaveBeenCalledWith(`/api/v1/alerts/rules/${id}`, {
        method: 'PUT',
        body: JSON.stringify(request),
      })
      expect(result.data).toEqual(mockUpdated)
    })

    it('updates alert rule level', async () => {
      const id = 'rule-2'
      const request = { level: 'P2' as const }
      vi.mocked(apiClient).mockResolvedValueOnce({ data: { id, level: 'P2' } })

      await updateAlertRule(id, request)

      expect(apiClient).toHaveBeenCalledWith(`/api/v1/alerts/rules/${id}`, {
        method: 'PUT',
        body: JSON.stringify(request),
      })
    })
  })

  describe('deleteAlertRule', () => {
    it('deletes an alert rule', async () => {
      const id = 'rule-1'
      vi.mocked(apiClient).mockResolvedValueOnce({ message: 'Alert rule deleted' })

      const result = await deleteAlertRule(id)

      expect(apiClient).toHaveBeenCalledWith(`/api/v1/alerts/rules/${id}`, {
        method: 'DELETE',
      })
      expect(result.message).toBe('Alert rule deleted')
    })
  })

  describe('fetchAlertRecords', () => {
    it('fetches all alert records without filters', async () => {
      const mockRecords = [{ id: 'record-1', level: 'P1', status: 'pending' }]
      vi.mocked(apiClient).mockResolvedValueOnce({ data: mockRecords })

      const result = await fetchAlertRecords()

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records')
      expect(result.data).toEqual(mockRecords)
    })

    it('fetches alert records with empty filters object', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchAlertRecords({})

      expect(apiClient).toHaveBeenCalledWith('/api/v1/alerts/records')
    })

    it('fetches alert records with node_id filter', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchAlertRecords({ node_id: 'node-1' })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('node_id=node-1')
    })

    it('fetches alert records with level filter', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchAlertRecords({ level: 'P0' })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('level=P0')
    })

    it('fetches alert records with status filter', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchAlertRecords({ status: 'pending' })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('status=pending')
    })

    it('fetches alert records with time range filters', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchAlertRecords({
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-12-31T23:59:59Z',
      })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('start_time=')
      expect(call).toContain('end_time=')
    })

    it('fetches alert records with multiple filters', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({ data: [] })

      await fetchAlertRecords({ node_id: 'node-1', level: 'P1', status: 'in_progress' })

      const call = vi.mocked(apiClient).mock.calls[0][0] as string
      expect(call).toContain('node_id=node-1')
      expect(call).toContain('level=P1')
      expect(call).toContain('status=in_progress')
    })
  })
})
