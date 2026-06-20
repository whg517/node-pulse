import { describe, it, expect, vi, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useExportStore } from './exportStore'
import { createExport, getExportStatus, listExports } from '../api/export'
import type { ExportTask, CreateExportRequest } from '../types/export'

// Mock localStorage with proper storage interface for zustand persist
const localStorageMock = (() => {
  let store: Record<string, string> = {}

  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value.toString()
    },
    removeItem: (key: string) => {
      delete store[key]
    },
    clear: () => {
      store = {}
    },
    length: Object.keys(store).length,
    key: (index: number) => Object.keys(store)[index] || null,
  }
})()

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
})

// Mock the API functions
vi.mock('../api/export', () => ({
  createExport: vi.fn(),
  getExportStatus: vi.fn(),
  downloadExport: vi.fn(),
  listExports: vi.fn(),
}))

describe('useExportStore', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    vi.mocked(listExports).mockResolvedValue({ data: [], message: 'ok', timestamp: '2024-01-01T00:00:00Z' })
    // Clear localStorage before each test
    localStorage.clear()
  })

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

  describe('initial state', () => {
    it('has correct initial state', () => {
      const { result } = renderHook(() => useExportStore())

      expect(result.current.currentExports).toEqual([])
      expect(result.current.exportHistory).toEqual([])
      expect(result.current.isLoading).toBe(false)
      expect(result.current.error).toBe(null)
    })
  })

  describe('clearError', () => {
    it('clears error message', () => {
      const { result } = renderHook(() => useExportStore())

      act(() => {
        result.current.error = 'Some error'
      })

      act(() => {
        result.current.clearError()
      })

      expect(result.current.error).toBe(null)
    })
  })

  describe('stopPolling', () => {
    it('stops polling for specific export', () => {
      const { result } = renderHook(() => useExportStore())

      act(() => {
        const mockTimeout = setTimeout(() => {}, 1000)
        result.current.pollingIntervals = new Map([['export-123', mockTimeout]])
      })

      act(() => {
        result.current.stopPolling('export-123')
      })

      expect(result.current.pollingIntervals.has('export-123')).toBe(false)
    })
  })

  describe('stopAllPolling', () => {
    it('stops all polling intervals', () => {
      const { result } = renderHook(() => useExportStore())

      act(() => {
        const timeout1 = setTimeout(() => {}, 1000)
        const timeout2 = setTimeout(() => {}, 1000)
        result.current.pollingIntervals = new Map([
          ['export-1', timeout1],
          ['export-2', timeout2],
        ])
      })

      act(() => {
        result.current.stopAllPolling()
      })

      expect(result.current.pollingIntervals.size).toBe(0)
    })
  })

  describe('createExport', () => {
    it('creates export task and starts polling', async () => {
      vi.mocked(createExport).mockResolvedValueOnce({
        data: mockExportTask,
        message: 'Export task created successfully',
        timestamp: '2024-01-01T00:00:00Z',
      })

      // Mock getExportStatus since createExport triggers polling
      vi.mocked(getExportStatus).mockResolvedValue({
        data: mockExportTask,
        message: 'Export status retrieved',
        timestamp: '2024-01-01T00:00:00Z',
      })

      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        const request: CreateExportRequest = {
          node_ids: ['node-1', 'node-2'],
          start_time: '2024-01-01T00:00:00Z',
          end_time: '2024-01-07T23:59:59Z',
          metrics: ['latency'],
        }

        await result.current.createExport(request)
      })

      expect(result.current.currentExports).toHaveLength(1)
      expect(result.current.currentExports[0]).toEqual(mockExportTask)
      expect(result.current.isLoading).toBe(false)
      expect(result.current.error).toBe(null)
    })

    it('sets error on create export failure', async () => {
      vi.mocked(createExport).mockRejectedValueOnce(
        new Error('Failed to create export')
      )

      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        const request: CreateExportRequest = {
          node_ids: ['node-1'],
          start_time: '2024-01-01T00:00:00Z',
          end_time: '2024-01-02T00:00:00Z',
          metrics: ['latency'],
        }

        try {
          await result.current.createExport(request)
        } catch {
          // Expected to throw
        }
      })

      expect(result.current.error).toBe('Failed to create export')
      expect(result.current.isLoading).toBe(false)
    })
  })

  describe('pollExportStatus', () => {
    it('processes completed export without errors', async () => {
      const completedTask = {
        ...mockExportTask,
        status: 'completed' as const,
        file_path: '/tmp/export.csv',
        record_count: 100,
      }

      vi.mocked(getExportStatus).mockResolvedValueOnce({
        data: completedTask,
        message: 'Export completed',
        timestamp: '2024-01-01T00:01:00Z',
      })

      const { result } = renderHook(() => useExportStore())

      act(() => {
        useExportStore.setState({ currentExports: [mockExportTask] })
      })

      // Should not throw
      await act(async () => {
        await expect(result.current.pollExportStatus(mockExportTask.id)).resolves.toBeUndefined()
      })
    })

    it('processes failed export without errors', async () => {
      const failedTask = {
        ...mockExportTask,
        status: 'failed' as const,
        error: 'No data found',
      }

      vi.mocked(getExportStatus).mockResolvedValueOnce({
        data: failedTask,
        message: 'Export failed',
        timestamp: '2024-01-01T00:01:00Z',
      })

      const { result } = renderHook(() => useExportStore())

      act(() => {
        useExportStore.setState({ currentExports: [mockExportTask] })
      })

      // Should not throw
      await act(async () => {
        await expect(result.current.pollExportStatus(mockExportTask.id)).resolves.toBeUndefined()
      })
    })
  })

  describe('fetchExportHistory', () => {
    it('loads active and completed exports from backend', async () => {
      const completedTask = {
        ...mockExportTask,
        id: 'export-completed',
        status: 'completed' as const,
      }
      vi.mocked(listExports).mockResolvedValueOnce({
        data: [mockExportTask, completedTask],
        message: 'ok',
        timestamp: '2024-01-01T00:00:00Z',
      })
      vi.mocked(getExportStatus).mockResolvedValue({ data: mockExportTask, message: 'ok', timestamp: '2024-01-01T00:00:00Z' })

      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        await result.current.fetchExportHistory()
      })

      expect(result.current.currentExports).toHaveLength(1)
      expect(result.current.exportHistory).toHaveLength(1)
      expect(result.current.exportHistory[0].id).toBe('export-completed')
    })
  })
})
