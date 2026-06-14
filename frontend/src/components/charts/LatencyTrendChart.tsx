import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts'
import { useTranslation } from 'react-i18next'
import type { DataPoint } from '../dashboard/TrendChart'
import { useThemeColors } from '../../hooks/useThemeColors'

export interface LatencyTrendChartProps {
  data: DataPoint[]
  height?: string
  className?: string
  isLoading?: boolean
  showBaseline?: boolean
  baselineValue?: number
}

export function LatencyTrendChart({
  data,
  height = '300px',
  className = '',
  isLoading = false,
  showBaseline = false,
  baselineValue,
}: LatencyTrendChartProps) {
  const { t } = useTranslation()
  const themeColors = useThemeColors()

  const formatted = data.map((d) => ({
    ...d,
    time: new Date(d.timestamp).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    }),
  }))

  const latencyColor = themeColors.brand || 'var(--chart-1)'
  const baselineColor = themeColors.healthy || 'var(--chart-2)'

  return (
    <div className={`latency-trend-chart ${className}`} role="img" aria-label={t('dashboard.latencyTrendChart')}>
      {isLoading ? (
        <div className="flex items-center justify-center" style={{ height }}>
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" role="status" aria-label={t('common.loading')} />
        </div>
      ) : formatted.length === 0 ? (
        <div className="flex items-center justify-center text-muted-foreground" style={{ height }}>
          {t('dashboard.noData')}
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={parseInt(height) || 300} minWidth={0}>
          <AreaChart data={formatted}>
            <defs>
              <linearGradient id="latencyGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={latencyColor} stopOpacity={0.3} />
                <stop offset="100%" stopColor={latencyColor} stopOpacity={0.05} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
            <XAxis
              dataKey="time"
              tick={{ fontSize: 11 }}
              className="fill-muted-foreground"
              interval="preserveStartEnd"
            />
            <YAxis
              tick={{ fontSize: 11 }}
              className="fill-muted-foreground"
              unit="ms"
            />
            <Tooltip
              contentStyle={{
                backgroundColor: 'var(--popover)',
                border: '1px solid var(--border)',
                borderRadius: 'var(--radius)',
                fontSize: 12,
              }}
              formatter={(value) => [`${Number(value).toFixed(1)} ${t('units.ms')}`, t('metrics.latency')]}
            />
            {showBaseline && baselineValue !== undefined && (
              <ReferenceLine
                y={baselineValue}
                stroke={baselineColor}
                strokeDasharray="6 3"
                strokeWidth={2}
                label={{ value: t('dashboard.baseline'), position: 'right', fill: baselineColor, fontSize: 11 }}
              />
            )}
            <Area
              type="monotone"
              dataKey="value"
              stroke={latencyColor}
              strokeWidth={2}
              fill="url(#latencyGradient)"
              dot={false}
              activeDot={{ r: 4 }}
            />
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  )
}
