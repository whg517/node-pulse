/**
 * Latency Trend Chart Component
 *
 * ECharts-based line chart for displaying network latency trends over time.
 * Supports dark/light theme switching via useTheme hook.
 */

import { useEffect, useRef, useCallback } from 'react'
import echarts, { graphic } from '../../lib/echarts-core'
import type { ECharts, EChartsOption, SeriesOption } from '../../lib/echarts-core'
import { useTranslation } from 'react-i18next'
import type { DataPoint } from '../dashboard/TrendChart'

function getCSSVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

export interface LatencyTrendChartProps {
  data: DataPoint[]
  height?: string
  className?: string
  isLoading?: boolean
  showBaseline?: boolean
  baselineValue?: number
}

// Color palette from UI design
const COLORS = {
  latency: '#3b82f6',    // Blue-500
  baseline: '#10b981',   // Green-500
  areaGradientStart: 'rgba(59, 130, 246, 0.3)',
  areaGradientEnd: 'rgba(59, 130, 246, 0.05)',
}

/**
 * LatencyTrendChart Component
 *
 * @param data - Array of data points with timestamp and value
 * @param height - Chart height (default: 300px)
 * @param className - Additional CSS classes
 * @param isLoading - Loading state
 * @param showBaseline - Whether to show baseline reference line
 * @param baselineValue - Baseline value for reference
 */
export function LatencyTrendChart({
  data,
  height = '300px',
  className = '',
  isLoading = false,
  showBaseline = false,
  baselineValue,
}: LatencyTrendChartProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<ECharts | null>(null)
  const { t } = useTranslation()

  const getChartOptions = useCallback((): EChartsOption => {
    const textColor = getCSSVar('--color-chart-text') || '#374151'
    const axisLineColor = getCSSVar('--color-chart-axis') || '#e5e7eb'
    const splitLineColor = getCSSVar('--color-chart-grid') || '#f3f4f6'
    const tooltipBg = getCSSVar('--color-chart-tooltip-bg') || 'rgba(255,255,255,0.95)'
    const tooltipBorder = getCSSVar('--color-chart-tooltip-border') || '#e5e7eb'
    const tooltipText = getCSSVar('--color-chart-tooltip-text') || '#374151'

    const series: SeriesOption[] = [
      {
        name: t('metrics.latency'),
        type: 'line',
        data: data.map(d => d.value),
        smooth: 0.3,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: {
          color: COLORS.latency,
          width: 2,
        },
        itemStyle: {
          color: COLORS.latency,
        },
        areaStyle: {
          color: new graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: COLORS.areaGradientStart },
            { offset: 1, color: COLORS.areaGradientEnd },
          ]),
        },
      },
    ]

    if (showBaseline && baselineValue !== undefined) {
      series.push({
        name: t('dashboard.baseline'),
        type: 'line',
        data: Array(data.length).fill(baselineValue),
        lineStyle: {
          color: COLORS.baseline,
          width: 2,
          type: 'dashed',
        },
        itemStyle: {
          color: COLORS.baseline,
        },
        symbol: 'none',
      })
    }

    return {
      backgroundColor: 'transparent',
      grid: {
        left: 60,
        right: 40,
        top: 40,
        bottom: 60,
      },
      tooltip: {
        trigger: 'axis',
        backgroundColor: tooltipBg,
        borderColor: tooltipBorder,
        textStyle: {
          color: tooltipText,
        },
        formatter: (params: unknown) => {
          const items = params as Array<{ name: string; value: number; seriesName: string }>
          if (!items || items.length === 0) return ''
          const date = new Date(items[0].name)
          const formattedDate = date.toLocaleString()
          let result = `${formattedDate}<br/>`
          items.forEach(item => {
            result += `${item.seriesName}: ${item.value.toFixed(1)} ${t('units.ms')}<br/>`
          })
          return result
        },
      },
      xAxis: {
        type: 'category',
        data: data.map(d => d.timestamp),
        axisLabel: {
          color: textColor,
          fontSize: 11,
        },
        axisLine: {
          lineStyle: {
            color: axisLineColor,
          },
        },
      },
      yAxis: {
        type: 'value',
        name: t('metrics.latency'),
        nameLocation: 'middle',
        nameGap: 45,
        nameTextStyle: {
          color: textColor,
        },
        axisLabel: {
          color: textColor,
          formatter: (value: number) => `${value.toFixed(0)}`,
        },
        axisLine: {
          lineStyle: {
            color: axisLineColor,
          },
        },
        splitLine: {
          lineStyle: {
            color: splitLineColor,
          },
        },
      },
      series,
    }
  }, [data, showBaseline, baselineValue, t])

  // Initialize chart only once on mount
  useEffect(() => {
    if (!chartRef.current) return

    chartInstance.current = echarts.init(chartRef.current)

    return () => {
      if (chartInstance.current) {
        chartInstance.current.dispose()
        chartInstance.current = null
      }
    }
  }, []) // Empty deps - only run on mount/unmount

  // Update chart options when data or theme changes
  useEffect(() => {
    if (!chartInstance.current || isLoading || data.length === 0) return

    chartInstance.current.setOption(getChartOptions(), true)
  }, [data, isLoading, getChartOptions])

  // Handle resize
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
      className={`latency-trend-chart ${className}`}
      role="img"
      aria-label={t('dashboard.latencyTrendChart')}
    >
      <div
        ref={chartRef}
        style={{ height }}
      />
      {isLoading && (
        <div className="flex items-center justify-center h-full">
          <div
            className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"
            role="status"
            aria-label={t('common.loading')}
          />
        </div>
      )}
      {!isLoading && data.length === 0 && (
        <div className="flex items-center justify-center h-full text-gray-500">
          {t('dashboard.noData')}
        </div>
      )}
    </div>
  )
}
