/**
 * NodeComparisonTable Component
 *
 * Displays a side-by-side comparison of multiple nodes with their metrics.
 * Highlights differences and provides visual diff indicators.
 */

import { useTranslation } from 'react-i18next'

export interface NodeComparisonData {
  nodeId: string
  nodeName: string
  region: string
  status: 'online' | 'offline' | 'connecting'
  latency?: number
  packetLoss?: number
  jitter?: number
}

interface NodeComparisonTableProps {
  nodes: NodeComparisonData[]
  highlightDifferences?: boolean
}

// Color palette from UI design
const COLORS = {
  healthy: '#22C55E',
  warning: '#F59E0B',
  critical: '#EF4444',
  unknown: '#6B7280',
}

export function NodeComparisonTable({ nodes, highlightDifferences = true }: NodeComparisonTableProps) {
  const { t } = useTranslation()

  if (nodes.length === 0) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6 text-center">
        <p className="text-gray-500 dark:text-gray-400">{t('nodes.noNodes')}</p>
      </div>
    )
  }

  // Calculate best/worst values for highlighting
  const latencies = nodes.map((n) => n.latency).filter((v): v is number => v !== undefined)
  const packetLosses = nodes.map((n) => n.packetLoss).filter((v): v is number => v !== undefined)
  const jitters = nodes.map((n) => n.jitter).filter((v): v is number => v !== undefined)

  const bestLatency = latencies.length > 0 ? Math.min(...latencies) : undefined
  const worstLatency = latencies.length > 0 ? Math.max(...latencies) : undefined
  const bestPacketLoss = packetLosses.length > 0 ? Math.min(...packetLosses) : undefined
  const worstPacketLoss = packetLosses.length > 0 ? Math.max(...packetLosses) : undefined
  const bestJitter = jitters.length > 0 ? Math.min(...jitters) : undefined
  const worstJitter = jitters.length > 0 ? Math.max(...jitters) : undefined

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'online':
        return COLORS.healthy
      case 'offline':
        return COLORS.critical
      case 'connecting':
        return COLORS.warning
      default:
        return COLORS.unknown
    }
  }

  const getValueClass = (
    value: number | undefined,
    best: number | undefined,
    worst: number | undefined,
    isLowerBetter: boolean = true
  ): string => {
    if (!highlightDifferences || value === undefined || best === undefined || worst === undefined) {
      return ''
    }
    if (best === worst) return ''

    if (isLowerBetter) {
      if (value === best) return 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-300'
      if (value === worst) return 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300'
    } else {
      if (value === best) return 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-300'
      if (value === worst) return 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300'
    }
    return ''
  }

  const getLatencyStatus = (latency: number | undefined): 'good' | 'warning' | 'critical' => {
    if (latency === undefined) return 'warning'
    if (latency < 100) return 'good'
    if (latency < 200) return 'warning'
    return 'critical'
  }

  const getPacketLossStatus = (loss: number | undefined): 'good' | 'warning' | 'critical' => {
    if (loss === undefined) return 'warning'
    if (loss === 0) return 'good'
    if (loss < 2) return 'warning'
    return 'critical'
  }

  const getJitterStatus = (jitter: number | undefined): 'good' | 'warning' | 'critical' => {
    if (jitter === undefined) return 'warning'
    if (jitter < 20) return 'good'
    if (jitter < 50) return 'warning'
    return 'critical'
  }

  const getStatusIndicatorColor = (status: 'good' | 'warning' | 'critical'): string => {
    switch (status) {
      case 'good':
        return COLORS.healthy
      case 'warning':
        return COLORS.warning
      case 'critical':
        return COLORS.critical
    }
  }

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                {t('nodes.nodeName')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                {t('nodes.region')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                {t('common.status')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                {t('metrics.latency')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                {t('metrics.packetLoss')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                {t('metrics.jitter')}
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            {nodes.map((node) => (
              <tr key={node.nodeId} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {node.nodeName}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                  {node.region}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span
                    className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                    style={{
                      backgroundColor: `${getStatusColor(node.status)}20`,
                      color: getStatusColor(node.status),
                    }}
                  >
                    <span
                      className="w-1.5 h-1.5 rounded-full mr-1.5"
                      style={{ backgroundColor: getStatusColor(node.status) }}
                    />
                    {t(`status.${node.status}`)}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className={`flex items-center space-x-2 px-2 py-1 rounded ${getValueClass(node.latency, bestLatency, worstLatency)}`}>
                    <span
                      className="w-2 h-2 rounded-full"
                      style={{ backgroundColor: getStatusIndicatorColor(getLatencyStatus(node.latency)) }}
                    />
                    <span className="text-sm font-mono text-gray-900 dark:text-gray-100">
                      {node.latency !== undefined ? `${node.latency.toFixed(1)} ms` : 'N/A'}
                    </span>
                    {highlightDifferences && node.latency === bestLatency && bestLatency !== worstLatency && (
                      <span className="text-xs text-green-600 dark:text-green-400">Best</span>
                    )}
                    {highlightDifferences && node.latency === worstLatency && bestLatency !== worstLatency && (
                      <span className="text-xs text-red-600 dark:text-red-400">Worst</span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className={`flex items-center space-x-2 px-2 py-1 rounded ${getValueClass(node.packetLoss, bestPacketLoss, worstPacketLoss)}`}>
                    <span
                      className="w-2 h-2 rounded-full"
                      style={{ backgroundColor: getStatusIndicatorColor(getPacketLossStatus(node.packetLoss)) }}
                    />
                    <span className="text-sm font-mono text-gray-900 dark:text-gray-100">
                      {node.packetLoss !== undefined ? `${node.packetLoss.toFixed(2)}%` : 'N/A'}
                    </span>
                    {highlightDifferences && node.packetLoss === bestPacketLoss && bestPacketLoss !== worstPacketLoss && (
                      <span className="text-xs text-green-600 dark:text-green-400">Best</span>
                    )}
                    {highlightDifferences && node.packetLoss === worstPacketLoss && bestPacketLoss !== worstPacketLoss && (
                      <span className="text-xs text-red-600 dark:text-red-400">Worst</span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className={`flex items-center space-x-2 px-2 py-1 rounded ${getValueClass(node.jitter, bestJitter, worstJitter)}`}>
                    <span
                      className="w-2 h-2 rounded-full"
                      style={{ backgroundColor: getStatusIndicatorColor(getJitterStatus(node.jitter)) }}
                    />
                    <span className="text-sm font-mono text-gray-900 dark:text-gray-100">
                      {node.jitter !== undefined ? `${node.jitter.toFixed(1)} ms` : 'N/A'}
                    </span>
                    {highlightDifferences && node.jitter === bestJitter && bestJitter !== worstJitter && (
                      <span className="text-xs text-green-600 dark:text-green-400">Best</span>
                    )}
                    {highlightDifferences && node.jitter === worstJitter && bestJitter !== worstJitter && (
                      <span className="text-xs text-red-600 dark:text-red-400">Worst</span>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Legend */}
      {highlightDifferences && (
        <div className="px-6 py-3 bg-gray-50 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700">
          <div className="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
            <div className="flex items-center space-x-1">
              <span className="w-3 h-3 rounded bg-green-100 dark:bg-green-900/30" />
              <span>{t('metrics.good')}</span>
            </div>
            <div className="flex items-center space-x-1">
              <span className="w-3 h-3 rounded bg-red-100 dark:bg-red-900/30" />
              <span>{t('status.warning')}</span>
            </div>
            <span className="text-gray-300 dark:text-gray-600">|</span>
            <span>{t('nodes.comparison')} - {t('nodes.selectedCount', { count: nodes.length, max: 5 })}</span>
          </div>
        </div>
      )}
    </div>
  )
}
