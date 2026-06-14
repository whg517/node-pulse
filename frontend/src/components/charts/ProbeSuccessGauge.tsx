import { PieChart, Pie, Cell, ResponsiveContainer } from 'recharts'
import { useTranslation } from 'react-i18next'
import { useThemeColors } from '../../hooks/useThemeColors'

export interface ProbeSuccessGaugeProps {
  value: number
  height?: string
  className?: string
  isLoading?: boolean
}

function getColorForValue(value: number, themeColors: ReturnType<typeof useThemeColors>): string {
  if (value >= 95) return themeColors.healthy || 'var(--chart-2)'
  if (value >= 80) return themeColors.warning || 'var(--chart-4)'
  if (value >= 50) return themeColors.critical || 'var(--destructive)'
  return themeColors.unknown || 'var(--muted-foreground)'
}

export function ProbeSuccessGauge({
  value,
  height = '200px',
  className = '',
  isLoading = false,
}: ProbeSuccessGaugeProps) {
  const { t } = useTranslation()
  const themeColors = useThemeColors()
  const color = getColorForValue(value, themeColors)
  const clampedValue = Math.max(0, Math.min(100, value))

  const data = [
    { name: 'success', value: clampedValue },
    { name: 'remaining', value: 100 - clampedValue },
  ]

  const trackColor = 'var(--muted)'

  return (
    <div className={`probe-success-gauge ${className}`} role="img" aria-label={`${t('dashboard.probeSuccessRate')}: ${value.toFixed(1)}${t('units.percent')}`}>
      {isLoading ? (
        <div className="flex items-center justify-center" style={{ height }}>
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-healthy" role="status" aria-label={t('common.loading')} />
        </div>
      ) : (
        <div className="relative" style={{ height }}>
          <ResponsiveContainer width="100%" height="100%" minWidth={0}>
            <PieChart>
              <Pie
                data={data}
                dataKey="value"
                cx="50%"
                cy="50%"
                innerRadius="65%"
                outerRadius="85%"
                startAngle={200}
                endAngle={-20}
                strokeWidth={0}
                isAnimationActive={false}
                animationBegin={0}
                animationDuration={600}
              >
                <Cell fill={color} />
                <Cell fill={trackColor} />
              </Pie>
            </PieChart>
          </ResponsiveContainer>
          <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
            <span className="text-2xl font-bold" style={{ color }}>
              {Math.round(clampedValue)}{t('units.percent')}
            </span>
            <span className="text-xs text-muted-foreground mt-0.5">
              {t('dashboard.probeSuccessRate')}
            </span>
          </div>
        </div>
      )}
    </div>
  )
}
