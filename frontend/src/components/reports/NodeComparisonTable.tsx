/**
 * NodeComparisonTable Component
 *
 * Displays a side-by-side comparison of multiple nodes with their metrics.
 * Highlights differences and provides visual diff indicators.
 */

import { useTranslation } from 'react-i18next'
import { useThemeColors } from '../../hooks/useThemeColors'

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

// Color palette from UI design - now using theme colors
const getColors = (themeColors: ReturnType<typeof useThemeColors>) => ({
  healthy: themeColors.healthy,
  warning: themeColors.warning,
  critical: themeColors.critical,
  unknown: themeColors.unknown,
})

export function NodeComparisonTable({ nodes, highlightDifferences = true }: NodeComparisonTableProps) {
  const { t } = useTranslation()
  const themeColors = useThemeColors()

  if (nodes.length === 0) {
    return (
      <div className="bg-card rounded-lg shadow-sm p-6 text-center">
        <p className="text-muted-foreground">{t('nodes.noNodes')}</p>
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
    const colors = getColors(themeColors)
    switch (status) {
      case 'online':
        return colors.healthy
      case 'offline':
        return colors.critical
      case 'connecting':
        return colors.warning
      default:
        return colors.unknown
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
      if (value === best) return 'bg-healthy-bg text-healthy-text'
      if (value === worst) return 'bg-destructive/10 text-destructive'
    } else {
      if (value === best) return 'bg-healthy-bg text-healthy-text'
      if (value === worst) return 'bg-destructive/10 text-destructive'
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
    const colors = getColors(themeColors)
    switch (status) {
      case 'good':
        return colors.healthy
      case 'warning':
        return colors.warning
      case 'critical':
        return colors.critical
    }
  }

  return (
    <div className="bg-card rounded-lg shadow-sm overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-border">
          <thead className="bg-muted">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                {t('nodes.nodeName')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                {t('nodes.region')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                {t('common.status')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                {t('metrics.latency')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                {t('metrics.packetLoss')}
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                {t('metrics.jitter')}
              </th>
            </tr>
          </thead>
          <tbody className="bg-card divide-y divide-border">
            {nodes.map((node) => (
              <tr key={node.nodeId} className="hover:bg-accent/10">
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="text-sm font-medium text-foreground">
                    {node.nodeName}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
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
                    <span className="text-sm font-mono text-foreground">
                      {node.latency !== undefined ? `${node.latency.toFixed(1)} ms` : 'N/A'}
                    </span>
                    {highlightDifferences && node.latency === bestLatency && bestLatency !== worstLatency && (
                      <span className="text-xs text-healthy">Best</span>
                    )}
                    {highlightDifferences && node.latency === worstLatency && bestLatency !== worstLatency && (
                      <span className="text-xs text-destructive">Worst</span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className={`flex items-center space-x-2 px-2 py-1 rounded ${getValueClass(node.packetLoss, bestPacketLoss, worstPacketLoss)}`}>
                    <span
                      className="w-2 h-2 rounded-full"
                      style={{ backgroundColor: getStatusIndicatorColor(getPacketLossStatus(node.packetLoss)) }}
                    />
                    <span className="text-sm font-mono text-foreground">
                      {node.packetLoss !== undefined ? `${node.packetLoss.toFixed(2)}%` : 'N/A'}
                    </span>
                    {highlightDifferences && node.packetLoss === bestPacketLoss && bestPacketLoss !== worstPacketLoss && (
                      <span className="text-xs text-healthy">Best</span>
                    )}
                    {highlightDifferences && node.packetLoss === worstPacketLoss && bestPacketLoss !== worstPacketLoss && (
                      <span className="text-xs text-destructive">Worst</span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className={`flex items-center space-x-2 px-2 py-1 rounded ${getValueClass(node.jitter, bestJitter, worstJitter)}`}>
                    <span
                      className="w-2 h-2 rounded-full"
                      style={{ backgroundColor: getStatusIndicatorColor(getJitterStatus(node.jitter)) }}
                    />
                    <span className="text-sm font-mono text-foreground">
                      {node.jitter !== undefined ? `${node.jitter.toFixed(1)} ms` : 'N/A'}
                    </span>
                    {highlightDifferences && node.jitter === bestJitter && bestJitter !== worstJitter && (
                      <span className="text-xs text-healthy">Best</span>
                    )}
                    {highlightDifferences && node.jitter === worstJitter && bestJitter !== worstJitter && (
                      <span className="text-xs text-destructive">Worst</span>
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
        <div className="px-6 py-3 bg-muted border-t border-border">
          <div className="flex items-center space-x-4 text-xs text-muted-foreground">
            <div className="flex items-center space-x-1">
              <span className="w-3 h-3 rounded bg-healthy-bg" />
              <span>{t('metrics.good')}</span>
            </div>
            <div className="flex items-center space-x-1">
              <span className="w-3 h-3 rounded bg-destructive/10" />
              <span>{t('status.warning')}</span>
            </div>
            <span className="text-border">|</span>
            <span>{t('nodes.comparison')} - {t('nodes.selectedCount', { count: nodes.length, max: 5 })}</span>
          </div>
        </div>
      )}
    </div>
  )
}
