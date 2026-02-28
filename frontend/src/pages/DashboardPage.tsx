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
import { useTheme } from '../hooks/useTheme'
import { NodeListTable } from '../components/dashboard/NodeListTable'
import { TopAnomaliesList } from '../components/dashboard/TopAnomaliesList'
import { MetricsSummaryCards } from '../components/dashboard/MetricsSummaryCards'
import { NodeSummaryCard } from '../components/dashboard/NodeSummaryCard'
import { LatencyTrendChart, PacketLossChart, ProbeSuccessGauge } from '../components/charts'
import type { DataPoint } from '../components/dashboard/TrendChart'

// Color palette from UI design
const HEALTH_COLORS = {
  healthy: '#22C55E',
  warning: '#F59E0B',
  critical: '#EF4444',
  unknown: '#6B7280',
}

/**
 * Generate sample trend data for charts
 * Note: In production, this data would come from the API
 */
function generateTrendData(baseValue: number, variance: number, points: number = 24): DataPoint[] {
  const now = new Date()
  const data: DataPoint[] = []
  for (let i = points - 1; i >= 0; i--) {
    const timestamp = new Date(now.getTime() - i * 3600000) // hourly data
    const value = baseValue + (Math.random() - 0.5) * variance
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
  const { isDark } = useTheme()

  const latencyTrendData = generateTrendData(stats.averageLatency, 20)
  const packetLossTrendData = generateTrendData(stats.averagePacketLoss, 2)

  // Get top 6 nodes for display
  const topNodes = sortedByAnomaly.slice(0, 6)

  return (
    <div className={`min-h-screen ${isDark ? 'bg-slate-900' : 'bg-gray-50'}`}>
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Page Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h2 className={`text-3xl font-bold ${isDark ? 'text-white' : 'text-gray-900'}`}>
                {t('dashboard.title')}
              </h2>
              <p className={`mt-2 ${isDark ? 'text-gray-400' : 'text-gray-600'}`}>
                {t('dashboard.realTimeMetrics')}
              </p>
            </div>
            <div className="flex items-center space-x-3">
              <button
                onClick={() => refetch()}
                className={`inline-flex items-center px-4 py-2 border rounded-md shadow-sm text-sm font-medium transition-colors ${
                  isDark
                    ? 'border-slate-600 bg-slate-700 text-white hover:bg-slate-600'
                    : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
                }`}
              >
                <svg
                  className="-ml-1 mr-2 h-4 w-4"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
                {t('dashboard.refreshData')}
              </button>
              <button
                onClick={() => navigate('/nodes')}
                className="inline-flex items-center px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md shadow-sm text-sm font-medium transition-colors"
              >
                {t('dashboard.viewAllNodes')}
              </button>
            </div>
          </div>
        </div>

        {/* Error State */}
        {error && (
          <div className="mb-6 bg-red-50 dark:bg-red-900/20 border-l-4 border-red-400 p-4 rounded-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-400"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                  aria-hidden="true"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-700 dark:text-red-300">{error.message}</p>
              </div>
              <div className="ml-auto pl-3">
                <button
                  onClick={() => refetch()}
                  className="inline-flex bg-red-50 dark:bg-red-900/50 rounded-md p-1.5 text-red-500 hover:bg-red-100 dark:hover:bg-red-900"
                >
                  <svg
                    className="h-5 w-5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    aria-hidden="true"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                    />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Health Overview Stats */}
        <div className="mb-8 grid grid-cols-2 md:grid-cols-4 gap-4">
          {/* Online Rate */}
          <div className={`rounded-lg border ${isDark ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'}`}>
            <div className={`px-4 py-3 border-b ${isDark ? 'border-slate-700' : 'border-gray-200'}`}>
              <h3 className={`text-sm font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}>
                {t('metrics.onlineRate')}
              </h3>
            </div>
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className={`text-2xl font-bold ${isDark ? 'text-white' : 'text-gray-900'}`}>
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
                <span className={isDark ? 'text-gray-500' : 'text-gray-400'}>
                  {stats.onlineNodes}/{stats.totalNodes} {t('status.online').toLowerCase()}
                </span>
              </div>
            </div>
          </div>

          {/* Anomaly Rate */}
          <div className={`rounded-lg border ${isDark ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'}`}>
            <div className={`px-4 py-3 border-b ${isDark ? 'border-slate-700' : 'border-gray-200'}`}>
              <h3 className={`text-sm font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}>
                {t('metrics.anomalyRate')}
              </h3>
            </div>
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p
                    className="text-2xl font-bold"
                    style={{
                      color: stats.anomalyRate > 10
                        ? HEALTH_COLORS.critical
                        : stats.anomalyRate > 5
                          ? HEALTH_COLORS.warning
                          : isDark ? '#fff' : '#111',
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
                <span className={`mx-1 ${isDark ? 'text-gray-600' : 'text-gray-300'}`}>|</span>
                <span style={{ color: HEALTH_COLORS.critical }}>{stats.criticalNodes}</span>
                <span className={`ml-1 ${isDark ? 'text-gray-500' : 'text-gray-400'}`}>{t('dashboard.nodesRequiringAttention')}</span>
              </div>
            </div>
          </div>

          {/* Average Latency */}
          <div className={`rounded-lg border ${isDark ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'}`}>
            <div className={`px-4 py-3 border-b ${isDark ? 'border-slate-700' : 'border-gray-200'}`}>
              <h3 className={`text-sm font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}>
                {t('metrics.avgLatency')}
              </h3>
            </div>
            <div className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p
                    className="text-2xl font-bold"
                    style={{
                      color: stats.averageLatency > 200
                        ? HEALTH_COLORS.critical
                        : stats.averageLatency > 160
                          ? HEALTH_COLORS.warning
                          : isDark ? '#fff' : '#111',
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
          <div className={`rounded-lg border ${isDark ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'}`}>
            <div className={`px-4 py-3 border-b ${isDark ? 'border-slate-700' : 'border-gray-200'}`}>
              <h3 className={`text-sm font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}>
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
          <div className={`rounded-lg border ${isDark ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'}`}>
            <div className={`px-4 py-3 border-b ${isDark ? 'border-slate-700' : 'border-gray-200'}`}>
              <h3 className={`text-lg font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}>
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
          <div className={`rounded-lg border ${isDark ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'}`}>
            <div className={`px-4 py-3 border-b ${isDark ? 'border-slate-700' : 'border-gray-200'}`}>
              <h3 className={`text-lg font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}>
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
          <h3 className={`text-lg font-semibold mb-4 ${isDark ? 'text-white' : 'text-gray-900'}`}>
            {t('dashboard.averageMetrics')}
          </h3>
          <MetricsSummaryCards metrics={metrics} isLoading={isLoading} />
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Node Health Cards (2/3 width on large screens) */}
          <div className="lg:col-span-2">
            <div className="mb-4 flex items-center justify-between">
              <h3 className={`text-lg font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}>
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
                <div className={`col-span-2 text-center py-8 ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
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
          <div className={`mt-6 text-center text-sm flex items-center justify-center ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
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
      </main>
    </div>
  )
}
