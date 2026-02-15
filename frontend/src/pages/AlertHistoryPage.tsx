/**
 * Alert History Page
 *
 * Displays historical view of past alert triggers with filtering,
 * search, and status management capabilities.
 */

import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import {
  getAlertRecords,
  updateAlertRecordStatus,
  isValidStatusTransition,
  type AlertRecordDTO,
  type AlertRecordFilters,
  type AlertRecordStatus,
  type AlertLevel,
} from '../api/alertRecords'
import { fetchNodes } from '../api/nodes'

export default function AlertHistoryPage() {
  const navigate = useNavigate()
  const { user, logout: storeLogout, clearAuth } = useAuthStore()
  const [records, setRecords] = useState<AlertRecordDTO[]>([])
  const [nodes, setNodes] = useState<Array<{ id: string; name: string; ip: string }>>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [statusUpdateError, setStatusUpdateError] = useState<string | null>(null)

  // Filters
  const [filters, setFilters] = useState<AlertRecordFilters>({})
  const [tempFilters, setTempFilters] = useState<AlertRecordFilters>({})

  // Pagination
  const [page, setPage] = useState(1)
  const pageSize = 20


  // Check if user can edit (admin only)
  const canEdit = user?.role === 'admin'

  useEffect(() => {
    loadRecords()
    loadNodes()
  }, [filters])

  const loadRecords = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await getAlertRecords({
        ...filters,
        limit: pageSize,
        offset: (page - 1) * pageSize,
      })
      setRecords(response.data)
    } catch (err) {
      setError(err as Error)
      console.error('Failed to load alert records:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const loadNodes = async () => {
    try {
      const response = await fetchNodes()
      setNodes(response.data)
    } catch (err) {
      console.error('Failed to load nodes:', err)
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

  const handleFilterChange = (
    key: keyof AlertRecordFilters,
    value: string | number | undefined
  ) => {
    setTempFilters((prev) => ({ ...prev, [key]: value || undefined }))
  }

  const applyFilters = () => {
    setFilters(tempFilters)
    setPage(1)
  }

  const clearFilters = () => {
    setTempFilters({})
    setFilters({})
    setPage(1)
  }

  const handleStatusUpdate = async (id: string, currentStatus: AlertRecordStatus, newStatus: AlertRecordStatus) => {
    // Clear previous error
    setStatusUpdateError(null)

    // Validate status transition
    if (!isValidStatusTransition(currentStatus, newStatus)) {
      setStatusUpdateError(`Invalid status transition: ${currentStatus} -> ${newStatus}`)
      return
    }

    try {
      await updateAlertRecordStatus(id, newStatus)
      // Reload records
      await loadRecords()
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to update status'
      setStatusUpdateError(message)
      console.error('Failed to update status:', error)
    }
  }

  const getNodeName = (nodeId: string): string => {
    const node = nodes.find((n) => n.id === nodeId)
    return node?.name || nodeId
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
              >
                Dashboard
              </a>
            </li>
            <li className="text-gray-400">/</li>
            <li>
              <a
                href="/alerts/rules"
                className="text-blue-600 hover:text-blue-800"
              >
                Alerts
              </a>
            </li>
            <li className="text-gray-400">/</li>
            <li className="text-gray-700 font-medium">History</li>
          </ol>
        </nav>

        {/* Page Header */}
        <div className="mb-8">
          <h2 className="text-3xl font-bold text-gray-900">Alert History</h2>
          <p className="mt-2 text-gray-600">
            View and manage historical alert records. Filter by node, level, status, or time range.
          </p>
        </div>

        {/* Error State */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex items-center">
              <svg
                className="w-5 h-5 text-red-600 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <p className="text-red-800">{error.message}</p>
              <button
                onClick={() => loadRecords()}
                className="ml-auto px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700 transition-colors text-sm"
              >
                Retry
              </button>
            </div>
          </div>
        )}

        {/* Status Update Error */}
        {statusUpdateError && (
          <div className="mb-6 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
            <div className="flex items-center">
              <svg
                className="w-5 h-5 text-yellow-600 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                />
              </svg>
              <p className="text-yellow-800">{statusUpdateError}</p>
              <button
                onClick={() => setStatusUpdateError(null)}
                className="ml-auto px-3 py-1 bg-yellow-600 text-white rounded hover:bg-yellow-700 transition-colors text-sm"
              >
                Dismiss
              </button>
            </div>
          </div>
        )}

        {/* Filters */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Filters</h3>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Node Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Node
              </label>
              <select
                value={tempFilters.node_id || ''}
                onChange={(e) =>
                  handleFilterChange('node_id', e.target.value || undefined)
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">All Nodes</option>
                {nodes.map((node) => (
                  <option key={node.id} value={node.id}>
                    {node.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Level Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Level
              </label>
              <select
                value={tempFilters.level || ''}
                onChange={(e) =>
                  handleFilterChange(
                    'level',
                    (e.target.value as AlertLevel) || undefined
                  )
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">All Levels</option>
                <option value="P0">P0 - Critical</option>
                <option value="P1">P1 - Warning</option>
                <option value="P2">P2 - Info</option>
              </select>
            </div>

            {/* Status Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Status
              </label>
              <select
                value={tempFilters.status || ''}
                onChange={(e) =>
                  handleFilterChange(
                    'status',
                    (e.target.value as AlertRecordStatus) || undefined
                  )
                }
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">All Statuses</option>
                <option value="pending">Pending</option>
                <option value="in_progress">In Progress</option>
                <option value="resolved">Resolved</option>
              </select>
            </div>

            {/* Actions */}
            <div className="flex items-end space-x-2">
              <button
                type="button"
                onClick={applyFilters}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
              >
                Apply
              </button>
              <button
                type="button"
                onClick={clearFilters}
                className="px-4 py-2 bg-gray-200 text-gray-800 rounded-md hover:bg-gray-300 transition-colors"
              >
                Clear
              </button>
            </div>
          </div>
        </div>

        {/* Alert Records Table */}
        <div className="bg-white rounded-lg shadow-sm overflow-hidden">
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <div
                className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
                role="status"
                aria-label="Loading alert records"
              />
            </div>
          ) : records.length === 0 ? (
            <div className="text-center py-12">
              <svg
                className="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">No alerts found</h3>
              <p className="mt-1 text-sm text-gray-500">
                {Object.keys(filters).length > 0
                  ? 'Try adjusting your filters to see more results.'
                  : 'No alert records have been created yet.'}
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Time
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Node
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Metric
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Level
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    {canEdit && (
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    )}
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {records.map((record) => (
                    <tr key={record.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {formatDateTime(record.created_at)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {getNodeName(record.node_id)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {formatMetric(record.metric)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <LevelBadge level={record.level} />
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <StatusBadge status={record.status} />
                      </td>
                      {canEdit && (
                        <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                          <StatusActions
                            currentStatus={record.status}
                            onStatusChange={(newStatus) =>
                              handleStatusUpdate(record.id, record.status, newStatus)
                            }
                          />
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Pagination */}
        {!isLoading && records.length > 0 && (
          <div className="mt-6 flex items-center justify-between">
            <div className="text-sm text-gray-700">
              Showing page {page} of results
            </div>
            <div className="flex space-x-2">
              <button
                type="button"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-4 py-2 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:bg-gray-100 disabled:cursor-not-allowed"
              >
                Previous
              </button>
              <button
                type="button"
                onClick={() => setPage((p) => p + 1)}
                disabled={records.length < pageSize}
                className="px-4 py-2 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:bg-gray-100 disabled:cursor-not-allowed"
              >
                Next
              </button>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}

/**
 * Level badge component
 */
function LevelBadge({ level }: { level: AlertLevel }) {
  const levelConfig = {
    P0: {
      bgColor: 'bg-red-100',
      textColor: 'text-red-800',
      label: 'P0 - Critical',
    },
    P1: {
      bgColor: 'bg-yellow-100',
      textColor: 'text-yellow-800',
      label: 'P1 - Warning',
    },
    P2: {
      bgColor: 'bg-blue-100',
      textColor: 'text-blue-800',
      label: 'P2 - Info',
    },
  }

  const config = levelConfig[level]

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.bgColor} ${config.textColor}`}
    >
      {config.label}
    </span>
  )
}

/**
 * Status badge component
 */
function StatusBadge({ status }: { status: AlertRecordStatus }) {
  const statusConfig = {
    pending: {
      bgColor: 'bg-red-100',
      textColor: 'text-red-800',
      label: 'Pending',
    },
    in_progress: {
      bgColor: 'bg-yellow-100',
      textColor: 'text-yellow-800',
      label: 'In Progress',
    },
    resolved: {
      bgColor: 'bg-green-100',
      textColor: 'text-green-800',
      label: 'Resolved',
    },
  }

  const config = statusConfig[status]

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.bgColor} ${config.textColor}`}
    >
      {config.label}
    </span>
  )
}

/**
 * Status actions component
 */
function StatusActions({
  currentStatus,
  onStatusChange,
}: {
  currentStatus: AlertRecordStatus
  onStatusChange: (status: AlertRecordStatus) => void
}) {
  const [isUpdating, setIsUpdating] = useState(false)

  const handleAction = async (newStatus: AlertRecordStatus) => {
    setIsUpdating(true)
    try {
      await onStatusChange(newStatus)
    } finally {
      setIsUpdating(false)
    }
  }

  if (currentStatus === 'resolved') {
    return (
      <span className="text-sm text-gray-500">Completed</span>
    )
  }

  if (currentStatus === 'pending') {
    return (
      <div className="flex justify-end space-x-2">
        <button
          type="button"
          onClick={() => handleAction('in_progress')}
          disabled={isUpdating}
          className="text-blue-600 hover:text-blue-900 text-sm disabled:text-blue-300 disabled:cursor-not-allowed"
        >
          {isUpdating ? 'Starting...' : 'Start'}
        </button>
        <button
          type="button"
          onClick={() => handleAction('resolved')}
          disabled={isUpdating}
          className="text-green-600 hover:text-green-900 text-sm disabled:text-green-300 disabled:cursor-not-allowed"
        >
          {isUpdating ? 'Resolving...' : 'Resolve'}
        </button>
      </div>
    )
  }

  if (currentStatus === 'in_progress') {
    return (
      <button
        type="button"
        onClick={() => handleAction('resolved')}
        disabled={isUpdating}
        className="text-green-600 hover:text-green-900 text-sm disabled:text-green-300 disabled:cursor-not-allowed"
      >
        {isUpdating ? 'Resolving...' : 'Resolve'}
      </button>
    )
  }

  return null
}

/**
 * Format date/time
 */
function formatDateTime(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleString()
}

/**
 * Format metric name
 */
function formatMetric(metric: string): string {
  const metricMap: Record<string, string> = {
    latency: 'Latency',
    packet_loss_rate: 'Packet Loss',
    jitter: 'Jitter',
  }
  return metricMap[metric] || metric
}
