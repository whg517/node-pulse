import { useEffect, useMemo, useRef, useState } from 'react'
import * as echarts from 'echarts'

export type TimeRange = '24h' | '7d' | '30d' | 'custom'
export type MetricType = 'latency_ms' | 'packet_loss_rate' | 'jitter_ms'
export type GroupByType = 'region' | 'isp' | 'none'
export type ComparisonMode = 'node' | 'timeRange'

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

export interface TimeRangeComparisonData {
  baseline: {
    start: string
    end: string
    label?: string
    data: ComparisonDataPoint[]
  }
  current: {
    start: string
    end: string
    label?: string
    data: ComparisonDataPoint[]
  }
  metric: MetricType
}

export interface StatisticalSummary {
  avg: number
  median: number // P50
  p95: number
  p99: number
  min: number
  max: number
  stdDev?: number
}

export interface ComparisonModeChange {
  mode: ComparisonMode
  nodes?: NodeComparisonData[]
  timeRangeData?: TimeRangeComparisonData
}

export interface ComparisonChartProps {
  nodes?: NodeComparisonData[]  // 2-5 nodes (for node comparison mode)
  timeRangeData?: TimeRangeComparisonData  // For time range comparison mode
  mode?: ComparisonMode  // Default to 'node' for backward compatibility
  metric: MetricType
  timeRange?: TimeRange  // For node comparison mode
  showStatistics?: boolean  // Show avg/max/min
  highlightDifferences?: boolean  // Highlight significant differences
  groupBy?: GroupByType
  height?: string
  className?: string
  onTimeRangeChange?: (range: TimeRange) => void
  onModeChange?: (change: ComparisonModeChange) => void
  isLoading?: boolean
  showPercentileStats?: boolean  // Show P50, P95, P99 values
  onExportPdf?: () => void  // Callback for PDF export
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
 * - Time range comparison (baseline vs current)
 * - Comparison mode toggle (node vs time range)
 * - Time range selection (24h/7d/30d/custom)
 * - Multi-metric support (latency/packet loss/jitter)
 * - Statistics panel (avg/max/min/diff)
 * - Percentile statistics (P50, P95, P99)
 * - Percentage change calculation
 * - Difference highlighting with color coding
 * - Grouping by region or ISP
 * - Hover tooltips with all node values
 * - Zoom functionality (mouse wheel)
 * - PDF export functionality
 * - Responsive design
 * - WCAG 2.1 AA accessibility compliance
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
  timeRangeData,
  mode = 'node',
  metric,
  timeRange,
  showStatistics = true,
  highlightDifferences = true,
  groupBy = 'none',
  height = '400px',
  className = '',
  onTimeRangeChange,
  onModeChange,
  isLoading = false,
  showPercentileStats = false,
  onExportPdf,
}: ComparisonChartProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<echarts.ECharts | null>(null)
  const [localTimeRange, setLocalTimeRange] = useState<TimeRange>(timeRange || '24h')
  const [localMode, setLocalMode] = useState<ComparisonMode>(mode)

  // Time range comparison state
  const [baselineStart, setBaselineStart] = useState('')
  const [baselineEnd, setBaselineEnd] = useState('')
  const [currentStart, setCurrentStart] = useState('')
  const [currentEnd, setCurrentEnd] = useState('')
  const [customBaselineTimeRange, setCustomBaselineTimeRange] = useState<TimeRange>('7d')
  const [customCurrentTimeRange, setCustomCurrentTimeRange] = useState<TimeRange>('7d')

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

  // Calculate percentile value from sorted array
  const calculatePercentile = (sortedValues: number[], percentile: number): number => {
    if (sortedValues.length === 0) return 0
    const index = (percentile / 100) * (sortedValues.length - 1)
    const lower = Math.floor(index)
    const upper = Math.ceil(index)
    if (lower === upper) return sortedValues[lower]
    const weight = index - lower
    return sortedValues[lower] * (1 - weight) + sortedValues[upper] * weight
  }

  // Calculate statistical summary with percentiles
  const calculateStatisticalSummary = (data: ComparisonDataPoint[]): StatisticalSummary | null => {
    if (!data || data.length === 0) return null

    const values = data.map((d) => d.value).sort((a, b) => a - b)
    const n = values.length

    const avg = values.reduce((sum, v) => sum + v, 0) / n
    const min = values[0]
    const max = values[n - 1]
    const median = calculatePercentile(values, 50)
    const p95 = calculatePercentile(values, 95)
    const p99 = calculatePercentile(values, 99)

    // Calculate standard deviation
    const variance = values.reduce((sum, v) => sum + Math.pow(v - avg, 2), 0) / n
    const stdDev = Math.sqrt(variance)

    return { avg, median, p95, p99, min, max, stdDev }
  }

  // Calculate percentage change between baseline and current
  const calculatePercentageChange = (baseline: number, current: number): number => {
    if (baseline === 0) return current > 0 ? 100 : 0
    return ((current - baseline) / Math.abs(baseline)) * 100
  }

  // Get time range dates based on selection
  const getTimeRangeDates = (range: TimeRange): { start: string; end: string } => {
    const now = new Date()
    const end = now.toISOString()
    let start: Date

    switch (range) {
      case '24h':
        start = new Date(now.getTime() - 24 * 60 * 60 * 1000)
        break
      case '7d':
        start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
        break
      case '30d':
        start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
        break
      case 'custom':
        start = new Date(baselineStart || now.getTime() - 7 * 24 * 60 * 60 * 1000)
        break
      default:
        start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
    }

    return {
      start: start.toISOString(),
      end: range === 'custom' ? (baselineEnd || end) : end,
    }
  }


  // Time range options
  const timeRangeOptions: { value: TimeRange; label: string }[] = [
    { value: '24h', label: '24 Hours' },
    { value: '7d', label: '7 Days' },
    { value: '30d', label: '30 Days' },
    { value: 'custom', label: 'Custom' },
  ]

  // Handle mode change
  const handleModeChange = (newMode: ComparisonMode) => {
    setLocalMode(newMode)
    if (onModeChange) {
      onModeChange({ mode: newMode, nodes: nodes || [], timeRangeData })
    }
  }

  // Handle PDF export using browser print
  const handleExportPdf = () => {
    if (onExportPdf) {
      onExportPdf()
    } else {
      // Fallback to browser print
      window.print()
    }
  }

  // Handle baseline time range change
  const handleBaselineTimeRangeChange = (range: TimeRange) => {
    setCustomBaselineTimeRange(range)
    if (range !== 'custom') {
      const dates = getTimeRangeDates(range)
      setBaselineStart(dates.start)
      setBaselineEnd(dates.end)
    }
  }

  // Handle current time range change
  const handleCurrentTimeRangeChange = (range: TimeRange) => {
    setCustomCurrentTimeRange(range)
    if (range !== 'custom') {
      const dates = getTimeRangeDates(range)
      setCurrentStart(dates.start)
      setCurrentEnd(dates.end)
    }
  }
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
    if (localMode === 'timeRange' || !nodes || nodes.length === 0) return {}

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
  }, [nodes, localMode])

  // Calculate statistics for time range comparison
  const timeRangeStatistics = useMemo(() => {
    if (localMode !== 'timeRange' || !timeRangeData) return null

    const baselineStats = calculateStatisticalSummary(timeRangeData.baseline.data)
    const currentStats = calculateStatisticalSummary(timeRangeData.current.data)

    if (!baselineStats || !currentStats) return null

    return {
      baseline: baselineStats,
      current: currentStats,
      changes: {
        avg: calculatePercentageChange(baselineStats.avg, currentStats.avg),
        median: calculatePercentageChange(baselineStats.median, currentStats.median),
        p95: calculatePercentageChange(baselineStats.p95, currentStats.p95),
        p99: calculatePercentageChange(baselineStats.p99, currentStats.p99),
        min: calculatePercentageChange(baselineStats.min, currentStats.min),
        max: calculatePercentageChange(baselineStats.max, currentStats.max),
      },
    }
  }, [localMode, timeRangeData])

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
  }, [nodes, localTimeRange, config, showStatistics, groupBy, highlightDifferences, localMode, timeRangeData, timeRangeStatistics])

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
      {/* Comparison Mode Toggle */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900">{config.label} Comparison</h3>
        <div className="flex items-center space-x-4">
          {/* Mode Toggle */}
          <div className="flex items-center space-x-2" role="group" aria-label="Comparison mode selector">
            <button
              onClick={() => handleModeChange('node')}
              className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                localMode === 'node'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
              aria-pressed={localMode === 'node'}
              disabled={isLoading}
            >
              Node Comparison
            </button>
            <button
              onClick={() => handleModeChange('timeRange')}
              className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                localMode === 'timeRange'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
              aria-pressed={localMode === 'timeRange'}
              disabled={isLoading}
            >
              Time Range Comparison
            </button>
          </div>

          {/* PDF Export Button */}
          <button
            onClick={handleExportPdf}
            disabled={isLoading}
            className="flex items-center space-x-2 px-3 py-1.5 rounded text-sm font-medium transition-colors bg-gray-100 text-gray-700 hover:bg-gray-200 disabled:opacity-50"
            aria-label="Export chart as PDF"
            title="Export as PDF"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <span>Export PDF</span>
          </button>
        </div>
      </div>

      {/* Time Range Comparison Mode - Dual Time Range Selectors */}
      {localMode === 'timeRange' && (
        <div className="mb-4 p-4 bg-gray-50 rounded-lg" role="group" aria-label="Time range selection">
          <div className="grid grid-cols-2 gap-6">
            {/* Baseline Time Range */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2" id="baseline-label">
                Baseline Period
              </label>
              <div className="space-y-2">
                <div className="flex space-x-2" role="group" aria-labelledby="baseline-label">
                  {['24h', '7d', '30d', 'custom'].map((range) => (
                    <button
                      key={range}
                      onClick={() => handleBaselineTimeRangeChange(range as TimeRange)}
                      className={`px-2 py-1 rounded text-xs font-medium transition-colors ${
                        customBaselineTimeRange === range
                          ? 'bg-green-600 text-white'
                          : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                      }`}
                      aria-pressed={customBaselineTimeRange === range}
                      disabled={isLoading}
                    >
                      {range === '24h' ? '24 Hours' : range === '7d' ? '7 Days' : range === '30d' ? '30 Days' : 'Custom'}
                    </button>
                  ))}
                </div>
                {customBaselineTimeRange === 'custom' && (
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label htmlFor="baselineStart" className="sr-only">Baseline Start Date</label>
                      <input
                        id="baselineStart"
                        type="date"
                        value={baselineStart ? baselineStart.split('T')[0] : ''}
                        onChange={(e) => setBaselineStart(e.target.value ? new Date(e.target.value).toISOString() : '')}
                        max={baselineEnd || new Date().toISOString().split('T')[0]}
                        disabled={isLoading}
                        className="w-full px-2 py-1 text-xs border border-gray-300 rounded focus:ring-green-500 focus:border-green-500"
                      />
                    </div>
                    <div>
                      <label htmlFor="baselineEnd" className="sr-only">Baseline End Date</label>
                      <input
                        id="baselineEnd"
                        type="date"
                        value={baselineEnd ? baselineEnd.split('T')[0] : ''}
                        onChange={(e) => setBaselineEnd(e.target.value ? new Date(e.target.value).toISOString() : '')}
                        min={baselineStart}
                        max={new Date().toISOString().split('T')[0]}
                        disabled={isLoading}
                        className="w-full px-2 py-1 text-xs border border-gray-300 rounded focus:ring-green-500 focus:border-green-500"
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Current Time Range */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2" id="current-label">
                Current Period
              </label>
              <div className="space-y-2">
                <div className="flex space-x-2" role="group" aria-labelledby="current-label">
                  {['24h', '7d', '30d', 'custom'].map((range) => (
                    <button
                      key={range}
                      onClick={() => handleCurrentTimeRangeChange(range as TimeRange)}
                      className={`px-2 py-1 rounded text-xs font-medium transition-colors ${
                        customCurrentTimeRange === range
                          ? 'bg-blue-600 text-white'
                          : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
                      }`}
                      aria-pressed={customCurrentTimeRange === range}
                      disabled={isLoading}
                    >
                      {range === '24h' ? '24 Hours' : range === '7d' ? '7 Days' : range === '30d' ? '30 Days' : 'Custom'}
                    </button>
                  ))}
                </div>
                {customCurrentTimeRange === 'custom' && (
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label htmlFor="currentStart" className="sr-only">Current Start Date</label>
                      <input
                        id="currentStart"
                        type="date"
                        value={currentStart ? currentStart.split('T')[0] : ''}
                        onChange={(e) => setCurrentStart(e.target.value ? new Date(e.target.value).toISOString() : '')}
                        max={currentEnd || new Date().toISOString().split('T')[0]}
                        disabled={isLoading}
                        className="w-full px-2 py-1 text-xs border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
                      />
                    </div>
                    <div>
                      <label htmlFor="currentEnd" className="sr-only">Current End Date</label>
                      <input
                        id="currentEnd"
                        type="date"
                        value={currentEnd ? currentEnd.split('T')[0] : ''}
                        onChange={(e) => setCurrentEnd(e.target.value ? new Date(e.target.value).toISOString() : '')}
                        min={currentStart}
                        max={new Date().toISOString().split('T')[0]}
                        disabled={isLoading}
                        className="w-full px-2 py-1 text-xs border border-gray-300 rounded focus:ring-blue-500 focus:border-blue-500"
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Node Comparison Mode - Single Time Range Selector */}
      {localMode === 'node' && (
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">{config.label} Comparison</h3>
          <div className="flex space-x-2" role="group" aria-label="Time range selector">
            {timeRangeOptions.filter((opt) => opt.value !== 'custom').map((option) => (
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
      )}

      {/* Statistics Panel - Node Comparison Mode */}
      {localMode === 'node' && showStatistics && overallStats && nodes && nodes.length > 0 && (
        <div className="mb-4 p-3 bg-gray-50 rounded-lg" role="region" aria-label="Node comparison statistics">
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
          {showPercentileStats && nodes.length > 0 && (
            <div className="mt-3 pt-3 border-t border-gray-200">
              <div className="text-sm font-medium text-gray-700 mb-2">Percentile Statistics</div>
              <div className="grid grid-cols-3 gap-4 text-center">
                <div>
                  <div className="text-sm text-gray-600">P50 (Median)</div>
                  <div className="text-lg font-semibold text-gray-900">
                    {calculateStatisticalSummary(nodes[0].data)?.median.toFixed(2) || '0.00'} {config.unit}
                  </div>
                </div>
                <div>
                  <div className="text-sm text-gray-600">P95</div>
                  <div className="text-lg font-semibold text-gray-900">
                    {calculateStatisticalSummary(nodes[0].data)?.p95.toFixed(2) || '0.00'} {config.unit}
                  </div>
                </div>
                <div>
                  <div className="text-sm text-gray-600">P99</div>
                  <div className="text-lg font-semibold text-gray-900">
                    {calculateStatisticalSummary(nodes[0].data)?.p99.toFixed(2) || '0.00'} {config.unit}
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Statistics Panel - Time Range Comparison Mode */}
      {localMode === 'timeRange' && showStatistics && timeRangeStatistics && (
        <div className="mb-4 p-4 bg-gray-50 rounded-lg" role="region" aria-label="Time range comparison statistics">
          <div className="grid grid-cols-2 gap-6">
            {/* Baseline Statistics */}
            <div>
              <h4 className="text-sm font-semibold text-gray-700 mb-3">Baseline Period</h4>
              <div className="space-y-2">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Average:</span>
                  <span className="text-base font-semibold text-gray-900">
                    {timeRangeStatistics.baseline.avg.toFixed(2)} {config.unit}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">P50 (Median):</span>
                  <span className="text-base font-semibold text-gray-900">
                    {timeRangeStatistics.baseline.median.toFixed(2)} {config.unit}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">P95:</span>
                  <span className="text-base font-semibold text-gray-900">
                    {timeRangeStatistics.baseline.p95.toFixed(2)} {config.unit}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">P99:</span>
                  <span className="text-base font-semibold text-gray-900">
                    {timeRangeStatistics.baseline.p99.toFixed(2)} {config.unit}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Min - Max:</span>
                  <span className="text-base font-semibold text-gray-900">
                    {timeRangeStatistics.baseline.min.toFixed(2)} - {timeRangeStatistics.baseline.max.toFixed(2)} {config.unit}
                  </span>
                </div>
              </div>
            </div>

            {/* Current Statistics with Percentage Change */}
            <div>
              <h4 className="text-sm font-semibold text-gray-700 mb-3">Current Period</h4>
              <div className="space-y-2">
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Average:</span>
                  <div className="text-right">
                    <span className="text-base font-semibold text-gray-900 mr-2">
                      {timeRangeStatistics.current.avg.toFixed(2)} {config.unit}
                    </span>
                    <span className={`text-xs font-medium ${
                      timeRangeStatistics.changes.avg > 0 ? 'text-red-600' : timeRangeStatistics.changes.avg < 0 ? 'text-green-600' : 'text-gray-600'
                    }`}>
                      {timeRangeStatistics.changes.avg > 0 ? '↑' : timeRangeStatistics.changes.avg < 0 ? '↓' : '•'} {Math.abs(timeRangeStatistics.changes.avg).toFixed(1)}%
                    </span>
                  </div>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">P50 (Median):</span>
                  <div className="text-right">
                    <span className="text-base font-semibold text-gray-900 mr-2">
                      {timeRangeStatistics.current.median.toFixed(2)} {config.unit}
                    </span>
                    <span className={`text-xs font-medium ${
                      timeRangeStatistics.changes.median > 0 ? 'text-red-600' : timeRangeStatistics.changes.median < 0 ? 'text-green-600' : 'text-gray-600'
                    }`}>
                      {timeRangeStatistics.changes.median > 0 ? '↑' : timeRangeStatistics.changes.median < 0 ? '↓' : '•'} {Math.abs(timeRangeStatistics.changes.median).toFixed(1)}%
                    </span>
                  </div>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">P95:</span>
                  <div className="text-right">
                    <span className="text-base font-semibold text-gray-900 mr-2">
                      {timeRangeStatistics.current.p95.toFixed(2)} {config.unit}
                    </span>
                    <span className={`text-xs font-medium ${
                      timeRangeStatistics.changes.p95 > 0 ? 'text-red-600' : timeRangeStatistics.changes.p95 < 0 ? 'text-green-600' : 'text-gray-600'
                    }`}>
                      {timeRangeStatistics.changes.p95 > 0 ? '↑' : timeRangeStatistics.changes.p95 < 0 ? '↓' : '•'} {Math.abs(timeRangeStatistics.changes.p95).toFixed(1)}%
                    </span>
                  </div>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">P99:</span>
                  <div className="text-right">
                    <span className="text-base font-semibold text-gray-900 mr-2">
                      {timeRangeStatistics.current.p99.toFixed(2)} {config.unit}
                    </span>
                    <span className={`text-xs font-medium ${
                      timeRangeStatistics.changes.p99 > 0 ? 'text-red-600' : timeRangeStatistics.changes.p99 < 0 ? 'text-green-600' : 'text-gray-600'
                    }`}>
                      {timeRangeStatistics.changes.p99 > 0 ? '↑' : timeRangeStatistics.changes.p99 < 0 ? '↓' : '•'} {Math.abs(timeRangeStatistics.changes.p99).toFixed(1)}%
                    </span>
                  </div>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm text-gray-600">Min - Max:</span>
                  <span className="text-base font-semibold text-gray-900">
                    {timeRangeStatistics.current.min.toFixed(2)} - {timeRangeStatistics.current.max.toFixed(2)} {config.unit}
                  </span>
                </div>
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
        aria-label={`${config.label} comparison chart showing ${localMode === 'timeRange' ? 'baseline vs current' : `${nodes?.length || 0} nodes`}`}
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

      {/* Legend - Node Comparison Mode */}
      {localMode === 'node' && nodes && nodes.length > 0 && (
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

      {/* Legend - Time Range Comparison Mode */}
      {localMode === 'timeRange' && timeRangeData && (
        <div className="mt-4 flex items-center justify-center space-x-6 text-sm">
          <div className="flex items-center space-x-2">
            <div
              className="w-4 h-1 bg-green-600"
              aria-hidden="true"
            />
            <span className="text-gray-700">Baseline Period</span>
          </div>
          <div className="flex items-center space-x-2">
            <div
              className="w-4 h-1 bg-blue-600"
              aria-hidden="true"
            />
            <span className="text-gray-700">Current Period</span>
          </div>
        </div>
      )}
    </div>
  )
}
