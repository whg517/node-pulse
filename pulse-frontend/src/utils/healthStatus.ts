/**
 * Health Status Determination Logic
 *
 * Provides utilities for determining node health status based on metrics
 * and detecting offline nodes.
 */

export type HealthStatus = 'healthy' | 'warning' | 'critical' | 'offline'

export interface HealthThresholds {
  latency: number      // default: 200ms
  packetLoss: number   // default: 5%
  jitter: number       // default: 50ms
}

export interface NodeMetrics {
  latency_ms: number
  packet_loss_rate: number
  jitter_ms: number
  last_heartbeat: string
}

/**
 * Default health thresholds
 */
export const DEFAULT_THRESHOLDS: HealthThresholds = {
  latency: 200,
  packetLoss: 5,
  jitter: 50,
}

/**
 * Determine if a node is offline based on last heartbeat timestamp
 *
 * A node is considered offline if no heartbeat has been received
 * in the last 120 seconds (2 heartbeat cycles) or more.
 *
 * @param lastHeartbeat - ISO timestamp of last heartbeat
 * @returns true if node is offline, false otherwise
 *
 * @example
 * isNodeOffline('2026-01-26T10:00:00Z') // true if current time > 120s later
 */
export function isNodeOffline(lastHeartbeat: string): boolean {
  if (!lastHeartbeat || lastHeartbeat === '') {
    return true // No heartbeat data = offline
  }

  const heartbeatTime = new Date(lastHeartbeat).getTime()
  const now = Date.now()
  const offlineThreshold = 120 * 1000 // 120 seconds

  return (now - heartbeatTime) >= offlineThreshold
}

/**
 * Determine health status based on node metrics
 *
 * Health status determination rules:
 * - Offline: No heartbeat for >=120 seconds
 * - Critical: Any metric exceeds threshold
 * - Warning: Any metric within 80-100% of threshold OR heartbeat 60-120 seconds ago
 * - Healthy: All metrics below 80% of threshold and recent heartbeat (<60 seconds)
 *
 * @param metrics - Node metrics including latency, packet loss, jitter, heartbeat
 * @param thresholds - Health threshold values (defaults to DEFAULT_THRESHOLDS)
 * @returns Health status classification
 *
 * @example
 * determineHealthStatus({
 *   latency_ms: 150,
 *   packet_loss_rate: 2,
 *   jitter_ms: 30,
 *   last_heartbeat: '2026-01-26T10:00:00Z'
 * }) // 'healthy'
 */
export function determineHealthStatus(
  metrics: NodeMetrics,
  thresholds: HealthThresholds = DEFAULT_THRESHOLDS
): HealthStatus {
  // Handle missing or invalid metrics
  if (!metrics) {
    return 'offline'
  }

  // Extract metrics with defaults for missing values
  const latency = metrics.latency_ms ?? 0
  const packetLoss = metrics.packet_loss_rate ?? 0
  const jitter = metrics.jitter_ms ?? 0

  // Offline check (highest priority) - >=120 seconds
  if (isNodeOffline(metrics.last_heartbeat)) {
    return 'offline'
  }

  // Critical check - any metric exceeds threshold
  if (
    latency > thresholds.latency ||
    packetLoss > thresholds.packetLoss ||
    jitter > thresholds.jitter
  ) {
    return 'critical'
  }

  // Warning check for stale heartbeat (>60 seconds but <120 seconds)
  const heartbeatTime = new Date(metrics.last_heartbeat).getTime()
  const now = Date.now()
  const heartbeatAge = now - heartbeatTime
  const staleHeartbeatThreshold = 60 * 1000 // 60 seconds

  if (heartbeatAge > staleHeartbeatThreshold) {
    return 'warning'
  }

  // Warning check - any metric within 80-100% of threshold
  const warningThreshold = 0.8 // 80%

  if (
    latency > thresholds.latency * warningThreshold ||
    packetLoss > thresholds.packetLoss * warningThreshold ||
    jitter > thresholds.jitter * warningThreshold
  ) {
    return 'warning'
  }

  // All checks passed - node is healthy
  return 'healthy'
}
