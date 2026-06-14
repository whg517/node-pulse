/**
 * HealthReportPDF Component
 *
 * Generates a printable PDF health report with all required sections:
 * 1. Node Information
 * 2. Key Metrics
 * 3. MTR Path Analysis
 * 4. 7-Day Baseline Comparison
 * 5. Root Cause Analysis
 * 6. Event Timeline
 *
 * Uses browser print dialog for PDF generation (no external libraries).
 * Optimized for ≤ 5 pages, ≤ 2MB, ≤ 10 seconds generation time.
 */

import { useTranslation } from 'react-i18next'
import type { NodeDTO } from '../../api/types'

export interface HealthMetrics {
  latency: {
    current: number
    baseline: number
    trend: 'improved' | 'degraded' | 'stable'
    data: Array<{ timestamp: string; value: number }>
  }
  packetLoss: {
    current: number
    baseline: number
    trend: 'improved' | 'degraded' | 'stable'
    data: Array<{ timestamp: string; value: number }>
  }
  jitter: {
    current: number
    baseline: number
    trend: 'improved' | 'degraded' | 'stable'
    data: Array<{ timestamp: string; value: number }>
  }
  uptime: number
  totalProbes: number
  failedProbes: number
}

export interface MTRHop {
  hop: number
  ip: string
  location?: string
  avgLatency: number
  lossRate: number
}

export interface RootCauseAnalysis {
  probableCause: string
  confidence: 'high' | 'medium' | 'low'
  impact: string
  recommendation: string
}

export interface TimelineEvent {
  timestamp: string
  event: string
  severity: 'critical' | 'warning' | 'info'
}

interface HealthReportPDFProps {
  node: NodeDTO
  metrics: HealthMetrics
  mtrPath?: MTRHop[]
  rootCause?: RootCauseAnalysis
  timeline?: TimelineEvent[]
  reportPeriod: {
    start: string
    end: string
  }
  onClose?: () => void
}

export function HealthReportPDF({
  node,
  metrics,
  mtrPath = [],
  rootCause,
  timeline = [],
  reportPeriod,
  onClose,
}: HealthReportPDFProps) {
  const { t, i18n } = useTranslation()
  const isZh = i18n.language === 'zh-CN'

  const getTrendIcon = (trend: 'improved' | 'degraded' | 'stable') => {
    switch (trend) {
      case 'improved':
        return '↓'
      case 'degraded':
        return '↑'
      case 'stable':
        return '→'
    }
  }

  const getTrendClass = (trend: 'improved' | 'degraded' | 'stable') => {
    switch (trend) {
      case 'improved':
        return 'text-healthy'
      case 'degraded':
        return 'text-destructive'
      case 'stable':
        return 'text-muted-foreground'
    }
  }

  const getSeverityClass = (severity: 'critical' | 'warning' | 'info') => {
    switch (severity) {
      case 'critical':
        return 'bg-destructive/10 text-destructive'
      case 'warning':
        return 'bg-warning-bg text-warning-text'
      case 'info':
        return 'bg-primary/10 text-primary'
    }
  }

  const getConfidenceClass = (confidence: 'high' | 'medium' | 'low') => {
    switch (confidence) {
      case 'high':
        return 'bg-healthy-bg text-healthy-text'
      case 'medium':
        return 'bg-warning-bg text-warning-text'
      case 'low':
        return 'bg-muted text-muted-foreground'
    }
  }

  const getNodeHealthStatus = () => {
    const { latency, packetLoss, jitter } = metrics
    const degradedCount = [latency, packetLoss, jitter].filter((m) => m.trend === 'degraded').length

    if (degradedCount === 0) {
      return { status: t('status.healthy'), class: 'text-healthy' }
    } else if (degradedCount === 1) {
      return { status: t('status.warning'), class: 'text-warning' }
    } else {
      return { status: t('status.critical'), class: 'text-destructive' }
    }
  }

  const healthStatus = getNodeHealthStatus()

  const handlePrint = () => {
    window.print()
    onClose?.()
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString(isZh ? 'zh-CN' : 'en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className="health-report-container">
      {/* Print Styles */}
      <style>{`
        @media print {
          @page {
            size: A4;
            margin: 15mm;
          }
          body {
            -webkit-print-color-adjust: exact;
            print-color-adjust: exact;
          }
          .no-print {
            display: none !important;
          }
          .health-report-container {
            background: white !important;
            color: black !important;
          }
          .report-section {
            page-break-inside: avoid;
          }
        }
        .health-report-container {
          background: white;
          color: #1a1a1a;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
        }
        .report-header {
          border-bottom: 3px solid #2563eb;
          padding-bottom: 1rem;
          margin-bottom: 1.5rem;
        }
        .section-title {
          font-size: 1.125rem;
          font-weight: 600;
          color: #1e293b;
          border-left: 4px solid #2563eb;
          padding-left: 0.75rem;
          margin-bottom: 1rem;
        }
        .metric-card {
          background: #f8fafc;
          border-radius: 0.5rem;
          padding: 1rem;
          border: 1px solid #e2e8f0;
        }
        .metric-value {
          font-size: 1.5rem;
          font-weight: 700;
          color: #1e293b;
        }
        .metric-label {
          font-size: 0.875rem;
          color: #64748b;
        }
        .trend-indicator {
          display: inline-flex;
          align-items: center;
          gap: 0.25rem;
          font-size: 0.75rem;
          font-weight: 500;
          padding: 0.125rem 0.5rem;
          border-radius: 9999px;
        }
        .timeline-item {
          position: relative;
          padding-left: 1.5rem;
          padding-bottom: 0.75rem;
          border-left: 2px solid #e2e8f0;
        }
        .timeline-item:last-child {
          border-left-color: transparent;
        }
        .timeline-item::before {
          content: '';
          position: absolute;
          left: -0.375rem;
          top: 0;
          width: 0.75rem;
          height: 0.75rem;
          border-radius: 50%;
          background: #2563eb;
        }
        .mtr-hop {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          padding: 0.5rem;
          background: #f8fafc;
          border-radius: 0.375rem;
          margin-bottom: 0.25rem;
        }
        .hop-number {
          width: 2rem;
          height: 2rem;
          display: flex;
          align-items: center;
          justify-content: center;
          background: #2563eb;
          color: white;
          border-radius: 50%;
          font-weight: 600;
          font-size: 0.875rem;
        }
      `}</style>

      {/* Action Buttons (hidden in print) */}
      <div className="no-print flex justify-end gap-2 mb-4">
        <button
          onClick={handlePrint}
          className="bg-primary hover:bg-primary/85 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-150 flex items-center gap-2"
        >
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
          </svg>
          {t('reports.printReport')}
        </button>
        {onClose && (
          <button
            onClick={onClose}
            className="bg-muted/50 hover:bg-muted text-muted-foreground font-medium py-2 px-4 rounded-lg transition-colors duration-150"
          >
            {t('common.close')}
          </button>
        )}
      </div>

      {/* Report Header */}
      <header className="report-header">
        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-2xl font-bold text-foreground">{t('reports.healthReportTitle')}</h1>
            <p className="text-muted-foreground mt-1">{node.name} - {node.region}</p>
          </div>
          <div className="text-right">
            <p className="text-sm text-muted-foreground">{t('reports.generatedOn')}</p>
            <p className="text-sm font-medium text-foreground">{formatDate(new Date().toISOString())}</p>
            <p className="text-xs text-muted-foreground mt-1">
              {formatDate(reportPeriod.start)} - {formatDate(reportPeriod.end)}
            </p>
          </div>
        </div>
      </header>

      {/* Section 1: Node Information */}
      <section className="report-section mb-6">
        <h2 className="section-title">{t('reports.nodeInfo')}</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="metric-card">
            <p className="metric-label">{t('nodes.nodeName')}</p>
            <p className="metric-value text-base">{node.name}</p>
          </div>
          <div className="metric-card">
            <p className="metric-label">{t('nodes.nodeId')}</p>
            <p className="metric-value text-base font-mono text-xs">{node.id}</p>
          </div>
          <div className="metric-card">
            <p className="metric-label">{t('nodes.region')}</p>
            <p className="metric-value text-base">{node.region}</p>
          </div>
          <div className="metric-card">
            <p className="metric-label">{t('reports.nodeHealth')}</p>
            <p className={`metric-value text-base ${healthStatus.class}`}>{healthStatus.status}</p>
          </div>
        </div>
      </section>

      {/* Section 2: Key Metrics */}
      <section className="report-section mb-6">
        <h2 className="section-title">{t('reports.keyMetrics')}</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {/* Latency */}
          <div className="metric-card">
            <div className="flex justify-between items-start mb-2">
              <p className="metric-label">{t('metrics.latency')}</p>
              <span className={`trend-indicator ${getTrendClass(metrics.latency.trend)}`}>
                {getTrendIcon(metrics.latency.trend)}
                {t(`reports.${metrics.latency.trend}`)}
              </span>
            </div>
            <p className="metric-value">{metrics.latency.current.toFixed(1)} ms</p>
            <p className="text-xs text-muted-foreground mt-1">
              {t('reports.baseline')}: {metrics.latency.baseline.toFixed(1)} ms
            </p>
          </div>

          {/* Packet Loss */}
          <div className="metric-card">
            <div className="flex justify-between items-start mb-2">
              <p className="metric-label">{t('metrics.packetLoss')}</p>
              <span className={`trend-indicator ${getTrendClass(metrics.packetLoss.trend)}`}>
                {getTrendIcon(metrics.packetLoss.trend)}
                {t(`reports.${metrics.packetLoss.trend}`)}
              </span>
            </div>
            <p className="metric-value">{metrics.packetLoss.current.toFixed(2)}%</p>
            <p className="text-xs text-muted-foreground mt-1">
              {t('reports.baseline')}: {metrics.packetLoss.baseline.toFixed(2)}%
            </p>
          </div>

          {/* Jitter */}
          <div className="metric-card">
            <div className="flex justify-between items-start mb-2">
              <p className="metric-label">{t('metrics.jitter')}</p>
              <span className={`trend-indicator ${getTrendClass(metrics.jitter.trend)}`}>
                {getTrendIcon(metrics.jitter.trend)}
                {t(`reports.${metrics.jitter.trend}`)}
              </span>
            </div>
            <p className="metric-value">{metrics.jitter.current.toFixed(1)} ms</p>
            <p className="text-xs text-muted-foreground mt-1">
              {t('reports.baseline')}: {metrics.jitter.baseline.toFixed(1)} ms
            </p>
          </div>

          {/* Uptime */}
          <div className="metric-card">
            <p className="metric-label">{t('metrics.uptime')}</p>
            <p className="metric-value">{metrics.uptime.toFixed(1)}%</p>
            <p className="text-xs text-muted-foreground mt-1">
              {metrics.totalProbes.toLocaleString()} {t('reports.totalProbes').toLowerCase()}
            </p>
          </div>
        </div>

        {/* Probe Statistics */}
        <div className="grid grid-cols-2 gap-4 mt-4">
          <div className="metric-card">
            <p className="metric-label">{t('reports.totalProbes')}</p>
            <p className="metric-value text-lg">{metrics.totalProbes.toLocaleString()}</p>
          </div>
          <div className="metric-card">
            <p className="metric-label">{t('reports.failedProbes')}</p>
            <p className="metric-value text-lg text-destructive">{metrics.failedProbes.toLocaleString()}</p>
          </div>
        </div>
      </section>

      {/* Section 3: MTR Path Analysis */}
      {mtrPath.length > 0 && (
        <section className="report-section mb-6">
          <h2 className="section-title">{t('reports.mtrPath')}</h2>
          <div className="space-y-2">
            {mtrPath.map((hop) => (
              <div key={hop.hop} className="mtr-hop">
                <div className="hop-number">{hop.hop}</div>
                <div className="flex-1">
                  <p className="font-medium text-sm">{hop.ip}</p>
                  {hop.location && <p className="text-xs text-muted-foreground">{hop.location}</p>}
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium">{hop.avgLatency.toFixed(1)} ms</p>
                  {hop.lossRate > 0 && (
                    <p className="text-xs text-destructive">{hop.lossRate.toFixed(1)}% loss</p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* Section 4: 7-Day Baseline Comparison */}
      <section className="report-section mb-6">
        <h2 className="section-title">{t('reports.baselineComparison')}</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-2 px-3 font-medium text-muted-foreground">{t('metrics.latency')}</th>
                <th className="text-left py-2 px-3 font-medium text-muted-foreground">{t('metrics.packetLoss')}</th>
                <th className="text-left py-2 px-3 font-medium text-muted-foreground">{t('metrics.jitter')}</th>
                <th className="text-left py-2 px-3 font-medium text-muted-foreground">{t('reports.change')}</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-b border-border">
                <td className="py-3 px-3">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{metrics.latency.current.toFixed(1)} ms</span>
                    <span className={`text-xs ${getTrendClass(metrics.latency.trend)}`}>
                      ({metrics.latency.current > metrics.latency.baseline ? '+' : ''}
                      {((metrics.latency.current - metrics.latency.baseline) / metrics.latency.baseline * 100).toFixed(1)}%)
                    </span>
                  </div>
                </td>
                <td className="py-3 px-3">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{metrics.packetLoss.current.toFixed(2)}%</span>
                    <span className={`text-xs ${getTrendClass(metrics.packetLoss.trend)}`}>
                      ({metrics.packetLoss.current > metrics.packetLoss.baseline ? '+' : ''}
                      {((metrics.packetLoss.current - metrics.packetLoss.baseline) / metrics.packetLoss.baseline * 100).toFixed(1)}%)
                    </span>
                  </div>
                </td>
                <td className="py-3 px-3">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{metrics.jitter.current.toFixed(1)} ms</span>
                    <span className={`text-xs ${getTrendClass(metrics.jitter.trend)}`}>
                      ({metrics.jitter.current > metrics.jitter.baseline ? '+' : ''}
                      {((metrics.jitter.current - metrics.jitter.baseline) / metrics.jitter.baseline * 100).toFixed(1)}%)
                    </span>
                  </div>
                </td>
                <td className="py-3 px-3">
                  <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${
                    [metrics.latency, metrics.packetLoss, metrics.jitter].some(m => m.trend === 'degraded')
                      ? 'bg-warning-bg text-warning-text'
                      : 'bg-healthy-bg text-healthy-text'
                  }`}>
                    {getTrendIcon(
                      [metrics.latency, metrics.packetLoss, metrics.jitter].some(m => m.trend === 'degraded')
                        ? 'degraded'
                        : 'improved'
                    )}
                    {t(`reports.${
                      [metrics.latency, metrics.packetLoss, metrics.jitter].some(m => m.trend === 'degraded')
                        ? 'degraded'
                        : 'improved'
                    }`)}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      {/* Section 5: Root Cause Analysis */}
      {rootCause && (
        <section className="report-section mb-6">
          <h2 className="section-title">{t('reports.rootCauseAnalysis')}</h2>
          {rootCause.probableCause === t('reports.noIssues') ? (
            <div className="bg-healthy-bg border border-healthy-bg rounded-lg p-4">
              <div className="flex items-start gap-3">
                <svg className="w-5 h-5 text-healthy mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <div>
                  <p className="font-medium text-healthy-text">{t('reports.healthyStatus')}</p>
                  <p className="text-sm text-healthy mt-1">{t('reports.noIssues')}</p>
                </div>
              </div>
            </div>
          ) : (
            <div className="space-y-3">
              <div className="metric-card">
                <p className="text-sm text-muted-foreground mb-1">{t('reports.probableCause')}</p>
                <p className="font-medium text-foreground">{rootCause.probableCause}</p>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="metric-card">
                  <p className="text-sm text-muted-foreground mb-1">{t('reports.confidenceLevel')}</p>
                  <span className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${getConfidenceClass(rootCause.confidence)}`}>
                    {t(`reports.${rootCause.confidence}`)}
                  </span>
                </div>
                <div className="metric-card">
                  <p className="text-sm text-muted-foreground mb-1">{t('reports.impact')}</p>
                  <p className="font-medium text-foreground text-sm">{rootCause.impact}</p>
                </div>
              </div>
              <div className="metric-card bg-primary/10 border border-primary/10">
                <p className="text-sm text-primary mb-1 font-medium">{t('reports.recommendation')}</p>
                <p className="text-foreground">{rootCause.recommendation}</p>
              </div>
            </div>
          )}
        </section>
      )}

      {/* Section 6: Event Timeline */}
      {timeline.length > 0 && (
        <section className="report-section mb-6">
          <h2 className="section-title">{t('reports.eventTimeline')}</h2>
          <div className="space-y-3">
            {timeline.map((event, index) => (
              <div key={index} className="timeline-item">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1">
                    <p className="font-medium text-foreground">{event.event}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">{formatDate(event.timestamp)}</p>
                  </div>
                  <span className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${getSeverityClass(event.severity)}`}>
                    {t(`status.${event.severity}`)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* Report Footer */}
      <footer className="border-t border-border pt-4 mt-8">
        <div className="flex justify-between items-center text-xs text-muted-foreground">
          <p>Node-Pulse Health Report</p>
          <p>
            {t('reports.generatedOn')}: {formatDate(new Date().toISOString())}
          </p>
        </div>
      </footer>
    </div>
  )
}
