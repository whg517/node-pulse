/**
 * Probe Success Rate Gauge Component
 *
 * ECharts-based gauge chart for displaying probe success rate.
 * Supports dark/light theme switching via useTheme hook.
 */

import { useEffect, useRef, useCallback } from 'react'
import echarts from '../../lib/echarts-core'
import type { ECharts, EChartsOption } from '../../lib/echarts-core'
import { useTranslation } from 'react-i18next'
import { useThemeColors } from '../../hooks/useThemeColors'

function getCSSVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

export interface ProbeSuccessGaugeProps {
  value: number // 0-100 percentage
  height?: string
  className?: string
  isLoading?: boolean
}

/**
 * Get color based on success rate value
 */
function getColorForValue(value: number, themeColors: ReturnType<typeof useThemeColors>): string {
  if (value >= 95) return themeColors.healthy
  if (value >= 80) return themeColors.warning
  if (value >= 50) return themeColors.critical
  return themeColors.unknown
}

/**
 * ProbeSuccessGauge Component
 *
 * @param value - Success rate percentage (0-100)
 * @param height - Chart height (default: 200px)
 * @param className - Additional CSS classes
 * @param isLoading - Loading state
 */
export function ProbeSuccessGauge({
  value,
  height = '200px',
  className = '',
  isLoading = false,
}: ProbeSuccessGaugeProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<ECharts | null>(null)
  const { t } = useTranslation()
  const themeColors = useThemeColors()

  const getChartOptions = useCallback((): EChartsOption => {
    const textColor = getCSSVar('--color-chart-text') || '#374151'
    const trackColor = getCSSVar('--color-chart-grid') || '#e5e7eb'
    const color = getColorForValue(value, themeColors)

    return {
      backgroundColor: 'transparent',
      series: [
        {
          type: 'gauge',
          startAngle: 200,
          endAngle: -20,
          min: 0,
          max: 100,
          splitNumber: 10,
          itemStyle: {
            color,
          },
          progress: {
            show: true,
            width: 18,
          },
          pointer: {
            show: false,
          },
          axisLine: {
            lineStyle: {
              width: 18,
              color: [[1, trackColor || '#e5e7eb']],
            },
          },
          axisTick: {
            show: false,
          },
          splitLine: {
            show: false,
          },
          axisLabel: {
            show: false,
          },
          anchor: {
            show: false,
          },
          title: {
            show: true,
            offsetCenter: [0, '35%'],
            fontSize: 12,
            color: textColor,
            formatter: t('dashboard.probeSuccessRate'),
          },
          detail: {
            valueAnimation: true,
            width: '60%',
            lineHeight: 32,
            borderRadius: 8,
            offsetCenter: [0, '20%'],
            fontSize: 28,
            fontWeight: 'bold',
            formatter: `{value}${t('units.percent')}`,
            color,
          },
          data: [
            {
              value: Math.round(value),
            },
          ],
        },
      ],
    }
  }, [value, t, themeColors])

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

  // Update chart options when value or theme changes
  useEffect(() => {
    if (!chartInstance.current || isLoading) return

    chartInstance.current.setOption(getChartOptions(), true)
  }, [value, isLoading, getChartOptions])

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
      className={`probe-success-gauge ${className}`}
      role="img"
      aria-label={`${t('dashboard.probeSuccessRate')}: ${value.toFixed(1)}${t('units.percent')}`}
    >
      <div
        ref={chartRef}
        style={{ height }}
      />
      {isLoading && (
        <div className="flex items-center justify-center h-full">
          <div
            className="animate-spin rounded-full h-8 w-8 border-b-2 border-[var(--color-healthy)]"
            role="status"
            aria-label={t('common.loading')}
          />
        </div>
      )}
    </div>
  )
}

// Export empty for backward compatibility
export const GAUGE_COLORS: Record<string, never> = {}
