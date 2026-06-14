/**
 * Utility functions for formatting data in the UI
 */

/**
 * Format a timestamp as a human-readable string
 *
 * @param timestamp - ISO timestamp string
 * @returns Formatted timestamp string
 *
 * @example
 * formatTimestamp('2024-01-01T12:00:00Z')
 * // Returns: "1 minute ago" or "2 hours ago" or "2024-01-01 12:00:00"
 */
export function formatTimestamp(timestamp: string | undefined | null): string {
  if (!timestamp) return 'N/A'

  try {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`

    const diffHours = Math.floor(diffMins / 60)
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`

    const diffDays = Math.floor(diffHours / 24)
    if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`

    return date.toLocaleString()
  } catch {
    return 'Invalid timestamp'
  }
}

/**
 * Format a number with specified decimal places
 *
 * @param value - Number to format
 * @param decimals - Number of decimal places (default: 2)
 * @returns Formatted number string
 *
 * @example
 * formatNumber(45.6789, 2)
 * // Returns: "45.68"
 */
export function formatNumber(value: number | undefined | null, decimals = 2): string {
  if (value === undefined || value === null) return 'N/A'
  return value.toFixed(decimals)
}

/**
 * Format a percentage value
 *
 * @param value - Percentage value (0-100)
 * @param decimals - Number of decimal places (default: 2)
 * @returns Formatted percentage string
 *
 * @example
 * formatPercentage(0.5, 2)
 * // Returns: "0.50%"
 */
export function formatPercentage(value: number | undefined | null, decimals = 2): string {
  if (value === undefined || value === null) return 'N/A'
  return `${value.toFixed(decimals)}%`
}

/**
 * Determine health status based on metrics
 *
 * @param latency - Latency in ms
 * @param packetLossRate - Packet loss rate (0-100)
 * @param jitter - Jitter in ms
 * @returns Health status
 *
 * @example
 * getHealthStatus(45, 0, 5)
 * // Returns: "good"
 */
export function getHealthStatus(
  latency: number | undefined | null,
  packetLossRate: number | undefined | null,
  jitter: number | undefined | null
): 'good' | 'warning' | 'critical' | 'neutral' {
  if (latency === null || latency === undefined) return 'neutral'
  if (packetLossRate === null || packetLossRate === undefined) return 'neutral'

  if (packetLossRate > 5 || latency > 500) return 'critical'
  if (packetLossRate > 2 || latency > 200 || (jitter !== null && jitter !== undefined && jitter > 50)) {
    return 'warning'
  }
  if (packetLossRate === 0 && latency < 100) return 'good'

  return 'neutral'
}

/**
 * Get status badge color classes
 *
 * @param status - Node status
 * @returns Tailwind CSS classes for status badge
 *
 * @example
 * getStatusBadgeClasses('online')
 * // Returns: "bg-green-100 text-green-800"
 */
export function getStatusBadgeClasses(
  status: 'online' | 'offline' | 'connecting' | string
): string {
  const statusClasses: Record<string, string> = {
    online: 'bg-healthy-bg text-healthy-text',
    offline: 'bg-destructive/10 text-destructive',
    connecting: 'bg-warning-bg text-warning-text',
  }

  return statusClasses[status] || 'bg-muted text-muted-foreground'
}

/**
 * Get status indicator color classes
 *
 * @param status - Node status
 * @returns Tailwind CSS classes for status indicator
 *
 * @example
 * getStatusIndicatorClasses('online')
 * // Returns: "bg-green-500"
 */
export function getStatusIndicatorClasses(
  status: 'online' | 'offline' | 'connecting' | string
): string {
  const indicatorClasses: Record<string, string> = {
    online: 'bg-healthy',
    offline: 'bg-destructive',
    connecting: 'bg-warning',
  }

  return indicatorClasses[status] || 'bg-muted-foreground'
}
