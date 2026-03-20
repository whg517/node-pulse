import { useEffect, useRef, useCallback } from 'react'
import echarts from '../../lib/echarts-core'
import type { ECharts, EChartsOption } from '../../lib/echarts-core'
import { useTranslation } from 'react-i18next'
import { useThemeColors } from '../../hooks/useThemeColors'

export type HealthStatus = 'healthy' | 'warning' | 'critical' | 'offline'

export interface NodeLocation {
  id: string
  name: string
  lat: number
  lng: number
  region: string
  healthStatus: HealthStatus
  avgLatency: number
  packetLoss: number
}

export interface WorldMapProps {
  nodes: NodeLocation[]
  onNodeClick?: (nodeId: string) => void
  height?: string
  className?: string
  isLoading?: boolean
  refreshInterval?: number
}

// Status color configuration - will use theme colors dynamically
const statusConfig = {
  healthy: {
    labelKey: 'status.healthy',
  },
  warning: {
    labelKey: 'status.warning',
  },
  critical: {
    labelKey: 'status.critical',
  },
  offline: {
    labelKey: 'status.offline',
  },
}

/**
 * WorldMap component for displaying node health distribution on an interactive map
 *
 * Features:
 * - Interactive world map with zoom and pan
 * - Node markers colored by health status
 * - Pulsing animation for critical nodes
 * - Tooltip on hover showing node details
 * - Click handler for node selection
 * - Responsive design with window resize handling
 *
 * @param props - WorldMap props
 * @returns WorldMap component
 *
 * @example
 * <WorldMap
 *   nodes={nodeLocations}
 *   onNodeClick={(nodeId) => console.log(nodeId)}
 *   height="500px"
 * />
 */
export default function WorldMap({
  nodes,
  onNodeClick,
  height = '400px',
  className = '',
  isLoading = false,
}: WorldMapProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<ECharts | null>(null)
  const { t } = useTranslation()
  const themeColors = useThemeColors()

  // Get status color
  const getStatusColor = useCallback((status: HealthStatus): string => {
    switch (status) {
      case 'healthy':
        return themeColors.healthy
      case 'warning':
        return themeColors.warning
      case 'critical':
        return themeColors.critical
      case 'offline':
        return themeColors.unknown
      default:
        return themeColors.unknown
    }
  }, [themeColors])

  // Get status label
  const getStatusLabel = useCallback(
    (status: HealthStatus): string => {
      const labelKey = statusConfig[status]?.labelKey || 'status.unknown'
      return t(labelKey)
    },
    [t]
  )

  // Initialize chart
  useEffect(() => {
    if (!chartRef.current) return

    // Only initialize if we have nodes and not loading
    if (!nodes || nodes.length === 0 || isLoading) {
      // Dispose existing chart if no data or loading
      if (chartInstance.current) {
        chartInstance.current.dispose()
        chartInstance.current = null
      }
      return
    }

    // Initialize ECharts instance if not already initialized
    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current)
    }

    // Cleanup on unmount
    return () => {
      if (chartInstance.current) {
        chartInstance.current.dispose()
        chartInstance.current = null
      }
    }
  }, [nodes, isLoading])

  // Update chart when nodes change
  useEffect(() => {
    if (!chartInstance.current || !nodes || nodes.length === 0) return

    // Separate nodes into regular and critical for effectScatter
    const regularNodes = nodes.filter((node) => node.healthStatus !== 'critical')
    const criticalNodes = nodes.filter((node) => node.healthStatus === 'critical')

    // Format node data for scatter series
    const formatNodeData = (node: NodeLocation) => ({
      name: node.name,
      value: [node.lng, node.lat, node.avgLatency, node.packetLoss],
      id: node.id,
      itemStyle: {
        color: getStatusColor(node.healthStatus),
      },
      status: node.healthStatus,
      region: node.region,
    })

    const option: EChartsOption = {
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'item',
        formatter: (params: unknown) => {
          const p = params as {
            name?: string
            value?: number[]
            data?: { status?: HealthStatus; region?: string }
          }
          if (!p || !p.value) return ''

          const [, , avgLatency, packetLoss] = p.value
          const status: HealthStatus = p.data?.status || 'offline'
          const region = p.data?.region || 'Unknown'

          return `
            <div style="padding: 8px;">
              <div style="font-weight: bold; margin-bottom: 4px;">${p.name || 'Unknown'}</div>
              <div style="color: #666; margin-bottom: 4px;">${t('nodes.region')}: ${region}</div>
              <div style="display: flex; align-items: center; margin-bottom: 2px;">
                <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: ${getStatusColor(status)}; margin-right: 6px;"></span>
                <span>${t('common.status')}: ${getStatusLabel(status)}</span>
              </div>
              <div>${t('metrics.avgLatency')}: ${avgLatency?.toFixed(1) ?? 'N/A'} ${t('units.ms')}</div>
              <div>${t('metrics.packetLoss')}: ${packetLoss?.toFixed(1) ?? 'N/A'}${t('units.percent')}</div>
            </div>
          `
        },
      },
      geo: {
        map: 'world',
        roam: true,
        zoom: 1.2,
        center: [0, 20],
        scaleLimit: {
          min: 1,
          max: 10,
        },
        itemStyle: {
          areaColor: '#f3f4f6',
          borderColor: '#d1d5db',
          borderWidth: 0.5,
        },
        emphasis: {
          itemStyle: {
            areaColor: '#e5e7eb',
          },
          label: {
            show: false,
          },
        },
        select: {
          disabled: true,
        },
      },
      series: [
        // Regular nodes as scatter
        {
          name: t('nodes.title'),
          type: 'scatter',
          coordinateSystem: 'geo',
          data: regularNodes.map(formatNodeData),
          symbolSize: 14,
          itemStyle: {
            borderWidth: 2,
            borderColor: '#fff',
          },
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowColor: 'rgba(0, 0, 0, 0.3)',
            },
          },
        },
        // Critical nodes with pulsing effect
        {
          name: t('status.critical'),
          type: 'effectScatter',
          coordinateSystem: 'geo',
          data: criticalNodes.map(formatNodeData),
          symbolSize: 16,
          showEffectOn: 'render',
          rippleEffect: {
            brushType: 'stroke',
            scale: 3,
            period: 4,
          },
          itemStyle: {
            color: themeColors.critical,
            borderWidth: 2,
            borderColor: '#fff',
            shadowBlur: 10,
            shadowColor: 'rgba(220, 38, 38, 0.5)',
          },
          emphasis: {
            itemStyle: {
              shadowBlur: 15,
              shadowColor: 'rgba(220, 38, 38, 0.7)',
            },
          },
        },
      ],
    }

    chartInstance.current.setOption(option, true)

    // Handle click events
    chartInstance.current.on('click', (params: unknown) => {
      const p = params as { seriesType?: string; data?: { id?: string } }
      if (p.seriesType === 'scatter' || p.seriesType === 'effectScatter') {
        const nodeId = p.data?.id
        if (nodeId && onNodeClick) {
          onNodeClick(nodeId)
        }
      }
    })
  }, [nodes, getStatusColor, getStatusLabel, onNodeClick, t])

  // Handle window resize
  useEffect(() => {
    const handleResize = () => {
      if (chartInstance.current) {
        chartInstance.current.resize()
      }
    }

    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
    }
  }, [])

  // Register world map (simple GeoJSON for world outline)
  useEffect(() => {
    // Check if world map is already registered
    if (!echarts.getMap('world')) {
      // Simple world map GeoJSON outline
      const worldJson = {
        type: 'FeatureCollection',
        features: [
          {
            type: 'Feature',
            properties: { name: 'World' },
            geometry: {
              type: 'MultiPolygon',
              coordinates: [
                // Simplified world outline coordinates
                [
                  [
                    [-180, -60],
                    [-180, 90],
                    [180, 90],
                    [180, -60],
                    [-180, -60],
                  ],
                ],
              ],
            },
          },
        ],
      }
      echarts.registerMap('world', worldJson as Parameters<typeof echarts.registerMap>[1])
    }
  }, [])

  // Render loading state
  if (isLoading) {
    return (
      <div
        className={`world-map bg-[var(--color-bg-surface)] rounded-lg shadow-sm p-4 ${className}`}
        role="region"
        aria-label={t('dashboard.nodeDistribution')}
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
            {t('dashboard.nodeDistribution')}
          </h3>
        </div>
        <div
          className="relative flex items-center justify-center"
          style={{ height }}
          role="status"
          aria-label={t('common.loading')}
        >
          <div className="flex flex-col items-center">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--color-brand)]" />
            <p className="mt-2 text-[var(--color-text-secondary)]">{t('common.loading')}</p>
          </div>
        </div>
      </div>
    )
  }

  // Render empty state
  if (!nodes || nodes.length === 0) {
    return (
      <div
        className={`world-map bg-[var(--color-bg-surface)] rounded-lg shadow-sm p-4 ${className}`}
        role="region"
        aria-label={t('dashboard.nodeDistribution')}
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
            {t('dashboard.nodeDistribution')}
          </h3>
        </div>
        <div
          className="relative flex items-center justify-center"
          style={{ height }}
          role="img"
          aria-label={t('dashboard.noData')}
        >
          <div className="text-center">
            <svg
              className="mx-auto h-12 w-12 text-[var(--color-text-muted)]"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-[var(--color-text-primary)]">{t('nodes.noNodes')}</h3>
            <p className="mt-1 text-sm text-[var(--color-text-secondary)]">{t('dashboard.noData')}</p>
          </div>
        </div>
      </div>
    )
  }

  // Count nodes by status for legend
  const statusCounts = {
    healthy: nodes.filter((n) => n.healthStatus === 'healthy').length,
    warning: nodes.filter((n) => n.healthStatus === 'warning').length,
    critical: nodes.filter((n) => n.healthStatus === 'critical').length,
    offline: nodes.filter((n) => n.healthStatus === 'offline').length,
  }

  return (
    <div
      className={`world-map bg-[var(--color-bg-surface)] rounded-lg shadow-sm p-4 ${className}`}
      role="region"
      aria-label={t('dashboard.nodeDistribution')}
    >
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
          {t('dashboard.nodeDistribution')}
        </h3>
        <div className="text-sm text-[var(--color-text-muted)]">
          {nodes.length} {t('metrics.totalNodes').toLowerCase()}
        </div>
      </div>

      {/* Chart Container */}
      <div
        ref={chartRef}
        className="relative"
        style={{ height }}
        role="img"
        aria-label={`${t('dashboard.nodeDistribution')} showing ${nodes.length} nodes`}
        tabIndex={0}
      />

      {/* Legend */}
      <div className="mt-4 flex items-center justify-center flex-wrap gap-4 text-sm">
        {Object.entries(statusCounts).map(([status, count]) => (
          <div key={status} className="flex items-center space-x-2">
            <div
              className="w-3 h-3 rounded-full"
              style={{ backgroundColor: getStatusColor(status as HealthStatus) }}
              aria-hidden="true"
            />
            <span className="text-[var(--color-text-secondary)]">
              {getStatusLabel(status as HealthStatus)} ({count})
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}
