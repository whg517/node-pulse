import { useEffect, useState } from 'react'
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
import { Textarea } from '@/components/ui/textarea'
import { addAlertNote, getAlertTimeline } from '@/api/alertRecords'
import type { AlertRecordDTO, AlertRecordStatus, AlertTimelineItemDTO } from '@/api/alertRecords'
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
  const [isTimelineLoading, setIsTimelineLoading] = useState(false)
  const [timeline, setTimeline] = useState<AlertTimelineItemDTO[]>([])
  const [error, setError] = useState<string | null>(null)
  const [note, setNote] = useState('')
  const [isAddingNote, setIsAddingNote] = useState(false)

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

  const formatLocalTime = (value: string) => new Date(value).toLocaleString()
  const formatUTCTime = (value: string) => new Date(value).toLocaleString('en-US', {
    timeZone: 'UTC',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }) + ' UTC'

  const getTimelineTitle = (item: AlertTimelineItemDTO) => {
    if (item.type === 'created') return t('alerts.timelineCreated', 'Alert created')
    if (item.type === 'status_changed') return t('alerts.timelineStatusChanged', 'Status changed')
    if (item.type === 'note') return t('alerts.timelineNoteAdded', 'Note added')
    return item.title
  }

  useEffect(() => {
    if (!open || !record.id) return

    let cancelled = false
    setIsTimelineLoading(true)
    void getAlertTimeline(record.id)
      .then((response) => {
        if (!cancelled) setTimeline(response.data || [])
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : t('alerts.timelineLoadFailed', 'Failed to load alert timeline'))
      })
      .finally(() => {
        if (!cancelled) setIsTimelineLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [open, record.id, t])

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

  const refreshTimeline = (id: string) => {
    setIsTimelineLoading(true)
    void getAlertTimeline(id)
      .then((response) => setTimeline(response.data || []))
      .catch((err) => setError(err instanceof Error ? err.message : t('alerts.timelineLoadFailed', 'Failed to load alert timeline')))
      .finally(() => setIsTimelineLoading(false))
  }

  const handleAddNote = async () => {
    const trimmed = note.trim()
    if (!trimmed) return
    setIsAddingNote(true)
    setError(null)
    try {
      await addAlertNote(record.id, trimmed)
      setNote('')
      refreshTimeline(record.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('alerts.addNoteFailed', 'Failed to add note'))
    } finally {
      setIsAddingNote(false)
    }
  }

  const handleNoteKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault()
      void handleAddNote()
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
              <p className="text-sm">{formatLocalTime(record.created_at)}</p>
            </div>
            {record.updated_at !== record.created_at && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">{t('alerts.updated', 'Updated')}</p>
                <p className="text-sm">{formatLocalTime(record.updated_at)}</p>
              </div>
            )}
          </div>

          <div className="pt-4 border-t">
            <p className="text-sm font-medium text-muted-foreground mb-3">{t('alerts.timeline', 'Timeline')}</p>
            {isTimelineLoading && (
              <div className="text-sm text-muted-foreground">{t('common.loading', 'Loading...')}</div>
            )}
            {!isTimelineLoading && timeline.length === 0 && (
              <div className="text-sm text-muted-foreground">{t('alerts.noTimelineEvents', 'No timeline events yet')}</div>
            )}
            {!isTimelineLoading && timeline.length > 0 && (
              <div className="space-y-3">
                {timeline.map((item, index) => (
                  <div key={item.id} className="flex gap-3">
                    <div className="flex flex-col items-center pt-1">
                      <span className="h-2 w-2 rounded-full bg-primary" />
                      {index < timeline.length - 1 && <span className="mt-1 min-h-8 w-px flex-1 bg-border" />}
                    </div>
                    <div className="min-w-0 flex-1 pb-2">
                      <div className="flex flex-wrap items-center gap-2">
                        <p className="text-sm font-medium text-foreground">{getTimelineTitle(item)}</p>
                        {item.type === 'status_changed' && item.to_status && (
                          <Badge variant={statusVariant(item.to_status)}>{statusLabel(item.to_status)}</Badge>
                        )}
                      </div>
                      {item.type === 'status_changed' && (
                        <p className="mt-1 text-sm text-muted-foreground">
                          {statusLabel(item.from_status || 'pending')} -&gt; {statusLabel(item.to_status || item.status || 'pending')}
                          {item.user_name ? ` by ${item.user_name}` : ''}
                        </p>
                      )}
                      {item.type === 'note' && item.content && (
                        <p className="mt-1 whitespace-pre-wrap text-sm text-muted-foreground">
                          {item.content}
                        </p>
                      )}
                      <p className="mt-1 text-xs text-muted-foreground">
                        {formatLocalTime(item.created_at)} | {formatUTCTime(item.created_at)}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {canEdit && (
            <div className="pt-4 border-t">
              <label htmlFor="alert-note-input" className="block text-sm font-medium text-muted-foreground mb-2">
                {t('alerts.addNote', 'Add Note')}
              </label>
              <Textarea
                id="alert-note-input"
                value={note}
                onChange={(e) => setNote(e.target.value)}
                onKeyDown={handleNoteKeyDown}
                placeholder={t('alerts.addNotePlaceholder', 'Add investigation details (Ctrl/Cmd+Enter to submit)')}
                rows={3}
                className="resize-none"
                disabled={isAddingNote}
              />
              <div className="mt-2 flex items-center justify-between">
                <p className="text-xs text-muted-foreground">{t('alerts.addNoteHint', 'Press Ctrl/Cmd+Enter to submit')}</p>
                <Button size="sm" onClick={() => void handleAddNote()} disabled={!note.trim() || isAddingNote}>
                  {isAddingNote ? t('common.saving', 'Saving...') : t('alerts.submitNote', 'Submit Note')}
                </Button>
              </div>
            </div>
          )}

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
