import { useTranslation } from 'react-i18next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { AlertRecordDTO } from '@/api/alertRecords'
import type { NodeDTO } from '@/api/types'

type SortField = 'timestamp' | 'level' | 'status' | null
type SortOrder = 'asc' | 'desc'

interface AlertRecordsTableProps {
  records: AlertRecordDTO[]
  nodes: NodeDTO[]
  onViewDetail: (record: AlertRecordDTO) => void
  page: number
  pageSize: number
  totalCount: number
  onPageChange: (page: number) => void
  sortField: SortField
  sortOrder: SortOrder
  onSort: (field: SortField) => void
}

export function AlertRecordsTable({
  records,
  nodes,
  onViewDetail,
  page,
  pageSize,
  totalCount,
  onPageChange,
  sortField,
  sortOrder,
  onSort,
}: AlertRecordsTableProps) {
  const { t } = useTranslation()

  const getNodeName = (nodeId: string) => nodes.find((n) => n.id === nodeId)?.name || nodeId

  const statusLabel = (status: string) => {
    switch (status) {
      case 'pending': return t('alerts.pending', 'Pending')
      case 'in_progress': return t('alerts.inProgress', 'In Progress')
      case 'resolved': return t('alerts.resolved', 'Resolved')
      default: return status
    }
  }

  const statusVariant = (status: string): 'destructive' | 'secondary' | 'default' => {
    if (status === 'pending') return 'destructive'
    if (status === 'in_progress') return 'secondary'
    return 'default'
  }

  const levelVariant = (level: string): 'destructive' | 'secondary' | 'outline' => {
    if (level === 'P0') return 'destructive'
    if (level === 'P1') return 'secondary'
    return 'outline'
  }

  const getMetricDisplayName = (metric: string) => {
    switch (metric) {
      case 'latency': return t('metrics.latency', 'Latency')
      case 'packet_loss_rate': return t('metrics.packetLoss', 'Packet Loss')
      case 'jitter': return t('metrics.jitter', 'Jitter')
      default: return metric
    }
  }

  const formatTimestamp = (timestamp: string) =>
    new Date(timestamp).toLocaleString(undefined, { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })

  const totalPages = Math.ceil(totalCount / pageSize)

  if (records.length === 0) {
    return (
      <div className="text-center py-12">
        <h3 className="text-sm font-medium">{t('alertHistory.noRecords')}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{t('alertHistory.noRecordsHint')}</p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border shadow-sm overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-border">
          <thead className="bg-muted/50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('alerts.node', 'Node')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('alerts.metric', 'Metric')}</th>
              <th
                className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted"
                onClick={() => onSort('level')}
              >
                <div className="flex items-center gap-1">
                  {t('alerts.level', 'Level')}
                  {sortField === 'level' && <span className="text-xs">{sortOrder === 'asc' ? '↑' : '↓'}</span>}
                </div>
              </th>
              <th
                className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted"
                onClick={() => onSort('status')}
              >
                <div className="flex items-center gap-1">
                  {t('common.status')}
                  {sortField === 'status' && <span className="text-xs">{sortOrder === 'asc' ? '↑' : '↓'}</span>}
                </div>
              </th>
              <th
                className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground cursor-pointer hover:bg-muted"
                onClick={() => onSort('timestamp')}
              >
                <div className="flex items-center gap-1">
                  {t('alerts.timestamp', 'Timestamp')}
                  {sortField === 'timestamp' && <span className="text-xs">{sortOrder === 'asc' ? '↑' : '↓'}</span>}
                </div>
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('common.actions')}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {records.map((record) => (
              <tr key={record.id} className="hover:bg-muted/50">
                <td className="px-6 py-4 whitespace-nowrap text-sm">{getNodeName(record.node_id)}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{getMetricDisplayName(record.metric)}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <Badge variant={levelVariant(record.level)}>{record.level}</Badge>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <Badge variant={statusVariant(record.status)}>{statusLabel(record.status)}</Badge>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">{formatTimestamp(record.created_at)}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  <Button variant="link" size="sm" onClick={() => onViewDetail(record)}>
                    {t('alerts.viewDetail', 'View Detail')}
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="border-t px-4 py-3 sm:px-6">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              {page * pageSize + 1} - {Math.min((page + 1) * pageSize, totalCount)} / {totalCount}
            </p>
            <div className="flex gap-1">
              <Button variant="outline" size="sm" onClick={() => onPageChange(page - 1)} disabled={page === 0}>
                {t('common.previous')}
              </Button>
              {[...Array(totalPages)].map((_, i) => (
                <Button key={i} variant={i === page ? 'default' : 'outline'} size="sm" onClick={() => onPageChange(i)}>
                  {i + 1}
                </Button>
              ))}
              <Button variant="outline" size="sm" onClick={() => onPageChange(page + 1)} disabled={page >= totalPages - 1}>
                {t('common.next')}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
