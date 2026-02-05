import type { HealthStatus } from '../../utils/healthStatus'

interface HealthStatusBadgeProps {
  status: HealthStatus
}

const statusConfig = {
  healthy: {
    label: '健康',
    bgColor: 'bg-green-100',
    textColor: 'text-green-800',
    dotColor: 'bg-green-500',
  },
  warning: {
    label: '预警',
    bgColor: 'bg-yellow-100',
    textColor: 'text-yellow-800',
    dotColor: 'bg-yellow-500',
  },
  critical: {
    label: '异常',
    bgColor: 'bg-red-100',
    textColor: 'text-red-800',
    dotColor: 'bg-red-500',
  },
  offline: {
    label: '离线',
    bgColor: 'bg-gray-100',
    textColor: 'text-gray-800',
    dotColor: 'bg-gray-500',
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
