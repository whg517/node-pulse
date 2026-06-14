import { useNavigate } from 'react-router-dom'
import { memo } from 'react'
import { useTranslation } from 'react-i18next'
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

export const NodeListTable = memo(function NodeListTable({ nodes, metrics, isLoading }: NodeListTableProps) {
  const navigate = useNavigate()
  const { t } = useTranslation()

  const safeNodes = Array.isArray(nodes) ? nodes : []
  const safeMetrics = Array.isArray(metrics) ? metrics : []

  const metricsMap = new Map(safeMetrics.map(m => [m.node_id, m]))

  const handleRowClick = (nodeId: string, nodeName: string) => {
    navigate(`/nodes/${nodeId}`, { state: { breadcrumbLabel: nodeName } })
  }

  if (isLoading) {
    return (
      <div className="bg-card shadow rounded-lg p-6">
        <div className="animate-pulse">
          <div className="h-8 bg-muted rounded w-1/4 mb-4"></div>
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="h-12 bg-muted rounded"></div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  if (safeNodes.length === 0) {
    return (
      <div className="bg-card shadow rounded-lg p-6">
        <h3 className="text-lg font-medium text-foreground mb-4">{t('dashboard.nodeList')}</h3>
        <div className="text-center py-12">
          <svg className="mx-auto h-12 w-12 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-foreground">{t('dashboard.noNodes')}</h3>
          <p className="mt-1 text-sm text-muted-foreground">{t('dashboard.noNodesDescription')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-card shadow rounded-lg overflow-hidden">
      <div className="px-6 py-4 border-b border-border">
        <h3 className="text-lg font-medium text-foreground">{t('dashboard.nodeList')}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{t('dashboard.nodeListDescription')}</p>
      </div>
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-border">
          <thead className="bg-muted">
            <tr>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('dashboard.nodeName')}</th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('nodes.ipAddress')}</th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('nodes.region')}</th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('dashboard.status')}</th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">{t('dashboard.health')}</th>
            </tr>
          </thead>
          <tbody className="bg-card divide-y divide-border">
            {safeNodes.map(node => {
              const nodeMetrics = metricsMap.get(node.id)
              const healthStatus = nodeMetrics
                ? determineHealthStatus({
                    latency_ms: nodeMetrics.latency_ms,
                    packet_loss_rate: nodeMetrics.packet_loss_rate,
                    jitter_ms: nodeMetrics.jitter_ms,
                    last_heartbeat: nodeMetrics.timestamp,
                  })
                : 'offline'

              return (
                <tr
                  key={node.id}
                  onClick={() => handleRowClick(node.id, node.name)}
                  className="hover:bg-accent/10 cursor-pointer transition-colors duration-150"
                >
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm font-medium text-foreground">{node.name}</div>
                    <div className="text-sm text-muted-foreground">{node.id}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="text-sm text-foreground">{node.ip}</div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary">{node.region}</span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${node.status === 'online' ? 'bg-healthy-bg text-healthy-text' : 'bg-muted text-muted-foreground'}`}>
                      {node.status === 'online' ? t('dashboard.online') : t('dashboard.offline')}
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
