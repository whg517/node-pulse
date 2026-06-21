import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/authStore'
import {
  getAlertRecords,
  updateAlertRecordStatus,
  isValidStatusTransition,
  type AlertRecordDTO,
  type AlertRecordFilters,
  type AlertRecordStatus,
  type AlertLevel,
} from '@/api/alertRecords'
import { fetchNodes } from '@/api/nodes'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'

export default function AlertHistoryPage() {
  const { user } = useAuthStore()
  const { t } = useTranslation()
  const [records, setRecords] = useState<AlertRecordDTO[]>([])
  const [nodes, setNodes] = useState<Array<{ id: string; name: string; ip: string }>>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [statusUpdateError, setStatusUpdateError] = useState<string | null>(null)

  const [filters, setFilters] = useState<AlertRecordFilters>({})
  const [tempFilters, setTempFilters] = useState<AlertRecordFilters>({})
  const [page, setPage] = useState(1)
  const pageSize = 20

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
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [filters, page])

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

  const handleFilterChange = (key: keyof AlertRecordFilters, value: string | number | undefined) => {
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
    setStatusUpdateError(null)
    if (!isValidStatusTransition(currentStatus, newStatus)) {
      setStatusUpdateError(`Invalid status transition: ${currentStatus} -> ${newStatus}`)
      return
    }
    try {
      await updateAlertRecordStatus(id, newStatus)
      await loadRecords()
    } catch (err) {
      setStatusUpdateError(err instanceof Error ? err.message : 'Failed to update status')
    }
  }

  const getNodeName = (nodeId: string): string => {
    const node = nodes.find((n) => n.id === nodeId)
    return node?.name || nodeId
  }

  const levelVariant = (level: AlertLevel): 'destructive' | 'secondary' | 'outline' => {
    if (level === 'P0') return 'destructive'
    if (level === 'P1') return 'secondary'
    return 'outline'
  }

  const statusVariant = (status: AlertRecordStatus): 'destructive' | 'secondary' | 'default' => {
    if (status === 'pending') return 'destructive'
    if (status === 'in_progress') return 'secondary'
    return 'default'
  }

  return (
    <div className="space-y-6">
      <PageHeader title={t('alertHistory.title')} subtitle={t('alertHistory.description')} />

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
          <Button variant="link" size="sm" onClick={loadRecords}>{t('common.retry')}</Button>
        </div>
      )}

      {statusUpdateError && (
        <div className="rounded-md border-l-4 border-yellow-500 bg-yellow-50 p-4 dark:bg-yellow-950">
          <div className="flex items-center gap-2">
            <p className="text-sm font-medium">{statusUpdateError}</p>
            <Button variant="link" size="sm" onClick={() => setStatusUpdateError(null)}>
              {t('alertHistory.dismiss')}
            </Button>
          </div>
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardContent className="p-6">
          <h3 className="text-lg font-medium mb-4">{t('alertHistory.filters')}</h3>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="space-y-2">
              <Label>{t('alertHistory.node')}</Label>
              <select
                value={tempFilters.node_id || ''}
                onChange={(e) => handleFilterChange('node_id', e.target.value || undefined)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="">{t('alertHistory.allNodes')}</option>
                {nodes.map((node) => (
                  <option key={node.id} value={node.id}>{node.name}</option>
                ))}
              </select>
            </div>

            <div className="space-y-2">
              <Label>{t('alertHistory.level')}</Label>
              <select
                value={tempFilters.level || ''}
                onChange={(e) => handleFilterChange('level', (e.target.value as AlertLevel) || undefined)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="">{t('alertHistory.allLevels')}</option>
                <option value="P0">{t('alertHistory.p0Critical')}</option>
                <option value="P1">{t('alertHistory.p1Warning')}</option>
                <option value="P2">{t('alertHistory.p2Info')}</option>
              </select>
            </div>

            <div className="space-y-2">
              <Label>{t('common.status')}</Label>
              <select
                value={tempFilters.status || ''}
                onChange={(e) => handleFilterChange('status', (e.target.value as AlertRecordStatus) || undefined)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="">{t('alertHistory.allStatuses')}</option>
                <option value="pending">{t('alertHistory.pending')}</option>
                <option value="in_progress">{t('alertHistory.inProgress')}</option>
                <option value="resolved">{t('alertHistory.resolved')}</option>
              </select>
            </div>

            <div className="flex items-end gap-2">
              <Button onClick={applyFilters} className="flex-1">{t('alertHistory.apply')}</Button>
              <Button variant="outline" onClick={clearFilters}>{t('alertHistory.clear')}</Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card>
        {isLoading ? (
          <div className="flex justify-center py-12">
            <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          </div>
        ) : records.length === 0 ? (
          <CardContent className="py-12 text-center">
            <h3 className="text-sm font-medium">{t('alertHistory.noAlerts')}</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              {Object.keys(filters).length > 0 ? t('alertHistory.noAlertsFiltered') : t('alertHistory.noAlertsEmpty')}
            </p>
          </CardContent>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border">
              <thead className="bg-muted/50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('alertHistory.time')}</th>
                  <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('alertHistory.node')}</th>
                  <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('alertHistory.metric')}</th>
                  <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('alertHistory.level')}</th>
                  <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('common.status')}</th>
                  {canEdit && (
                    <th className="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('common.actions')}</th>
                  )}
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {records.map((record) => (
                  <tr key={record.id} className="hover:bg-muted/50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm">{new Date(record.created_at).toLocaleString()}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">{getNodeName(record.node_id)}</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">{formatMetric(record.metric)}</td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <Badge variant={levelVariant(record.level)}>{record.level}</Badge>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <Badge variant={statusVariant(record.status)}>{record.status}</Badge>
                    </td>
                    {canEdit && (
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                        <StatusActions
                          currentStatus={record.status}
                          onStatusChange={(newStatus) => handleStatusUpdate(record.id, record.status, newStatus)}
                        />
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Pagination */}
      {!isLoading && records.length > 0 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">{t('alertHistory.showingPage', { page })}</p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page === 1}>
              {t('common.previous')}
            </Button>
            <Button variant="outline" size="sm" onClick={() => setPage((p) => p + 1)} disabled={records.length < pageSize}>
              {t('common.next')}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

function StatusActions({ currentStatus, onStatusChange }: { currentStatus: AlertRecordStatus; onStatusChange: (status: AlertRecordStatus) => void }) {
  const [isUpdating, setIsUpdating] = useState(false)

  const handleAction = async (newStatus: AlertRecordStatus) => {
    setIsUpdating(true)
    try { await onStatusChange(newStatus) } finally { setIsUpdating(false) }
  }

  if (currentStatus === 'resolved') return <span className="text-sm text-muted-foreground">Completed</span>

  return (
    <div className="flex justify-end gap-2">
      {currentStatus === 'pending' && (
        <Button variant="link" size="sm" onClick={() => handleAction('in_progress')} disabled={isUpdating}>
          {isUpdating ? 'Starting...' : 'Start'}
        </Button>
      )}
      <Button variant="link" size="sm" onClick={() => handleAction('resolved')} disabled={isUpdating}>
        {isUpdating ? 'Resolving...' : 'Resolve'}
      </Button>
    </div>
  )
}

function formatMetric(metric: string): string {
  const map: Record<string, string> = { latency: 'Latency', packet_loss_rate: 'Packet Loss', jitter: 'Jitter' }
  return map[metric] || metric
}
