/**
 * Dashboard Page
 *
 * Enhanced dashboard with node health overview, real-time metrics,
 * world map, alert stream, charts from historical API, and quick actions.
 */

import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useDashboardData } from '../hooks/useDashboardData'
import { useDashboard } from '../hooks/useDashboard'
import { useDashboardHistory } from '../hooks/useDashboardHistory'
import { useThemeColors } from '../hooks/useThemeColors'
import { PageContainer, ErrorBanner, ActionButton } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { NodeListTable } from '../components/dashboard/NodeListTable'
import { AlertStream } from '../components/dashboard/AlertStream'
import { MetricsSummaryCards } from '../components/dashboard/MetricsSummaryCards'
import { NodeSummaryCard } from '../components/dashboard/NodeSummaryCard'
import WorldMap from '../components/dashboard/WorldMap'
import { LatencyTrendChart, PacketLossChart, ProbeSuccessGauge } from '../components/charts'
import type { NodeLocation } from '../components/dashboard/WorldMap'
import { useAlertsStore } from '../stores/alertsStore'
import { estimateRegionBaseCoordinates, scatterAroundBase } from '../utils/regionCoordinates'

export default function DashboardPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const themeColors = useThemeColors()
  const { nodes, metrics, isLoading, error, refetch } = useDashboardData()
  const { stats, sortedByAnomaly, nodeHealthSummaries } = useDashboard(nodes, metrics)
  const fetchAlertRecords = useAlertsStore((s) => s.fetchAlertRecords)

  const [historyRefreshToken, setHistoryRefreshToken] = useState(0)

  const nodeIdsForHistory = useMemo(() => nodes.map((n) => n.id), [nodes])
  const { latencyTrend, packetLossTrend, isLoading: historyLoading } = useDashboardHistory(
    nodeIdsForHistory,
    historyRefreshToken
  )

  useEffect(() => {
    void fetchAlertRecords()
  }, [fetchAlertRecords])

  const handleRefetch = useCallback(async () => {
    await refetch()
    setHistoryRefreshToken((n) => n + 1)
    void fetchAlertRecords()
  }, [refetch, fetchAlertRecords])

  const nodeMapLocations = useMemo((): NodeLocation[] => {
    return nodeHealthSummaries.map(({ node, metrics: nodeMetrics, healthStatus }) => {
      const base = estimateRegionBaseCoordinates(node.region)
      const { lat, lng } = scatterAroundBase(base, node.id)
      return {
        id: node.id,
        name: node.name,
        lat,
        lng,
        region: node.region,
        healthStatus,
        avgLatency: nodeMetrics?.latency_ms ?? 0,
        packetLoss: nodeMetrics?.packet_loss_rate ?? 0,
      }
    })
  }, [nodeHealthSummaries])

  const chartsLoading = isLoading || historyLoading
  const latencyBaseline =
    stats.averageLatency > 0 ? stats.averageLatency : 100

  // Get top 6 nodes for display
  const topNodes = sortedByAnomaly.slice(0, 6)

  return (
    <PageContainer>
      <PageHeader
        title={t('dashboard.title')}
        subtitle={t('dashboard.realTimeMetrics')}
        actions={
          <div className="flex items-center space-x-3">
            <ActionButton variant="secondary" onClick={() => void handleRefetch()}>
              <svg className="-ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              {t('dashboard.refreshData')}
            </ActionButton>
            <ActionButton onClick={() => navigate('/nodes')}>
              {t('dashboard.viewAllNodes')}
            </ActionButton>
          </div>
        }
      />

      {error && (
        <ErrorBanner error={error} onRetry={() => void handleRefetch()} className="mb-6" />
      )}

      {/* World map — node distribution (UI design §4.1–4.2) */}
      <div className="mb-8">
        <WorldMap
          nodes={nodeMapLocations}
          height="420px"
          isLoading={isLoading}
          onNodeClick={(nodeId) => navigate(`/nodes/${nodeId}`)}
        />
      </div>

      {/* Health Overview Stats */}
      <div className="mb-8 grid grid-cols-2 md:grid-cols-4 gap-4">
        {/* Online Rate */}
        <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
              {t('metrics.onlineRate')}
            </h3>
          </div>
          <div className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-2xl font-bold text-[var(--color-text-primary)]">
                  {stats.onlineRate.toFixed(1)}{t('units.percent')}
                </p>
              </div>
              <div
                className="w-12 h-12 rounded-full flex items-center justify-center"
                style={{ backgroundColor: `${themeColors.healthy}20` }}
              >
                <svg
                  className="w-6 h-6"
                  style={{ color: themeColors.healthy }}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </div>
            </div>
            <div className="mt-3 flex items-center text-xs">
              <span className="text-[var(--color-text-muted)]">
                {stats.onlineNodes}/{stats.totalNodes} {t('status.online').toLowerCase()}
              </span>
            </div>
          </div>
        </div>

        {/* Anomaly Rate */}
        <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
              {t('metrics.anomalyRate')}
            </h3>
          </div>
          <div className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p
                  className="text-2xl font-bold text-[var(--color-text-primary)]"
                  style={{
                    color: stats.anomalyRate > 10
                      ? themeColors.critical
                      : stats.anomalyRate > 5
                        ? themeColors.warning
                        : undefined,
                  }}
                >
                  {stats.anomalyRate.toFixed(1)}{t('units.percent')}
                </p>
              </div>
              <div
                className="w-12 h-12 rounded-full flex items-center justify-center"
                style={{
                  backgroundColor: stats.anomalyRate > 10
                    ? `${themeColors.critical}20`
                    : stats.anomalyRate > 5
                      ? `${themeColors.warning}20`
                      : `${themeColors.unknown}20`,
                }}
              >
                <svg
                  className="w-6 h-6"
                  style={{
                    color: stats.anomalyRate > 10
                      ? themeColors.critical
                      : stats.anomalyRate > 5
                        ? themeColors.warning
                        : themeColors.unknown,
                  }}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                  />
                </svg>
              </div>
            </div>
            <div className="mt-3 flex items-center text-xs">
              <span style={{ color: themeColors.warning }}>{stats.warningNodes}</span>
              <span className="mx-1 text-[var(--color-border-strong)]">|</span>
              <span style={{ color: themeColors.critical }}>{stats.criticalNodes}</span>
              <span className="ml-1 text-[var(--color-text-muted)]">{t('dashboard.nodesRequiringAttention')}</span>
            </div>
          </div>
        </div>

        {/* Average Latency */}
        <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
              {t('metrics.avgLatency')}
            </h3>
          </div>
          <div className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p
                  className="text-2xl font-bold text-[var(--color-text-primary)]"
                  style={{
                    color: stats.averageLatency > 200
                      ? themeColors.critical
                      : stats.averageLatency > 160
                        ? themeColors.warning
                        : undefined,
                  }}
                >
                  {stats.averageLatency.toFixed(0)}{t('units.ms')}
                </p>
              </div>
              <div
                className="w-12 h-12 rounded-full flex items-center justify-center"
                style={{ backgroundColor: `${themeColors.brand}20` }}
              >
                <svg
                  className="w-6 h-6"
                  style={{ color: themeColors.brand }}
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
              </div>
            </div>
          </div>
        </div>

        {/* Probe Success Rate Gauge */}
        <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
              {t('dashboard.probeSuccessRate')}
            </h3>
          </div>
          <div className="p-4">
            <ProbeSuccessGauge
              value={100 - stats.averagePacketLoss}
              height="140px"
              isLoading={isLoading}
            />
          </div>
        </div>
      </div>

      {/* Charts Row — last 24h cluster average from /data/history */}
      <div className="mb-8 grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
              {t('dashboard.latencyTrendChart')}
            </h3>
          </div>
          <div className="p-4">
            <LatencyTrendChart
              data={latencyTrend}
              height="250px"
              isLoading={chartsLoading}
              showBaseline
              baselineValue={latencyBaseline}
            />
          </div>
        </div>

        <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
              {t('dashboard.packetLossChart')}
            </h3>
          </div>
          <div className="p-4">
            <PacketLossChart
              data={packetLossTrend}
              height="250px"
              isLoading={chartsLoading}
              warningThreshold={3}
              criticalThreshold={5}
            />
          </div>
        </div>
      </div>

      {/* Metrics Summary Cards */}
      <div className="mb-8">
        <h3 className="text-lg font-semibold mb-4 text-[var(--color-text-primary)]">
          {t('dashboard.averageMetrics')}
        </h3>
        <MetricsSummaryCards metrics={metrics} isLoading={isLoading} />
      </div>

      {/* Node health cards + alert stream (UI design §4.1) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
              {t('dashboard.nodeHealthOverview')}
            </h3>
            <button
              type="button"
              onClick={() => navigate('/nodes')}
              className="text-[var(--color-brand)] hover:text-[var(--color-brand-hover)] text-sm font-medium"
            >
              {t('dashboard.viewAllNodes')} &rarr;
            </button>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {topNodes.map(({ node, metrics: nodeMetrics, healthStatus }) => (
              <NodeSummaryCard
                key={node.id}
                node={node}
                healthStatus={healthStatus}
                lastSeen={nodeMetrics?.timestamp}
                latency={nodeMetrics?.latency_ms}
                packetLoss={nodeMetrics?.packet_loss_rate}
              />
            ))}
            {topNodes.length === 0 && !isLoading && (
              <div className="col-span-2 text-center py-8 text-[var(--color-text-secondary)]">
                {t('nodes.noNodes')}
              </div>
            )}
          </div>
        </div>

        <div className="lg:col-span-1">
          <AlertStream isLoading={isLoading} />
        </div>
      </div>

      {/* Node List Table */}
      <div className="mt-8">
        <NodeListTable nodes={nodes} metrics={metrics} isLoading={isLoading} />
      </div>

      {/* Auto-refresh indicator */}
      {!isLoading && nodes.length > 0 && (
        <div className="mt-6 text-center text-sm flex items-center justify-center text-[var(--color-text-muted)]">
          <svg
            className="inline-block h-4 w-4 mr-1"
            style={{ color: themeColors.brand }}
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
          {t('dashboard.autoRefresh', { interval: 5 })}
        </div>
      )}
    </PageContainer>
  )
}
