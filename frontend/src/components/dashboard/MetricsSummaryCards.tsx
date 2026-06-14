import { memo } from 'react'
import { useTranslation } from 'react-i18next'
import type { MetricsDTO } from '../../api/types'
import { memoCompare } from '../../utils/deepEqual'

interface MetricsSummaryCardsProps {
  metrics: MetricsDTO[]
  isLoading?: boolean
}

interface MetricSummary {
  averageLatency: number
  averagePacketLoss: number
  averageJitter: number
  totalNodes: number
  onlineNodes: number
  offlineNodes: number
}

function calculateMetrics(metrics: MetricsDTO[]): MetricSummary {
  if (metrics.length === 0) {
    return { averageLatency: 0, averagePacketLoss: 0, averageJitter: 0, totalNodes: 0, onlineNodes: 0, offlineNodes: 0 }
  }
  const totalLatency = metrics.reduce((sum, m) => sum + m.latency_ms, 0)
  const totalPacketLoss = metrics.reduce((sum, m) => sum + m.packet_loss_rate, 0)
  const totalJitter = metrics.reduce((sum, m) => sum + m.jitter_ms, 0)
  return {
    averageLatency: totalLatency / metrics.length,
    averagePacketLoss: totalPacketLoss / metrics.length,
    averageJitter: totalJitter / metrics.length,
    totalNodes: metrics.length,
    onlineNodes: metrics.length,
    offlineNodes: 0,
  }
}

function getMetricColor(value: number, threshold: number) {
  if (value >= threshold) return { bg: 'bg-destructive/10', text: 'text-destructive', icon: 'text-destructive' }
  if (value >= threshold * 0.8) return { bg: 'bg-warning-bg', text: 'text-warning-text', icon: 'text-warning' }
  return { bg: 'bg-healthy-bg', text: 'text-healthy-text', icon: 'text-healthy' }
}

export const MetricsSummaryCards = memo(function MetricsSummaryCards({ metrics, isLoading }: MetricsSummaryCardsProps) {
  const { t } = useTranslation()
  const safeMetrics = Array.isArray(metrics) ? metrics : []
  const summary = calculateMetrics(safeMetrics)
  const latencyColor = getMetricColor(summary.averageLatency, 200)
  const packetLossColor = getMetricColor(summary.averagePacketLoss, 5)
  const jitterColor = getMetricColor(summary.averageJitter, 50)

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3].map(i => (
          <div key={i} className="bg-card shadow rounded-lg p-6">
            <div className="animate-pulse">
              <div className="h-4 bg-muted rounded w-1/2 mb-4"></div>
              <div className="h-8 bg-muted rounded w-3/4"></div>
            </div>
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
      <div className={`${latencyColor.bg} rounded-lg shadow overflow-hidden`}>
        <div className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className={`h-6 w-6 ${latencyColor.icon}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="ml-5 w-0 flex-1">
              <dl>
                <dt className={`text-sm font-medium truncate ${latencyColor.text}`}>{t('metrics.avgLatency')}</dt>
                <dd><div className={`text-2xl font-semibold ${latencyColor.text}`}>{summary.averageLatency.toFixed(1)}<span className="text-sm font-normal ml-1">{t('units.ms')}</span></div></dd>
              </dl>
            </div>
          </div>
        </div>
        <div className={`${latencyColor.bg} px-6 py-3`}>
          <div className="text-xs text-muted-foreground">{t('dashboard.acrossNodes', { count: summary.totalNodes })}</div>
        </div>
      </div>

      <div className={`${packetLossColor.bg} rounded-lg shadow overflow-hidden`}>
        <div className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className={`h-6 w-6 ${packetLossColor.icon}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div className="ml-5 w-0 flex-1">
              <dl>
                <dt className={`text-sm font-medium truncate ${packetLossColor.text}`}>{t('metrics.avgPacketLoss')}</dt>
                <dd><div className={`text-2xl font-semibold ${packetLossColor.text}`}>{summary.averagePacketLoss.toFixed(2)}<span className="text-sm font-normal ml-1">{t('units.percent')}</span></div></dd>
              </dl>
            </div>
          </div>
        </div>
        <div className={`${packetLossColor.bg} px-6 py-3`}>
          <div className="text-xs text-muted-foreground">{t('dashboard.acrossNodes', { count: summary.totalNodes })}</div>
        </div>
      </div>

      <div className={`${jitterColor.bg} rounded-lg shadow overflow-hidden`}>
        <div className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className={`h-6 w-6 ${jitterColor.icon}`} fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <div className="ml-5 w-0 flex-1">
              <dl>
                <dt className={`text-sm font-medium truncate ${jitterColor.text}`}>{t('metrics.avgJitter')}</dt>
                <dd><div className={`text-2xl font-semibold ${jitterColor.text}`}>{summary.averageJitter.toFixed(1)}<span className="text-sm font-normal ml-1">{t('units.ms')}</span></div></dd>
              </dl>
            </div>
          </div>
        </div>
        <div className={`${jitterColor.bg} px-6 py-3`}>
          <div className="text-xs text-muted-foreground">{t('dashboard.acrossNodes', { count: summary.totalNodes })}</div>
        </div>
      </div>
    </div>
  )
}, memoCompare)
