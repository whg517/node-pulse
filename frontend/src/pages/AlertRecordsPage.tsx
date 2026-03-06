import { useEffect, useState, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../stores/authStore'
import { useTheme } from '../hooks/useTheme'
import { getAlertRecords, updateAlertRecordStatus, isValidStatusTransition } from '../api/alertRecords'
import { fetchNodes } from '../api/nodes'
import { exportData } from '../api/data'
import type { AlertRecordDTO, AlertRecordFilters, AlertRecordStatus } from '../api/alertRecords'
import type { NodeDTO } from '../api/types'
import { AlertRecordsTable } from '../components/alerts/AlertRecordsTable'
import { AlertRecordsFilter } from '../components/alerts/AlertRecordsFilter'
import { AlertRecordDetailModal } from '../components/alerts/AlertRecordDetailModal'
import { PageContainer, ErrorBanner, LoadingSpinner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'

type ToastType = 'success' | 'error' | 'warning' | 'info'

interface Toast {
  id: string
  type: ToastType
  title: string
  message?: string
}

type SortField = 'timestamp' | 'level' | 'status' | null
type SortOrder = 'asc' | 'desc'

export default function AlertRecordsPage() {
  const { t } = useTranslation()
  const { isDark } = useTheme()
  const { user } = useAuthStore()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [records, setRecords] = useState<AlertRecordDTO[]>([])
  const [nodes, setNodes] = useState<NodeDTO[]>([])
  const [allRecords, setAllRecords] = useState<AlertRecordDTO[]>([]) // For client-side sorting/filtering

  // Pagination state
  const [page, setPage] = useState(0)
  const [pageSize] = useState(20)
  const [totalCount, setTotalCount] = useState(0)

  // Filter state
  const [filters, setFilters] = useState<AlertRecordFilters>({})
  const [searchQuery, setSearchQuery] = useState('')

  // Sorting state
  const [sortField, setSortField] = useState<SortField>(null)
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc')

  // Modal state
  const [selectedRecord, setSelectedRecord] = useState<AlertRecordDTO | null>(null)

  // Toast state
  const [toasts, setToasts] = useState<Toast[]>([])

  // Refs for cleanup
  const isMounted = useRef(true)

  // Check if user can edit (admin or operator)
  const canEdit = user?.role === 'admin' || user?.role === 'operator'

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      // Close modal with Escape
      if (e.key === 'Escape' && selectedRecord) {
        setSelectedRecord(null)
      }
      // Refresh with R key
      if (e.key === 'r' || e.key === 'R') {
        if (!selectedRecord) {
          loadData()
          showToast('success', 'Refreshed', 'Alert records have been refreshed')
        }
      }
    }

    window.addEventListener('keydown', handleKeyPress)
    return () => window.removeEventListener('keydown', handleKeyPress)
  }, [selectedRecord])

  useEffect(() => {
    loadData()
    return () => {
      isMounted.current = false
    }
  }, [filters, page, pageSize])

  const showToast = (type: ToastType, title: string, message?: string) => {
    const id = Date.now().toString()
    setToasts((prev) => [...prev, { id, type, title, message }])
  }

  const removeToast = (id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }

  const loadData = async () => {
    setIsLoading(true)
    setError(null)
    try {
      // Fetch alert records and nodes
      const [recordsResponse] = await Promise.all([
        getAlertRecords({ ...filters, limit: 1000, offset: 0 }), // Fetch all for client-side sorting
        loadNodes(),
      ])

      if (!isMounted.current) return

      const fetchedRecords = recordsResponse.data || []
      setAllRecords(fetchedRecords)

      // Apply sorting and search
      let processedRecords = [...fetchedRecords]

      // Apply search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase()
        processedRecords = processedRecords.filter((record) => {
          const node = nodes.find((n) => n.id === record.node_id)
          const nodeName = node?.name.toLowerCase() || ''
          const metricName = record.metric.toLowerCase()
          return nodeName.includes(query) || metricName.includes(query)
        })
      }

      // Apply sorting
      if (sortField) {
        processedRecords.sort((a, b) => {
          let comparison = 0
          switch (sortField) {
            case 'timestamp':
              comparison = new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
              break
            case 'level':
              const levelOrder = { P0: 3, P1: 2, P2: 1 }
              comparison = levelOrder[a.level] - levelOrder[b.level]
              break
            case 'status':
              const statusOrder = { pending: 1, in_progress: 2, resolved: 3 }
              comparison = statusOrder[a.status] - statusOrder[b.status]
              break
          }
          return sortOrder === 'asc' ? comparison : -comparison
        })
      }

      // Apply pagination
      const startIndex = page * pageSize
      const endIndex = startIndex + pageSize
      const paginatedRecords = processedRecords.slice(startIndex, endIndex)

      setRecords(paginatedRecords)
      setTotalCount(processedRecords.length)
    } catch (err) {
      if (isMounted.current) {
        setError(err as Error)
        console.error('Failed to load data:', err)
      }
    } finally {
      if (isMounted.current) {
        setIsLoading(false)
      }
    }
  }

  const loadNodes = async () => {
    if (nodes.length > 0) return nodes // Cache nodes

    try {
      const response = await fetchNodes()
      if (isMounted.current) {
        setNodes(response.data.nodes)
        return response.data.nodes
      }
      return []
    } catch (err) {
      console.error('Failed to load nodes:', err)
      throw err
    }
  }

  const handleFilterChange = (newFilters: AlertRecordFilters) => {
    setFilters(newFilters)
    setPage(0) // Reset to first page when filters change
  }

  const handleResetFilters = () => {
    setFilters({})
    setSearchQuery('')
    setPage(0)
    setSortField(null)
    setSortOrder('desc')
  }

  const handleSearchChange = (query: string) => {
    setSearchQuery(query)
    setPage(0)
  }

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortOrder('desc')
    }
    setPage(0)
  }

  const handlePageChange = (newPage: number) => {
    setPage(newPage)
  }

  const handleViewDetail = (record: AlertRecordDTO) => {
    setSelectedRecord(record)
  }

  const handleCloseModal = () => {
    setSelectedRecord(null)
  }

  const handleStatusUpdate = async (id: string, newStatus: AlertRecordStatus) => {
    // Find current record to validate transition
    const record = allRecords.find((r) => r.id === id)
    if (!record) {
      showToast('error', 'Error', 'Alert record not found')
      return
    }

    // Validate status transition
    if (!isValidStatusTransition(record.status, newStatus)) {
      showToast(
        'error',
        'Invalid Status Transition',
        `Cannot change status from "${record.status}" to "${newStatus}"`
      )
      return
    }

    try {
      await updateAlertRecordStatus(id, newStatus)
      showToast('success', 'Status Updated', `Alert record status changed to ${newStatus}`)
      setSelectedRecord(null)
      await loadData()
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to update status'
      showToast('error', 'Update Failed', errorMessage)
      console.error('Failed to update alert record status:', error)
      throw error
    }
  }

  const handleExportCSV = async () => {
    try {
      // Generate time range for export (last 7 days)
      const endTime = new Date().toISOString()
      const startTime = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString()

      // Get node IDs from current filters or all nodes
      const nodeIds = filters.node_id ? [filters.node_id] : nodes.map((n) => n.id)

      const response = await exportData({
        node_ids: nodeIds,
        start_time: startTime,
        end_time: endTime,
        format: 'csv',
      })

      // Open download URL in new tab
      window.open(response.data.download_url, '_blank')
      showToast('success', 'Export Started', 'Your CSV file is being prepared')
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to export data'
      showToast('error', 'Export Failed', errorMessage)
      console.error('Failed to export data:', error)
    }
  }

  return (
    <PageContainer>
      {/* Toast Notifications */}
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`fixed top-4 right-4 z-50 p-4 rounded-lg border shadow-lg transition-all duration-300 ${
            toast.type === 'success'
              ? 'bg-green-100 text-green-800 border-green-300 dark:bg-green-900/30 dark:text-green-300 dark:border-green-700'
              : toast.type === 'error'
              ? 'bg-red-100 text-red-800 border-red-300 dark:bg-red-900/30 dark:text-red-300 dark:border-red-700'
              : toast.type === 'warning'
              ? 'bg-yellow-100 text-yellow-800 border-yellow-300 dark:bg-yellow-900/30 dark:text-yellow-300 dark:border-yellow-700'
              : 'bg-blue-100 text-blue-800 border-blue-300 dark:bg-blue-900/30 dark:text-blue-300 dark:border-blue-700'
          }`}
          role="alert"
        >
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <h3 className="font-semibold text-sm">{toast.title}</h3>
              {toast.message && <p className="text-sm mt-1">{toast.message}</p>}
            </div>
            <button
              onClick={() => removeToast(toast.id)}
              className="ml-4 text-current opacity-60 hover:opacity-100"
              aria-label="Close notification"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      ))}

      <main className={isDark ? 'bg-slate-900' : 'bg-gray-50'}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <PageHeader
            title={t('alerts.alertHistory')}
            subtitle={t('alerts.viewManageHistory')}
            showBreadcrumb
            actions={
              <button
                type="button"
                onClick={handleExportCSV}
                className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md transition-colors"
              >
                {t('common.export')} CSV
              </button>
            }
          />

          {/* Error State */}
          {error && (
            <ErrorBanner
              error={error}
              onRetry={loadData}
              className="mb-6"
            />
          )}

          {/* Filters */}
          <AlertRecordsFilter
            filters={filters}
            nodes={nodes}
            searchQuery={searchQuery}
            onFilterChange={handleFilterChange}
            onSearchChange={handleSearchChange}
            onReset={handleResetFilters}
          />

          {/* Loading State */}
          {isLoading && !error && (
            <div className="py-12">
              <LoadingSpinner />
            </div>
          )}

          {/* Content */}
          {!isLoading && !error && (
            <AlertRecordsTable
              records={records}
              nodes={nodes}
              onViewDetail={handleViewDetail}
              page={page}
              pageSize={pageSize}
              totalCount={totalCount}
              onPageChange={handlePageChange}
              sortField={sortField}
              sortOrder={sortOrder}
              onSort={handleSort}
            />
          )}

          {/* Detail Modal */}
          {selectedRecord && (
            <AlertRecordDetailModal
              record={selectedRecord}
              nodes={nodes}
              canEdit={canEdit}
              onClose={handleCloseModal}
              onStatusUpdate={handleStatusUpdate}
            />
          )}
        </div>
      </main>
    </PageContainer>
  )
}
