/**
 * useDashboard Hook
 *
 * Provides dashboard-specific data aggregation and statistics.
 * Works alongside useDashboardData for enhanced dashboard functionality.
 */

import { useMemo } from 'react'
import type { NodeDTO, MetricsDTO } from '../api/types'
import { determineHealthStatus, type HealthStatus } from '../utils/healthStatus'

export interface DashboardStats {
  totalNodes: number
  onlineNodes: number
  offlineNodes: number
  healthyNodes: number
  warningNodes: number
  criticalNodes: number
  unknownNodes: number
  onlineRate: number
  anomalyRate: number
  averageLatency: number
  averagePacketLoss: number
  averageJitter: number
}

export interface NodeHealthSummary {
  node: NodeDTO
  metrics?: MetricsDTO
  healthStatus: HealthStatus
}

export interface UseDashboardResult {
  stats: DashboardStats
  nodeHealthSummaries: NodeHealthSummary[]
  sortedByAnomaly: NodeHealthSummary[]
}

/**
 * Hook for computing dashboard statistics and node health summaries
 *
 * @param nodes - Array of nodes
 * @param metrics - Array of metrics
 * @returns Computed dashboard data
 *
 * @example
 * const { stats, nodeHealthSummaries, sortedByAnomaly } = useDashboard(nodes, metrics)
 */
export function useDashboard(nodes: NodeDTO[], metrics: MetricsDTO[]): UseDashboardResult {
  // Create metrics map for quick lookup
  const metricsMap = useMemo(() => {
    const safeMetrics = Array.isArray(metrics) ? metrics : []
    return new Map(safeMetrics.map(m => [m.node_id, m]))
  }, [metrics])

  // Compute health status for each node
  const nodeHealthSummaries = useMemo((): NodeHealthSummary[] => {
    const safeNodes = Array.isArray(nodes) ? nodes : []
    return safeNodes.map(node => {
      const nodeMetrics = metricsMap.get(node.id)
      const healthStatus = nodeMetrics
        ? determineHealthStatus({
            latency_ms: nodeMetrics.latency_ms,
            packet_loss_rate: nodeMetrics.packet_loss_rate,
            jitter_ms: nodeMetrics.jitter_ms,
            last_heartbeat: nodeMetrics.timestamp,
          })
        : 'offline'

      return {
        node,
        metrics: nodeMetrics,
        healthStatus,
      }
    })
  }, [nodes, metricsMap])

  // Compute overall statistics
  const stats = useMemo((): DashboardStats => {
    const totalNodes = nodeHealthSummaries.length

    if (totalNodes === 0) {
      return {
        totalNodes: 0,
        onlineNodes: 0,
        offlineNodes: 0,
        healthyNodes: 0,
        warningNodes: 0,
        criticalNodes: 0,
        unknownNodes: 0,
        onlineRate: 0,
        anomalyRate: 0,
        averageLatency: 0,
        averagePacketLoss: 0,
        averageJitter: 0,
      }
    }

    const onlineNodes = nodeHealthSummaries.filter(
      s => s.healthStatus !== 'offline'
    ).length
    const offlineNodes = totalNodes - onlineNodes
    const healthyNodes = nodeHealthSummaries.filter(
      s => s.healthStatus === 'healthy'
    ).length
    const warningNodes = nodeHealthSummaries.filter(
      s => s.healthStatus === 'warning'
    ).length
    const criticalNodes = nodeHealthSummaries.filter(
      s => s.healthStatus === 'critical'
    ).length
    const unknownNodes = nodeHealthSummaries.filter(
      s => s.healthStatus === 'offline'
    ).length

    const onlineRate = (onlineNodes / totalNodes) * 100
    const anomalyRate = ((warningNodes + criticalNodes) / totalNodes) * 100

    // Compute average metrics from online nodes only
    const onlineMetrics = nodeHealthSummaries
      .filter(s => s.metrics)
      .map(s => s.metrics!)

    const averageLatency = onlineMetrics.length > 0
      ? onlineMetrics.reduce((sum, m) => sum + m.latency_ms, 0) / onlineMetrics.length
      : 0
    const averagePacketLoss = onlineMetrics.length > 0
      ? onlineMetrics.reduce((sum, m) => sum + m.packet_loss_rate, 0) / onlineMetrics.length
      : 0
    const averageJitter = onlineMetrics.length > 0
      ? onlineMetrics.reduce((sum, m) => sum + m.jitter_ms, 0) / onlineMetrics.length
      : 0

    return {
      totalNodes,
      onlineNodes,
      offlineNodes,
      healthyNodes,
      warningNodes,
      criticalNodes,
      unknownNodes,
      onlineRate,
      anomalyRate,
      averageLatency,
      averagePacketLoss,
      averageJitter,
    }
  }, [nodeHealthSummaries])

  // Sort nodes by anomaly (critical first, then warning)
  const sortedByAnomaly = useMemo(() => {
    const severityOrder: Record<HealthStatus, number> = {
      critical: 0,
      warning: 1,
      healthy: 2,
      offline: 3,
    }

    return [...nodeHealthSummaries].sort(
      (a, b) => severityOrder[a.healthStatus] - severityOrder[b.healthStatus]
    )
  }, [nodeHealthSummaries])

  return {
    stats,
    nodeHealthSummaries,
    sortedByAnomaly,
  }
}
