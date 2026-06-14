import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { AlertRecordDTO, AlertRecordStatus } from '@/api/alertRecords'
import type { NodeDTO } from '@/api/types'

interface AlertRecordDetailModalProps {
  record: AlertRecordDTO
  nodes: NodeDTO[]
  canEdit: boolean
  open: boolean
  onClose: () => void
  onStatusUpdate: (id: string, status: AlertRecordStatus) => Promise<void>
}

export function AlertRecordDetailModal({ record, nodes, canEdit, open, onClose, onStatusUpdate }: AlertRecordDetailModalProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [isUpdating, setIsUpdating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const node = nodes.find((n) => n.id === record.node_id)

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

  const handleStatusUpdate = async (newStatus: AlertRecordStatus) => {
    setIsUpdating(true)
    setError(null)
    try {
      await onStatusUpdate(record.id, newStatus)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : t('alerts.updateFailed', 'Failed to update status'))
    } finally {
      setIsUpdating(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose() }}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t('alerts.recordDetail', 'Alert Record Detail')}</DialogTitle>
        </DialogHeader>

        {error && (
          <div className="rounded-md bg-destructive/10 px-4 py-2 text-sm text-destructive">{error}</div>
        )}

        <div className="space-y-4">
          <div>
            <p className="text-sm font-medium text-muted-foreground">{t('alerts.alertId', 'Alert ID')}</p>
            <p className="text-sm font-mono">{record.id}</p>
          </div>

          <div>
            <p className="text-sm font-medium text-muted-foreground">{t('alerts.node', 'Node')}</p>
            <div className="flex items-center gap-2">
              <p className="text-sm">{node?.name || record.node_id}</p>
              <Button
                variant="link"
                size="sm"
                onClick={() => { onClose(); navigate(`/nodes/${record.node_id}`, { state: { breadcrumbLabel: node?.name } }) }}
              >
                {t('alerts.viewNodeDetail', 'View Node')}
              </Button>
            </div>
            {node && <p className="text-xs text-muted-foreground">IP: {node.ip}</p>}
          </div>

          <div>
            <p className="text-sm font-medium text-muted-foreground">{t('alerts.metric', 'Metric')}</p>
            <p className="text-sm">{getMetricDisplayName(record.metric)}</p>
          </div>

          <div className="flex gap-6">
            <div>
              <p className="text-sm font-medium text-muted-foreground">{t('alerts.level', 'Level')}</p>
              <Badge variant={levelVariant(record.level)}>{record.level}</Badge>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">{t('common.status')}</p>
              <Badge variant={statusVariant(record.status)}>{statusLabel(record.status)}</Badge>
            </div>
          </div>

          <div className="flex gap-6">
            <div>
              <p className="text-sm font-medium text-muted-foreground">{t('alerts.created', 'Created')}</p>
              <p className="text-sm">{new Date(record.created_at).toLocaleString()}</p>
            </div>
            {record.updated_at !== record.created_at && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">{t('alerts.updated', 'Updated')}</p>
                <p className="text-sm">{new Date(record.updated_at).toLocaleString()}</p>
              </div>
            )}
          </div>

          {canEdit && record.status !== 'resolved' && (
            <div className="pt-4 border-t">
              <p className="text-sm font-medium text-muted-foreground mb-2">{t('alerts.updateStatus', 'Update Status')}</p>
              <div className="flex gap-3">
                {record.status === 'pending' && (
                  <Button variant="secondary" onClick={() => handleStatusUpdate('in_progress')} disabled={isUpdating}>
                    {t('alerts.markInProgress', 'Mark In Progress')}
                  </Button>
                )}
                <Button onClick={() => handleStatusUpdate('resolved')} disabled={isUpdating}>
                  {t('alerts.markResolved', 'Mark Resolved')}
                </Button>
              </div>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>{t('common.close', 'Close')}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
