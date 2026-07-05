/**
 * Export Store
 *
 * Zustand store for managing export tasks and history.
 * Handles export creation, status polling, and download.
 */

import { create } from 'zustand'
import { createExport, getExportStatus, downloadExport, listExports, deleteExport as deleteExportApi } from '../api/export'
import type { ExportTask, CreateExportRequest } from '../types/export'

interface ExportStore {
  // State
  currentExports: ExportTask[]
  exportHistory: ExportTask[]
  isLoading: boolean
  error: string | null
  pollingIntervals: Map<string, ReturnType<typeof setTimeout>>
  _activePolls: Set<string> // Track active polls to prevent race conditions

  // Actions
  createExport: (request: CreateExportRequest) => Promise<ExportTask>
  pollExportStatus: (exportId: string) => Promise<void>
  downloadExport: (exportId: string) => Promise<void>
  fetchExportHistory: () => Promise<void>
  deleteExport: (exportId: string) => Promise<void>
  clearError: () => void
  stopPolling: (exportId: string) => void
  stopAllPolling: () => void
  _loadHistoryFromStorage: () => void
  _saveHistoryToStorage: () => void
}

// Load history from localStorage on store creation
const loadHistoryFromStorage = (): ExportTask[] => {
  try {
    const stored = localStorage.getItem('export-history')
    return stored ? JSON.parse(stored) : []
  } catch {
    return []
  }
}

// Save history to localStorage
const saveHistoryToStorage = (history: ExportTask[]) => {
  try {
    localStorage.setItem('export-history', JSON.stringify(history))
  } catch (error) {
    console.error('Failed to save export history to localStorage:', error)
  }
}

export const useExportStore = create<ExportStore>()((set, get) => ({
  // Initial state
  currentExports: [],
  exportHistory: loadHistoryFromStorage(),
  isLoading: false,
  error: null,
  pollingIntervals: new Map<string, ReturnType<typeof setTimeout>>(),
  _activePolls: new Set<string>(), // Track active polls to prevent race conditions

      /**
       * Create a new export task
       */
      createExport: async (request: CreateExportRequest) => {
        set({ isLoading: true, error: null })

        try {
          const response = await createExport(request)
          const exportTask = response.data

          // Add to current exports
          set((state) => ({
            currentExports: [...state.currentExports, exportTask],
            isLoading: false,
          }))

          // Start polling for status updates
          get().pollExportStatus(exportTask.id)

          return exportTask
        } catch (error) {
          const errorMessage =
            error instanceof Error ? error.message : 'Failed to create export'
          set({ error: errorMessage, isLoading: false })
          throw error
        }
      },

      /**
       * Poll export task status
       * Auto-polls every 5 seconds for pending/processing tasks
       */
      pollExportStatus: async (exportId: string) => {
        const { _activePolls } = get()

        // Prevent race condition: skip if already polling this export
        if (_activePolls.has(exportId)) {
          return
        }

        // Mark as actively polling
        set((state) => ({
          _activePolls: new Set(state._activePolls).add(exportId),
        }))

        try {
          const response = await getExportStatus(exportId)
          const updatedTask = response.data

          // Update current exports
          set((state) => ({
            currentExports: state.currentExports.map((task) =>
              task.id === exportId ? updatedTask : task
            ),
          }))

          // If task is completed or failed, stop polling and move to history
          if (
            updatedTask.status === 'completed' ||
            updatedTask.status === 'failed'
          ) {
            get().stopPolling(exportId)

            // Move to history
            set((state) => {
              const newActivePolls = new Set(state._activePolls)
              newActivePolls.delete(exportId)
              const newHistory = [updatedTask, ...state.exportHistory]
              // Save to localStorage
              saveHistoryToStorage(newHistory)
              return {
                currentExports: state.currentExports.filter(
                  (task) => task.id !== exportId
                ),
                exportHistory: newHistory,
                _activePolls: newActivePolls,
              }
            })
          } else {
            // Continue polling every 5 seconds
            const intervalId = setTimeout(() => {
              get().pollExportStatus(exportId)
            }, 5000)

            // Store interval ID for cleanup
            set((state) => {
              const newIntervals = new Map(state.pollingIntervals)
              newIntervals.set(exportId, intervalId)
              return { pollingIntervals: newIntervals }
            })
          }
        } catch (error) {
          const errorMessage =
            error instanceof Error ? error.message : 'Failed to fetch export status'
          set({ error: errorMessage })

          // Stop polling on error and remove from active polls
          get().stopPolling(exportId)
        }
      },

      /**
       * Download exported file
       */
      downloadExport: async (exportId: string) => {
        set({ isLoading: true, error: null })

        try {
          const blob = await downloadExport(exportId)

          // Create download link
          const url = window.URL.createObjectURL(blob)
          const a = document.createElement('a')
          a.href = url
          a.download = `metrics_export_${exportId}.csv`
          document.body.appendChild(a)
          a.click()
          window.URL.revokeObjectURL(url)
          document.body.removeChild(a)

          set({ isLoading: false })
        } catch (error) {
          const errorMessage =
            error instanceof Error ? error.message : 'Failed to download export'
          set({ error: errorMessage, isLoading: false })
          throw error
        }
      },

      /**
       * Fetch export tasks from backend and split them into active/history lists.
       */
      fetchExportHistory: async () => {
        set({ error: null })

        try {
          const response = await listExports()
          const tasks = response.data ?? []
          const activeTasks = tasks.filter((task) => task.status === 'pending' || task.status === 'processing')
          const historyTasks = tasks.filter((task) => task.status === 'completed' || task.status === 'failed')

          set({
            currentExports: activeTasks,
            exportHistory: historyTasks,
          })
          saveHistoryToStorage(historyTasks)

          activeTasks.forEach((task) => {
            void get().pollExportStatus(task.id)
          })
        } catch (error) {
          const fallbackHistory = loadHistoryFromStorage()
          const errorMessage =
            error instanceof Error ? error.message : 'Failed to fetch export history'
          set({ exportHistory: fallbackHistory, error: errorMessage })
        }
      },

      /**
       * Delete an export task: stop its poll, remove from local state + storage,
       * then call the backend DELETE endpoint. Errors surface via `error` state.
       */
      deleteExport: async (exportId: string) => {
        set({ error: null })
        try {
          // Stop any active polling first to avoid races with the deletion.
          get().stopPolling(exportId)

          // Optimistically remove from local state.
          const { currentExports, exportHistory } = get()
          set({
            currentExports: currentExports.filter((t) => t.id !== exportId),
            exportHistory: exportHistory.filter((t) => t.id !== exportId),
          })
          saveHistoryToStorage(exportHistory.filter((t) => t.id !== exportId))

          await deleteExportApi(exportId)
        } catch (error) {
          const errorMessage =
            error instanceof Error ? error.message : 'Failed to delete export'
          set({ error: errorMessage })
          // Re-fetch to restore consistent state after a failed delete.
          void get().fetchExportHistory()
          throw error
        }
      },

      /**
       * Clear error message
       */
      clearError: () => {
        set({ error: null })
      },

      /**
       * Stop polling for a specific export
       */
      stopPolling: (exportId: string) => {
        const { pollingIntervals, _activePolls } = get()
        const intervalId = pollingIntervals.get(exportId)

        if (intervalId) {
          clearTimeout(intervalId)

          const newIntervals = new Map(pollingIntervals)
          newIntervals.delete(exportId)
          set({ pollingIntervals: newIntervals })
        }

        // Also remove from active polls
        const newActivePolls = new Set(_activePolls)
        newActivePolls.delete(exportId)
        set({ _activePolls: newActivePolls })
      },

      /**
       * Stop all polling intervals
       */
      stopAllPolling: () => {
        const { pollingIntervals } = get()

        pollingIntervals.forEach((intervalId) => {
          clearTimeout(intervalId)
        })

        set({
          pollingIntervals: new Map<string, ReturnType<typeof setTimeout>>(),
          _activePolls: new Set<string>(),
        })
      },

      /**
       * Load history from localStorage
       */
      _loadHistoryFromStorage: () => {
        const history = loadHistoryFromStorage()
        set({ exportHistory: history })
      },

      /**
       * Save history to localStorage
       */
      _saveHistoryToStorage: () => {
        const { exportHistory } = get()
        saveHistoryToStorage(exportHistory)
      },
    })
)
