import { useCallback, useEffect, useRef, useState } from 'react'
import echarts from '../../lib/echarts-core'
import type { ECharts, EChartsOption, SeriesOption } from '../../lib/echarts-core'

interface TooltipParam {
  name: string
  value: number
}

function normalizeTooltipParams(params: unknown): TooltipParam[] {
  if (!Array.isArray(params)) {
    return []
  }

  return params.filter((param): param is TooltipParam => {
    return typeof param === 'object' && param !== null && 'name' in param && 'value' in param
  })
}

export type TimeRange = '24h' | '7d' | '30d'
export type MetricType = 'latency_ms' | 'packet_loss_rate' | 'jitter_ms'

export interface DataPoint {
  timestamp: string
  value: number
}

export interface TrendChartProps {
  data: DataPoint[]
  metric: MetricType
  timeRange: TimeRange
  showBaseline?: boolean
  baselineValue?: number
  height?: string
  className?: string
  onTimeRangeChange?: (range: TimeRange) => void
  isLoading?: boolean
}

/**
 * TrendChart component for displaying time-series data using ECharts
 *
 * Features:
 * - Time range selection (24h/7d/30d)
 * - Multi-metric support (latency/packet loss/jitter)
 * - Hover tooltips with exact values
 * - Zoom functionality (mouse wheel)
 * - Baseline reference line
 * - Responsive design
 *
 * @param props - TrendChart props
 * @returns TrendChart component
 *
 * @example
 * <TrendChart
 *   data={dataPoints}
 *   metric="latency_ms"
 *   timeRange="7d"
 *   showBaseline={true}
 *   onTimeRangeChange={(range) => console.log(range)}
 * />
 */
export default function TrendChart({
  data,
  metric,
  timeRange,
  showBaseline = false,
  baselineValue,
  height = '400px',
  className = '',
  onTimeRangeChange,
  isLoading = false,
}: TrendChartProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<ECharts | null>(null)
  const [localTimeRange, setLocalTimeRange] = useState<TimeRange>(timeRange)

  // Metric configuration
  const metricConfig = {
    latency_ms: {
      label: 'Latency',
      unit: 'ms',
      color: '#3b82f6', // blue-500
      yAxisLabel: 'Latency (ms)',
    },
    packet_loss_rate: {
      label: 'Packet Loss Rate',
      unit: '%',
      color: '#ef4444', // red-500
      yAxisLabel: 'Packet Loss Rate (%)',
    },
    jitter_ms: {
      label: 'Jitter',
      unit: 'ms',
      color: '#8b5cf6', // purple-500
      yAxisLabel: 'Jitter (ms)',
    },
  }

  const config = metricConfig[metric]

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
  const getXAxisFormatter = useCallback((): string => {
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
  }, [localTimeRange])

  // Initialize chart only when data exists, not loading, and ref is available
  useEffect(() => {
    if (!chartRef.current) return

    // Only initialize if we have data and not loading
    if (!data || data.length === 0 || isLoading) {
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
  }, [data, isLoading])

  // Update chart when data or time range changes
  useEffect(() => {
    if (!chartInstance.current || !data || data.length === 0) return

    const series: SeriesOption[] = [
      {
        name: config.label,
        type: 'line',
        data: data.map((point) => point.value),
        smooth: true,
        lineStyle: {
          color: config.color,
          width: 2,
        },
        itemStyle: {
          color: config.color,
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: `${config.color}80` },
            { offset: 1, color: `${config.color}10` },
          ]),
        },
      },
    ]

    const option: EChartsOption = {
      title: {
        text: `${config.label} Trend`,
        left: 'center',
        textStyle: {
          fontSize: 16,
          fontWeight: 'bold',
        },
      },
      tooltip: {
        trigger: 'axis',
        formatter: (params: unknown) => {
          const tooltipParams = normalizeTooltipParams(params)
          if (tooltipParams.length === 0) return ''

          const param = tooltipParams[0]
          const date = new Date(param.name)
          const formattedDate = date.toLocaleString()
          return `${formattedDate}<br/>${config.label}: ${param.value} ${config.unit}`
        },
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
        top: '80px',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: data.map((point) => point.timestamp),
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

    // Add baseline if enabled
    if (showBaseline && baselineValue !== undefined) {
      series.push({
        name: 'Baseline',
        type: 'line',
        data: Array(data.length).fill(baselineValue),
        lineStyle: {
          color: '#10b981',
          width: 2,
          type: 'dashed',
        },
        itemStyle: {
          color: '#10b981',
        },
        symbol: 'none',
      })
    }

    chartInstance.current.setOption(option, true)
  }, [data, localTimeRange, config, showBaseline, baselineValue, getXAxisFormatter])

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

  return (
    <div
      className={`trend-chart bg-[var(--color-bg-surface)] rounded-lg shadow-sm p-4 ${className}`}
      role="region"
      aria-label={`${config.label} trend chart`}
    >
      {/* Time Range Selector */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">{config.label} Trend</h3>
        <div className="flex space-x-2" role="group" aria-label="Time range selector">
          {timeRangeOptions.map((option) => (
            <button
              key={option.value}
              onClick={() => handleTimeRangeChange(option.value)}
              className={`px-3 py-1 rounded text-sm font-medium transition-colors ${
                localTimeRange === option.value
                  ? 'bg-blue-600 text-white'
                  : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-subtle)]'
              }`}
              aria-pressed={localTimeRange === option.value}
              disabled={isLoading}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>

      {/* Chart Container */}
      <div
        ref={chartRef}
        className="relative"
        style={{ height }}
        role="img"
        aria-label={`${config.label} trend chart showing ${data.length} data points`}
      >
        {/* Loading Overlay */}
        {isLoading && (
          <div className="absolute inset-0 flex items-center justify-center bg-[var(--color-bg-surface)] bg-opacity-75 z-10">
            <div className="flex flex-col items-center">
              <div
                className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
                role="status"
                aria-label="Loading chart data"
              />
              <p className="mt-2 text-[var(--color-text-secondary)]">Loading chart data...</p>
            </div>
          </div>
        )}

        {/* Empty State */}
        {!isLoading && (!data || data.length === 0) && (
          <div className="absolute inset-0 flex items-center justify-center">
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
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
              <h3 className="mt-2 text-sm font-medium text-[var(--color-text-primary)]">No Data Available</h3>
              <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
                No trend data available for the selected time range.
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Legend */}
      {showBaseline && baselineValue !== undefined && (
        <div className="mt-4 flex items-center justify-center space-x-6 text-sm">
          <div className="flex items-center space-x-2">
            <div
              className="w-4 h-1"
              style={{ backgroundColor: config.color }}
              aria-hidden="true"
            />
            <span className="text-[var(--color-text-secondary)]">{config.label}</span>
          </div>
          <div className="flex items-center space-x-2">
            <div
              className="w-4 h-1 bg-green-500"
              style={{ borderStyle: 'dashed', borderWidth: '2px' }}
              aria-hidden="true"
            />
            <span className="text-[var(--color-text-secondary)]">Baseline ({baselineValue} {config.unit})</span>
          </div>
        </div>
      )}
    </div>
  )
}
