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
  const statusColors = {
    good: 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-800 dark:text-green-300',
    warning: 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-300',
    critical: 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-800 dark:text-red-300',
    neutral: 'bg-[var(--color-bg-muted)] border-[var(--color-border)] text-[var(--color-text-secondary)]',
  }

  const statusIndicatorColors = {
    good: 'bg-green-500',
    warning: 'bg-yellow-500',
    critical: 'bg-red-500',
    neutral: 'bg-[var(--color-text-muted)]',
  }

  return (
    <div
      className={`metric-card rounded-lg border-2 p-4 shadow-sm hover:shadow-md transition-shadow duration-200 ${statusColors[status]} ${className}`}
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
            className={`font-medium ${trend.isPositive ? 'text-green-600' : 'text-red-600'}`}
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
