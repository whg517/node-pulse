import { useCallback, useEffect, useState, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/authStore'
import { getAlertRecords, updateAlertRecordStatus, isValidStatusTransition } from '@/api/alertRecords'
import { fetchNodes } from '@/api/nodes'
import { exportData } from '@/api/data'
import type { AlertRecordDTO, AlertRecordFilters, AlertRecordStatus } from '@/api/alertRecords'
import type { NodeDTO } from '@/api/types'
import { AlertRecordsTable } from '@/components/alerts/AlertRecordsTable'
import { AlertRecordsFilter } from '@/components/alerts/AlertRecordsFilter'
import { AlertRecordDetailModal } from '@/components/alerts/AlertRecordDetailModal'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'

type SortField = 'timestamp' | 'level' | 'status' | null
type SortOrder = 'asc' | 'desc'

export default function AlertRecordsPage() {
  const { t } = useTranslation()
  const { user } = useAuthStore()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [records, setRecords] = useState<AlertRecordDTO[]>([])
  const [nodes, setNodes] = useState<NodeDTO[]>([])
  const [allRecords, setAllRecords] = useState<AlertRecordDTO[]>([])

  const [page, setPage] = useState(0)
  const [pageSize] = useState(20)
  const [totalCount, setTotalCount] = useState(0)

  const [filters, setFilters] = useState<AlertRecordFilters>({})
  const [searchQuery, setSearchQuery] = useState('')

  const [sortField, setSortField] = useState<SortField>(null)
  const [sortOrder, setSortOrder] = useState<SortOrder>('desc')

  const [selectedRecord, setSelectedRecord] = useState<AlertRecordDTO | null>(null)

  const isMounted = useRef(true)
  const canEdit = user?.role === 'admin' || user?.role === 'operator'

  const loadNodes = useCallback(async () => {
    if (nodes.length > 0) return nodes
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
  }, [nodes])

  const loadData = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [recordsResponse] = await Promise.all([
        getAlertRecords({ ...filters, limit: 100, offset: page * pageSize }),
        loadNodes(),
      ])

      if (!isMounted.current) return

      const fetchedRecords = recordsResponse.data || []
      setAllRecords(fetchedRecords)

      let processedRecords = [...fetchedRecords]

      if (searchQuery) {
        const query = searchQuery.toLowerCase()
        processedRecords = processedRecords.filter((record) => {
          const node = nodes.find((n) => n.id === record.node_id)
          const nodeName = node?.name.toLowerCase() || ''
          const metricName = record.metric.toLowerCase()
          return nodeName.includes(query) || metricName.includes(query)
        })
      }

      if (sortField) {
        processedRecords.sort((a, b) => {
          let comparison = 0
          switch (sortField) {
            case 'timestamp':
              comparison = new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
              break
            case 'level': {
              const levelOrder = { P0: 3, P1: 2, P2: 1 }
              comparison = levelOrder[a.level] - levelOrder[b.level]
              break
            }
            case 'status': {
              const statusOrder = { pending: 1, in_progress: 2, resolved: 3 }
              comparison = statusOrder[a.status] - statusOrder[b.status]
              break
            }
          }
          return sortOrder === 'asc' ? comparison : -comparison
        })
      }

      const startIndex = page * pageSize
      const paginatedRecords = processedRecords.slice(startIndex, startIndex + pageSize)

      setRecords(paginatedRecords)
      setTotalCount(processedRecords.length)
    } catch (err) {
      if (isMounted.current) setError(err instanceof Error ? err.message : String(err))
    } finally {
      if (isMounted.current) setIsLoading(false)
    }
  }, [filters, loadNodes, nodes, page, pageSize, searchQuery, sortField, sortOrder])

  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && selectedRecord) setSelectedRecord(null)
      if ((e.key === 'r' || e.key === 'R') && !selectedRecord) void loadData()
    }
    window.addEventListener('keydown', handleKeyPress)
    return () => window.removeEventListener('keydown', handleKeyPress)
  }, [loadData, selectedRecord])

  useEffect(() => {
    void loadData()
    return () => { isMounted.current = false }
  }, [loadData])

  const handleFilterChange = (newFilters: AlertRecordFilters) => {
    setFilters(newFilters)
    setPage(0)
  }

  const handleResetFilters = () => {
    setFilters({})
    setSearchQuery('')
    setPage(0)
    setSortField(null)
    setSortOrder('desc')
  }

  const handleSort = (field: SortField) => {
    if (sortField === field) setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc')
    else { setSortField(field); setSortOrder('desc') }
    setPage(0)
  }

  const handleStatusUpdate = async (id: string, newStatus: AlertRecordStatus) => {
    const record = allRecords.find((r) => r.id === id)
    if (!record) return
    if (!isValidStatusTransition(record.status, newStatus)) return
    try {
      await updateAlertRecordStatus(id, newStatus)
      setSelectedRecord(null)
      await loadData()
    } catch {
      throw new Error('Failed to update status')
    }
  }

  const handleExportCSV = async () => {
    try {
      const endTime = new Date().toISOString()
      const startTime = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString()
      const nodeIds = filters.node_id ? [filters.node_id] : nodes.map((n) => n.id)
      const response = await exportData({ node_ids: nodeIds, start_time: startTime, end_time: endTime, format: 'csv' })
      window.open(response.data.download_url, '_blank')
    } catch (error) {
      console.error('Failed to export data:', error)
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('alerts.alertHistory')}
        subtitle={t('alerts.viewManageHistory')}
        actions={
          <Button variant="outline" onClick={handleExportCSV}>
            {t('common.export')} CSV
          </Button>
        }
      />

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
          <Button variant="link" size="sm" onClick={loadData}>{t('common.retry')}</Button>
        </div>
      )}

      <AlertRecordsFilter
        filters={filters}
        nodes={nodes}
        searchQuery={searchQuery}
        onFilterChange={handleFilterChange}
        onSearchChange={setSearchQuery}
        onReset={handleResetFilters}
      />

      {isLoading && !error && (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      )}

      {!isLoading && !error && (
        <AlertRecordsTable
          records={records}
          nodes={nodes}
          onViewDetail={setSelectedRecord}
          page={page}
          pageSize={pageSize}
          totalCount={totalCount}
          onPageChange={setPage}
          sortField={sortField}
          sortOrder={sortOrder}
          onSort={handleSort}
        />
      )}

      {selectedRecord && (
        <AlertRecordDetailModal
          record={selectedRecord}
          nodes={nodes}
          canEdit={canEdit}
          open={!!selectedRecord}
          onClose={() => setSelectedRecord(null)}
          onStatusUpdate={handleStatusUpdate}
        />
      )}
    </div>
  )
}
