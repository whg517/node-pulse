import React from 'react'

export interface MetricCardProps {
  title: string
  value: string | number
  unit?: string
  status?: 'good' | 'warning' | 'critical' | 'neutral'
  icon?: React.ReactNode
  trend?: {
    value: number
    isPositive: boolean
  }
  className?: string
}

/**
 * MetricCard component for displaying individual metrics
 *
 * Shows a metric with title, value, unit, status indicator, and optional trend.
 * Uses color coding to indicate health status.
 *
 * @param props - MetricCard props
 * @returns MetricCard component
 *
 * @example
 * <MetricCard
 *   title="Latency"
 *   value={45}
 *   unit="ms"
 *   status="good"
 *   trend={{ value: 5, isPositive: false }}
 * />
 */
export default function MetricCard({
  title,
  value,
  unit = '',
  status = 'neutral',
  icon,
  trend,
  className = '',
}: MetricCardProps) {
  const statusBorderColors = {
    good: 'border-l-healthy',
    warning: 'border-l-warning',
    critical: 'border-l-destructive',
    neutral: 'border-l-muted-foreground',
  }

  const statusIndicatorColors = {
    good: 'bg-healthy',
    warning: 'bg-warning',
    critical: 'bg-destructive',
    neutral: 'bg-muted-foreground',
  }

  return (
    <div
      className={`metric-card rounded-lg border border-border border-l-[3px] p-4 shadow-sm hover:shadow-md transition-shadow duration-200 bg-card ${statusBorderColors[status]} ${className}`}
      role="region"
      aria-label={`${title} metric`}
    >
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2">
          {icon && <div className="text-muted-foreground">{icon}</div>}
          <h3 className="text-sm font-semibold uppercase tracking-wide text-muted-foreground">
            {title}
          </h3>
        </div>
        <div
          className={`w-3 h-3 rounded-full ${statusIndicatorColors[status]}`}
          aria-hidden="true"
        />
      </div>

      <div className="flex items-baseline space-x-2">
        <span className="text-3xl font-bold text-foreground" aria-label={`${title} value`}>
          {value}
        </span>
        {unit && (
          <span className="text-sm text-muted-foreground" aria-label="unit">
            {unit}
          </span>
        )}
      </div>

      {trend && (
        <div className="mt-2 flex items-center text-sm">
          <span
            className={`font-medium ${trend.isPositive ? 'text-healthy' : 'text-destructive'}`}
            aria-label={`trend ${trend.isPositive ? 'up' : 'down'} by ${Math.abs(trend.value)}%`}
          >
            {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}%
          </span>
          <span className="text-muted-foreground ml-1">vs. previous period</span>
        </div>
      )}
    </div>
  )
}
