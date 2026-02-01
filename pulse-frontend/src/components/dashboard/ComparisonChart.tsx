import { useEffect, useMemo, useRef, useState } from 'react'
import * as echarts from 'echarts'

export type TimeRange = '24h' | '7d' | '30d'
export type MetricType = 'latency_ms' | 'packet_loss_rate' | 'jitter_ms'
export type GroupByType = 'region' | 'isp' | 'none'

export interface ComparisonDataPoint {
  timestamp: string
  value: number
}

export interface NodeComparisonData {
  node_id: string
  node_name: string
  region?: string
  isp?: string
  data: ComparisonDataPoint[]
}

export interface ComparisonChartProps {
  nodes: NodeComparisonData[]  // 2-5 nodes
  metric: MetricType
  timeRange: TimeRange
  showStatistics?: boolean  // Show avg/max/min
  highlightDifferences?: boolean  // Highlight significant differences
  groupBy?: GroupByType
  height?: string
  className?: string
  onTimeRangeChange?: (range: TimeRange) => void
  isLoading?: boolean
}

interface TooltipParams {
  name: string
  value: number
  seriesName: string
  marker: string
  data?: any
}

interface StatisticsResult {
  avg: number
  max: number
  min: number
  diff: number
  diffPercent: number
  maxNode?: string
  minNode?: string
}

/**
 * ComparisonChart component for displaying multi-node time-series comparison using ECharts
 *
 * Features:
 * - Multi-node comparison (2-5 nodes)
 * - Time range selection (24h/7d/30d)
 * - Multi-metric support (latency/packet loss/jitter)
 * - Statistics panel (avg/max/min/diff)
 * - Difference highlighting with color coding
 * - Grouping by region or ISP
 * - Hover tooltips with all node values
 * - Zoom functionality (mouse wheel)
 * - Responsive design
 *
 * @param props - ComparisonChart props
 * @returns ComparisonChart component
 *
 * @example
 * <ComparisonChart
 *   nodes={nodeData}
 *   metric="latency_ms"
 *   timeRange="7d"
 *   showStatistics={true}
 *   highlightDifferences={true}
 *   groupBy="region"
 *   onTimeRangeChange={(range) => console.log(range)}
 * />
 */
export default function ComparisonChart({
  nodes,
  metric,
  timeRange,
  showStatistics = true,
  highlightDifferences = true,
  groupBy = 'none',
  height = '400px',
  className = '',
  onTimeRangeChange,
  isLoading = false,
}: ComparisonChartProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<echarts.ECharts | null>(null)
  const [localTimeRange, setLocalTimeRange] = useState<TimeRange>(timeRange)

  // Metric configuration
  const metricConfig = {
    latency_ms: {
      label: 'Latency',
      unit: 'ms',
      color: '#3b82f6', // blue-500
      yAxisLabel: 'Latency (ms)',
      warningThreshold: 0.2, // 20% difference
      criticalThreshold: 0.5, // 50% difference
    },
    packet_loss_rate: {
      label: 'Packet Loss Rate',
      unit: '%',
      color: '#ef4444', // red-500
      yAxisLabel: 'Packet Loss Rate (%)',
      warningThreshold: 0.05, // 5% difference
      criticalThreshold: 0.1, // 10% difference
    },
    jitter_ms: {
      label: 'Jitter',
      unit: 'ms',
      color: '#8b5cf6', // purple-500
      yAxisLabel: 'Jitter (ms)',
      warningThreshold: 0.3, // 30% difference
      criticalThreshold: 0.6, // 60% difference
    },
  }

  const config = metricConfig[metric]

  // Color palette for multiple nodes
  const nodeColors = [
    '#3b82f6', // blue-500
    '#10b981', // green-500
    '#f59e0b', // amber-500
    '#ef4444', // red-500
    '#8b5cf6', // purple-500
  ]

  // Time range options
  const timeRangeOptions: { value: TimeRange; label: string }[] = [
    { value: '24h', label: '24 Hours' },
    { value: '7d', label: '7 Days' },
    { value: '30d', label: '30 Days' },
  ]

  // Handle time range change
  const handleTimeRangeChange = (newRange: TimeRange) => {
    setLocalTimeRange(newRange)
    if (onTimeRangeChange) {
      onTimeRangeChange(newRange)
    }
  }

  // Format X-axis labels based on time range
  const getXAxisFormatter = (): string => {
    switch (localTimeRange) {
      case '24h':
        return '{HH}:{mm}'
      case '7d':
        return '{MM}-{dd}\n{HH}:{mm}'
      case '30d':
        return '{MM}-{dd}'
      default:
        return '{HH}:{mm}'
    }
  }

  // Pre-calculate all statistics for all timestamps (performance optimization)
  const allStatistics = useMemo(() => {
    if (!nodes || nodes.length === 0) return {}

    // Get all unique timestamps from all nodes
    const allTimestamps = Array.from(
      new Set(nodes.flatMap((node) => node.data.map((d) => d.timestamp)))
    ).sort()

    const stats: Record<string, StatisticsResult> = {}

    allTimestamps.forEach((timestamp) => {
      const valuesWithNodes = nodes
        .map((node) => ({
          node: node.node_name,
          value: node.data.find((d) => d.timestamp === timestamp)?.value,
        }))
        .filter((v): v is { node: string; value: number } => v.value !== undefined)

      if (valuesWithNodes.length === 0) {
        stats[timestamp] = {
          avg: 0,
          max: 0,
          min: 0,
          diff: 0,
          diffPercent: 0,
        }
        return
      }

      const values = valuesWithNodes.map((v) => v.value)
      const avg = values.reduce((sum, v) => sum + v, 0) / values.length
      const max = Math.max(...values)
      const min = Math.min(...values)
      const diff = max - min
      const diffPercent = avg > 0 ? (diff / avg) * 100 : 0

      // Find which node has max and min
      const maxNode = valuesWithNodes.find((v) => v.value === max)?.node
      const minNode = valuesWithNodes.find((v) => v.value === min)?.node

      stats[timestamp] = { avg, max, min, diff, diffPercent, maxNode, minNode }
    })

    return stats
  }, [nodes])

  // Calculate statistics for a specific timestamp (used in tooltip)
  const calculateStatistics = (timestamp: string): StatisticsResult | null => {
    return allStatistics[timestamp] || null
  }

  // Detect outlier data points for markPoint highlighting
  const outlierDataPoints = useMemo(() => {
    if (!highlightDifferences || !nodes || nodes.length === 0) return {}

    const outliers: Record<string, Array<{ timestamp: string; value: number; type: 'max' | 'min' }>> = {}

    Object.entries(allStatistics).forEach(([timestamp, stats]) => {
      if (!stats.avg) return

      // Check if difference exceeds warning threshold
      if (stats.diffPercent > config.warningThreshold * 100) {
        nodes.forEach((node) => {
          const point = node.data.find((d) => d.timestamp === timestamp)
          if (!point) return

          // Mark this node if it has the max or min value
          if (point.value === stats.max && stats.maxNode === node.node_name) {
            if (!outliers[node.node_id]) outliers[node.node_id] = []
            outliers[node.node_id].push({ timestamp, value: point.value, type: 'max' })
          } else if (point.value === stats.min && stats.minNode === node.node_name) {
            if (!outliers[node.node_id]) outliers[node.node_id] = []
            outliers[node.node_id].push({ timestamp, value: point.value, type: 'min' })
          }
        })
      }
    })

    return outliers
  }, [highlightDifferences, nodes, allStatistics, config.warningThreshold])

  // Initialize chart
  useEffect(() => {
    if (!chartRef.current) return

    // Initialize ECharts instance
    chartInstance.current = echarts.init(chartRef.current)

    // Cleanup on unmount
    return () => {
      try {
        if (chartInstance.current) {
          chartInstance.current.dispose()
          chartInstance.current = null
        }
      } catch (error) {
        console.warn('Error disposing ECharts instance:', error)
      }
    }
  }, [])

  // Update chart when data or time range changes
  useEffect(() => {
    if (!chartInstance.current || !nodes || nodes.length === 0) return

    // Validate node count
    if (nodes.length < 2 || nodes.length > 5) {
      console.warn('ComparisonChart requires 2-5 nodes')
      return
    }

    // Get all unique timestamps from all nodes
    const allTimestamps = Array.from(
      new Set(
        nodes.flatMap((node) => node.data.map((d) => d.timestamp))
      )
    ).sort()

    // Build series for each node
    const series = nodes.map((node, index) => {
      const nodeOutliers = outlierDataPoints[node.node_id] || []

      return {
        name: node.node_name,
        type: 'line' as const,
        data: allTimestamps.map((timestamp) => {
          const point = node.data.find((d) => d.timestamp === timestamp)
          return point ? point.value : null
        }),
        smooth: true,
        lineStyle: {
          color: nodeColors[index],
          width: 2,
        },
        itemStyle: {
          color: nodeColors[index],
        },
        // Add markPoints for outlier highlighting when enabled
        ...(highlightDifferences && nodeOutliers.length > 0
          ? {
              markPoint: {
                symbol: 'circle',
                symbolSize: 10,
                data: nodeOutliers.map((outlier) => ({
                  name: outlier.type === 'max' ? '▲ Max' : '▼ Min',
                  coord: [outlier.timestamp, outlier.value],
                  itemStyle: {
                    color:
                      outlier.type === 'max'
                        ? config.criticalThreshold * 100 < 50
                          ? '#ef4444'
                          : '#f59e0b'
                        : '#3b82f6',
                  },
                  label: {
                    show: true,
                    formatter: outlier.type === 'max' ? '▲' : '▼',
                    fontSize: 14,
                    offset: [0, -5],
                  },
                })),
              },
            }
          : {}),
      }
    })

    // Build group information for legend
    const getGroupLabel = (node: NodeComparisonData): string => {
      if (groupBy === 'region' && node.region) return ` (${node.region})`
      if (groupBy === 'isp' && node.isp) return ` (${node.isp})`
      return ''
    }

    const legendData = nodes.map((node) => node.node_name + getGroupLabel(node))

    const option: echarts.EChartsOption = {
      title: {
        text: `${config.label} Comparison`,
        left: 'center',
        textStyle: {
          fontSize: 16,
          fontWeight: 'bold',
        },
      },
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          if (!params || params.length === 0) return ''
          const timestamp = params[0].name
          const date = new Date(timestamp)
          const formattedDate = date.toLocaleString()

          let tooltip = `${formattedDate}<br/><br/>`

          // Add all node values
          params.forEach((param: any) => {
            tooltip += `${param.marker}${param.seriesName}: ${param.value} ${config.unit}<br/>`
          })

          // Add statistics if enabled
          if (showStatistics) {
            const stats = calculateStatistics(timestamp)
            if (stats) {
              tooltip += `<br/><strong>Statistics:</strong><br/>`
              tooltip += `Avg: ${stats.avg.toFixed(2)} ${config.unit}<br/>`
              tooltip += `Max: ${stats.max.toFixed(2)} ${config.unit}${stats.maxNode ? ` (${stats.maxNode})` : ''}<br/>`
              tooltip += `Min: ${stats.min.toFixed(2)} ${config.unit}${stats.minNode ? ` (${stats.minNode})` : ''}<br/>`
              tooltip += `Diff: ${stats.diff.toFixed(2)} ${config.unit} (${stats.diffPercent.toFixed(1)}%)`
            }
          }

          return tooltip
        },
      },
      legend: {
        data: legendData,
        top: 40,
        type: 'scroll',
      },
      toolbox: {
        feature: {
          dataZoom: {
            yAxisIndex: 'none',
          },
          restore: {},
          saveAsImage: {},
        },
        right: 20,
        top: 10,
      },
      grid: {
        left: '60px',
        right: '60px',
        bottom: '80px',
        top: '100px',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: allTimestamps,
        axisLabel: {
          formatter: getXAxisFormatter(),
          rotate: localTimeRange === '7d' ? 45 : 0,
        },
      },
      yAxis: {
        type: 'value',
        name: config.yAxisLabel,
        nameLocation: 'middle',
        nameGap: 50,
        axisLabel: {
          formatter: (value: number) => value.toFixed(2),
        },
      },
      dataZoom: [
        {
          type: 'inside',
          start: 0,
          end: 100,
        },
        {
          start: 0,
          end: 100,
          height: 20,
          bottom: 10,
        },
      ],
      series,
    }

    chartInstance.current.setOption(option, true)
  }, [nodes, localTimeRange, config, showStatistics, groupBy, highlightDifferences])

  // Handle window resize with debounce for performance
  useEffect(() => {
    let resizeTimer: number | undefined = undefined

    const handleResize = () => {
      if (resizeTimer !== undefined) {
        clearTimeout(resizeTimer)
      }
      resizeTimer = window.setTimeout(() => {
        if (chartInstance.current) {
          chartInstance.current.resize()
        }
      }, 150) // Debounce by 150ms
    }

    window.addEventListener('resize', handleResize)

    return () => {
      if (resizeTimer !== undefined) {
        clearTimeout(resizeTimer)
      }
      window.removeEventListener('resize', handleResize)
    }
  }, [])

  // Calculate overall statistics
  const getOverallStatistics = () => {
    if (!showStatistics || !nodes || nodes.length === 0) return null

    const allValues = nodes.flatMap((node) => node.data.map((d) => d.value))
    if (allValues.length === 0) return null

    const avg = allValues.reduce((sum, v) => sum + v, 0) / allValues.length
    const max = Math.max(...allValues)
    const min = Math.min(...allValues)
    const diff = max - min
    const diffPercent = avg > 0 ? (diff / avg) * 100 : 0

    return { avg, max, min, diff, diffPercent }
  }

  const overallStats = getOverallStatistics()

  return (
    <div
      className={`comparison-chart bg-white rounded-lg shadow-sm p-4 ${className}`}
      role="region"
      aria-label={`${config.label} comparison chart`}
    >
      {/* Time Range Selector */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900">{config.label} Comparison</h3>
        <div className="flex space-x-2" role="group" aria-label="Time range selector">
          {timeRangeOptions.map((option) => (
            <button
              key={option.value}
              onClick={() => handleTimeRangeChange(option.value)}
              className={`px-3 py-1 rounded text-sm font-medium transition-colors ${
                localTimeRange === option.value
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
              aria-pressed={localTimeRange === option.value}
              disabled={isLoading}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* Statistics Panel - only show when there are nodes */}
      {showStatistics && overallStats && nodes && nodes.length > 0 && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg">
          <div className="grid grid-cols-5 gap-4 text-center">
            <div>
              <div className="text-sm text-gray-600">Average</div>
              <div className="text-lg font-semibold text-gray-900">
                {overallStats.avg.toFixed(2)} {config.unit}
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Maximum</div>
              <div className="text-lg font-semibold text-gray-900">
                {overallStats.max.toFixed(2)} {config.unit}
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Minimum</div>
              <div className="text-lg font-semibold text-gray-900">
                {overallStats.min.toFixed(2)} {config.unit}
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Difference</div>
              <div className="text-lg font-semibold text-gray-900">
                {overallStats.diff.toFixed(2)} {config.unit}
              </div>
            </div>
            <div>
              <div className="text-sm text-gray-600">Diff %</div>
              <div className={`text-lg font-semibold ${
                overallStats.diffPercent > config.criticalThreshold * 100
                  ? 'text-red-600'
                  : overallStats.diffPercent > config.warningThreshold * 100
                  ? 'text-yellow-600'
                  : 'text-green-600'
              }`}>
                {overallStats.diffPercent.toFixed(1)}%
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Chart Container */}
      <div
        ref={chartRef}
        className="relative"
        style={{ height }}
        role="img"
        aria-label={`${config.label} comparison chart showing ${nodes.length} nodes`}
      >
        {/* Loading Overlay */}
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-white bg-opacity-75 z-10">
            <div className="flex flex-col items-center">
              <div
                className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
                role="status"
                aria-label="Loading chart data"
              />
              <p className="mt-2 text-gray-600">Loading chart data...</p>
            </div>
          </div>
        )}

        {/* Empty State */}
        {!isLoading && (!nodes || nodes.length === 0) && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="text-center">
              <svg
                className="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-gray-900">No Data Available</h3>
              <p className="mt-1 text-sm text-gray-500">
                No comparison data available for the selected time range.
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Legend - only show when there are nodes */}
      {nodes && nodes.length > 0 && (
        <div className="mt-4 flex items-center justify-center space-x-6 text-sm flex-wrap gap-y-2">
          {nodes.map((node, index) => (
            <div key={node.node_id} className="flex items-center space-x-2">
              <div
                className="w-4 h-1"
                style={{ backgroundColor: nodeColors[index] }}
                aria-hidden="true"
              />
              <span className="text-gray-700">
                {node.node_name}
                {groupBy === 'region' && node.region && ` (${node.region})`}
                {groupBy === 'isp' && node.isp && ` (${node.isp})`}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
