import { useEffect, useRef } from 'react'
import echarts from '../../lib/echarts-core'
import type { ECharts, EChartsOption, SeriesOption } from '../../lib/echarts-core'
import type { MetricTrendData } from '../../api/performance'

interface TooltipParam {
  name: string
  value: number
  seriesName: string
}

function normalizeTooltipParams(params: unknown): TooltipParam[] {
  if (!Array.isArray(params)) {
    return []
  }

  return params.filter((param): param is TooltipParam => {
    return typeof param === 'object' && param !== null && 'name' in param && 'value' in param && 'seriesName' in param
  })
}

interface PerformanceTrendChartProps {
  trendData: MetricTrendData[]
  targetP99?: number
  targetP95?: number
  height?: string
  className?: string
  isLoading?: boolean
}

/**
 * PerformanceTrendChart displays performance metrics trends over time
 * with P99 and P95 lines, plus target value reference lines.
 */
export function PerformanceTrendChart({
  trendData,
  targetP99,
  targetP95,
  height = '400px',
  className = '',
  isLoading = false,
}: PerformanceTrendChartProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<ECharts | null>(null)

  // Initialize chart
  useEffect(() => {
    if (!chartRef.current) return

    // Initialize ECharts instance
    chartInstance.current = echarts.init(chartRef.current)

    // Cleanup on unmount
    return () => {
      if (chartInstance.current) {
        chartInstance.current.dispose()
        chartInstance.current = null
      }
    }
  }, [])

  // Update chart when data changes
  useEffect(() => {
    if (!chartInstance.current || !trendData || trendData.length === 0) return

    // Prepare series data for each metric
    const series: SeriesOption[] = []

    // Color palette for different metrics
    const colors = ['#3b82f6', '#10b981', '#f59e0b'] // blue, green, amber

    trendData.forEach((metric, index) => {
      if (!metric.data_points || metric.data_points.length === 0) return

      const color = colors[index % colors.length]

      // Extract P99 and P95 values
      const p99Values = metric.data_points.map((dp) => dp.p99)
      const p95Values = metric.data_points.map((dp) => dp.p95)

      // Add P99 line
      series.push({
        name: `${metric.metric_name} P99`,
        type: 'line',
        data: p99Values,
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: {
          width: 2,
          color,
        },
        itemStyle: {
          color,
        },
      })

      // Add P95 line (dashed)
      series.push({
        name: `${metric.metric_name} P95`,
        type: 'line',
        data: p95Values,
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: {
          width: 2,
          type: 'dashed',
          color,
        },
        itemStyle: {
          color,
        },
      })
    })

    // Add target reference lines if provided
    if (targetP99 !== undefined) {
      series.push({
        name: 'P99 目标值',
        type: 'line',
        data: trendData[0]?.data_points?.map(() => targetP99) || [],
        lineStyle: {
          width: 2,
          type: 'dashed',
          color: '#10b981', // green
        },
        itemStyle: {
          opacity: 0,
        },
        markLine: {
          silent: true,
          lineStyle: {
            color: '#10b981',
            width: 2,
            type: 'dashed',
          },
          label: {
            formatter: 'P99 目标',
            position: 'end',
          },
        },
      })
    }

    if (targetP95 !== undefined) {
      series.push({
        name: 'P95 目标值',
        type: 'line',
        data: trendData[0]?.data_points?.map(() => targetP95) || [],
        lineStyle: {
          width: 2,
          type: 'dashed',
          color: '#34d399', // lighter green
        },
        itemStyle: {
          opacity: 0,
        },
        markLine: {
          silent: true,
          lineStyle: {
            color: '#34d399',
            width: 2,
            type: 'dashed',
          },
          label: {
            formatter: 'P95 目标',
            position: 'end',
          },
        },
      })
    }

    // Use first metric's timestamps for X-axis
    const xAxisData =
      trendData[0]?.data_points?.map((dp) => {
        const date = new Date(dp.timestamp)
        return date.toLocaleTimeString('zh-CN', {
          hour: '2-digit',
          minute: '2-digit',
          month: '2-digit',
          day: '2-digit',
        })
      }) || []

    // Build legend data from series names
    const legendData = series
      .map((s) => s.name)
      .filter((name): name is string => name !== undefined)

    const option: EChartsOption = {
      title: {
        text: '性能趋势',
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

          const date = new Date(tooltipParams[0].name)

          let tooltip = `<div style="font-weight: bold; margin-bottom: 8px;">
            ${date.toLocaleString('zh-CN')}
          </div>`

          tooltipParams.forEach((param) => {
            tooltip += `<div style="display: flex; justify-content: space-between; gap: 16px;">
              <span>${param.seriesName}:</span>
              <span style="font-weight: bold;">${param.value.toFixed(2)} ms</span>
            </div>`
          })

          return tooltip
        },
      },
      legend: {
        data: legendData as string[],
        top: 30,
        type: 'scroll' as const,
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: 80,
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        data: xAxisData,
        boundaryGap: false,
        axisLabel: {
          rotate: 45,
          fontSize: 10,
        },
      },
      yAxis: {
        type: 'value',
        name: '响应时间 (ms)',
        axisLabel: {
          formatter: '{value} ms',
        },
      },
      series,
      dataZoom: [
        {
          type: 'inside',
          start: 0,
          end: 100,
        },
        {
          type: 'slider',
          start: 0,
          end: 100,
          height: 20,
          bottom: 10,
        },
      ],
    }

    chartInstance.current.setOption(option)
  }, [trendData, targetP99, targetP95])

  // Resize chart on window resize
  useEffect(() => {
    const handleResize = () => {
      if (chartInstance.current) {
        chartInstance.current.resize()
      }
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  if (isLoading) {
    return (
      <div
        className={`flex items-center justify-center bg-gray-50 rounded-lg ${className}`}
        style={{ height }}
      >
        <div className="text-gray-500">加载中...</div>
      </div>
    )
  }

  if (!trendData || trendData.length === 0) {
    return (
      <div
        className={`flex items-center justify-center bg-gray-50 rounded-lg ${className}`}
        style={{ height }}
      >
        <div className="text-gray-500">暂无趋势数据</div>
      </div>
    )
  }

  return <div ref={chartRef} className={className} style={{ height }} />
}
