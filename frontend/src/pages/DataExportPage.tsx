import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { useExportStore } from '../stores/exportStore'
import { fetchNodes } from '../api/nodes'
import { ExportForm, ExportStatusCard, ExportHistoryTable } from '../components/export'
import type { NodeDTO } from '../api/types'
import type { CreateExportRequest } from '../types/export'

export default function DataExportPage() {
  const navigate = useNavigate()
  const { user, logout: storeLogout, clearAuth } = useAuthStore()
  const {
    createExport,
    currentExports,
    exportHistory,
    downloadExport,
    fetchExportHistory,
    isLoading: exportLoading,
  } = useExportStore()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [nodes, setNodes] = useState<NodeDTO[]>([])

  useEffect(() => {
    loadNodes()
    fetchExportHistory()

    // Cleanup function - stop all polling when component unmounts
    return () => {
      useExportStore.getState().stopAllPolling()
    }
  }, [])

  const loadNodes = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await fetchNodes()
      setNodes(response.data)
    } catch (err) {
      setError(err as Error)
      console.error('Failed to load nodes:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleLogout = async () => {
    try {
      await storeLogout()
      clearAuth()
      navigate('/login')
    } catch (error) {
      console.error('Logout failed:', error)
    }
  }

  const handleExportSubmit = async (request: CreateExportRequest) => {
    try {
      await createExport(request)
    } catch (error) {
      console.error('Failed to create export:', error)
      throw error
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-bold text-gray-900">Node Pulse</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-700">
                Welcome, {user?.username || 'Guest'}
              </span>
              <button
                type="button"
                onClick={handleLogout}
                className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors duration-150"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="mb-4 text-sm">
          <ol className="flex items-center space-x-2">
            <li>
              <a
                href="/dashboard"
                className="text-blue-600 hover:text-blue-800"
                onClick={(e) => {
                  e.preventDefault()
                  navigate('/dashboard')
                }}
              >
                Dashboard
              </a>
            </li>
            <li className="text-gray-400">/</li>
            <li className="text-gray-700 font-medium">Export</li>
          </ol>
        </nav>

        {/* Page Header */}
        <div className="mb-8">
          <h2 className="text-3xl font-bold text-gray-900">Data Export</h2>
          <p className="mt-2 text-gray-600">
            Configure export parameters and download metric data as CSV files
          </p>
        </div>

        {/* Error State */}
        {error && (
          <div className="mb-6 bg-red-50 border-l-4 border-red-400 p-4 rounded-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-400"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                  aria-hidden="true"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">Failed to load data</h3>
                <p className="text-sm text-red-700">{error.message}</p>
              </div>
              <div className="ml-auto pl-3">
                <div className="-mx-1.5 -my-1.5">
                  <button
                    onClick={() => loadNodes()}
                    className="inline-flex bg-red-50 rounded-md p-1.5 text-red-500 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-red-50 focus:ring-red-600"
                  >
                    <svg
                      className="h-5 w-5"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                      />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Access Warning for Non-Admin Users */}
        {!isLoading && !error && user?.role !== 'admin' && (
          <div className="access-warning bg-yellow-50 border-l-4 border-yellow-400 p-4 rounded-md mb-6">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-yellow-400"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                  aria-hidden="true"
                >
                  <path
                    fillRule="evenodd"
                    d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-yellow-800">Admin Only</h3>
                <p className="text-sm text-yellow-700 mt-1">
                  Data export is restricted to administrators. Please contact your administrator if you need access to this feature.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Loading State */}
        {isLoading && !error && (
          <div className="flex justify-center items-center py-12" data-testid="loading-spinner">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          </div>
        )}

        {/* Content - Export Form and Current Exports (Admin Only) */}
        {!isLoading && !error && user?.role === 'admin' && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Export Form */}
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Create New Export</h3>
              <ExportForm
                nodes={nodes}
                onSubmit={handleExportSubmit}
                loading={exportLoading}
              />
            </div>

            {/* Current Exports */}
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-gray-900">Active Exports</h3>
              {currentExports.length === 0 ? (
                <div className="bg-white rounded-lg shadow-md p-6 text-center text-gray-500">
                  <p>No active exports</p>
                </div>
              ) : (
                currentExports.map((exportTask) => (
                  <ExportStatusCard
                    key={exportTask.id}
                    exportTask={exportTask}
                    onDownload={downloadExport}
                  />
                ))
              )}
            </div>
          </div>
        )}

        {/* Export History */}
        {!isLoading && !error && exportHistory.length > 0 && (
          <div className="mt-8 bg-white rounded-lg shadow-md p-6">
            <ExportHistoryTable
              exports={exportHistory}
              onDownload={downloadExport}
              onDelete={() => {
                /* Delete not implemented in MVP */
              }}
            />
          </div>
        )}
      </main>
    </div>
  )
}
