/**
 * Node Summary Card Component
 *
 * Displays a summary card for a node with health status indicator,
 * node name, region, and last seen timestamp.
 */

import { memo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
// Note: useTimezone is available for future time formatting needs
import type { NodeDTO } from '../../api/types'
import type { HealthStatus } from '../../utils/healthStatus'

// Color palette — CSS variables handle both light and dark themes
const HEALTH_COLORS = {
  healthy: {
    bg: 'bg-[var(--color-healthy-bg)]',
    text: 'text-[var(--color-healthy-text)]',
    dot: 'bg-[var(--color-healthy)]',
    border: 'border-[var(--color-healthy-bg)]',
  },
  warning: {
    bg: 'bg-[var(--color-warning-bg)]',
    text: 'text-[var(--color-warning-text)]',
    dot: 'bg-[var(--color-warning)]',
    border: 'border-[var(--color-warning-bg)]',
  },
  critical: {
    bg: 'bg-[var(--color-critical-bg)]',
    text: 'text-[var(--color-critical-text)]',
    dot: 'bg-[var(--color-critical)]',
    border: 'border-[var(--color-critical-bg)]',
  },
  offline: {
    bg: 'bg-[var(--color-unknown-bg)]',
    text: 'text-[var(--color-unknown-text)]',
    dot: 'bg-[var(--color-unknown)]',
    border: 'border-[var(--color-border)]',
  },
} as const

export interface NodeSummaryCardProps {
  node: NodeDTO
  healthStatus: HealthStatus
  lastSeen?: string
  latency?: number
  packetLoss?: number
  className?: string
}

/**
 * NodeSummaryCard Component
 *
 * @param node - Node data
 * @param healthStatus - Health status of the node
 * @param lastSeen - ISO timestamp of last heartbeat
 * @param latency - Current latency in ms
 * @param packetLoss - Current packet loss percentage
 * @param className - Additional CSS classes
 */
export const NodeSummaryCard = memo(function NodeSummaryCard({
  node,
  healthStatus,
  lastSeen,
  latency,
  packetLoss,
  className = '',
}: NodeSummaryCardProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  // Timezone formatting available via useTimezone hook if needed

  const colors = HEALTH_COLORS[healthStatus] || HEALTH_COLORS.offline

  const handleClick = () => {
    navigate(`/nodes/${node.id}`, { state: { breadcrumbLabel: node.name } })
  }

  const formatLastSeen = (timestamp: string | undefined): string => {
    if (!timestamp) return t('status.unknown')

    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMinutes = Math.floor(diffMs / 60000)

    if (diffMinutes < 1) return t('time.justNow')
    if (diffMinutes < 60) return t('time.minutesAgo', { count: diffMinutes })

    const diffHours = Math.floor(diffMinutes / 60)
    if (diffHours < 24) return t('time.hoursAgo', { count: diffHours })

    const diffDays = Math.floor(diffHours / 24)
    return t('time.daysAgo', { count: diffDays })
  }

  return (
    <div
      onClick={handleClick}
      className={`
        node-summary-card
        bg-[var(--color-bg-muted)]
        rounded-lg border
        ${colors.border}
        p-4 cursor-pointer
        hover:shadow-md transition-shadow duration-200
        ${className}
      `}
      role="button"
      tabIndex={0}
      aria-label={`${node.name} - ${t(`status.${healthStatus}`)}`}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          handleClick()
        }
      }}
    >
      {/* Header with status indicator */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center space-x-2">
          <span
            className={`w-3 h-3 rounded-full ${colors.dot}`}
            aria-hidden="true"
          />
          <h4 className="font-semibold text-[var(--color-text-primary)] truncate">
            {node.name}
          </h4>
        </div>
        <span
          className={`
            px-2 py-0.5 rounded-full text-xs font-medium
            ${colors.bg} ${colors.text}
          `}
        >
          {t(`status.${healthStatus}`)}
        </span>
      </div>

      {/* Region */}
      <div className="text-sm text-[var(--color-text-secondary)] mb-2">
        <span className="font-medium">{t('nodes.region')}:</span> {node.region}
      </div>

      {/* Metrics row */}
      {(latency !== undefined || packetLoss !== undefined) && (
        <div className="flex items-center space-x-4 text-sm">
          {latency !== undefined && (
            <div className="flex items-center space-x-1">
              <svg
                className="w-4 h-4 text-[var(--color-text-muted)]"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <span className="text-[var(--color-text-secondary)]">
                {latency.toFixed(0)}{t('units.ms')}
              </span>
            </div>
          )}
          {packetLoss !== undefined && (
            <div className="flex items-center space-x-1">
              <svg
                className="w-4 h-4 text-[var(--color-text-muted)]"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
              <span className="text-[var(--color-text-secondary)]">
                {packetLoss.toFixed(1)}{t('units.percent')}
              </span>
            </div>
          )}
        </div>
      )}

      {/* Last seen */}
      <div className="mt-3 pt-3 border-t border-[var(--color-border)] text-xs text-[var(--color-text-muted)]">
        <span>{t('nodes.lastSeen')}:</span>{' '}
        <span>{formatLastSeen(lastSeen)}</span>
      </div>

      {/* Tags */}
      {Array.isArray(node.tags) && node.tags.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1">
          {node.tags.slice(0, 3).map((tag) => (
            <span
              key={tag}
              className="px-2 py-0.5 bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] text-xs rounded"
            >
              {tag}
            </span>
          ))}
          {node.tags.length > 3 && (
            <span className="text-xs text-[var(--color-text-muted)]">
              +{node.tags.length - 3}
            </span>
          )}
        </div>
      )}
    </div>
  )
})
