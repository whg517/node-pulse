import { useNavigate } from 'react-router-dom'
import { memo } from 'react'
import { HealthStatusBadge } from './HealthStatusBadge'
import { determineHealthStatus, type HealthStatus } from '../../utils/healthStatus'
import { memoCompare } from '../../utils/deepEqual'
import type { NodeDTO } from '../../api/types'
import type { MetricsDTO } from '../../api/types'

interface NodeWithHealth {
  node: NodeDTO
  metrics?: MetricsDTO
  healthStatus: HealthStatus
  severityScore: number
}

interface TopAnomaliesListProps {
  nodes: NodeDTO[]
  metrics: MetricsDTO[]
  isLoading?: boolean
}

/**
 * Calculate severity score for sorting anomalies
 * Critical = 3, Warning = 2, Healthy = 1, Offline = 0
 */
function getSeverityScore(status: HealthStatus): number {
  switch (status) {
    case 'critical':
      return 3
    case 'warning':
      return 2
    case 'healthy':
      return 1
    case 'offline':
      return 0
  }
}

/**
 * Get a key metric value to display for each node
 * Returns the metric that's most indicative of the health status
 */
function getKeyMetric(_node: NodeDTO, metrics?: MetricsDTO): string {
  if (!metrics) {
    return 'No data'
  }

  const healthStatus = determineHealthStatus({
    latency_ms: metrics.latency_ms,
    packet_loss_rate: metrics.packet_loss_rate,
    jitter_ms: metrics.jitter_ms,
    last_heartbeat: metrics.timestamp,
  })

  if (healthStatus === 'critical' || healthStatus === 'warning') {
    // Show the worst metric
    if (metrics.packet_loss_rate > 5) {
      return `${metrics.packet_loss_rate.toFixed(1)}% loss`
    }
    if (metrics.latency_ms > 200) {
      return `${metrics.latency_ms.toFixed(0)}ms`
    }
    if (metrics.jitter_ms > 50) {
      return `${metrics.jitter_ms.toFixed(0)}ms jitter`
    }
  }

  return `${metrics.latency_ms.toFixed(0)}ms`
}

/**
 * TopAnomaliesList Component
 *
 * Displays the top 5 nodes with the most critical health status.
 * Sorted by severity (critical > warning > healthy).
 *
 * @param nodes - Array of all nodes
 * @param metrics - Array of metrics for health determination
 * @param isLoading - Optional loading state
 *
 * @example
 * <TopAnomaliesList nodes={nodes} metrics={metrics} />
 */
export const TopAnomaliesList = memo(function TopAnomaliesList({ nodes, metrics, isLoading }: TopAnomaliesListProps) {
  const navigate = useNavigate()

  // Defensive check: ensure nodes and metrics are arrays
  const safeNodes = Array.isArray(nodes) ? nodes : []
  const safeMetrics = Array.isArray(metrics) ? metrics : []

  // Create a map of node_id to metrics
  const metricsMap = new Map(safeMetrics.map(m => [m.node_id, m]))

  // Calculate health status and severity for each node
  const nodesWithHealth: NodeWithHealth[] = safeNodes.map(node => {
    const metrics = metricsMap.get(node.id)
    const healthStatus = metrics
      ? determineHealthStatus({
          latency_ms: metrics.latency_ms,
          packet_loss_rate: metrics.packet_loss_rate,
          jitter_ms: metrics.jitter_ms,
          last_heartbeat: metrics.timestamp,
        })
      : 'offline'

    return {
      node,
      metrics,
      healthStatus,
      severityScore: getSeverityScore(healthStatus),
    }
  })

  // Sort by severity score (highest first) and take top 5
  const topAnomalies = nodesWithHealth
    .sort((a, b) => b.severityScore - a.severityScore)
    .slice(0, 5)

  const handleNodeClick = (nodeId: string) => {
    navigate(`/nodes/${nodeId}`)
  }

  if (isLoading) {
    return (
      <div className="bg-[var(--color-bg-surface)] shadow rounded-lg p-6">
        <div className="animate-pulse">
          <div className="h-8 bg-[var(--color-bg-muted)] rounded w-1/3 mb-4"></div>
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="h-16 bg-[var(--color-bg-muted)] rounded"></div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  // Filter to show only critical and warning nodes
  const criticalAndWarning = topAnomalies.filter(
    n => n.healthStatus === 'critical' || n.healthStatus === 'warning'
  )

  if (criticalAndWarning.length === 0) {
    return (
      <div className="bg-[var(--color-bg-surface)] shadow rounded-lg p-6">
        <div className="px-6 py-4 border-b border-[var(--color-border)]">
          <h3 className="text-lg font-medium text-[var(--color-text-primary)]">Top Anomalies</h3>
          <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
            Nodes with critical or warning health status
          </p>
        </div>
        <div className="text-center py-12">
          <svg
            className="mx-auto h-12 w-12 text-green-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-[var(--color-text-primary)]">All systems normal</h3>
          <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
            No critical or warning issues detected across all nodes.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-[var(--color-bg-surface)] shadow rounded-lg overflow-hidden">
      <div className="px-6 py-4 border-b border-[var(--color-border)]">
        <h3 className="text-lg font-medium text-[var(--color-text-primary)]">Top Anomalies</h3>
        <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
          Nodes requiring attention (sorted by severity)
        </p>
      </div>
      <ul className="divide-y divide-[var(--color-border)]">
        {criticalAndWarning.map(({ node, metrics, healthStatus }) => (
          <li
            key={node.id}
            onClick={() => handleNodeClick(node.id)}
            className="hover:bg-[var(--color-hover-overlay)] cursor-pointer transition-colors duration-150"
          >
            <div className="px-6 py-4">
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-[var(--color-text-primary)] truncate">
                    {node.name}
                  </p>
                  <p className="text-sm text-[var(--color-text-secondary)] truncate">{node.ip}</p>
                </div>
                <div className="ml-4 flex items-center space-x-3">
                  <div className="text-right">
                    <p className="text-sm font-medium text-[var(--color-text-primary)]">
                      {metrics ? getKeyMetric(node, metrics) : 'Offline'}
                    </p>
                    <p className="text-xs text-[var(--color-text-muted)]">{node.region}</p>
                  </div>
                  <HealthStatusBadge status={healthStatus} />
                </div>
              </div>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}, memoCompare)
