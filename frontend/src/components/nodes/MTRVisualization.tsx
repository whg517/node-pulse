import { useTranslation } from 'react-i18next'

/**
 * MTRHop represents a single hop in an MTR trace
 */
export interface MTRHop {
  hopNumber: number
  ip: string
  hostname?: string
  asNumber?: string
  sent: number
  received: number
  lossRate: number
  lastRTTMs: number
  avgRTTMs: number
  bestRTTMs: number
  worstRTTMs: number
  stdDevMs: number
  location?: string
}

/**
 * MTRResult represents the complete result of an MTR probe
 */
export interface MTRResult {
  target: string
  totalHops: number
  hops: MTRHop[]
  completedAt: string
  success: boolean
  errorMessage?: string
}

export interface MTRVisualizationProps {
  /** MTR result data to display */
  data?: MTRResult | null
  /** Loading state when MTR is running */
  isLoading?: boolean
  /** Error state for failed MTR */
  error?: string | null
  /** Additional CSS classes */
  className?: string
  /** Callback when a hop is clicked */
  onHopClick?: (hop: MTRHop) => void
}

/** Hop health status based on packet loss rate */
type HopHealthStatus = 'healthy' | 'degraded' | 'problematic'

/**
 * Get hop health status based on packet loss rate
 * - healthy: < 5% loss
 * - degraded: 5-20% loss
 * - problematic: > 20% loss
 */
function getHopHealthStatus(lossRate: number): HopHealthStatus {
  if (lossRate > 20) return 'problematic'
  if (lossRate > 5) return 'degraded'
  return 'healthy'
}

/** Health status configuration for visual styling */
const healthStatusConfig: Record<
  HopHealthStatus,
  {
    borderColor: string
    bgColor: string
    textColor: string
    badgeBgColor: string
    badgeTextColor: string
    dotColor: string
  }
> = {
  healthy: {
    borderColor: 'border-green-300',
    bgColor: 'bg-green-50',
    textColor: 'text-green-800',
    badgeBgColor: 'bg-green-100',
    badgeTextColor: 'text-green-700',
    dotColor: 'bg-green-500',
  },
  degraded: {
    borderColor: 'border-yellow-300',
    bgColor: 'bg-yellow-50',
    textColor: 'text-yellow-800',
    badgeBgColor: 'bg-yellow-100',
    badgeTextColor: 'text-yellow-700',
    dotColor: 'bg-yellow-500',
  },
  problematic: {
    borderColor: 'border-red-300',
    bgColor: 'bg-red-50',
    textColor: 'text-red-800',
    badgeBgColor: 'bg-red-100',
    badgeTextColor: 'text-red-700',
    dotColor: 'bg-red-500',
  },
}

/**
 * Format RTT value for display
 */
function formatRTT(ms: number): string {
  if (ms === 0) return '-'
  return `${ms.toFixed(1)}ms`
}

/**
 * Format packet loss rate for display
 */
function formatLossRate(rate: number): string {
  return `${rate.toFixed(1)}%`
}

/**
 * Calculate overall path health status based on hops
 */
function getPathHealthStatus(hops: MTRHop[]): HopHealthStatus {
  if (hops.length === 0) return 'healthy'

  const problematicCount = hops.filter((h) => getHopHealthStatus(h.lossRate) === 'problematic').length
  const degradedCount = hops.filter((h) => getHopHealthStatus(h.lossRate) === 'degraded').length

  // If any hop is problematic, path is problematic
  if (problematicCount > 0) return 'problematic'
  // If multiple hops are degraded, path is problematic
  if (degradedCount >= 2) return 'problematic'
  // If any hop is degraded, path is degraded
  if (degradedCount > 0) return 'degraded'
  return 'healthy'
}

/**
 * MTRVisualization Component
 *
 * Displays MTR (My Traceroute) results as a vertical timeline/hop list with
 * color-coded health status indicators and RTT statistics.
 *
 * Features:
 * - Vertical timeline visualization of network hops
 * - Color-coded health status (green < 5%, yellow 5-20%, red > 20% packet loss)
 * - RTT statistics (avg, min, max, std dev) in compact format
 * - AS number display when available
 * - Loading, error, and empty states
 * - Fully accessible with ARIA attributes
 * - Internationalized with i18n
 *
 * @param props - MTRVisualization props
 * @returns MTRVisualization component
 *
 * @example
 * <MTRVisualization
 *   data={{
 *     target: '8.8.8.8',
 *     totalHops: 10,
 *     hops: [...],
 *     completedAt: '2024-01-15T10:30:00Z',
 *     success: true
 *   }}
 * />
 */
export default function MTRVisualization({
  data,
  isLoading = false,
  error = null,
  className = '',
  onHopClick,
}: MTRVisualizationProps) {
  const { t } = useTranslation()

  // Loading state
  if (isLoading) {
    return (
      <div
        className={`mtr-visualization bg-white rounded-lg shadow-sm p-4 ${className}`}
        role="region"
        aria-label={t('mtr.title')}
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">{t('mtr.title')}</h3>
        </div>
        <div
          className="flex flex-col items-center justify-center py-12"
          role="status"
          aria-label={t('common.loading')}
        >
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
          <p className="mt-3 text-gray-600">{t('mtr.running')}</p>
        </div>
      </div>
    )
  }

  // Empty state
  if (!data) {
    return (
      <div
        className={`mtr-visualization bg-white rounded-lg shadow-sm p-4 ${className}`}
        role="region"
        aria-label={t('mtr.title')}
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">{t('mtr.title')}</h3>
        </div>
        <div className="flex flex-col items-center justify-center py-12 text-center" role="img" aria-label={t('mtr.noData')}>
          <svg
            className="h-12 w-12 text-gray-400 mb-3"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
            />
          </svg>
          <p className="text-gray-500">{t('mtr.noData')}</p>
        </div>
      </div>
    )
  }

  // Error state (from error prop or data.success check)
  if (error || !data.success) {
    const errorMessage = error || data.errorMessage
    return (
      <div
        className={`mtr-visualization bg-white rounded-lg shadow-sm p-4 ${className}`}
        role="region"
        aria-label={t('mtr.title')}
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">{t('mtr.title')}</h3>
        </div>
        <div
          className="flex flex-col items-center justify-center py-12 text-center"
          role="alert"
          aria-label={t('mtr.error')}
        >
          <svg
            className="h-12 w-12 text-red-500 mb-3"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
          <p className="text-red-600 font-medium mb-1">{t('mtr.error')}</p>
          {errorMessage && (
            <p className="text-gray-500 text-sm">{errorMessage}</p>
          )}
        </div>
      </div>
    )
  }

  const pathHealth = getPathHealthStatus(data.hops)
  const pathHealthConfig = healthStatusConfig[pathHealth]

  return (
    <div
      className={`mtr-visualization bg-white rounded-lg shadow-sm p-4 ${className}`}
      role="region"
      aria-label={t('mtr.title')}
    >
      {/* Summary Header */}
      <div className="mb-4 pb-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-gray-900">{t('mtr.title')}</h3>
            <p className="text-sm text-gray-500 mt-1">
              {t('mtr.target')}: <span className="font-mono font-medium text-gray-700">{data.target}</span>
            </p>
          </div>
          <div className="flex items-center gap-3">
            {/* Path Health Status Badge */}
            <span
              className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${pathHealthConfig.badgeBgColor} ${pathHealthConfig.badgeTextColor}`}
              role="status"
              aria-label={`${t('mtr.pathStatus')}: ${t(`mtr.healthStatus.${pathHealth}`)}`}
            >
              <span className={`w-2 h-2 rounded-full ${pathHealthConfig.dotColor}`} aria-hidden="true" />
              {t(`mtr.healthStatus.${pathHealth}`)}
            </span>
          </div>
        </div>
        <div className="flex items-center gap-4 mt-3 text-sm text-gray-600">
          <span>
            {t('mtr.totalHops')}: <span className="font-medium text-gray-700">{data.totalHops}</span>
          </span>
          {data.completedAt && (
            <span>
              {t('mtr.completedAt')}:{' '}
              <span className="font-medium text-gray-700">
                {new Date(data.completedAt).toLocaleString()}
              </span>
            </span>
          )}
        </div>
      </div>

      {/* Hop List */}
      <div className="space-y-3" role="list" aria-label={t('mtr.hopList')}>
        {data.hops.map((hop, index) => {
          const hopHealth = getHopHealthStatus(hop.lossRate)
          const hopConfig = healthStatusConfig[hopHealth]
          const isLastHop = index === data.hops.length - 1

          const handleHopClick = () => {
            onHopClick?.(hop)
          }

          const handleHopKeyDown = (event: React.KeyboardEvent) => {
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault()
              onHopClick?.(hop)
            }
          }

          return (
            <div
              key={`${hop.hopNumber}-${hop.ip}-${hop.hostname ?? 'unknown'}-${index}`}
              className={`relative flex items-start gap-3 p-3 rounded-lg border ${hopConfig.borderColor} ${hopConfig.bgColor} ${onHopClick ? 'cursor-pointer hover:opacity-80 transition-opacity' : ''}`}
              role={onHopClick ? 'button' : 'listitem'}
              tabIndex={onHopClick ? 0 : undefined}
              onClick={handleHopClick}
              onKeyDown={handleHopKeyDown}
              aria-label={t('mtr.hopAriaLabel', { number: hop.hopNumber, ip: hop.ip })}
            >
              {/* Timeline connector */}
              <div className="flex flex-col items-center">
                {/* Hop number badge */}
                <div
                  className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold ${hopConfig.badgeBgColor} ${hopConfig.textColor}`}
                  aria-hidden="true"
                >
                  {hop.hopNumber}
                </div>
                {/* Connector line */}
                {!isLastHop && (
                  <div className="w-0.5 h-full bg-gray-300 mt-2 min-h-[20px]" aria-hidden="true" />
                )}
              </div>

              {/* Hop details */}
              <div className="flex-1 min-w-0">
                {/* IP and Hostname */}
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-mono font-medium text-gray-900">{hop.ip}</span>
                  {hop.hostname && hop.hostname !== hop.ip && (
                    <span className="text-gray-600 text-sm">({hop.hostname})</span>
                  )}
                  {hop.asNumber && (
                    <span className="text-xs text-gray-500 font-mono bg-gray-200 px-1.5 py-0.5 rounded">
                      AS{hop.asNumber}
                    </span>
                  )}
                </div>

                {/* Location */}
                {hop.location && (
                  <p className="text-sm text-gray-500 mt-0.5">{hop.location}</p>
                )}

                {/* Statistics Row */}
                <div className="flex items-center gap-4 mt-2 text-sm flex-wrap">
                  {/* Packet Loss */}
                  <div className="flex items-center gap-1.5">
                    <span className={`inline-block w-2 h-2 rounded-full ${hopConfig.dotColor}`} aria-hidden="true" />
                    <span className="text-gray-600">{t('mtr.loss')}:</span>
                    <span className={`font-medium ${hopConfig.textColor}`}>
                      {formatLossRate(hop.lossRate)}
                    </span>
                  </div>

                  {/* RTT Stats */}
                  <div className="flex items-center gap-2 text-gray-600">
                    <span>{t('mtr.avg')}:</span>
                    <span className="font-medium text-gray-700">{formatRTT(hop.avgRTTMs)}</span>
                  </div>

                  <div className="flex items-center gap-2 text-gray-600">
                    <span>{t('mtr.min')}:</span>
                    <span className="font-medium text-gray-700">{formatRTT(hop.bestRTTMs)}</span>
                  </div>

                  <div className="flex items-center gap-2 text-gray-600">
                    <span>{t('mtr.max')}:</span>
                    <span className="font-medium text-gray-700">{formatRTT(hop.worstRTTMs)}</span>
                  </div>

                  {hop.stdDevMs > 0 && (
                    <div className="flex items-center gap-2 text-gray-600">
                      <span>{t('mtr.stdDev')}:</span>
                      <span className="font-medium text-gray-700">{formatRTT(hop.stdDevMs)}</span>
                    </div>
                  )}
                </div>

                {/* Packets info (collapsed by default) */}
                <div className="text-xs text-gray-400 mt-1">
                  {t('mtr.packets')}: {hop.received}/{hop.sent} {t('mtr.received')}
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Legend */}
      <div className="mt-4 pt-4 border-t border-gray-200">
        <p className="text-xs text-gray-500 mb-2">{t('mtr.legend')}:</p>
        <div className="flex items-center gap-4 flex-wrap">
          {(['healthy', 'degraded', 'problematic'] as const).map((status) => {
            const config = healthStatusConfig[status]
            return (
              <div key={status} className="flex items-center gap-1.5">
                <span className={`w-3 h-3 rounded-full ${config.dotColor}`} aria-hidden="true" />
                <span className="text-xs text-gray-600">
                  {t(`mtr.healthStatus.${status}`)}: {t(`mtr.legendDescription.${status}`)}
                </span>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
