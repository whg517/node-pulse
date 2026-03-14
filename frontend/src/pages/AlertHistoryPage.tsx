/**
 * Alert History Page
 *
 * Displays historical view of past alert triggers with filtering,
 * search, and status management capabilities.
 */

import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
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
import { PageContainer, ErrorBanner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'

export default function AlertHistoryPage() {
  const { user } = useAuthStore()
  const { t } = useTranslation()
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

  const loadRecords = useCallback(async () => {
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
  }, [filters, page, pageSize])

  const loadNodes = useCallback(async () => {
    try {
      const response = await fetchNodes()
      setNodes(response.data.nodes)
    } catch (err) {
      console.error('Failed to load nodes:', err)
    }
  }, [])

  useEffect(() => {
    void loadRecords()
    void loadNodes()
  }, [loadRecords, loadNodes])

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
    <PageContainer>
      <PageHeader
        title={t('alertHistory.title')}
        subtitle={t('alertHistory.description')}
        showBreadcrumb
      />

      {/* Error State */}
      {error && (
        <ErrorBanner error={error} onRetry={loadRecords} />
      )}

      {/* Status Update Error */}
      {statusUpdateError && (
        <div className="mb-6 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
          <div className="flex items-center">
            <svg
              className="w-5 h-5 text-yellow-600 dark:text-yellow-400 mr-2"
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
            <p className="text-yellow-800 dark:text-yellow-300">{statusUpdateError}</p>
            <button
              onClick={() => setStatusUpdateError(null)}
              className="ml-auto px-3 py-1 bg-yellow-600 text-white rounded hover:bg-yellow-700 transition-colors text-sm"
            >
              {t('alertHistory.dismiss')}
            </button>
          </div>
        </div>
      )}

      {/* Filters */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 mb-6 border border-gray-200 dark:border-gray-700">
        <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">{t('alertHistory.filters')}</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {/* Node Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('alertHistory.node')}
            </label>
            <select
              value={tempFilters.node_id || ''}
              onChange={(e) =>
                handleFilterChange('node_id', e.target.value || undefined)
              }
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            >
              <option value="">{t('alertHistory.allNodes')}</option>
              {nodes.map((node) => (
                <option key={node.id} value={node.id}>
                  {node.name}
                </option>
              ))}
            </select>
          </div>

          {/* Level Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('alertHistory.level')}
            </label>
            <select
              value={tempFilters.level || ''}
              onChange={(e) =>
                handleFilterChange(
                  'level',
                  (e.target.value as AlertLevel) || undefined
                )
              }
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            >
              <option value="">{t('alertHistory.allLevels')}</option>
              <option value="P0">{t('alertHistory.p0Critical')}</option>
              <option value="P1">{t('alertHistory.p1Warning')}</option>
              <option value="P2">{t('alertHistory.p2Info')}</option>
            </select>
          </div>

          {/* Status Filter */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('common.status')}
            </label>
            <select
              value={tempFilters.status || ''}
              onChange={(e) =>
                handleFilterChange(
                  'status',
                  (e.target.value as AlertRecordStatus) || undefined
                )
              }
              className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100"
            >
              <option value="">{t('alertHistory.allStatuses')}</option>
              <option value="pending">{t('alertHistory.pending')}</option>
              <option value="in_progress">{t('alertHistory.inProgress')}</option>
              <option value="resolved">{t('alertHistory.resolved')}</option>
            </select>
          </div>

          {/* Actions */}
          <div className="flex items-end space-x-2">
            <button
              type="button"
              onClick={applyFilters}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
            >
              {t('alertHistory.apply')}
            </button>
            <button
              type="button"
              onClick={clearFilters}
              className="px-4 py-2 bg-gray-200 dark:bg-gray-600 text-gray-800 dark:text-gray-200 rounded-md hover:bg-gray-300 dark:hover:bg-gray-500 transition-colors"
            >
              {t('alertHistory.clear')}
            </button>
          </div>
        </div>
      </div>

      {/* Alert Records Table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden border border-gray-200 dark:border-gray-700">
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
            <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">{t('alertHistory.noAlerts')}</h3>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {Object.keys(filters).length > 0
                ? t('alertHistory.noAlertsFiltered')
                : t('alertHistory.noAlertsEmpty')}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-900/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    {t('alertHistory.time')}
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    {t('alertHistory.node')}
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    {t('alertHistory.metric')}
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    {t('alertHistory.level')}
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                    {t('common.status')}
                  </th>
                  {canEdit && (
                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                      {t('common.actions')}
                    </th>
                  )}
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {records.map((record) => (
                  <tr key={record.id} className="hover:bg-gray-50 dark:hover:bg-gray-700/50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                      {formatDateTime(record.created_at)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                      {getNodeName(record.node_id)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
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
          <div className="text-sm text-gray-700 dark:text-gray-300">
            {t('alertHistory.showingPage', { page })}
          </div>
          <div className="flex space-x-2">
            <button
              type="button"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-4 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:cursor-not-allowed text-gray-900 dark:text-gray-100"
            >
              {t('common.previous')}
            </button>
            <button
              type="button"
              onClick={() => setPage((p) => p + 1)}
              disabled={records.length < pageSize}
              className="px-4 py-2 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700 disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:cursor-not-allowed text-gray-900 dark:text-gray-100"
            >
              {t('common.next')}
            </button>
          </div>
        </div>
      )}
    </PageContainer>
  )
}

/**
 * Level badge component
 */
function LevelBadge({ level }: { level: AlertLevel }) {
  const levelConfig = {
    P0: {
      bgColor: 'bg-red-100 dark:bg-red-900/30',
      textColor: 'text-red-800 dark:text-red-300',
      label: 'P0 - Critical',
    },
    P1: {
      bgColor: 'bg-yellow-100 dark:bg-yellow-900/30',
      textColor: 'text-yellow-800 dark:text-yellow-300',
      label: 'P1 - Warning',
    },
    P2: {
      bgColor: 'bg-blue-100 dark:bg-blue-900/30',
      textColor: 'text-blue-800 dark:text-blue-300',
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
      bgColor: 'bg-red-100 dark:bg-red-900/30',
      textColor: 'text-red-800 dark:text-red-300',
      label: 'Pending',
    },
    in_progress: {
      bgColor: 'bg-yellow-100 dark:bg-yellow-900/30',
      textColor: 'text-yellow-800 dark:text-yellow-300',
      label: 'In Progress',
    },
    resolved: {
      bgColor: 'bg-green-100 dark:bg-green-900/30',
      textColor: 'text-green-800 dark:text-green-300',
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
      <span className="text-sm text-gray-500 dark:text-gray-400">Completed</span>
    )
  }

  if (currentStatus === 'pending') {
    return (
      <div className="flex justify-end space-x-2">
        <button
          type="button"
          onClick={() => handleAction('in_progress')}
          disabled={isUpdating}
          className="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 text-sm disabled:text-blue-300 dark:disabled:text-blue-600 disabled:cursor-not-allowed"
        >
          {isUpdating ? 'Starting...' : 'Start'}
        </button>
        <button
          type="button"
          onClick={() => handleAction('resolved')}
          disabled={isUpdating}
          className="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300 text-sm disabled:text-green-300 dark:disabled:text-green-600 disabled:cursor-not-allowed"
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
        className="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300 text-sm disabled:text-green-300 dark:disabled:text-green-600 disabled:cursor-not-allowed"
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
