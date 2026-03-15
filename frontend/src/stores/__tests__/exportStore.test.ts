import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { useExportStore } from '../exportStore'
import * as exportApi from '../../api/export'
import type { ExportTask } from '../../types/export'

// Mock the export API
vi.mock('../../api/export', () => ({
  createExport: vi.fn(),
  getExportStatus: vi.fn(),
  downloadExport: vi.fn(),
}))

const mockExportTask: ExportTask = {
  id: 'export-1',
  user_id: 'user-1',
  node_ids: ['node-1'],
  start_time: '2024-01-01T00:00:00Z',
  end_time: '2024-01-07T23:59:59Z',
  metrics: ['latency'],
  format: 'csv',
  status: 'pending',
  created_at: '2024-01-07T00:00:00Z',
  updated_at: '2024-01-07T00:00:00Z',
}

const mockCompletedTask: ExportTask = {
  ...mockExportTask,
  status: 'completed',
  download_url: '/exports/file.csv',
}

describe('useExportStore', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    // Reset store state
    useExportStore.setState({
      currentExports: [],
      exportHistory: [],
      isLoading: false,
      error: null,
      pollingIntervals: new Map(),
      _activePolls: new Set(),
    })
    // Mock URL methods
    window.URL.createObjectURL = vi.fn(() => 'blob:url')
    window.URL.revokeObjectURL = vi.fn()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('has correct initial state', () => {
    const { result } = renderHook(() => useExportStore())
    expect(result.current.currentExports).toEqual([])
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  describe('createExport', () => {
    it('creates an export task and starts polling', async () => {
      vi.mocked(exportApi.createExport).mockResolvedValueOnce({ data: mockExportTask })
      vi.mocked(exportApi.getExportStatus).mockResolvedValue({ data: mockExportTask })

      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        await result.current.createExport({
          node_ids: ['node-1'],
          start_time: '2024-01-01T00:00:00Z',
          end_time: '2024-01-07T23:59:59Z',
          metrics: ['latency'],
          format: 'csv',
        })
      })

      expect(result.current.currentExports).toHaveLength(1)
      expect(result.current.currentExports[0].id).toBe('export-1')
      expect(result.current.isLoading).toBe(false)
    })

    it('sets error on failure', async () => {
      vi.mocked(exportApi.createExport).mockRejectedValueOnce(new Error('Network error'))

      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        await expect(result.current.createExport({
          node_ids: ['node-1'],
          start_time: '2024-01-01T00:00:00Z',
          end_time: '2024-01-07T23:59:59Z',
          metrics: ['latency'],
          format: 'csv',
        })).rejects.toThrow('Network error')
      })

      expect(result.current.error).toBe('Network error')
      expect(result.current.isLoading).toBe(false)
    })
  })

  describe('pollExportStatus', () => {
    it('moves completed task to history', async () => {
      vi.mocked(exportApi.getExportStatus).mockResolvedValueOnce({ data: mockCompletedTask })

      useExportStore.setState({ currentExports: [mockExportTask] })
      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        await result.current.pollExportStatus('export-1')
      })

      expect(result.current.currentExports).toHaveLength(0)
      expect(result.current.exportHistory).toHaveLength(1)
      expect(result.current.exportHistory[0].status).toBe('completed')
    })

    it('continues polling for pending task', async () => {
      vi.mocked(exportApi.getExportStatus).mockResolvedValueOnce({ data: mockExportTask })

      useExportStore.setState({ currentExports: [mockExportTask] })
      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        await result.current.pollExportStatus('export-1')
      })

      // Should still be in currentExports since status is 'pending'
      expect(result.current.currentExports).toHaveLength(1)
    })

    it('prevents duplicate polling for same export', async () => {
      vi.mocked(exportApi.getExportStatus).mockResolvedValue({ data: mockExportTask })

      useExportStore.setState({
        currentExports: [mockExportTask],
        _activePolls: new Set(['export-1']),
      })
      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        await result.current.pollExportStatus('export-1')
      })

      // Should not call API since already polling
      expect(exportApi.getExportStatus).not.toHaveBeenCalled()
    })

    it('sets error and stops polling on failure', async () => {
      vi.mocked(exportApi.getExportStatus).mockRejectedValueOnce(new Error('API error'))

      useExportStore.setState({ currentExports: [mockExportTask] })
      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        await result.current.pollExportStatus('export-1')
      })

      expect(result.current.error).toBe('API error')
    })

    it('moves failed task to history', async () => {
      const failedTask = { ...mockExportTask, status: 'failed' as const }
      vi.mocked(exportApi.getExportStatus).mockResolvedValueOnce({ data: failedTask })

      useExportStore.setState({ currentExports: [mockExportTask] })
      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        await result.current.pollExportStatus('export-1')
      })

      expect(result.current.currentExports).toHaveLength(0)
      expect(result.current.exportHistory[0].status).toBe('failed')
    })
  })

  describe('clearError', () => {
    it('clears the error state', () => {
      useExportStore.setState({ error: 'Some error' })
      const { result } = renderHook(() => useExportStore())

      act(() => {
        result.current.clearError()
      })

      expect(result.current.error).toBeNull()
    })
  })

  describe('fetchExportHistory', () => {
    it('clears error on fetch', async () => {
      useExportStore.setState({ error: 'Old error' })
      const { result } = renderHook(() => useExportStore())

      await act(async () => {
        await result.current.fetchExportHistory()
      })

      expect(result.current.error).toBeNull()
    })
  })

  describe('stopPolling', () => {
    it('stops polling for a specific export', () => {
      const mockInterval = setTimeout(() => {}, 10000)
      useExportStore.setState({
        pollingIntervals: new Map([['export-1', mockInterval]]),
        _activePolls: new Set(['export-1']),
      })
      const { result } = renderHook(() => useExportStore())

      act(() => {
        result.current.stopPolling('export-1')
      })

      expect(result.current.pollingIntervals.size).toBe(0)
      expect(result.current._activePolls.has('export-1')).toBe(false)
    })

    it('handles stopping non-existent poll gracefully', () => {
      const { result } = renderHook(() => useExportStore())

      expect(() => {
        act(() => {
          result.current.stopPolling('non-existent')
        })
      }).not.toThrow()
    })
  })

  describe('stopAllPolling', () => {
    it('stops all polling intervals', () => {
      const interval1 = setTimeout(() => {}, 10000)
      const interval2 = setTimeout(() => {}, 10000)
      useExportStore.setState({
        pollingIntervals: new Map([['export-1', interval1], ['export-2', interval2]]),
        _activePolls: new Set(['export-1', 'export-2']),
      })
      const { result } = renderHook(() => useExportStore())

      act(() => {
        result.current.stopAllPolling()
      })

      expect(result.current.pollingIntervals.size).toBe(0)
      expect(result.current._activePolls.size).toBe(0)
    })
  })

  describe('_loadHistoryFromStorage and _saveHistoryToStorage', () => {
    it('loads history from storage', () => {
      const { result } = renderHook(() => useExportStore())
      expect(() => {
        act(() => {
          result.current._loadHistoryFromStorage()
        })
      }).not.toThrow()
    })

    it('saves history to storage', () => {
      useExportStore.setState({ exportHistory: [mockCompletedTask] })
      const { result } = renderHook(() => useExportStore())
      expect(() => {
        act(() => {
          result.current._saveHistoryToStorage()
        })
      }).not.toThrow()
      expect(localStorage.setItem).toHaveBeenCalledWith(
        'export-history',
        expect.stringContaining('export-1')
      )
    })

    it('handles localStorage.getItem throwing error gracefully', () => {
      vi.spyOn(localStorage, 'getItem').mockImplementationOnce(() => {
        throw new Error('Storage error')
      })
      const { result } = renderHook(() => useExportStore())
      expect(() => {
        act(() => {
          result.current._loadHistoryFromStorage()
        })
      }).not.toThrow()
    })

    it('handles localStorage.setItem throwing error gracefully', () => {
      vi.spyOn(localStorage, 'setItem').mockImplementationOnce(() => {
        throw new Error('Storage full')
      })
      useExportStore.setState({ exportHistory: [mockCompletedTask] })
      const { result } = renderHook(() => useExportStore())
      expect(() => {
        act(() => {
          result.current._saveHistoryToStorage()
        })
      }).not.toThrow()
    })
  })

  describe('downloadExport', () => {
    it('downloads an export file', async () => {
      const mockBlob = new Blob(['csv data'], { type: 'text/csv' })
      vi.mocked(exportApi.downloadExport).mockResolvedValueOnce(mockBlob)

      const { result } = renderHook(() => useExportStore())

      // Mock DOM methods after renderHook to avoid breaking React's container
      const mockAnchor = {
        href: '',
        download: '',
        click: vi.fn(),
      }
      const createElementSpy = vi.spyOn(document, 'createElement').mockReturnValueOnce(mockAnchor as unknown as HTMLAnchorElement)
      const appendChildSpy = vi.spyOn(document.body, 'appendChild').mockImplementationOnce((el) => el)
      const removeChildSpy = vi.spyOn(document.body, 'removeChild').mockImplementationOnce((el) => el)

      await act(async () => {
        await result.current.downloadExport('export-1')
      })

      expect(mockAnchor.click).toHaveBeenCalled()
      expect(result.current.isLoading).toBe(false)

      createElementSpy.mockRestore()
      appendChildSpy.mockRestore()
      removeChildSpy.mockRestore()
    })

    it('sets error on download failure', async () => {
      vi.mocked(exportApi.downloadExport).mockRejectedValueOnce(new Error('Download failed'))

      const { result } = renderHook(() => useExportStore())

      let thrownError: Error | null = null
      await act(async () => {
        try {
          await result.current.downloadExport('export-1')
        } catch (e) {
          thrownError = e as Error
        }
      })

      expect(result.current.error).toBe('Download failed')
      expect(result.current.isLoading).toBe(false)
      expect(thrownError?.message).toBe('Download failed')
    })
  })
})
