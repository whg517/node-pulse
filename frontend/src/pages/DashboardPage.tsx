/**
 * Dashboard Page
 *
 * Enhanced dashboard with node health overview, real-time metrics,
 * charts, and quick action buttons.
 */

import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useDashboardData } from '../hooks/useDashboardData'
import { useDashboard } from '../hooks/useDashboard'
import { PageContainer, ErrorBanner, ActionButton } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { statusColors } from '../config/designTokens'
import { NodeListTable } from '../components/dashboard/NodeListTable'
import { TopAnomaliesList } from '../components/dashboard/TopAnomaliesList'
import { MetricsSummaryCards } from '../components/dashboard/MetricsSummaryCards'
import { NodeSummaryCard } from '../components/dashboard/NodeSummaryCard'
import { LatencyTrendChart, PacketLossChart, ProbeSuccessGauge } from '../components/charts'
import type { DataPoint } from '../components/dashboard/TrendChart'

// Use design tokens for consistent color palette
const HEALTH_COLORS = {
  healthy: statusColors.healthy.main,
  warning: statusColors.warning.main,
  critical: statusColors.critical.main,
  unknown: statusColors.unknown.main,
}

/**
 * Generate sample trend data for charts
 * Note: In production, this data would come from the API
 */
function generateTrendData(baseValue: number, variance: number, points: number = 24): DataPoint[] {
  const now = new Date('2026-01-01T00:00:00Z')
  const data: DataPoint[] = []
  for (let i = points - 1; i >= 0; i--) {
    const timestamp = new Date(now.getTime() - i * 3600000) // hourly data
    const phase = (points - i) / points * Math.PI * 2
    const value = baseValue + Math.sin(phase) * (variance * 0.5)
    data.push({
      timestamp: timestamp.toISOString(),
      value: Math.max(0, value),
    })
  }
  return data
}

export default function DashboardPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const { nodes, metrics, isLoading, error, refetch } = useDashboardData()
  const { stats, sortedByAnomaly } = useDashboard(nodes, metrics)

  const latencyTrendData = generateTrendData(stats.averageLatency, 20)
  const packetLossTrendData = generateTrendData(stats.averagePacketLoss, 2)

  // Get top 6 nodes for display
  const topNodes = sortedByAnomaly.slice(0, 6)

  return (
    <PageContainer>
      <PageHeader
        title={t('dashboard.title')}
        subtitle={t('dashboard.realTimeMetrics')}
        actions={
          <div className="flex items-center space-x-3">
            <ActionButton variant="secondary" onClick={() => refetch()}>
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
        <ErrorBanner error={error} onRetry={refetch} className="mb-6" />
      )}

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
                style={{ backgroundColor: `${HEALTH_COLORS.healthy}20` }}
              >
                <svg
                  className="w-6 h-6"
                  style={{ color: HEALTH_COLORS.healthy }}
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
                      ? HEALTH_COLORS.critical
                      : stats.anomalyRate > 5
                        ? HEALTH_COLORS.warning
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
                    ? `${HEALTH_COLORS.critical}20`
                    : stats.anomalyRate > 5
                      ? `${HEALTH_COLORS.warning}20`
                      : `${HEALTH_COLORS.unknown}20`,
                }}
              >
                <svg
                  className="w-6 h-6"
                  style={{
                    color: stats.anomalyRate > 10
                      ? HEALTH_COLORS.critical
                      : stats.anomalyRate > 5
                        ? HEALTH_COLORS.warning
                        : HEALTH_COLORS.unknown,
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
              <span style={{ color: HEALTH_COLORS.warning }}>{stats.warningNodes}</span>
              <span className="mx-1 text-[var(--color-border-strong)]">|</span>
              <span style={{ color: HEALTH_COLORS.critical }}>{stats.criticalNodes}</span>
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
                      ? HEALTH_COLORS.critical
                      : stats.averageLatency > 160
                        ? HEALTH_COLORS.warning
                        : undefined,
                  }}
                >
                  {stats.averageLatency.toFixed(0)}{t('units.ms')}
                </p>
              </div>
              <div
                className="w-12 h-12 rounded-full flex items-center justify-center"
                style={{ backgroundColor: '#3b82f620' }}
              >
                <svg
                  className="w-6 h-6 text-blue-500"
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

      {/* Charts Row */}
      <div className="mb-8 grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Latency Trend Chart */}
        <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
              {t('dashboard.latencyTrendChart')}
            </h3>
          </div>
          <div className="p-4">
            <LatencyTrendChart
              data={latencyTrendData}
              height="250px"
              isLoading={isLoading}
              showBaseline
              baselineValue={100}
            />
          </div>
        </div>

        {/* Packet Loss Chart */}
        <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
              {t('dashboard.packetLossChart')}
            </h3>
          </div>
          <div className="p-4">
            <PacketLossChart
              data={packetLossTrendData}
              height="250px"
              isLoading={isLoading}
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

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Node Health Cards (2/3 width on large screens) */}
        <div className="lg:col-span-2">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">
              {t('dashboard.nodeHealthOverview')}
            </h3>
            <button
              onClick={() => navigate('/nodes')}
              className="text-blue-600 hover:text-blue-700 text-sm font-medium"
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

        {/* Top Anomalies List (1/3 width on large screens) */}
        <div className="lg:col-span-1">
          <TopAnomaliesList nodes={nodes} metrics={metrics} isLoading={isLoading} />
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
            className="inline-block h-4 w-4 mr-1 text-blue-500"
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
