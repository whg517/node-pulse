import { useNavigate } from 'react-router-dom'
import { memo, useMemo } from 'react'
import { useAlertsStore } from '../../stores/alertsStore'
import { useTheme } from '../../hooks/useTheme'
import { memoCompare } from '../../utils/deepEqual'
import type { AlertRecord } from '../../stores/types'

interface AlertStreamProps {
  maxItems?: number
  className?: string
  isLoading?: boolean
}

type AlertLevel = 'P0' | 'P1' | 'P2'

interface AlertWithNode extends AlertRecord {
  nodeName: string
}

/**
 * Get severity color classes based on alert level
 * P0 (Critical): Red background/border
 * P1 (Warning): Amber/Yellow background/border
 * P2 (Notice): Blue background/border
 */
function getSeverityStyles(level: AlertLevel, isDark: boolean): string {
  const styles: Record<AlertLevel, { light: string; dark: string }> = {
    P0: {
      light: 'bg-red-50 border-red-200 hover:bg-red-100',
      dark: 'bg-red-900/30 border-red-700 hover:bg-red-900/50',
    },
    P1: {
      light: 'bg-amber-50 border-amber-200 hover:bg-amber-100',
      dark: 'bg-amber-900/30 border-amber-700 hover:bg-amber-900/50',
    },
    P2: {
      light: 'bg-blue-50 border-blue-200 hover:bg-blue-100',
      dark: 'bg-blue-900/30 border-blue-700 hover:bg-blue-900/50',
    },
  }
  return isDark ? styles[level].dark : styles[level].light
}

/**
 * Get level badge styles based on alert level
 */
function getLevelBadgeStyles(level: AlertLevel, isDark: boolean): string {
  const styles: Record<AlertLevel, { light: string; dark: string }> = {
    P0: {
      light: 'bg-red-100 text-red-800',
      dark: 'bg-red-900/50 text-red-300',
    },
    P1: {
      light: 'bg-amber-100 text-amber-800',
      dark: 'bg-amber-900/50 text-amber-300',
    },
    P2: {
      light: 'bg-blue-100 text-blue-800',
      dark: 'bg-blue-900/50 text-blue-300',
    },
  }
  return isDark ? styles[level].dark : styles[level].light
}

/**
 * Format relative time (e.g., "2 minutes ago")
 */
function formatTimeAgo(timestamp: string): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  return `${diffDays}d ago`
}

/**
 * Get display text for metric type
 */
function getMetricDisplay(metric: string): string {
  const displays: Record<string, string> = {
    latency: 'Latency',
    packet_loss_rate: 'Packet Loss',
    jitter: 'Jitter',
  }
  return displays[metric] || metric
}

/**
 * AlertStream Component
 *
 * Displays the latest active alerts in a real-time scrollable list.
 * New alerts appear at the top. Color-coded by severity level.
 *
 * @param maxItems - Maximum number of alerts to display (default: 10)
 * @param className - Optional additional CSS classes
 * @param isLoading - Loading state indicator
 *
 * @example
 * <AlertStream maxItems={10} />
 */
export const AlertStream = memo(function AlertStream({
  maxItems = 10,
  className = '',
  isLoading = false,
}: AlertStreamProps) {
  const navigate = useNavigate()
  const { isDark } = useTheme()
  const alertRecords = useAlertsStore((state) => state.alertRecords)

  // Filter and sort alerts - only 'new' or 'acknowledged' status
  // Note: The store uses 'pending' | 'processing' | 'resolved'
  // We treat 'pending' as 'new' and 'processing' as 'acknowledged'
  const activeAlerts = useMemo(() => {
    // Defensive check: ensure alertRecords is an array
    const safeRecords = Array.isArray(alertRecords) ? alertRecords : []

    return safeRecords
      .filter((record) => record.status !== 'resolved')
      .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
      .slice(0, maxItems) as AlertWithNode[]
  }, [alertRecords, maxItems])

  const handleAlertClick = (alertId: string) => {
    navigate(`/alerts?highlight=${alertId}`)
  }

  // Loading skeleton
  if (isLoading) {
    return (
      <div
        className={`rounded-lg border shadow-sm overflow-hidden ${
          isDark ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'
        } ${className}`}
      >
        <div className="px-4 py-3 border-b border-gray-200 dark:border-slate-700">
          <div className="animate-pulse">
            <div className={`h-5 rounded w-32 ${isDark ? 'bg-slate-700' : 'bg-gray-200'}`}></div>
          </div>
        </div>
        <div className="p-3 space-y-2 max-h-96 overflow-y-auto">
          {[...Array(5)].map((_, i) => (
            <div
              key={i}
              className={`h-16 rounded animate-pulse ${isDark ? 'bg-slate-700' : 'bg-gray-100'}`}
            ></div>
          ))}
        </div>
      </div>
    )
  }

  // Empty state
  if (activeAlerts.length === 0) {
    return (
      <div
        className={`rounded-lg border shadow-sm overflow-hidden ${
          isDark ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'
        } ${className}`}
      >
        <div
          className={`px-4 py-3 border-b ${isDark ? 'border-slate-700' : 'border-gray-200'}`}
        >
          <h3
            className={`text-sm font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}
          >
            Active Alerts
          </h3>
        </div>
        <div className="text-center py-8">
          <svg
            className={`mx-auto h-10 w-10 ${isDark ? 'text-green-400' : 'text-green-500'}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
          <p
            className={`mt-2 text-sm font-medium ${isDark ? 'text-gray-300' : 'text-gray-900'}`}
          >
            No active alerts
          </p>
          <p
            className={`mt-1 text-xs ${isDark ? 'text-gray-500' : 'text-gray-500'}`}
          >
            All systems are operating normally
          </p>
        </div>
      </div>
    )
  }

  return (
    <div
      className={`rounded-lg border shadow-sm overflow-hidden ${
        isDark ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'
      } ${className}`}
    >
      {/* Header */}
      <div
        className={`px-4 py-3 border-b ${isDark ? 'border-slate-700' : 'border-gray-200'}`}
      >
        <div className="flex items-center justify-between">
          <h3
            className={`text-sm font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}
          >
            Active Alerts
          </h3>
          <span
            className={`px-2 py-0.5 text-xs font-medium rounded-full ${
              isDark ? 'bg-slate-700 text-gray-300' : 'bg-gray-100 text-gray-600'
            }`}
          >
            {activeAlerts.length}
          </span>
        </div>
      </div>

      {/* Alert List */}
      <ul className="p-3 space-y-2 max-h-96 overflow-y-auto">
        {activeAlerts.map((alert) => {
          const level = (alert.level || 'P2').toUpperCase() as AlertLevel
          const severityStyles = getSeverityStyles(level, isDark)
          const badgeStyles = getLevelBadgeStyles(level, isDark)

          return (
            <li
              key={alert.id}
              onClick={() => handleAlertClick(alert.id)}
              className={`cursor-pointer rounded-lg border p-3 transition-all duration-150 ${severityStyles}`}
              role="button"
              tabIndex={0}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  handleAlertClick(alert.id)
                }
              }}
            >
              <div className="flex items-start justify-between gap-2">
                {/* Left side: Level badge + Alert info */}
                <div className="flex items-start gap-2 min-w-0 flex-1">
                  {/* Level Badge */}
                  <span
                    className={`inline-flex items-center px-2 py-0.5 text-xs font-semibold rounded shrink-0 ${badgeStyles}`}
                  >
                    {level}
                  </span>

                  {/* Alert details */}
                  <div className="min-w-0">
                    <p
                      className={`text-sm font-medium truncate ${
                        isDark ? 'text-gray-100' : 'text-gray-900'
                      }`}
                    >
                      {getMetricDisplay(alert.metric)}
                    </p>
                    <p
                      className={`text-xs truncate ${isDark ? 'text-gray-400' : 'text-gray-500'}`}
                    >
                      Node: {alert.nodeId.slice(0, 8)}...
                    </p>
                  </div>
                </div>

                {/* Right side: Time ago */}
                <span
                  className={`text-xs shrink-0 ${
                    isDark ? 'text-gray-500' : 'text-gray-400'
                  }`}
                >
                  {formatTimeAgo(alert.timestamp)}
                </span>
              </div>
            </li>
          )
        })}
      </ul>

      {/* Footer with view all link */}
      <div
        className={`px-4 py-2 border-t text-center ${
          isDark ? 'border-slate-700' : 'border-gray-200'
        }`}
      >
        <button
          onClick={() => navigate('/alerts')}
          className={`text-xs font-medium hover:underline ${
            isDark ? 'text-blue-400 hover:text-blue-300' : 'text-blue-600 hover:text-blue-700'
          }`}
        >
          View all alerts
        </button>
      </div>
    </div>
  )
}, memoCompare)
