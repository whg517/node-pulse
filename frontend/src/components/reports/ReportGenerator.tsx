/**
 * ReportGenerator Component
 *
 * Form component for generating reports with node selection, time range,
 * metrics selection, and export format options.
 */

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { NodeDTO } from '../../api/types'
import { fetchHistory, fetchLatestMTR, fetchMetrics } from '@/api/data'
import { HealthReportPDF, type HealthMetrics, type MTRHop, type RootCauseAnalysis, type TimelineEvent } from './HealthReportPDF'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'

export type ReportType = 'health' | 'performance' | 'comparison'
export type ExportFormat = 'csv' | 'pdf' | 'excel'
export type DateRange = '7d' | '30d' | 'custom'

export interface ReportConfig {
  type: ReportType
  nodeIds: string[]
  dateRange: DateRange
  customStartDate?: string
  customEndDate?: string
  metrics: ('latency' | 'packet_loss_rate' | 'jitter')[]
  format: ExportFormat
  includeCharts: boolean
  includeSummary: boolean
}

interface ReportGeneratorProps {
  nodes: NodeDTO[]
  onSubmit: (config: ReportConfig) => Promise<void>
  loading?: boolean
  /** Pre-select these node IDs when the component first mounts */
  defaultNodeIds?: string[]
}

export function ReportGenerator({ nodes, onSubmit, loading = false, defaultNodeIds }: ReportGeneratorProps) {
  const { t } = useTranslation()

  // Form state
  const [reportType, setReportType] = useState<ReportType>('health')
  const [selectedNodeIds, setSelectedNodeIds] = useState<string[]>(defaultNodeIds ?? [])
  const [dateRange, setDateRange] = useState<DateRange>('7d')
  const [customStartDate, setCustomStartDate] = useState('')
  const [customEndDate, setCustomEndDate] = useState('')
  const [selectedMetrics, setSelectedMetrics] = useState<('latency' | 'packet_loss_rate' | 'jitter')[]>(['latency'])
  const [format, setFormat] = useState<ExportFormat>('csv')
  const [includeCharts, setIncludeCharts] = useState(true)
  const [includeSummary, setIncludeSummary] = useState(true)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isPreparingPdf, setIsPreparingPdf] = useState(false)

  // PDF Preview state
  const [showPdfPreview, setShowPdfPreview] = useState(false)
  const [pdfReportData, setPdfReportData] = useState<{
    node: NodeDTO
    metrics: HealthMetrics
    mtrPath?: MTRHop[]
    rootCause?: RootCauseAnalysis
    timeline?: TimelineEvent[]
    reportPeriod: { start: string; end: string }
  } | null>(null)

  const metricOptions = [
    { key: 'latency' as const, label: t('metrics.latency') },
    { key: 'packet_loss_rate' as const, label: t('metrics.packetLoss') },
    { key: 'jitter' as const, label: t('metrics.jitter') },
  ]

  const reportTypeOptions = [
    { key: 'health' as const, label: t('reports.healthReport') },
    { key: 'performance' as const, label: t('reports.performanceReport') },
    { key: 'comparison' as const, label: t('reports.comparisonReport') },
  ]

  const toggleNode = (nodeId: string) => {
    setSelectedNodeIds((prev) =>
      prev.includes(nodeId) ? prev.filter((id) => id !== nodeId) : [...prev, nodeId]
    )
    if (errors.nodeIds) {
      setErrors((prev) => ({ ...prev, nodeIds: '' }))
    }
  }

  const selectAllNodes = () => {
    setSelectedNodeIds(nodes.map((n) => n.id))
    if (errors.nodeIds) {
      setErrors((prev) => ({ ...prev, nodeIds: '' }))
    }
  }

  const clearNodeSelection = () => {
    setSelectedNodeIds([])
  }

  const toggleMetric = (metric: 'latency' | 'packet_loss_rate' | 'jitter') => {
    setSelectedMetrics((prev) =>
      prev.includes(metric) ? prev.filter((m) => m !== metric) : [...prev, metric]
    )
    if (errors.metrics) {
      setErrors((prev) => ({ ...prev, metrics: '' }))
    }
  }

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {}

    if (selectedNodeIds.length === 0) {
      newErrors.nodeIds = t('reports.selectNodes')
    }

    if (dateRange === 'custom') {
      if (!customStartDate || !customEndDate) {
        newErrors.dateRange = t('reports.startDate') + ' / ' + t('reports.endDate')
      } else {
        const start = new Date(customStartDate)
        const end = new Date(customEndDate)
        if (start >= end) {
          newErrors.dateRange = t('errors.validationError')
        }
      }
    }

    if (selectedMetrics.length === 0) {
      newErrors.metrics = t('reports.selectMetrics')
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const fetchReportMTRPath = async (nodeId: string): Promise<MTRHop[]> => {
    const latestMTR = await fetchLatestMTR(nodeId)
    if (!latestMTR?.success) return []

    return latestMTR.hops.map((hop) => ({
      hop: hop.hopNumber,
      ip: hop.ip,
      location: hop.location || hop.hostname,
      avgLatency: hop.avgRTTMs,
      lossRate: hop.lossRate,
    }))
  }

  const getReportPeriod = () =>
    dateRange === 'custom'
      ? { start: customStartDate!, end: customEndDate! }
      : {
          start: new Date(Date.now() - (dateRange === '7d' ? 7 : 30) * 24 * 60 * 60 * 1000).toISOString(),
          end: new Date().toISOString(),
        }

  const getTrend = (current: number, baseline: number): HealthMetrics['latency']['trend'] => {
    if (baseline === 0) return current === 0 ? 'stable' : 'degraded'
    const change = (current - baseline) / baseline
    if (change > 0.05) return 'degraded'
    if (change < -0.05) return 'improved'
    return 'stable'
  }

  const fetchReportMetrics = async (
    nodeId: string,
    reportPeriod: { start: string; end: string }
  ): Promise<HealthMetrics> => {
    const aggregation = dateRange === '7d' || dateRange === '30d' ? '5m' : '1m'
    const [metricsResult, latencyResult, lossResult, jitterResult] = await Promise.all([
      fetchMetrics([nodeId]),
      fetchHistory({
        node_ids: [nodeId],
        start_time: reportPeriod.start,
        end_time: reportPeriod.end,
        metrics: ['latency'],
        aggregation,
      }),
      fetchHistory({
        node_ids: [nodeId],
        start_time: reportPeriod.start,
        end_time: reportPeriod.end,
        metrics: ['packet_loss_rate'],
        aggregation,
      }),
      fetchHistory({
        node_ids: [nodeId],
        start_time: reportPeriod.start,
        end_time: reportPeriod.end,
        metrics: ['jitter'],
        aggregation,
      }),
    ])

    const currentMetrics = metricsResult.data.find((metric) => metric.node_id === nodeId)
    const toPoints = (result: Awaited<ReturnType<typeof fetchHistory>>, metricName: string) =>
      result.data.find((series) => series.metric === metricName)?.data_points || []
    const average = (points: Array<{ value: number }>, fallback: number) =>
      points.length === 0 ? fallback : points.reduce((sum, point) => sum + point.value, 0) / points.length
    const latestValue = (points: Array<{ value: number }>, fallback = 0) =>
      currentMetrics ? fallback : points.at(-1)?.value ?? fallback

    const latencyData = toPoints(latencyResult, 'latency')
    const lossData = toPoints(lossResult, 'packet_loss_rate')
    const jitterData = toPoints(jitterResult, 'jitter')
    const currentLatency = currentMetrics?.latency_ms ?? latestValue(latencyData)
    const currentLoss = currentMetrics?.packet_loss_rate ?? latestValue(lossData)
    const currentJitter = currentMetrics?.jitter_ms ?? latestValue(jitterData)
    const totalProbes = Math.max(latencyData.length, lossData.length, jitterData.length)
    const failedProbes = lossData.filter((point) => point.value > 0).length

    return {
      latency: {
        current: currentLatency,
        baseline: average(latencyData, currentLatency),
        trend: getTrend(currentLatency, average(latencyData, currentLatency)),
        data: latencyData,
      },
      packetLoss: {
        current: currentLoss,
        baseline: average(lossData, currentLoss),
        trend: getTrend(currentLoss, average(lossData, currentLoss)),
        data: lossData,
      },
      jitter: {
        current: currentJitter,
        baseline: average(jitterData, currentJitter),
        trend: getTrend(currentJitter, average(jitterData, currentJitter)),
        data: jitterData,
      },
      uptime: totalProbes === 0 ? 100 : ((totalProbes - failedProbes) / totalProbes) * 100,
      totalProbes,
      failedProbes,
    }
  }

  const generateRootCauseAnalysis = (metrics: HealthMetrics): RootCauseAnalysis => {
    const degradedMetrics = [
      metrics.latency.trend === 'degraded' ? 'latency' : null,
      metrics.packetLoss.trend === 'degraded' ? 'packet loss' : null,
      metrics.jitter.trend === 'degraded' ? 'jitter' : null,
    ].filter(Boolean)

    if (degradedMetrics.length === 0) {
      return {
        probableCause: t('reports.noIssues'),
        confidence: 'high',
        impact: t('reports.healthyStatus'),
        recommendation: t('reports.healthyStatus'),
      }
    }

    return {
      probableCause: `Network congestion affecting ${degradedMetrics.join(' and ')}`,
      confidence: 'medium',
      impact: 'Moderate impact on application performance',
      recommendation: 'Monitor network path and consider routing optimization or bandwidth upgrade',
    }
  }

  const generateTimelineEvents = (metrics: HealthMetrics): TimelineEvent[] => {
    const events: TimelineEvent[] = []
    const addMetricEvent = (
      data: Array<{ timestamp: string; value: number }>,
      label: string,
      unit: string,
      warningThreshold: number,
      criticalThreshold: number
    ) => {
      if (data.length === 0) return

      const peak = data.reduce((max, point) => point.value > max.value ? point : max, data[0])
      if (peak.value < warningThreshold) return

      events.push({
        timestamp: peak.timestamp,
        event: `${label} threshold exceeded (${peak.value.toFixed(1)}${unit})`,
        severity: peak.value >= criticalThreshold ? 'critical' : 'warning',
      })
    }

    addMetricEvent(
      metrics.latency.data,
      'Latency',
      'ms',
      Math.max(metrics.latency.baseline * 1.5, 200),
      500
    )
    addMetricEvent(
      metrics.packetLoss.data,
      'Packet loss',
      '%',
      Math.max(metrics.packetLoss.baseline * 1.5, 1),
      5
    )
    addMetricEvent(
      metrics.jitter.data,
      'Jitter',
      'ms',
      Math.max(metrics.jitter.baseline * 1.5, 50),
      100
    )

    return events.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) {
      return
    }

    // Handle PDF format differently - show preview first
    if (format === 'pdf') {
      const selectedNode = nodes.find((n) => n.id === selectedNodeIds[0])
      if (selectedNode) {
        setIsPreparingPdf(true)
        const reportPeriod = getReportPeriod()

        try {
          const [reportMetrics, mtrPath] = await Promise.all([
            fetchReportMetrics(selectedNode.id, reportPeriod),
            fetchReportMTRPath(selectedNode.id),
          ])
          setPdfReportData({
            node: selectedNode,
            metrics: reportMetrics,
            mtrPath,
            rootCause: generateRootCauseAnalysis(reportMetrics),
            timeline: generateTimelineEvents(reportMetrics),
            reportPeriod,
          })
          setShowPdfPreview(true)
        } catch (err) {
          console.error('Failed to prepare PDF report:', err)
          setErrors((prev) => ({
            ...prev,
            submit: err instanceof Error ? err.message : t('errors.failedToLoad'),
          }))
        } finally {
          setIsPreparingPdf(false)
        }
      }
      return
    }

    await onSubmit({
      type: reportType,
      nodeIds: selectedNodeIds,
      dateRange,
      customStartDate: dateRange === 'custom' ? customStartDate : undefined,
      customEndDate: dateRange === 'custom' ? customEndDate : undefined,
      metrics: selectedMetrics,
      format,
      includeCharts,
      includeSummary,
    })
  }

  const handlePdfClose = () => {
    setShowPdfPreview(false)
    setPdfReportData(null)
  }

  if (showPdfPreview && pdfReportData) {
    return (
      <Dialog open={showPdfPreview} onOpenChange={(o) => { if (!o) handlePdfClose() }}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{t('reports.healthReportTitle')}</DialogTitle>
          </DialogHeader>
          <HealthReportPDF
            node={pdfReportData.node}
            metrics={pdfReportData.metrics}
            mtrPath={pdfReportData.mtrPath}
            rootCause={pdfReportData.rootCause}
            timeline={pdfReportData.timeline}
            reportPeriod={pdfReportData.reportPeriod}
            onClose={handlePdfClose}
          />
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {errors.submit && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {errors.submit}
        </div>
      )}

      {/* Report Type */}
      <div>
        <label className="block text-sm font-medium text-muted-foreground mb-2">
          {t('reports.reportType')} <span className="text-destructive">*</span>
        </label>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
          {reportTypeOptions.map((option) => (
            <button
              key={option.key}
              type="button"
              onClick={() => setReportType(option.key)}
              disabled={loading}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                reportType === option.key
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted-foreground hover:bg-accent/10'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* Node Selection */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <label className="block text-sm font-medium text-muted-foreground">
            {t('reports.selectNodes')} <span className="text-destructive">*</span>
          </label>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={selectAllNodes}
              disabled={loading}
              className="text-xs text-primary hover:text-primary dark:text-primary"
            >
              {t('reports.selectAll')}
            </button>
            <button
              type="button"
              onClick={clearNodeSelection}
              disabled={loading}
              className="text-xs text-muted-foreground hover:text-foreground"
            >
              {t('reports.clearSelection')}
            </button>
          </div>
        </div>
        <div className="text-xs text-muted-foreground mb-2">
          {t('reports.selectedNodes')}: {selectedNodeIds.length} / {nodes.length}
        </div>
        <div className="max-h-48 overflow-y-auto border border-input rounded-lg p-3 space-y-2 bg-card">
          {nodes.map((node) => (
            <label
              key={node.id}
              className="flex items-center space-x-2 cursor-pointer hover:bg-muted/50 dark:hover:bg-accent p-1 rounded"
            >
              <input
                type="checkbox"
                checked={selectedNodeIds.includes(node.id)}
                onChange={() => toggleNode(node.id)}
                disabled={loading}
                className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
              />
              <span className="text-sm text-foreground">
                {node.name}
              </span>
              <span className="text-xs text-muted-foreground">
                ({node.region})
              </span>
            </label>
          ))}
        </div>
        {errors.nodeIds && (
          <p className="mt-1 text-sm text-destructive">{errors.nodeIds}</p>
        )}
      </div>

      {/* Date Range */}
      <div>
        <label className="block text-sm font-medium text-muted-foreground mb-2">
          {t('reports.dateRange')} <span className="text-destructive">*</span>
        </label>
        <div className="flex flex-wrap gap-2 mb-2">
          {[
            { value: '7d' as const, label: t('reports.last7Days') },
            { value: '30d' as const, label: t('reports.last30Days') },
            { value: 'custom' as const, label: t('reports.customRange') },
          ].map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => setDateRange(option.value)}
              disabled={loading}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                dateRange === option.value
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted-foreground hover:bg-accent/10'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
        {dateRange === 'custom' && (
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-muted-foreground mb-1">
                {t('reports.startDate')}
              </label>
              <input
                type="date"
                value={customStartDate}
                onChange={(e) => setCustomStartDate(e.target.value)}
                max={customEndDate || new Date().toISOString().split('T')[0]}
                disabled={loading}
                className="w-full px-3 py-2 border border-input rounded-lg bg-card text-foreground focus:ring-2 focus:ring-primary focus:border-transparent"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-muted-foreground mb-1">
                {t('reports.endDate')}
              </label>
              <input
                type="date"
                value={customEndDate}
                onChange={(e) => setCustomEndDate(e.target.value)}
                min={customStartDate}
                max={new Date().toISOString().split('T')[0]}
                disabled={loading}
                className="w-full px-3 py-2 border border-input rounded-lg bg-card text-foreground focus:ring-2 focus:ring-primary focus:border-transparent"
              />
            </div>
          </div>
        )}
        {errors.dateRange && (
          <p className="mt-1 text-sm text-destructive">{errors.dateRange}</p>
        )}
      </div>

      {/* Metrics Selection */}
      <div>
        <label className="block text-sm font-medium text-muted-foreground mb-2">
          {t('reports.selectMetrics')} <span className="text-destructive">*</span>
        </label>
        <div className="flex flex-wrap gap-2">
          {metricOptions.map((option) => (
            <button
              key={option.key}
              type="button"
              onClick={() => toggleMetric(option.key)}
              disabled={loading}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                selectedMetrics.includes(option.key)
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted-foreground hover:bg-accent/10'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
        {errors.metrics && (
          <p className="mt-1 text-sm text-destructive">{errors.metrics}</p>
        )}
      </div>

      {/* Export Format */}
      <div>
        <label className="block text-sm font-medium text-muted-foreground mb-2">
          {t('reports.selectFormat')} <span className="text-destructive">*</span>
        </label>
        <div className="flex flex-wrap gap-2">
          {[
            { value: 'csv' as const, label: t('reports.csv') },
            { value: 'pdf' as const, label: t('reports.pdf') },
            { value: 'excel' as const, label: t('reports.excel') },
          ].map((option) => (
            <button
              key={option.value}
              type="button"
              onClick={() => setFormat(option.value)}
              disabled={loading || isPreparingPdf}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                format === option.value
                  ? 'bg-primary text-white'
                  : 'bg-muted text-muted-foreground hover:bg-accent/10'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* Report Options */}
      <div>
        <label className="block text-sm font-medium text-muted-foreground mb-2">
          {t('reports.reportOptions')}
        </label>
        <div className="space-y-2">
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="checkbox"
              checked={includeCharts}
              onChange={(e) => setIncludeCharts(e.target.checked)}
              disabled={loading}
              className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
            />
            <span className="text-sm text-muted-foreground">
              {t('reports.includeCharts')}
            </span>
          </label>
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="checkbox"
              checked={includeSummary}
              onChange={(e) => setIncludeSummary(e.target.checked)}
              disabled={loading}
              className="h-4 w-4 text-primary focus:ring-primary border-border rounded"
            />
            <span className="text-sm text-muted-foreground">
              {t('reports.includeSummary')}
            </span>
          </label>
        </div>
      </div>

      {/* Submit Button */}
      <div className="flex justify-end">
        <button
          type="submit"
          disabled={loading || isPreparingPdf}
          className="bg-primary hover:bg-primary/85 disabled:bg-primary/10 dark:disabled:bg-primary/10 text-white font-medium py-2 px-6 rounded-lg transition-colors duration-150"
        >
          {loading || isPreparingPdf ? t('reports.generating') : t('reports.generateReport')}
        </button>
      </div>
    </form>
  )
}
