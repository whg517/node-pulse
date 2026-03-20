import { useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import type { MTRHop } from './MTRVisualization'

/**
 * RiskCondition represents a detected risk condition for a hop
 */
export interface RiskCondition {
  type: 'high_loss' | 'high_latency' | 'high_jitter' | 'timeout'
  severity: 'critical' | 'warning'
  value: number
  threshold: number
  message: string
}

/**
 * HopRiskStatus represents the overall risk status for a hop
 */
export type HopRiskStatus = 'safe' | 'warning' | 'critical' | 'timeout'

/**
 * MTRPathVisualizationProps defines the props for the MTR path visualization component
 */
export interface MTRPathVisualizationProps {
  /** MTR hops data to visualize */
  hops: MTRHop[]
  /** Target destination */
  target?: string
  /** Callback when a hop is clicked */
  onHopClick?: (hop: MTRHop) => void
  /** Callback when viewing hop history */
  onViewHistory?: (hop: MTRHop) => void
  /** Additional CSS classes */
  className?: string
  /** Disable interactive features */
  disabled?: boolean
}

/**
 * Risk threshold constants based on PRD requirements
 */
const RISK_THRESHOLDS = {
  loss: {
    critical: 10, // >= 10% packet loss
  },
  latency: {
    critical: 200, // >= 200ms latency
  },
  jitter: {
    warning: 50, // >= 50ms jitter
  },
} as const

/**
 * Risk status configuration for visual styling
 */
const riskStatusConfig: Record<
  HopRiskStatus,
  {
    borderColor: string
    bgColor: string
    textColor: string
    badgeBgColor: string
    badgeTextColor: string
    indicatorColor: string
    shadowColor: string
  }
> = {
  safe: {
    borderColor: 'border-[var(--color-healthy-bg)]',
    bgColor: 'bg-[var(--color-healthy-bg)]',
    textColor: 'text-[var(--color-healthy-text)]',
    badgeBgColor: 'bg-[var(--color-healthy-bg)]',
    badgeTextColor: 'text-[var(--color-healthy-text)]',
    indicatorColor: 'bg-[var(--color-healthy)]',
    shadowColor: 'shadow-sm',
  },
  warning: {
    borderColor: 'border-[var(--color-warning-bg)]',
    bgColor: 'bg-[var(--color-warning-bg)]',
    textColor: 'text-[var(--color-warning-text)]',
    badgeBgColor: 'bg-[var(--color-warning-bg)]',
    badgeTextColor: 'text-[var(--color-warning-text)]',
    indicatorColor: 'bg-[var(--color-warning)]',
    shadowColor: 'shadow-sm',
  },
  critical: {
    borderColor: 'border-[var(--color-critical-bg)]',
    bgColor: 'bg-[var(--color-critical-bg)]',
    textColor: 'text-[var(--color-critical-text)]',
    badgeBgColor: 'bg-[var(--color-critical-bg)]',
    badgeTextColor: 'text-[var(--color-critical-text)]',
    indicatorColor: 'bg-[var(--color-critical)]',
    shadowColor: 'shadow-sm',
  },
  timeout: {
    borderColor: 'border-[var(--color-border)]',
    bgColor: 'bg-[var(--color-bg-muted)]',
    textColor: 'text-[var(--color-text-primary)]',
    badgeBgColor: 'bg-gray-100',
    badgeTextColor: 'text-gray-700',
    indicatorColor: 'bg-gray-400',
    shadowColor: 'shadow-sm',
  },
}

/**
 * Detect risk conditions for a hop based on metrics
 */
function detectRiskConditions(hop: MTRHop): RiskCondition[] {
  const conditions: RiskCondition[] = []

  // Check for timeout (no response)
  if (hop.sent > 0 && hop.received === 0) {
    conditions.push({
      type: 'timeout',
      severity: 'critical',
      value: hop.lossRate,
      threshold: 100,
      message: 'Hop timeout - no response',
    })
    return conditions
  }

  // Check packet loss
  if (hop.lossRate >= RISK_THRESHOLDS.loss.critical) {
    conditions.push({
      type: 'high_loss',
      severity: 'critical',
      value: hop.lossRate,
      threshold: RISK_THRESHOLDS.loss.critical,
      message: `High packet loss: ${hop.lossRate.toFixed(1)}% (threshold: ${RISK_THRESHOLDS.loss.critical}%)`,
    })
  }

  // Check latency (using avg RTT)
  if (hop.avgRTTMs >= RISK_THRESHOLDS.latency.critical) {
    conditions.push({
      type: 'high_latency',
      severity: 'critical',
      value: hop.avgRTTMs,
      threshold: RISK_THRESHOLDS.latency.critical,
      message: `High latency: ${hop.avgRTTMs.toFixed(1)}ms (threshold: ${RISK_THRESHOLDS.latency.critical}ms)`,
    })
  }

  // Check jitter (using standard deviation as proxy)
  if (hop.stdDevMs >= RISK_THRESHOLDS.jitter.warning) {
    conditions.push({
      type: 'high_jitter',
      severity: 'warning',
      value: hop.stdDevMs,
      threshold: RISK_THRESHOLDS.jitter.warning,
      message: `High jitter: ${hop.stdDevMs.toFixed(1)}ms (threshold: ${RISK_THRESHOLDS.jitter.warning}ms)`,
    })
  }

  return conditions
}

/**
 * Determine overall risk status for a hop based on detected conditions
 */
function getHopRiskStatus(hop: MTRHop): HopRiskStatus {
  const conditions = detectRiskConditions(hop)

  if (conditions.length === 0) {
    return 'safe'
  }

  // Check for timeout first
  if (conditions.some((c) => c.type === 'timeout')) {
    return 'timeout'
  }

  // Any critical condition makes the hop critical
  if (conditions.some((c) => c.severity === 'critical')) {
    return 'critical'
  }

  // Any warning condition makes the hop warning
  if (conditions.some((c) => c.severity === 'warning')) {
    return 'warning'
  }

  return 'safe'
}

/**
 * Format RTT value for display
 */
function formatRTT(ms: number): string {
  if (ms === 0) return '-'
  if (ms < 1) return `${ms.toFixed(1)}ms`
  return `${Math.round(ms)}ms`
}

/**
 * Format packet loss rate for display
 */
function formatLossRate(rate: number): string {
  return `${rate.toFixed(1)}%`
}

/**
 * Tooltip component for displaying detailed hop information
 */
interface HopTooltipProps {
  hop: MTRHop
  conditions: RiskCondition[]
  position: { x: number; y: number }
  onClose: () => void
}

function HopTooltip({ hop, conditions, position, onClose }: HopTooltipProps) {
  const { t } = useTranslation()

  return (
    <div
      className="fixed z-50 w-80 bg-white rounded-lg shadow-xl border border-gray-200 overflow-hidden"
      style={{
        top: Math.min(position.y + 10, window.innerHeight - 400),
        left: Math.min(position.x + 10, window.innerWidth - 350),
      }}
      role="tooltip"
      aria-label={`Detailed information for hop ${hop.hopNumber}`}
    >
      {/* Tooltip Header */}
      <div className="bg-gray-50 px-4 py-3 border-b border-gray-200 flex items-center justify-between">
        <h4 className="font-semibold text-gray-900">
          {t('mtr.hopDetail', 'Hop {{number}}', { number: hop.hopNumber })}
        </h4>
        <button
          onClick={onClose}
          className="text-gray-400 hover:text-gray-600 transition-colors"
          aria-label={t('common.close', 'Close')}
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {/* Tooltip Content */}
      <div className="p-4 space-y-3 max-h-96 overflow-y-auto">
        {/* IP and Hostname */}
        <div>
          <p className="text-xs text-gray-500 mb-1">{t('mtr.ipAddress', 'IP Address')}</p>
          <p className="font-mono text-sm font-medium text-gray-900">{hop.ip}</p>
          {hop.hostname && hop.hostname !== hop.ip && (
            <p className="text-xs text-gray-600 mt-0.5">{hop.hostname}</p>
          )}
        </div>

        {/* AS Number */}
        {hop.asNumber && (
          <div>
            <p className="text-xs text-gray-500 mb-1">{t('mtr.asNumber', 'AS Number')}</p>
            <p className="text-sm font-medium text-gray-900">AS{hop.asNumber}</p>
          </div>
        )}

        {/* Location */}
        {hop.location && (
          <div>
            <p className="text-xs text-gray-500 mb-1">{t('mtr.location', 'Location')}</p>
            <p className="text-sm text-gray-700">{hop.location}</p>
          </div>
        )}

        {/* Metrics Grid */}
        <div className="grid grid-cols-2 gap-3 pt-2 border-t border-gray-100">
          <div>
            <p className="text-xs text-gray-500 mb-1">{t('mtr.avgLatency', 'Avg Latency')}</p>
            <p className="text-sm font-semibold text-gray-900">{formatRTT(hop.avgRTTMs)}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 mb-1">{t('mtr.minLatency', 'Min Latency')}</p>
            <p className="text-sm font-medium text-gray-700">{formatRTT(hop.bestRTTMs)}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 mb-1">{t('mtr.maxLatency', 'Max Latency')}</p>
            <p className="text-sm font-medium text-gray-700">{formatRTT(hop.worstRTTMs)}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 mb-1">{t('mtr.jitter', 'Jitter')}</p>
            <p className="text-sm font-medium text-gray-700">{formatRTT(hop.stdDevMs)}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 mb-1">{t('mtr.packetLoss', 'Packet Loss')}</p>
            <p className="text-sm font-semibold text-gray-900">{formatLossRate(hop.lossRate)}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500 mb-1">{t('mtr.packets', 'Packets')}</p>
            <p className="text-sm font-medium text-gray-700">
              {hop.received}/{hop.sent}
            </p>
          </div>
        </div>

        {/* Risk Conditions */}
        {conditions.length > 0 && (
          <div className="pt-2 border-t border-gray-100">
            <p className="text-xs font-medium text-gray-700 mb-2">
              {t('mtr.riskConditions', 'Risk Conditions')}
            </p>
            <div className="space-y-1.5">
              {conditions.map((condition, idx) => (
                <div
                  key={idx}
                  className={`flex items-start gap-2 px-2 py-1.5 rounded text-xs ${
                    condition.severity === 'critical'
                      ? 'bg-[var(--color-critical-bg)] text-[var(--color-critical)]'
                      : 'bg-[var(--color-warning-bg)] text-[var(--color-warning)]'
                  }`}
                >
                  <svg
                    className="w-4 h-4 flex-shrink-0 mt-0.5"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d={
                        condition.severity === 'critical'
                          ? 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z'
                          : 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
                      }
                    />
                  </svg>
                  <span>{condition.message}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Tooltip Footer */}
      <div className="bg-gray-50 px-4 py-2 border-t border-gray-200">
        <button
          onClick={() => {
            onClose()
          }}
          className="w-full text-xs text-[var(--color-brand)] hover:text-[var(--color-brand-hover)] font-medium transition-colors"
        >
          {t('mtr.closeTooltip', 'Click outside to close')}
        </button>
      </div>
    </div>
  )
}

/**
 * MTRPathVisualization Component
 *
 * Provides an enhanced MTR path visualization with:
 * - Risk highlighting based on PRD requirements (loss >= 10%, latency >= 200ms, jitter >= 50ms)
 * - Detailed hop information with tooltips
 * - Interactive features (hover, click)
 * - Accessibility compliance (WCAG 2.1 AA)
 * - Color coding from design system (emerald/amber/red/gray)
 *
 * @param props - MTRPathVisualization props
 * @returns MTRPathVisualization component
 *
 * @example
 * <MTRPathVisualization
 *   hops={mtrHops}
 *   target="8.8.8.8"
 *   onHopClick={(hop) => console.log(hop)}
 *   onViewHistory={(hop) => navigate(`/nodes/${nodeId}/history?hop=${hop.ip}`)}
 * />
 */
export default function MTRPathVisualization({
  hops,
  target,
  onHopClick,
  onViewHistory,
  className = '',
  disabled = false,
}: MTRPathVisualizationProps) {
  const { t } = useTranslation()
  const [hoveredHop, setHoveredHop] = useState<MTRHop | null>(null)
  const [tooltipPosition, setTooltipPosition] = useState<{ x: number; y: number } | null>(null)

  const handleHopMouseEnter = useCallback(
    (hop: MTRHop, event: React.MouseEvent) => {
      if (disabled) return
      setHoveredHop(hop)
      setTooltipPosition({ x: event.clientX, y: event.clientY })
    },
    [disabled]
  )

  const handleHopMouseLeave = useCallback(() => {
    setHoveredHop(null)
    setTooltipPosition(null)
  }, [])

  const handleHopClick = useCallback(
    (hop: MTRHop) => {
      if (disabled) return
      onHopClick?.(hop)
    },
    [disabled, onHopClick]
  )

  const handleHopKeyDown = useCallback(
    (event: React.KeyboardEvent, hop: MTRHop) => {
      if (disabled) return
      if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault()
        onHopClick?.(hop)
      }
    },
    [disabled, onHopClick]
  )

  // Handle empty hops
  if (!hops || hops.length === 0) {
    return (
      <div
        className={`mtr-path-visualization bg-white rounded-lg shadow-sm p-6 ${className}`}
        role="region"
        aria-label={t('mtr.pathVisualization', 'MTR Path Visualization')}
      >
        <div className="flex flex-col items-center justify-center py-12 text-center" role="img">
          <svg
            className="h-16 w-16 text-gray-300 mb-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
            />
          </svg>
          <p className="text-gray-500 font-medium">{t('mtr.noHopData', 'No hop data available')}</p>
          <p className="text-gray-400 text-sm mt-1">{t('mtr.runMTRToSeePath', 'Run an MTR trace to see the network path')}</p>
        </div>
      </div>
    )
  }

  return (
    <div
      className={`mtr-path-visualization bg-white rounded-lg shadow-sm p-4 ${className}`}
      role="region"
      aria-label={t('mtr.pathVisualization', 'MTR Path Visualization')}
    >
      {/* Header with target info */}
      {target && (
        <div className="mb-4 pb-3 border-b border-gray-100">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-base font-semibold text-gray-900">
                {t('mtr.networkPath', 'Network Path')}
              </h3>
              <p className="text-sm text-gray-500 mt-0.5">
                {t('mtr.pathTo', 'Path to')}:{' '}
                <span className="font-mono font-medium text-gray-700">{target}</span>
              </p>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-500">
                {hops.length} {t('mtr.hops', 'hops')}
              </span>
            </div>
          </div>
        </div>
      )}

      {/* Path visualization */}
      <div className="space-y-2" role="list" aria-label={t('mtr.hopList', 'Network hops')}>
        {hops.map((hop, index) => {
          const riskStatus = getHopRiskStatus(hop)
          const conditions = detectRiskConditions(hop)
          const config = riskStatusConfig[riskStatus]
          const isLastHop = index === hops.length - 1
          const isInteractive = !disabled && (onHopClick || onViewHistory)

          return (
            <div key={`${hop.hopNumber}-${hop.ip}`} className="relative">
              {/* Hop Card */}
              <div
                className={`relative flex items-center gap-3 p-3 rounded-lg border-2 transition-all duration-200 ${
                  config.bgColor
                } ${config.borderColor} ${
                  isInteractive
                    ? 'cursor-pointer hover:shadow-md hover:scale-[1.01] active:scale-[0.99]'
                    : ''
                } ${config.shadowColor}`}
                role={isInteractive ? 'button' : 'listitem'}
                tabIndex={isInteractive ? 0 : undefined}
                aria-label={t('mtr.hopAriaLabel', 'Hop {{number}}: {{ip}} - {{status}}', {
                  number: hop.hopNumber,
                  ip: hop.ip,
                  status: t(`mtr.riskStatus.${riskStatus}`, riskStatus),
                })}
                aria-pressed={riskStatus === 'critical' || riskStatus === 'timeout'}
                onClick={() => handleHopClick(hop)}
                onKeyDown={(e) => handleHopKeyDown(e, hop)}
                onMouseEnter={(e) => handleHopMouseEnter(hop, e)}
                onMouseLeave={handleHopMouseLeave}
              >
                {/* Hop Number Indicator */}
                <div className="flex-shrink-0">
                  <div
                    className={`w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm ${
                      config.badgeBgColor
                    } ${config.textColor} shadow-sm`}
                    aria-hidden="true"
                  >
                    {hop.hopNumber}
                  </div>
                </div>

                {/* Connection Line */}
                {!isLastHop && (
                  <div
                    className={`absolute left-8 top-10 w-0.5 h-full ${
                      riskStatus === 'critical' || riskStatus === 'timeout'
                        ? 'bg-gray-300'
                        : config.indicatorColor
                    } opacity-50`}
                    style={{ height: 'calc(100% + 8px)' }}
                    aria-hidden="true"
                  />
                )}

                {/* Hop Details */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    {/* IP Address */}
                    <span className="font-mono font-medium text-gray-900 text-sm">
                      {hop.ip}
                    </span>

                    {/* Risk Indicator Badge */}
                    {conditions.length > 0 && (
                      <div className="flex items-center gap-1">
                        {conditions.map((condition, idx) => (
                          <span
                            key={idx}
                            className={`inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium ${
                              condition.severity === 'critical'
                                ? 'bg-[var(--color-critical-bg)] text-[var(--color-critical)]'
                                : 'bg-[var(--color-warning-bg)] text-[var(--color-warning)]'
                            }`}
                            title={condition.message}
                          >
                            {condition.type === 'high_loss' && '📉'}
                            {condition.type === 'high_latency' && '⏱️'}
                            {condition.type === 'high_jitter' && '〰️'}
                            {condition.type === 'timeout' && '⚠️'}
                          </span>
                        ))}
                      </div>
                    )}

                    {/* AS Number */}
                    {hop.asNumber && (
                      <span className="text-xs text-gray-500 font-mono bg-gray-100 px-1.5 py-0.5 rounded">
                        AS{hop.asNumber}
                      </span>
                    )}
                  </div>

                  {/* Location */}
                  {hop.location && (
                    <p className="text-xs text-gray-500 mt-0.5 flex items-center gap-1">
                      <svg
                        className="w-3 h-3"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        aria-hidden="true"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
                        />
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
                        />
                      </svg>
                      {hop.location}
                    </p>
                  )}

                  {/* Metrics Row */}
                  <div className="flex items-center gap-4 mt-2 text-xs flex-wrap">
                    {/* Packet Loss */}
                    <div className="flex items-center gap-1.5">
                      <span
                        className={`w-2 h-2 rounded-full ${config.indicatorColor}`}
                        aria-hidden="true"
                      />
                      <span className="text-gray-600">{t('mtr.loss', 'Loss')}:</span>
                      <span
                        className={`font-semibold ${
                          hop.lossRate >= 10
                            ? 'text-[var(--color-critical)]'
                            : hop.lossRate >= 5
                            ? 'text-[var(--color-warning)]'
                            : 'text-gray-700'
                        }`}
                      >
                        {formatLossRate(hop.lossRate)}
                      </span>
                    </div>

                    {/* Avg Latency */}
                    <div className="flex items-center gap-1">
                      <span className="text-gray-600">{t('mtr.avg', 'Avg')}:</span>
                      <span
                        className={`font-medium ${
                          hop.avgRTTMs >= 200
                            ? 'text-[var(--color-critical)]'
                            : hop.avgRTTMs >= 100
                            ? 'text-[var(--color-warning)]'
                            : 'text-gray-700'
                        }`}
                      >
                        {formatRTT(hop.avgRTTMs)}
                      </span>
                    </div>

                    {/* Min/Max Latency */}
                    <div className="flex items-center gap-1 text-gray-500">
                      <span>
                        {formatRTT(hop.bestRTTMs)} - {formatRTT(hop.worstRTTMs)}
                      </span>
                    </div>

                    {/* Jitter */}
                    {hop.stdDevMs > 0 && (
                      <div className="flex items-center gap-1">
                        <span className="text-gray-600">{t('mtr.jitter', 'Jitter')}:</span>
                        <span
                          className={`font-medium ${
                            hop.stdDevMs >= 50
                              ? 'text-[var(--color-warning)]'
                              : 'text-gray-700'
                          }`}
                        >
                          {formatRTT(hop.stdDevMs)}
                        </span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Status Indicator Icon */}
                <div className="flex-shrink-0" aria-hidden="true">
                  {riskStatus === 'critical' && (
                    <svg className="w-5 h-5 text-[var(--color-critical)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                      />
                    </svg>
                  )}
                  {riskStatus === 'warning' && (
                    <svg className="w-5 h-5 text-[var(--color-warning)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                  )}
                  {riskStatus === 'timeout' && (
                    <svg className="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636"
                      />
                    </svg>
                  )}
                  {riskStatus === 'safe' && (
                    <svg className="w-5 h-5 text-[var(--color-healthy)]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                  )}
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Legend */}
      <div className="mt-4 pt-3 border-t border-gray-100">
        <p className="text-xs text-gray-500 mb-2">{t('mtr.riskLegend', 'Risk indicators')}:</p>
        <div className="flex items-center gap-3 flex-wrap text-xs">
          <div className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-full bg-[var(--color-healthy)]" aria-hidden="true" />
            <span className="text-gray-600">
              {t('mtr.riskSafe', 'Normal')} ({t('mtr.riskSafeDesc', '< 10% loss, < 200ms latency')})
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-full bg-[var(--color-warning)]" aria-hidden="true" />
            <span className="text-gray-600">
              {t('mtr.riskWarning', 'Warning')} ({t('mtr.riskWarningDesc', '≥ 50ms jitter')})
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-full bg-[var(--color-critical)]" aria-hidden="true" />
            <span className="text-gray-600">
              {t('mtr.riskCritical', 'Critical')} ({t('mtr.riskCriticalDesc', '≥ 10% loss or ≥ 200ms latency')})
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="w-3 h-3 rounded-full bg-gray-400" aria-hidden="true" />
            <span className="text-gray-600">
              {t('mtr.riskTimeout', 'Timeout')} ({t('mtr.riskTimeoutDesc', 'No response')})
            </span>
          </div>
        </div>
      </div>

      {/* Tooltip */}
      {hoveredHop && tooltipPosition && (
        <HopTooltip
          hop={hoveredHop}
          conditions={detectRiskConditions(hoveredHop)}
          position={tooltipPosition}
          onClose={() => {
            setHoveredHop(null)
            setTooltipPosition(null)
          }}
        />
      )}
    </div>
  )
}
