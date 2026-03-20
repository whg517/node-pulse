/**
 * ReportGenerator Component
 *
 * Form component for generating reports with node selection, time range,
 * metrics selection, and export format options.
 */

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { NodeDTO } from '../../api/types'
import { HealthReportPDF, type HealthMetrics, type MTRHop, type RootCauseAnalysis, type TimelineEvent } from './HealthReportPDF'

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
}

export function ReportGenerator({ nodes, onSubmit, loading = false }: ReportGeneratorProps) {
  const { t } = useTranslation()

  // Form state
  const [reportType, setReportType] = useState<ReportType>('health')
  const [selectedNodeIds, setSelectedNodeIds] = useState<string[]>([])
  const [dateRange, setDateRange] = useState<DateRange>('7d')
  const [customStartDate, setCustomStartDate] = useState('')
  const [customEndDate, setCustomEndDate] = useState('')
  const [selectedMetrics, setSelectedMetrics] = useState<('latency' | 'packet_loss_rate' | 'jitter')[]>(['latency'])
  const [format, setFormat] = useState<ExportFormat>('csv')
  const [includeCharts, setIncludeCharts] = useState(true)
  const [includeSummary, setIncludeSummary] = useState(true)
  const [errors, setErrors] = useState<Record<string, string>>({})

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

  const generateSampleMTRPath = (): MTRHop[] => {
    return [
      { hop: 1, ip: '192.168.1.1', location: 'Local Gateway', avgLatency: 1.2, lossRate: 0 },
      { hop: 2, ip: '10.0.0.1', location: 'ISP Core', avgLatency: 5.4, lossRate: 0 },
      { hop: 3, ip: '203.0.113.1', location: 'Regional Hub', avgLatency: 12.8, lossRate: 0.1 },
      { hop: 4, ip: '198.51.100.1', location: 'International Gateway', avgLatency: 35.2, lossRate: 0.2 },
      { hop: 5, ip: '192.0.2.1', location: 'Destination', avgLatency: 45.2, lossRate: 0.5 },
    ]
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

  const generateTimelineEvents = (): TimelineEvent[] => {
    const events: TimelineEvent[] = []
    const now = Date.now()

    if (Math.random() > 0.5) {
      events.push({
        timestamp: new Date(now - 2 * 60 * 60 * 1000).toISOString(),
        event: 'Latency spike detected (85ms)',
        severity: 'warning',
      })
    }

    if (Math.random() > 0.7) {
      events.push({
        timestamp: new Date(now - 6 * 60 * 60 * 1000).toISOString(),
        event: 'Packet loss threshold exceeded (2.5%)',
        severity: 'critical',
      })
    }

    if (Math.random() > 0.6) {
      events.push({
        timestamp: new Date(now - 24 * 60 * 60 * 1000).toISOString(),
        event: 'Node reconnected after brief outage',
        severity: 'warning',
      })
    }

    events.push({
      timestamp: new Date(now - 48 * 60 * 60 * 1000).toISOString(),
      event: 'Scheduled maintenance completed',
      severity: 'info',
    })

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
        // Generate sample data for PDF report
        const reportMetrics: HealthMetrics = {
          latency: {
            current: 45.2,
            baseline: 42.1,
            trend: 'stable' as const,
            data: [],
          },
          packetLoss: {
            current: 0.5,
            baseline: 0.3,
            trend: 'degraded' as const,
            data: [],
          },
          jitter: {
            current: 12.3,
            baseline: 11.8,
            trend: 'stable' as const,
            data: [],
          },
          uptime: 99.5,
          totalProbes: 10080,
          failedProbes: 50,
        }

        const reportPeriod =
          dateRange === 'custom'
            ? { start: customStartDate!, end: customEndDate! }
            : {
                start: new Date(Date.now() - (dateRange === '7d' ? 7 : 30) * 24 * 60 * 60 * 1000).toISOString(),
                end: new Date().toISOString(),
              }

        setPdfReportData({
          node: selectedNode,
          metrics: reportMetrics,
          mtrPath: generateSampleMTRPath(),
          rootCause: generateRootCauseAnalysis(reportMetrics),
          timeline: generateTimelineEvents(),
          reportPeriod,
        })
        setShowPdfPreview(true)
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
      <div className="fixed inset-0 bg-black bg-opacity-50 z-50 overflow-y-auto p-4">
        <div className="min-h-full flex items-start justify-center py-8">
          <div className="bg-[var(--color-bg-surface)] rounded-lg shadow-xl max-w-4xl w-full">
            <HealthReportPDF
              node={pdfReportData.node}
              metrics={pdfReportData.metrics}
              mtrPath={pdfReportData.mtrPath}
              rootCause={pdfReportData.rootCause}
              timeline={pdfReportData.timeline}
              reportPeriod={pdfReportData.reportPeriod}
              onClose={handlePdfClose}
            />
          </div>
        </div>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Report Type */}
      <div>
        <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
          {t('reports.reportType')} <span className="text-[var(--color-critical)]">*</span>
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
                  ? 'bg-[var(--color-brand)] text-white'
                  : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]'
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
          <label className="block text-sm font-medium text-[var(--color-text-secondary)]">
            {t('reports.selectNodes')} <span className="text-[var(--color-critical)]">*</span>
          </label>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={selectAllNodes}
              disabled={loading}
              className="text-xs text-[var(--color-brand)] hover:text-[var(--color-brand-hover)] dark:text-[var(--color-brand)]"
            >
              {t('reports.selectAll')}
            </button>
            <button
              type="button"
              onClick={clearNodeSelection}
              disabled={loading}
              className="text-xs text-gray-600 hover:text-gray-800 dark:text-gray-400"
            >
              {t('reports.clearSelection')}
            </button>
          </div>
        </div>
        <div className="text-xs text-[var(--color-text-muted)] mb-2">
          {t('reports.selectedNodes')}: {selectedNodeIds.length} / {nodes.length}
        </div>
        <div className="max-h-48 overflow-y-auto border border-[var(--color-input-border)] rounded-lg p-3 space-y-2 bg-[var(--color-bg-surface)]">
          {nodes.map((node) => (
            <label
              key={node.id}
              className="flex items-center space-x-2 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700 p-1 rounded"
            >
              <input
                type="checkbox"
                checked={selectedNodeIds.includes(node.id)}
                onChange={() => toggleNode(node.id)}
                disabled={loading}
                className="h-4 w-4 text-[var(--color-brand)] focus:ring-[var(--color-brand)] border-gray-300 rounded"
              />
              <span className="text-sm text-[var(--color-text-primary)]">
                {node.name}
              </span>
              <span className="text-xs text-[var(--color-text-muted)]">
                ({node.region})
              </span>
            </label>
          ))}
        </div>
        {errors.nodeIds && (
          <p className="mt-1 text-sm text-[var(--color-critical)]">{errors.nodeIds}</p>
        )}
      </div>

      {/* Date Range */}
      <div>
        <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
          {t('reports.dateRange')} <span className="text-[var(--color-critical)]">*</span>
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
                  ? 'bg-[var(--color-brand)] text-white'
                  : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
        {dateRange === 'custom' && (
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                {t('reports.startDate')}
              </label>
              <input
                type="date"
                value={customStartDate}
                onChange={(e) => setCustomStartDate(e.target.value)}
                max={customEndDate || new Date().toISOString().split('T')[0]}
                disabled={loading}
                className="w-full px-3 py-2 border border-[var(--color-input-border)] rounded-lg bg-[var(--color-bg-surface)] text-[var(--color-text-primary)] focus:ring-2 focus:ring-[var(--color-brand)] focus:border-transparent"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                {t('reports.endDate')}
              </label>
              <input
                type="date"
                value={customEndDate}
                onChange={(e) => setCustomEndDate(e.target.value)}
                min={customStartDate}
                max={new Date().toISOString().split('T')[0]}
                disabled={loading}
                className="w-full px-3 py-2 border border-[var(--color-input-border)] rounded-lg bg-[var(--color-bg-surface)] text-[var(--color-text-primary)] focus:ring-2 focus:ring-[var(--color-brand)] focus:border-transparent"
              />
            </div>
          </div>
        )}
        {errors.dateRange && (
          <p className="mt-1 text-sm text-[var(--color-critical)]">{errors.dateRange}</p>
        )}
      </div>

      {/* Metrics Selection */}
      <div>
        <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
          {t('reports.selectMetrics')} <span className="text-[var(--color-critical)]">*</span>
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
                  ? 'bg-[var(--color-brand)] text-white'
                  : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
        {errors.metrics && (
          <p className="mt-1 text-sm text-[var(--color-critical)]">{errors.metrics}</p>
        )}
      </div>

      {/* Export Format */}
      <div>
        <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
          {t('reports.selectFormat')} <span className="text-[var(--color-critical)]">*</span>
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
              disabled={loading}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                format === option.value
                  ? 'bg-[var(--color-brand)] text-white'
                  : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]'
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* Report Options */}
      <div>
        <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
          {t('reports.reportOptions')}
        </label>
        <div className="space-y-2">
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="checkbox"
              checked={includeCharts}
              onChange={(e) => setIncludeCharts(e.target.checked)}
              disabled={loading}
              className="h-4 w-4 text-[var(--color-brand)] focus:ring-[var(--color-brand)] border-gray-300 rounded"
            />
            <span className="text-sm text-[var(--color-text-secondary)]">
              {t('reports.includeCharts')}
            </span>
          </label>
          <label className="flex items-center space-x-2 cursor-pointer">
            <input
              type="checkbox"
              checked={includeSummary}
              onChange={(e) => setIncludeSummary(e.target.checked)}
              disabled={loading}
              className="h-4 w-4 text-[var(--color-brand)] focus:ring-[var(--color-brand)] border-gray-300 rounded"
            />
            <span className="text-sm text-[var(--color-text-secondary)]">
              {t('reports.includeSummary')}
            </span>
          </label>
        </div>
      </div>

      {/* Submit Button */}
      <div className="flex justify-end">
        <button
          type="submit"
          disabled={loading}
          className="bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] disabled:bg-[var(--color-brand-muted)] dark:disabled:bg-[var(--color-brand-muted)] text-white font-medium py-2 px-6 rounded-lg transition-colors duration-150"
        >
          {loading ? t('reports.generating') : t('reports.generateReport')}
        </button>
      </div>
    </form>
  )
}
