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

export interface PacketLossChartProps {
  data: DataPoint[]
  height?: string
  className?: string
  isLoading?: boolean
  warningThreshold?: number
  criticalThreshold?: number
}

export function PacketLossChart({
  data,
  height = '300px',
  className = '',
  isLoading = false,
  warningThreshold = 3,
  criticalThreshold = 5,
}: PacketLossChartProps) {
  const { t } = useTranslation()
  const themeColors = useThemeColors()

  const formatted = data.map((d) => ({
    ...d,
    time: new Date(d.timestamp).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    }),
  }))

  const packetLossColor = themeColors.critical || 'var(--destructive)'
  const warningColor = themeColors.warning || 'var(--chart-4)'
  const criticalColor = themeColors.critical || 'var(--destructive)'

  return (
    <div className={`packet-loss-chart ${className}`} role="img" aria-label={t('dashboard.packetLossChart')}>
      {isLoading ? (
        <div className="flex items-center justify-center" style={{ height }}>
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-destructive" role="status" aria-label={t('common.loading')} />
        </div>
      ) : formatted.length === 0 ? (
        <div className="flex items-center justify-center text-muted-foreground" style={{ height }}>
          {t('dashboard.noData')}
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={parseInt(height) || 300} minWidth={0}>
          <AreaChart data={formatted}>
            <defs>
              <linearGradient id="packetLossGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={packetLossColor} stopOpacity={0.3} />
                <stop offset="100%" stopColor={packetLossColor} stopOpacity={0.05} />
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
              unit="%"
              domain={[0, 'auto']}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: 'var(--popover)',
                border: '1px solid var(--border)',
                borderRadius: 'var(--radius)',
                fontSize: 12,
              }}
              formatter={(value) => [`${Number(value).toFixed(2)}${t('units.percent')}`, t('metrics.packetLoss')]}
            />
            <ReferenceLine
              y={warningThreshold}
              stroke={warningColor}
              strokeDasharray="6 3"
              strokeWidth={2}
              label={{ value: t('status.warning'), position: 'right', fill: warningColor, fontSize: 11 }}
            />
            <ReferenceLine
              y={criticalThreshold}
              stroke={criticalColor}
              strokeDasharray="6 3"
              strokeWidth={2}
              label={{ value: t('status.critical'), position: 'right', fill: criticalColor, fontSize: 11 }}
            />
            <Area
              type="monotone"
              dataKey="value"
              stroke={packetLossColor}
              strokeWidth={2}
              fill="url(#packetLossGradient)"
              dot={false}
              activeDot={{ r: 4 }}
            />
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  )
}
