/**
 * Probe Success Rate Gauge Component
 *
 * ECharts-based gauge chart for displaying probe success rate.
 * Supports dark/light theme switching via useTheme hook.
 */

import { useEffect, useRef, useCallback } from 'react'
import * as echarts from 'echarts'
import { useTranslation } from 'react-i18next'
import { useTheme } from '../../hooks/useTheme'

export interface ProbeSuccessGaugeProps {
  value: number // 0-100 percentage
  height?: string
  className?: string
  isLoading?: boolean
}

// Color palette from UI design
const COLORS = {
  healthy: '#22C55E',   // Green (task specified)
  warning: '#F59E0B',   // Amber (task specified)
  critical: '#EF4444',  // Red (task specified)
  unknown: '#6B7280',   // Gray (task specified)
}

/**
 * Get color based on success rate value
 */
function getColorForValue(value: number): string {
  if (value >= 95) return COLORS.healthy
  if (value >= 80) return COLORS.warning
  if (value >= 50) return COLORS.critical
  return COLORS.unknown
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
  const chartInstance = useRef<echarts.ECharts | null>(null)
  const { t } = useTranslation()
  const { isDark } = useTheme()

  const getChartOptions = useCallback((): echarts.EChartsOption => {
    const textColor = isDark ? '#e5e7eb' : '#374151'
    const color = getColorForValue(value)

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
            width: 20,
          },
          pointer: {
            show: false,
          },
          axisLine: {
            lineStyle: {
              width: 20,
              color: isDark
                ? [[1, '#374151']]
                : [[1, '#e5e7eb']],
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
            offsetCenter: [0, '70%'],
            fontSize: 14,
            color: textColor,
            formatter: t('dashboard.probeSuccessRate'),
          },
          detail: {
            valueAnimation: true,
            width: '60%',
            lineHeight: 40,
            borderRadius: 8,
            offsetCenter: [0, '-10%'],
            fontSize: 36,
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
  }, [value, isDark, t])

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
            className="animate-spin rounded-full h-8 w-8 border-b-2 border-green-500"
            role="status"
            aria-label={t('common.loading')}
          />
        </div>
      )}
    </div>
  )
}

export { COLORS as GAUGE_COLORS }
