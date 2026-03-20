import type { HealthStatus } from '../../utils/healthStatus'

interface HealthStatusBadgeProps {
  status: HealthStatus
}

const statusConfig = {
  healthy: {
    label: '健康',
    bgColor: 'bg-[var(--color-healthy-bg)]',
    textColor: 'text-[var(--color-healthy-text)]',
    dotColor: 'bg-[var(--color-healthy)]',
  },
  warning: {
    label: '预警',
    bgColor: 'bg-[var(--color-warning-bg)]',
    textColor: 'text-[var(--color-warning-text)]',
    dotColor: 'bg-[var(--color-warning)]',
  },
  critical: {
    label: '异常',
    bgColor: 'bg-[var(--color-critical-bg)]',
    textColor: 'text-[var(--color-critical-text)]',
    dotColor: 'bg-[var(--color-critical)]',
  },
  offline: {
    label: '离线',
    bgColor: 'bg-[var(--color-unknown-bg)]',
    textColor: 'text-[var(--color-unknown-text)]',
    dotColor: 'bg-[var(--color-unknown)]',
  },
}

/**
 * HealthStatusBadge Component
 *
 * Displays a colored badge indicating node health status
 *
 * @param status - Health status to display
 *
 * @example
 * <HealthStatusBadge status="healthy" /> // Green "健康" badge
 * <HealthStatusBadge status="critical" /> // Red "异常" badge
 */
export function HealthStatusBadge({ status }: HealthStatusBadgeProps) {
  const config = statusConfig[status]

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium ${config.bgColor} ${config.textColor}`}
      role="status"
      aria-label={`Health status: ${config.label}`}
    >
      <span className={`w-2 h-2 rounded-full ${config.dotColor}`} aria-hidden="true" />
      {config.label}
    </span>
  )
}
