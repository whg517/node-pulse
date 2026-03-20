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
    good: 'border-l-[var(--color-healthy)]',
    warning: 'border-l-[var(--color-warning)]',
    critical: 'border-l-[var(--color-critical)]',
    neutral: 'border-l-[var(--color-text-muted)]',
  }

  const statusIndicatorColors = {
    good: 'bg-[var(--color-healthy)]',
    warning: 'bg-[var(--color-warning)]',
    critical: 'bg-[var(--color-critical)]',
    neutral: 'bg-[var(--color-text-muted)]',
  }

  return (
    <div
      className={`metric-card rounded-lg border border-[var(--color-border)] border-l-[3px] p-4 shadow-sm hover:shadow-md transition-shadow duration-200 bg-[var(--color-bg-surface)] ${statusBorderColors[status]} ${className}`}
      role="region"
      aria-label={`${title} metric`}
    >
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center space-x-2">
          {icon && <div className="text-[var(--color-text-secondary)]">{icon}</div>}
          <h3 className="text-sm font-semibold uppercase tracking-wide text-[var(--color-text-secondary)]">
            {title}
          </h3>
        </div>
        <div
          className={`w-3 h-3 rounded-full ${statusIndicatorColors[status]}`}
          aria-hidden="true"
        />
      </div>

      <div className="flex items-baseline space-x-2">
        <span className="text-3xl font-bold text-[var(--color-text-primary)]" aria-label={`${title} value`}>
          {value}
        </span>
        {unit && (
          <span className="text-sm text-[var(--color-text-secondary)]" aria-label="unit">
            {unit}
          </span>
        )}
      </div>

      {trend && (
        <div className="mt-2 flex items-center text-sm">
          <span
            className={`font-medium ${trend.isPositive ? 'text-[var(--color-healthy)]' : 'text-[var(--color-critical)]'}`}
            aria-label={`trend ${trend.isPositive ? 'up' : 'down'} by ${Math.abs(trend.value)}%`}
          >
            {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}%
          </span>
          <span className="text-[var(--color-text-secondary)] ml-1">vs. previous period</span>
        </div>
      )}
    </div>
  )
}
