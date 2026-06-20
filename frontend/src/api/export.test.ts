import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { createExport, getExportStatus, downloadExport, listExports } from './export'
import { apiClient } from './client'
import type { ExportTask, CreateExportRequest } from '../types/export'

// Mock the API client
vi.mock('./client', () => ({
  apiClient: vi.fn(),
}))

// Mock global fetch for downloadExport
global.fetch = vi.fn()

describe('Export API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('createExport', () => {
    const mockExportTask: ExportTask = {
      id: 'export-123',
      user_id: 'user-1',
      node_ids: ['node-1', 'node-2'],
      start_time: '2024-01-01T00:00:00Z',
      end_time: '2024-01-07T23:59:59Z',
      metrics: ['latency', 'packet_loss_rate'],
      format: 'csv',
      status: 'pending',
      created_at: '2024-01-01T00:00:00Z',
    }

    it('creates export task with correct parameters', async () => {
      const request: CreateExportRequest = {
        node_ids: ['node-1', 'node-2'],
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-07T23:59:59Z',
        metrics: ['latency', 'packet_loss_rate'],
        format: 'csv',
      }

      vi.mocked(apiClient).mockResolvedValueOnce({
        data: mockExportTask,
        message: 'Export task created successfully',
        timestamp: '2024-01-01T00:00:00Z',
      })

      const response = await createExport(request)

      expect(response.data).toEqual(mockExportTask)
      expect(response.message).toBe('Export task created successfully')
      expect(apiClient).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/data/export?'),
        expect.objectContaining({
          method: 'POST',
        })
      )
    })

    it('defaults format to csv if not provided', async () => {
      const request: CreateExportRequest = {
        node_ids: ['node-1'],
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-02T00:00:00Z',
        metrics: ['latency'],
      }

      vi.mocked(apiClient).mockResolvedValueOnce({
        data: mockExportTask,
        message: 'Export task created successfully',
        timestamp: '2024-01-01T00:00:00Z',
      })

      await createExport(request)

      const callArgs = vi.mocked(apiClient).mock.calls[0]
      expect(callArgs[0]).toContain('format=csv')
    })

    it('handles multiple node IDs', async () => {
      const request: CreateExportRequest = {
        node_ids: ['node-1', 'node-2', 'node-3'],
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-02T00:00:00Z',
        metrics: ['latency'],
      }

      vi.mocked(apiClient).mockResolvedValueOnce({
        data: mockExportTask,
        message: 'Export task created successfully',
        timestamp: '2024-01-01T00:00:00Z',
      })

      await createExport(request)

      const callArgs = vi.mocked(apiClient).mock.calls[0]
      expect(callArgs[0]).toContain('node_ids=node-1')
      expect(callArgs[0]).toContain('node_ids=node-2')
      expect(callArgs[0]).toContain('node_ids=node-3')
    })

    it('handles multiple metrics', async () => {
      const request: CreateExportRequest = {
        node_ids: ['node-1'],
        start_time: '2024-01-01T00:00:00Z',
        end_time: '2024-01-02T00:00:00Z',
        metrics: ['latency', 'packet_loss_rate', 'jitter'],
      }

      vi.mocked(apiClient).mockResolvedValueOnce({
        data: mockExportTask,
        message: 'Export task created successfully',
        timestamp: '2024-01-01T00:00:00Z',
      })

      await createExport(request)

      const callArgs = vi.mocked(apiClient).mock.calls[0]
      expect(callArgs[0]).toContain('metrics=latency')
      expect(callArgs[0]).toContain('metrics=packet_loss_rate')
      expect(callArgs[0]).toContain('metrics=jitter')
    })
  })

  describe('getExportStatus', () => {
    const mockExportTask: ExportTask = {
      id: 'export-123',
      user_id: 'user-1',
      node_ids: ['node-1'],
      start_time: '2024-01-01T00:00:00Z',
      end_time: '2024-01-07T23:59:59Z',
      metrics: ['latency'],
      format: 'csv',
      status: 'completed',
      file_path: '/tmp/exports/export-123.csv',
      file_size: 1024000,
      record_count: 500,
      created_at: '2024-01-01T00:00:00Z',
      completed_at: '2024-01-01T00:05:00Z',
    }

    it('retrieves export task status', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: mockExportTask,
        message: 'Export status retrieved',
        timestamp: '2024-01-01T00:05:00Z',
      })

      const response = await getExportStatus('export-123')

      expect(response.data).toEqual(mockExportTask)
      expect(apiClient).toHaveBeenCalledWith('/api/v1/data/export/export-123')
    })

    it('returns failed status with error message', async () => {
      const failedTask = {
        ...mockExportTask,
        status: 'failed' as const,
        error: 'No data found for specified time range',
      }

      vi.mocked(apiClient).mockResolvedValueOnce({
        data: failedTask,
        message: 'Export status retrieved',
        timestamp: '2024-01-01T00:05:00Z',
      })

      const response = await getExportStatus('export-123')

      expect(response.data.status).toBe('failed')
      expect(response.data.error).toBe('No data found for specified time range')
    })
  })

  describe('listExports', () => {
    it('lists recent export tasks with default limit', async () => {
      vi.mocked(apiClient).mockResolvedValueOnce({
        data: [],
        message: 'Export tasks retrieved',
        timestamp: '2024-01-01T00:00:00Z',
      })

      await listExports()

      expect(apiClient).toHaveBeenCalledWith('/api/v1/data/export?limit=50')
    })
  })

  describe('downloadExport', () => {
    it('downloads export file as blob', async () => {
      const mockBlob = new Blob(['csv,data,here'], { type: 'text/csv' })

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        blob: async () => mockBlob,
      } as Response)

      const result = await downloadExport('export-123')

      expect(result).toBeInstanceOf(Blob)
      expect(result.type).toBe('text/csv')
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/data/export/export-123/download'),
        expect.objectContaining({
          method: 'GET',
          credentials: 'include',
        })
      )
    })

    it('throws error on download failure', async () => {
      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: false,
        status: 404,
        json: async () => ({ message: 'Export file not found' }),
      } as Response)

      await expect(downloadExport('export-123')).rejects.toThrow('Export file not found')
    })

    it('throws error with status text on download failure when no JSON', async () => {
      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: async () => {
          throw new Error('No JSON')
        },
      } as Response)

      await expect(downloadExport('export-123')).rejects.toThrow('Internal Server Error')
    })
  })
})
