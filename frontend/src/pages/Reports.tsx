import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import { useExportStore } from '@/stores/exportStore'
import { useSettingsStore, type ReportSchedule } from '@/stores/settingsStore'
import { fetchNodes } from '@/api/nodes'
import { fetchMetrics } from '@/api/data'
import { ReportGenerator, NodeComparisonTable, type ReportConfig, type NodeComparisonData } from '@/components/reports'
import { ExportStatusCard, ExportHistoryTable } from '@/components/export'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import type { NodeDTO } from '@/api/types'
import type { CreateExportRequest } from '@/types/export'

export default function ReportsPage() {
  const { t } = useTranslation()
  const [searchParams] = useSearchParams()
  const preselectedNodeId = searchParams.get('nodeId')
  const defaultNodeIds = preselectedNodeId ? [preselectedNodeId] : undefined

  const {
    createExport,
    currentExports,
    exportHistory,
    downloadExport,
    fetchExportHistory,
    isLoading: exportLoading,
  } = useExportStore()

  const { reportSchedules, addReportSchedule, updateReportSchedule, deleteReportSchedule } = useSettingsStore()

  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [nodes, setNodes] = useState<NodeDTO[]>([])
  const [comparisonData, setComparisonData] = useState<NodeComparisonData[]>([])

  const [showScheduleDialog, setShowScheduleDialog] = useState(false)
  const [scheduleForm, setScheduleForm] = useState({
    name: '',
    frequency: 'daily' as ReportSchedule['frequency'],
    time: '09:00',
    format: 'pdf' as ReportSchedule['format'],
    nodeIds: [] as string[],
    enabled: true,
  })

  const loadNodes = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await fetchNodes()
      const nodesArray = response.data.nodes || []
      setNodes(nodesArray)

      const slice = nodesArray.slice(0, 5)
      const metricsByNode = new Map<string, { latency_ms: number; packet_loss_rate: number; jitter_ms: number }>()
      if (slice.length > 0) {
        try {
          const mr = await fetchMetrics(slice.map((n) => n.id))
          for (const m of mr.data) {
            metricsByNode.set(m.node_id, { latency_ms: m.latency_ms, packet_loss_rate: m.packet_loss_rate, jitter_ms: m.jitter_ms })
          }
        } catch { /* table lists nodes without live metrics */ }
      }

      setComparisonData(
        slice.map((node) => {
          const m = metricsByNode.get(node.id)
          return {
            nodeId: node.id,
            nodeName: node.name,
            region: node.region,
            status: node.status as 'online' | 'offline' | 'connecting',
            latency: m?.latency_ms,
            packetLoss: m?.packet_loss_rate,
            jitter: m?.jitter_ms,
          }
        })
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadNodes()
    void fetchExportHistory()
    return () => { useExportStore.getState().stopAllPolling() }
  }, [fetchExportHistory, loadNodes])

  const handleReportSubmit = async (config: ReportConfig) => {
    if (config.format === 'pdf') return
      const exportRequest: CreateExportRequest = {
        node_ids: config.nodeIds,
        start_time: config.dateRange === 'custom' && config.customStartDate
          ? new Date(config.customStartDate).toISOString()
          : new Date(Date.now() - (config.dateRange === '7d' ? 7 : 30) * 24 * 60 * 60 * 1000).toISOString(),
        end_time: config.dateRange === 'custom' && config.customEndDate
          ? new Date(config.customEndDate).toISOString()
          : new Date().toISOString(),
        metrics: config.metrics,
        format: config.format as 'csv' | 'excel',
      }
      await createExport(exportRequest)
  }

  const handleCreateSchedule = () => {
    const schedule: ReportSchedule = {
      id: crypto.randomUUID(),
      name: scheduleForm.name,
      frequency: scheduleForm.frequency,
      time: scheduleForm.time,
      nodeIds: scheduleForm.nodeIds,
      format: scheduleForm.format,
      enabled: scheduleForm.enabled,
    }
    addReportSchedule(schedule)
    setShowScheduleDialog(false)
    setScheduleForm({ name: '', frequency: 'daily', time: '09:00', format: 'pdf', nodeIds: [], enabled: true })
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('reports.title')}
        subtitle={`${t('reports.generateReport')} - ${t('reports.exportHistory')}`}
      />

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
          <Button variant="link" size="sm" onClick={loadNodes}>{t('common.retry')}</Button>
        </div>
      )}

      {isLoading && !error && (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      )}

      {!isLoading && !error && (
        <div className="space-y-8">
          <Card>
            <CardContent className="p-6">
              <h3 className="text-lg font-semibold mb-4">{t('reports.generateReport')}</h3>
              <ReportGenerator nodes={nodes} onSubmit={handleReportSubmit} loading={exportLoading} defaultNodeIds={defaultNodeIds} />
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <h3 className="text-lg font-semibold mb-4">{t('nodes.comparison')}</h3>
              <NodeComparisonTable nodes={comparisonData} highlightDifferences={true} />
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold">{t('reports.scheduled')}</h3>
                <Button onClick={() => setShowScheduleDialog(true)} size="sm">{t('reports.createSchedule')}</Button>
              </div>
              {reportSchedules.length === 0 ? (
                <div className="py-8 text-center">
                  <p className="text-sm text-muted-foreground">{t('reports.noSchedules')}</p>
                  <p className="text-xs text-muted-foreground mt-1">{t('reports.noSchedulesHint')}</p>
                </div>
              ) : (
                <div className="divide-y">
                  {reportSchedules.map((schedule) => (
                    <div key={schedule.id} className="py-3 flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <Switch
                          checked={schedule.enabled}
                          onCheckedChange={(checked) => updateReportSchedule(schedule.id, { enabled: checked })}
                        />
                        <div>
                          <p className="text-sm font-medium">{schedule.name}</p>
                          <p className="text-xs text-muted-foreground">
                            {t(`reports.${schedule.frequency}`)} · {schedule.time} · {schedule.format.toUpperCase()}
                            {schedule.lastRun && ` · ${t('reports.lastRun')}: ${new Date(schedule.lastRun).toLocaleString()}`}
                          </p>
                        </div>
                      </div>
                      <Button variant="link" size="sm" className="text-destructive" onClick={() => deleteReportSchedule(schedule.id)}>
                        {t('common.delete')}
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {currentExports.length > 0 && (
            <Card>
              <CardContent className="p-6">
                <h3 className="text-lg font-semibold mb-4">{t('reports.download')}</h3>
                <div className="space-y-4">
                  {currentExports.map((exportTask) => (
                    <ExportStatusCard key={exportTask.id} exportTask={exportTask} onDownload={downloadExport} />
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {exportHistory.length > 0 && (
            <Card>
              <CardContent className="p-6">
                <h3 className="text-lg font-semibold mb-4">{t('reports.exportHistory')}</h3>
                <ExportHistoryTable exports={exportHistory} onDownload={downloadExport} onDelete={() => {}} />
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Schedule Creation Dialog */}
      <Dialog open={showScheduleDialog} onOpenChange={setShowScheduleDialog}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t('reports.createSchedule')}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="schedule-name">{t('reports.scheduleName')}</Label>
              <Input id="schedule-name" value={scheduleForm.name} onChange={(e) => setScheduleForm({ ...scheduleForm, name: e.target.value })} />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="schedule-frequency">{t('reports.frequency')}</Label>
                <select
                  id="schedule-frequency"
                  value={scheduleForm.frequency}
                  onChange={(e) => setScheduleForm({ ...scheduleForm, frequency: e.target.value as ReportSchedule['frequency'] })}
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="daily">{t('reports.daily')}</option>
                  <option value="weekly">{t('reports.weekly')}</option>
                  <option value="monthly">{t('reports.monthly')}</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="schedule-time">{t('reports.scheduleTime')}</Label>
                <Input id="schedule-time" type="time" value={scheduleForm.time} onChange={(e) => setScheduleForm({ ...scheduleForm, time: e.target.value })} />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="schedule-format">{t('reports.format')}</Label>
              <select
                id="schedule-format"
                value={scheduleForm.format}
                onChange={(e) => setScheduleForm({ ...scheduleForm, format: e.target.value as ReportSchedule['format'] })}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="pdf">PDF</option>
                <option value="csv">CSV</option>
              </select>
            </div>
            <div className="flex items-center gap-2">
              <Switch
                id="schedule-enabled"
                checked={scheduleForm.enabled}
                onCheckedChange={(checked) => setScheduleForm({ ...scheduleForm, enabled: checked })}
              />
              <Label htmlFor="schedule-enabled">{t('status.enabled')}</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowScheduleDialog(false)}>{t('common.cancel')}</Button>
            <Button onClick={handleCreateSchedule} disabled={!scheduleForm.name.trim()}>{t('common.create')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
