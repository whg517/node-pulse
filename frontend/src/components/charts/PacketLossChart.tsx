/**
 * Packet Loss Chart Component
 *
 * ECharts-based line chart for displaying packet loss rate trends over time.
 * Supports dark/light theme switching via useTheme hook.
 */

import { useEffect, useRef, useCallback } from 'react'
import * as echarts from 'echarts'
import { useTranslation } from 'react-i18next'
import { useTheme } from '../../hooks/useTheme'
import type { DataPoint } from '../dashboard/TrendChart'

export interface PacketLossChartProps {
  data: DataPoint[]
  height?: string
  className?: string
  isLoading?: boolean
  warningThreshold?: number
  criticalThreshold?: number
}

// Color palette from UI design
const COLORS = {
  packetLoss: '#ef4444',  // Red-500
  warning: '#f59e0b',     // Amber
  critical: '#ef4444',    // Red
  areaGradientStart: 'rgba(239, 68, 68, 0.3)',
  areaGradientEnd: 'rgba(239, 68, 68, 0.05)',
}

/**
 * PacketLossChart Component
 *
 * @param data - Array of data points with timestamp and value (percentage)
 * @param height - Chart height (default: 300px)
 * @param className - Additional CSS classes
 * @param isLoading - Loading state
 * @param warningThreshold - Warning threshold line value (default: 3%)
 * @param criticalThreshold - Critical threshold line value (default: 5%)
 */
export function PacketLossChart({
  data,
  height = '300px',
  className = '',
  isLoading = false,
  warningThreshold = 3,
  criticalThreshold = 5,
}: PacketLossChartProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<echarts.ECharts | null>(null)
  const { t } = useTranslation()
  const { isDark } = useTheme()

  const getChartOptions = useCallback((): echarts.EChartsOption => {
    const textColor = isDark ? '#e5e7eb' : '#374151'
    const axisLineColor = isDark ? '#4b5563' : '#e5e7eb'
    const splitLineColor = isDark ? '#374151' : '#f3f4f6'

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
        backgroundColor: isDark ? 'rgba(17, 24, 39, 0.95)' : 'rgba(255, 255, 255, 0.95)',
        borderColor: isDark ? '#374151' : '#e5e7eb',
        textStyle: {
          color: isDark ? '#f9fafb' : '#374151',
        },
        formatter: (params: unknown) => {
          const items = params as Array<{ name: string; value: number; seriesName: string }>
          if (!items || items.length === 0) return ''
          const date = new Date(items[0].name)
          const formattedDate = date.toLocaleString()
          return `${formattedDate}<br/>${t('metrics.packetLoss')}: ${items[0].value.toFixed(2)}${t('units.percent')}`
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
        name: `${t('metrics.packetLoss')} (%)`,
        nameLocation: 'middle',
        nameGap: 45,
        nameTextStyle: {
          color: textColor,
        },
        axisLabel: {
          color: textColor,
          formatter: (value: number) => `${value.toFixed(0)}%`,
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
        max: 100,
      },
      series: [
        {
          name: t('metrics.packetLoss'),
          type: 'line',
          data: data.map(d => d.value),
          smooth: 0.3,
          symbol: 'circle',
          symbolSize: 4,
          lineStyle: {
            color: COLORS.packetLoss,
            width: 2,
          },
          itemStyle: {
            color: COLORS.packetLoss,
          },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: COLORS.areaGradientStart },
              { offset: 1, color: COLORS.areaGradientEnd },
            ]),
          },
          markLine: {
            silent: true,
            data: [
              {
                yAxis: warningThreshold,
                lineStyle: {
                  color: COLORS.warning,
                  type: 'dashed',
                },
                label: {
                  formatter: t('status.warning'),
                  color: COLORS.warning,
                },
              },
              {
                yAxis: criticalThreshold,
                lineStyle: {
                  color: COLORS.critical,
                  type: 'dashed',
                },
                label: {
                  formatter: t('status.critical'),
                  color: COLORS.critical,
                },
              },
            ],
          },
        },
      ],
    }
  }, [data, isDark, warningThreshold, criticalThreshold, t])

  // Initialize and update chart
  useEffect(() => {
    if (!chartRef.current || isLoading || data.length === 0) return

    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current)
    }

    chartInstance.current.setOption(getChartOptions(), true)

    return () => {
      if (chartInstance.current) {
        chartInstance.current.dispose()
        chartInstance.current = null
      }
    }
  }, [data, isLoading, getChartOptions])

  // Handle resize and theme change
  useEffect(() => {
    const handleResize = () => {
      if (chartInstance.current) {
        chartInstance.current.resize()
      }
    }

    window.addEventListener('resize', handleResize)

    if (chartInstance.current && data.length > 0) {
      chartInstance.current.setOption(getChartOptions(), true)
    }

    return () => {
      window.removeEventListener('resize', handleResize)
    }
  }, [isDark, getChartOptions, data.length])

  return (
    <div
      className={`packet-loss-chart ${className}`}
      role="img"
      aria-label={t('dashboard.packetLossChart')}
    >
      <div
        ref={chartRef}
        style={{ height }}
      />
      {isLoading && (
        <div className="flex items-center justify-center h-full">
          <div
            className="animate-spin rounded-full h-8 w-8 border-b-2 border-red-500"
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
