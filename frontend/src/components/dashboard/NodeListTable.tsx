import { useNavigate } from 'react-router-dom'
import { memo } from 'react'
import { HealthStatusBadge } from './HealthStatusBadge'
import { determineHealthStatus } from '../../utils/healthStatus'
import { memoCompare } from '../../utils/deepEqual'
import type { NodeDTO } from '../../api/types'
import type { MetricsDTO } from '../../api/types'

interface NodeListTableProps {
  nodes: NodeDTO[]
  metrics: MetricsDTO[]
  isLoading?: boolean
}

/**
 * NodeListTable Component
 *
 * Displays a table of all monitoring nodes with their health status.
 * Supports clicking on rows to navigate to node detail pages.
 *
 * @param nodes - Array of nodes to display
 * @param metrics - Array of metrics for health determination
 * @param isLoading - Optional loading state
 *
 * @example
 * <NodeListTable nodes={nodes} metrics={metrics} />
 */
export const NodeListTable = memo(function NodeListTable({ nodes, metrics, isLoading }: NodeListTableProps) {
  const navigate = useNavigate()

  // Defensive check: ensure nodes and metrics are arrays
  const safeNodes = Array.isArray(nodes) ? nodes : []
  const safeMetrics = Array.isArray(metrics) ? metrics : []

  // Create a map of node_id to metrics for quick lookup
  const metricsMap = new Map(safeMetrics.map(m => [m.node_id, m]))

  const handleRowClick = (nodeId: string) => {
    navigate(`/nodes/${nodeId}`)
  }

  if (isLoading) {
    return (
      <div className="bg-[var(--color-bg-surface)] shadow rounded-lg p-6">
        <div className="animate-pulse">
          <div className="h-8 bg-[var(--color-bg-muted)] rounded w-1/4 mb-4"></div>
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="h-12 bg-[var(--color-bg-muted)] rounded"></div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  if (safeNodes.length === 0) {
    return (
      <div className="bg-[var(--color-bg-surface)] shadow rounded-lg p-6">
        <h3 className="text-lg font-medium text-[var(--color-text-primary)] mb-4">Node List</h3>
        <div className="text-center py-12">
          <svg
            className="mx-auto h-12 w-12 text-[var(--color-text-muted)]"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-[var(--color-text-primary)]">No nodes</h3>
          <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
            No monitoring nodes configured yet. Get started by adding your first node.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-[var(--color-bg-surface)] shadow rounded-lg overflow-hidden">
      <div className="px-6 py-4 border-b border-[var(--color-border)]">
        <h3 className="text-lg font-medium text-[var(--color-text-primary)]">Node List</h3>
        <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
          Overview of all monitoring nodes and their health status
        </p>
      </div>
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-[var(--color-border)]">
          <thead className="bg-[var(--color-bg-muted)]">
            <tr>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                Node Name
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                IP Address
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                Region
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                Status
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                Health
              </th>
            </tr>
          </thead>
          <tbody className="bg-[var(--color-bg-surface)] divide-y divide-[var(--color-border)]">
            {safeNodes.map(node => {
              const metrics = metricsMap.get(node.id)
              const healthStatus = metrics
                ? determineHealthStatus({
                    latency_ms: metrics.latency_ms,
                    packet_loss_rate: metrics.packet_loss_rate,
                    jitter_ms: metrics.jitter_ms,
                    last_heartbeat: metrics.timestamp,
                  })
                : 'offline'

              return (
                <tr
                  key={node.id}
                  onClick={() => handleRowClick(node.id)}
                  className="hover:bg-[var(--color-hover-overlay)] cursor-pointer transition-colors duration-150"
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <div>
                        <div className="text-sm font-medium text-[var(--color-text-primary)]">
                          {node.name}
                        </div>
                        <div className="text-sm text-[var(--color-text-muted)]">{node.id}</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-[var(--color-text-primary)]">{node.ip}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 dark:bg-blue-900/40 text-blue-800 dark:text-blue-300">
                      {node.region}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        node.status === 'online'
                          ? 'bg-green-100 dark:bg-green-900/40 text-green-800 dark:text-green-300'
                          : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)]'
                      }`}
                    >
                      {node.status === 'online' ? 'Online' : 'Offline'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <HealthStatusBadge status={healthStatus} />
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}, memoCompare)
