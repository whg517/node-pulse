import type { MetricsDTO } from '../../api/types'
import { memo } from 'react'
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

/**
 * Calculate summary statistics from metrics array
 */
function calculateMetrics(metrics: MetricsDTO[]): MetricSummary {
  if (metrics.length === 0) {
    return {
      averageLatency: 0,
      averagePacketLoss: 0,
      averageJitter: 0,
      totalNodes: 0,
      onlineNodes: 0,
      offlineNodes: 0,
    }
  }

  const totalLatency = metrics.reduce((sum, m) => sum + m.latency_ms, 0)
  const totalPacketLoss = metrics.reduce((sum, m) => sum + m.packet_loss_rate, 0)
  const totalJitter = metrics.reduce((sum, m) => sum + m.jitter_ms, 0)

  return {
    averageLatency: totalLatency / metrics.length,
    averagePacketLoss: totalPacketLoss / metrics.length,
    averageJitter: totalJitter / metrics.length,
    totalNodes: metrics.length,
    onlineNodes: metrics.length, // All nodes with metrics are considered online
    offlineNodes: 0,
  }
}

/**
 * Get color class based on metric value and threshold
 */
function getMetricColor(value: number, threshold: number): {
  bg: string
  text: string
  icon: string
} {
  if (value >= threshold) {
    return {
      bg: 'bg-red-50',
      text: 'text-red-700',
      icon: 'text-red-500',
    }
  }
  if (value >= threshold * 0.8) {
    return {
      bg: 'bg-yellow-50',
      text: 'text-yellow-700',
      icon: 'text-yellow-500',
    }
  }
  return {
    bg: 'bg-green-50',
    text: 'text-green-700',
    icon: 'text-green-500',
  }
}

/**
 * MetricsSummaryCards Component
 *
 * Displays summary cards for core metrics:
 * - Average Latency
 * - Average Packet Loss Rate
 * - Average Jitter
 *
 * @param metrics - Array of metrics data points
 * @param isLoading - Optional loading state
 *
 * @example
 * <MetricsSummaryCards metrics={metrics} />
 */
export const MetricsSummaryCards = memo(function MetricsSummaryCards({
  metrics,
  isLoading,
}: MetricsSummaryCardsProps) {
  // Defensive check: ensure metrics is an array
  const safeMetrics = Array.isArray(metrics) ? metrics : []

  const summary = calculateMetrics(safeMetrics)

  const latencyColor = getMetricColor(summary.averageLatency, 200)
  const packetLossColor = getMetricColor(summary.averagePacketLoss, 5)
  const jitterColor = getMetricColor(summary.averageJitter, 50)

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3].map(i => (
          <div key={i} className="bg-white shadow rounded-lg p-6">
            <div className="animate-pulse">
              <div className="h-4 bg-gray-200 rounded w-1/2 mb-4"></div>
              <div className="h-8 bg-gray-200 rounded w-3/4"></div>
            </div>
          </div>
        ))}
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
      {/* Average Latency Card */}
      <div className={`${latencyColor.bg} rounded-lg shadow overflow-hidden`}>
        <div className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg
                className={`h-6 w-6 ${latencyColor.icon}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <div className="ml-5 w-0 flex-1">
              <dl>
                <dt
                  className={`text-sm font-medium truncate ${latencyColor.text}`}
                >
                  Avg Latency
                </dt>
                <dd>
                  <div
                    className={`text-2xl font-semibold ${latencyColor.text}`}
                  >
                    {summary.averageLatency.toFixed(1)}
                    <span className="text-sm font-normal ml-1">ms</span>
                  </div>
                </dd>
              </dl>
            </div>
          </div>
        </div>
        <div className={`${latencyColor.bg} px-6 py-3`}>
          <div className="text-xs text-gray-500">
            Across {summary.totalNodes} {summary.totalNodes === 1 ? 'node' : 'nodes'}
          </div>
        </div>
      </div>

      {/* Average Packet Loss Card */}
      <div className={`${packetLossColor.bg} rounded-lg shadow overflow-hidden`}>
        <div className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg
                className={`h-6 w-6 ${packetLossColor.icon}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
            </div>
            <div className="ml-5 w-0 flex-1">
              <dl>
                <dt
                  className={`text-sm font-medium truncate ${packetLossColor.text}`}
                >
                  Avg Packet Loss
                </dt>
                <dd>
                  <div
                    className={`text-2xl font-semibold ${packetLossColor.text}`}
                  >
                    {summary.averagePacketLoss.toFixed(2)}
                    <span className="text-sm font-normal ml-1">%</span>
                  </div>
                </dd>
              </dl>
            </div>
          </div>
        </div>
        <div className={`${packetLossColor.bg} px-6 py-3`}>
          <div className="text-xs text-gray-500">
            Across {summary.totalNodes} {summary.totalNodes === 1 ? 'node' : 'nodes'}
          </div>
        </div>
      </div>

      {/* Average Jitter Card */}
      <div className={`${jitterColor.bg} rounded-lg shadow overflow-hidden`}>
        <div className="p-6">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg
                className={`h-6 w-6 ${jitterColor.icon}`}
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
            </div>
            <div className="ml-5 w-0 flex-1">
              <dl>
                <dt
                  className={`text-sm font-medium truncate ${jitterColor.text}`}
                >
                  Avg Jitter
                </dt>
                <dd>
                  <div
                    className={`text-2xl font-semibold ${jitterColor.text}`}
                  >
                    {summary.averageJitter.toFixed(1)}
                    <span className="text-sm font-normal ml-1">ms</span>
                  </div>
                </dd>
              </dl>
            </div>
          </div>
        </div>
        <div className={`${jitterColor.bg} px-6 py-3`}>
          <div className="text-xs text-gray-500">
            Across {summary.totalNodes} {summary.totalNodes === 1 ? 'node' : 'nodes'}
          </div>
        </div>
      </div>
    </div>
  )
}, memoCompare)
