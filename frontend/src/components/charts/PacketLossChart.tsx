/**
 * Packet Loss Chart Component
 *
 * ECharts-based line chart for displaying packet loss rate trends over time.
 * Supports dark/light theme switching via useTheme hook.
 */

import { useEffect, useRef, useCallback } from 'react'
import echarts from '../../lib/echarts-core'
import type { ECharts, EChartsOption } from '../../lib/echarts-core'
import { useTranslation } from 'react-i18next'
import type { DataPoint } from '../dashboard/TrendChart'
import { useThemeColors } from '../../hooks/useThemeColors'

function getCSSVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

export interface PacketLossChartProps {
  data: DataPoint[]
  height?: string
  className?: string
  isLoading?: boolean
  warningThreshold?: number
  criticalThreshold?: number
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
  const chartInstance = useRef<ECharts | null>(null)
  const { t } = useTranslation()
  const themeColors = useThemeColors()

  const getChartOptions = useCallback((): EChartsOption => {
    const textColor = getCSSVar('--color-chart-text') || '#374151'
    const axisLineColor = getCSSVar('--color-chart-axis') || '#e5e7eb'
    const splitLineColor = getCSSVar('--color-chart-grid') || '#f3f4f6'
    const tooltipBg = getCSSVar('--color-chart-tooltip-bg') || 'rgba(255,255,255,0.95)'
    const tooltipBorder = getCSSVar('--color-chart-tooltip-border') || '#e5e7eb'
    const tooltipText = getCSSVar('--color-chart-tooltip-text') || '#374151'

    const packetLossColor = themeColors.critical

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
            color: packetLossColor,
            width: 2,
          },
          itemStyle: {
            color: packetLossColor,
          },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: `${packetLossColor}4D` },
              { offset: 1, color: `${packetLossColor}0D` },
            ]),
          },
          markLine: {
            silent: true,
            data: [
              {
                yAxis: warningThreshold,
                lineStyle: {
                  color: themeColors.warning,
                  type: 'dashed',
                },
                label: {
                  formatter: t('status.warning'),
                  color: themeColors.warning,
                },
              },
              {
                yAxis: criticalThreshold,
                lineStyle: {
                  color: themeColors.critical,
                  type: 'dashed',
                },
                label: {
                  formatter: t('status.critical'),
                  color: themeColors.critical,
                },
              },
            ],
          },
        },
      ],
    }
  }, [data, warningThreshold, criticalThreshold, t, themeColors])

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
            className="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--color-critical)]"
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
