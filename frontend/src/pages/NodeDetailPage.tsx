import { useParams, Link, useNavigate } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { useSetBreadcrumbLabel } from '@/components/layout/useBreadcrumb'
import { useNodeDetail } from '@/hooks/useNodeDetail'
import { useTimezone } from '@/hooks/useTimezone'
import MetricCard from '@/components/dashboard/MetricCard'
import ProblemDiagnosis, { type ProblemType, type ConfidenceLevel } from '@/components/dashboard/ProblemDiagnosis'
import TrendChart, { type TimeRange, type DataPoint } from '@/components/dashboard/TrendChart'
import { fetchHistory } from '@/api/data'
import MTRVisualization from '@/components/nodes/MTRVisualization'

export default function NodeDetailPage() {
  const { t } = useTranslation()
  const { formatTime } = useTimezone()
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { node, nodeStatus, metrics, isLoading, error, isPolling } = useNodeDetail(id || '')
  const { setDynamicLabel, clearDynamicLabels } = useSetBreadcrumbLabel()

  useEffect(() => {
    if (node && node.id === id) setDynamicLabel(0, node.name)
    return () => { clearDynamicLabels() }
  }, [node, id, setDynamicLabel, clearDynamicLabels])

  const [timeRange, setTimeRange] = useState<TimeRange>('24h')
  const [historyData, setHistoryData] = useState<{
    latency_ms: DataPoint[]; packet_loss_rate: DataPoint[]; jitter_ms: DataPoint[]
  }>({ latency_ms: [], packet_loss_rate: [], jitter_ms: [] })
  const [baselines, setBaselines] = useState<{ latency_ms?: number; packet_loss_rate?: number; jitter_ms?: number }>({})
  const [isLoadingHistory, setIsLoadingHistory] = useState(false)
  const [historyError, setHistoryError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    const fetchHistoricalData = async () => {
      setIsLoadingHistory(true)
      setHistoryError(null)
      try {
        const now = new Date()
        const hours = timeRange === '7d' ? 168 : timeRange === '30d' ? 720 : 24
        const startTime = new Date(now.getTime() - hours * 60 * 60 * 1000)
        const aggregation = timeRange === '7d' || timeRange === '30d' ? '5m' : '1m'

        const [latencyResult, lossResult, jitterResult] = await Promise.all([
          fetchHistory({ node_ids: [id], start_time: startTime.toISOString(), end_time: now.toISOString(), metrics: ['latency'], aggregation }),
          fetchHistory({ node_ids: [id], start_time: startTime.toISOString(), end_time: now.toISOString(), metrics: ['packet_loss_rate'], aggregation }),
          fetchHistory({ node_ids: [id], start_time: startTime.toISOString(), end_time: now.toISOString(), metrics: ['jitter'], aggregation }),
        ])

        const toDataPoints = (result: { data: { metric: string; data_points: { timestamp: string; value: number }[] }[] }, metricName: string) =>
          result.data.find((s) => s.metric === metricName)?.data_points.map((dp) => ({ timestamp: dp.timestamp, value: dp.value })) || []

        const latencyDP = toDataPoints(latencyResult, 'latency')
        const lossDP = toDataPoints(lossResult, 'packet_loss_rate')
        const jitterDP = toDataPoints(jitterResult, 'jitter')

        setHistoryData({ latency_ms: latencyDP, packet_loss_rate: lossDP, jitter_ms: jitterDP })

        if (timeRange === '7d' || timeRange === '30d') {
          const avg = (pts: DataPoint[]) => pts.length === 0 ? undefined : pts.reduce((s, dp) => s + dp.value, 0) / pts.length
          setBaselines({ latency_ms: avg(latencyDP), packet_loss_rate: avg(lossDP), jitter_ms: avg(jitterDP) })
        } else {
          setBaselines({})
        }
      } catch (err) {
        console.error('Failed to fetch historical data:', err)
        setHistoryError(t('errors.loadHistoricalError'))
      } finally {
        setIsLoadingHistory(false)
      }
    }
    fetchHistoricalData()
  }, [id, timeRange, t])

  const getProblemType = (): ProblemType => {
    if (!metrics || !nodeStatus) return 'none'
    const { packet_loss_rate, latency_ms } = metrics
    if (nodeStatus.status === 'offline' || packet_loss_rate > 50) return 'node_local'
    if (packet_loss_rate > 10 || latency_ms > 1000) return 'node_local'
    if (packet_loss_rate > 3 || latency_ms > 300 || metrics.jitter_ms > 100) return 'node_local'
    if (packet_loss_rate > 1 || latency_ms > 150 || metrics.jitter_ms > 50) return 'node_local'
    return 'none'
  }

  const getConfidence = (): ConfidenceLevel => {
    if (!metrics || !nodeStatus) return 'low'
    if (nodeStatus.status === 'offline' || metrics.packet_loss_rate > 10 || metrics.latency_ms > 500) return 'high'
    const score = Math.max(metrics.packet_loss_rate / 10, metrics.latency_ms / 200, metrics.jitter_ms / 50)
    if (score > 2) return 'medium'
    return 'low'
  }

  const formatTimestamp = (timestamp: string | undefined): string => {
    if (!timestamp) return 'N/A'
    try {
      const date = new Date(timestamp)
      const diffMins = Math.floor((Date.now() - date.getTime()) / 60000)
      if (diffMins < 1) return t('time.justNow')
      if (diffMins < 60) return t('time.minutesAgo', { count: diffMins })
      const diffHours = Math.floor(diffMins / 60)
      if (diffHours < 24) return t('time.hoursAgo', { count: diffHours })
      return formatTime(date)
    } catch { return 'N/A' }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="text-center">
          <div className="h-12 w-12 animate-spin rounded-full border-2 border-primary border-t-transparent mx-auto" />
          <p className="mt-4 text-muted-foreground">{t('common.loading')}</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <Card className="max-w-md">
        <CardContent className="p-6">
          <h2 className="text-xl font-semibold mb-2">{t('errors.failedToLoad')}</h2>
          <p className="text-muted-foreground mb-4">{error.message}</p>
          <Button asChild><Link to="/dashboard">{t('common.back')}</Link></Button>
        </CardContent>
      </Card>
    )
  }

  if (!node) {
    return (
      <Card className="max-w-md">
        <CardContent className="p-6">
          <h2 className="text-xl font-semibold mb-2">{t('errors.nodeNotFound')}</h2>
          <p className="text-muted-foreground mb-4">{t('errors.notFound')}</p>
          <Button asChild><Link to="/dashboard">{t('common.back')}</Link></Button>
        </CardContent>
      </Card>
    )
  }

  const statusColor = nodeStatus?.status === 'online' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    : nodeStatus?.status === 'connecting' ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
    : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'

  const dotColor = nodeStatus?.status === 'online' ? 'bg-green-500'
    : nodeStatus?.status === 'connecting' ? 'bg-yellow-500' : 'bg-red-500'

  return (
    <div className="space-y-6">
      <PageHeader
        title={node.name}
        subtitle={node.ip}
        actions={
          <div className="flex items-center gap-2">
            <div className={`flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium ${statusColor}`}>
              <span className={`w-2 h-2 rounded-full ${dotColor}`} />
              <span className="capitalize">{t(`status.${nodeStatus?.status || 'unknown'}`)}</span>
            </div>
            {isPolling && (
              <div className="flex items-center gap-1 text-sm text-muted-foreground">
                <span className="relative inline-flex rounded-full h-3 w-3 bg-primary" />
                <span>{t('nodes.live')}</span>
              </div>
            )}
            <Button size="sm" onClick={() => navigate(`/reports?nodeId=${id}`)}>{t('nodes.viewDiagnosticReport')}</Button>
          </div>
        }
      />

      <div className="space-y-6">
        {/* Node Info */}
        <Card>
          <CardContent className="p-6">
            <h2 className="text-lg font-semibold mb-4">{t('nodes.nodeInfo')}</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div><dt className="text-sm font-medium text-muted-foreground">{t('nodes.region')}</dt><dd className="mt-1 text-lg font-semibold">{node.region}</dd></div>
              <div><dt className="text-sm font-medium text-muted-foreground">{t('nodes.ipAddress')}</dt><dd className="mt-1 text-lg font-semibold font-mono">{node.ip}</dd></div>
              <div><dt className="text-sm font-medium text-muted-foreground">{t('nodes.lastHeartbeat')}</dt><dd className="mt-1 text-lg font-semibold">{formatTimestamp(nodeStatus?.last_heartbeat)}</dd></div>
              <div>
                <dt className="text-sm font-medium text-muted-foreground">{t('nodes.tags')}</dt>
                <dd className="mt-1 flex flex-wrap gap-2">
                  {node.tags && node.tags.length > 0
                    ? node.tags.map((tag, i) => <Badge key={i} variant="secondary">{tag}</Badge>)
                    : <span className="text-muted-foreground">{t('nodes.noTags')}</span>}
                </dd>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Metrics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <MetricCard
            title={t('metrics.latency')} value={metrics?.latency_ms ?? 'N/A'} unit="ms"
            status={metrics && metrics.latency_ms < 100 ? 'good' : metrics && metrics.latency_ms < 200 ? 'warning' : 'critical'}
            icon={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>}
          />
          <MetricCard
            title={t('metrics.packetLoss')} value={metrics?.packet_loss_rate ?? 'N/A'} unit="%"
            status={metrics && metrics.packet_loss_rate === 0 ? 'good' : metrics && metrics.packet_loss_rate < 2 ? 'warning' : 'critical'}
            icon={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>}
          />
          <MetricCard
            title={t('metrics.jitter')} value={metrics?.jitter_ms ?? 'N/A'} unit="ms"
            status={metrics && metrics.jitter_ms < 20 ? 'good' : metrics && metrics.jitter_ms < 50 ? 'warning' : 'critical'}
            icon={<svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" /></svg>}
          />
        </div>

        {/* Historical Trend Charts */}
        <div className="space-y-6">
          {historyError && (
            <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive flex items-center justify-between">
              <span>{historyError}</span>
              <Button variant="link" size="sm" onClick={() => { setHistoryError(null); window.location.reload() }}>{t('common.retry')}</Button>
            </div>
          )}

          <TrendChart data={historyData.latency_ms} metric="latency_ms" timeRange={timeRange}
            showBaseline={timeRange === '7d' || timeRange === '30d'} baselineValue={baselines.latency_ms}
            height="350px" onTimeRangeChange={setTimeRange} isLoading={isLoadingHistory} />
          <TrendChart data={historyData.packet_loss_rate} metric="packet_loss_rate" timeRange={timeRange}
            showBaseline={timeRange === '7d' || timeRange === '30d'} baselineValue={baselines.packet_loss_rate}
            height="350px" onTimeRangeChange={setTimeRange} isLoading={isLoadingHistory} />
          <TrendChart data={historyData.jitter_ms} metric="jitter_ms" timeRange={timeRange}
            showBaseline={timeRange === '7d' || timeRange === '30d'} baselineValue={baselines.jitter_ms}
            height="350px" onTimeRangeChange={setTimeRange} isLoading={isLoadingHistory} />
        </div>

        {/* Problem Diagnosis */}
        <Card>
          <CardContent className="p-6">
            <h2 className="text-lg font-semibold mb-4">{t('nodes.problemDiagnosis')}</h2>
            <ProblemDiagnosis
              problemType={getProblemType()} confidence={getConfidence()}
              details={metrics ? `${t('metrics.latency')}: ${metrics.latency_ms}ms, ${t('metrics.packetLoss')}: ${metrics.packet_loss_rate}%, ${t('metrics.jitter')}: ${metrics.jitter_ms}ms` : t('errors.notFound')}
              isExpanded={false}
            />
            <p className="mt-4 text-sm text-muted-foreground italic">{t('nodes.diagnosisNote')}</p>
          </CardContent>
        </Card>

        {/* MTR */}
        <Card>
          <CardContent className="p-6">
            <h2 className="text-lg font-semibold mb-4">{t('mtr.title')}</h2>
            <MTRVisualization />
            <p className="mt-3 text-sm text-muted-foreground">{t('nodes.mtrBackendNote')}</p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
