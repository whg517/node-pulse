import { useParams, Link } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Breadcrumb } from '../components/layout/Breadcrumb'
import { useNodeDetail } from '../hooks/useNodeDetail'
import { useTimezone } from '../hooks/useTimezone'
import MetricCard from '../components/dashboard/MetricCard'
import ProblemDiagnosis, {
  type ProblemType,
  type ConfidenceLevel,
} from '../components/dashboard/ProblemDiagnosis'
import TrendChart, {
  type TimeRange,
  type DataPoint,
} from '../components/dashboard/TrendChart'
import { fetchHistory } from '../api/data'

/**
 * NodeDetailPage component
 *
 * Displays detailed information about a single node including:
 * - Basic node information (name, IP, region, tags)
 * - Real-time metrics (latency, packet loss rate, jitter)
 * - Node status (online/offline/connecting)
 * - Last heartbeat timestamp
 * - Problem diagnosis with expandable details
 *
 * @returns NodeDetailPage component
 */
export default function NodeDetailPage() {
  const { t } = useTranslation()
  const { formatTime } = useTimezone()
  const { id } = useParams<{ id: string }>()
  const { node, nodeStatus, metrics, isLoading, error, isPolling } = useNodeDetail(id || '')

  // Historical data state
  const [timeRange, setTimeRange] = useState<TimeRange>('24h')
  const [historyData, setHistoryData] = useState<{
    latency_ms: DataPoint[]
    packet_loss_rate: DataPoint[]
    jitter_ms: DataPoint[]
  }>({
    latency_ms: [],
    packet_loss_rate: [],
    jitter_ms: [],
  })
  const [baselines, setBaselines] = useState<{
    latency_ms?: number
    packet_loss_rate?: number
    jitter_ms?: number
  }>({})
  const [isLoadingHistory, setIsLoadingHistory] = useState(false)
  const [historyError, setHistoryError] = useState<string | null>(null)

  // Fetch historical data when node ID or time range changes
  useEffect(() => {
    if (!id) return

    const fetchHistoricalData = async () => {
      setIsLoadingHistory(true)
      setHistoryError(null)

      try {
        const now = new Date()
        let startTime: Date

        // Calculate start time based on selected range
        switch (timeRange) {
          case '24h':
            startTime = new Date(now.getTime() - 24 * 60 * 60 * 1000)
            break
          case '7d':
            startTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
            break
          case '30d':
            startTime = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
            break
          default:
            startTime = new Date(now.getTime() - 24 * 60 * 60 * 1000)
        }

        // Determine aggregation based on time range
        let aggregation: '1m' | '5m' = '1m'
        if (timeRange === '7d' || timeRange === '30d') {
          aggregation = '5m'
        }

        // Fetch data for all three metrics in parallel
        const [latencyResult, lossResult, jitterResult] = await Promise.all([
          fetchHistory({
            node_ids: [id],
            start_time: startTime.toISOString(),
            end_time: now.toISOString(),
            metrics: ['latency'],
            aggregation,
          }),
          fetchHistory({
            node_ids: [id],
            start_time: startTime.toISOString(),
            end_time: now.toISOString(),
            metrics: ['packet_loss_rate'],
            aggregation,
          }),
          fetchHistory({
            node_ids: [id],
            start_time: startTime.toISOString(),
            end_time: now.toISOString(),
            metrics: ['jitter'],
            aggregation,
          }),
        ])

        // Process latency data
        const latencySeries = latencyResult.data.find(
          (series) => series.metric === 'latency'
        )
        const latencyDataPoints: DataPoint[] =
          latencySeries?.data_points.map((dp) => ({
            timestamp: dp.timestamp,
            value: dp.value,
          })) || []

        // Process packet loss data
        const lossSeries = lossResult.data.find(
          (series) => series.metric === 'packet_loss_rate'
        )
        const lossDataPoints: DataPoint[] =
          lossSeries?.data_points.map((dp) => ({
            timestamp: dp.timestamp,
            value: dp.value,
          })) || []

        // Process jitter data
        const jitterSeries = jitterResult.data.find(
          (series) => series.metric === 'jitter'
        )
        const jitterDataPoints: DataPoint[] =
          jitterSeries?.data_points.map((dp) => ({
            timestamp: dp.timestamp,
            value: dp.value,
          })) || []

        // Update history data state
        setHistoryData({
          latency_ms: latencyDataPoints,
          packet_loss_rate: lossDataPoints,
          jitter_ms: jitterDataPoints,
        })

        // Calculate baselines for longer time ranges (7d and 30d)
        if (timeRange === '7d' || timeRange === '30d') {
          const calculateBaseline = (dataPoints: DataPoint[]): number | undefined => {
            if (dataPoints.length === 0) return undefined

            // Calculate average (baseline)
            const sum = dataPoints.reduce((acc, dp) => acc + dp.value, 0)
            return sum / dataPoints.length
          }

          setBaselines({
            latency_ms: calculateBaseline(latencyDataPoints),
            packet_loss_rate: calculateBaseline(lossDataPoints),
            jitter_ms: calculateBaseline(jitterDataPoints),
          })
        } else {
          // No baseline for 24h view
          setBaselines({})
        }
      } catch (err) {
        console.error('Failed to fetch historical data:', err)
        setHistoryError(t('errors.loadHistoricalError'))
      } finally {
        setIsLoadingHistory(false)
      }
    }

    fetchHistoricalData()
  }, [id, timeRange, t])

  // Handle time range change
  const handleTimeRangeChange = (newRange: TimeRange) => {
    setTimeRange(newRange)
  }

  // Determine problem type based on metrics
  const getProblemType = (): ProblemType => {
    if (!metrics || !nodeStatus) return 'none'

    const { packet_loss_rate, latency_ms, jitter_ms } = metrics

    // Critical: Complete outage or severe degradation
    if (nodeStatus.status === 'offline' || packet_loss_rate > 50) {
      return 'node_local'
    }

    // Severe: Very high packet loss
    if (packet_loss_rate > 10) {
      return 'node_local'
    }

    // Severe: Very high latency
    if (latency_ms > 1000) {
      return 'node_local'
    }

    // Warning: Elevated metrics
    if (packet_loss_rate > 3 || latency_ms > 300 || jitter_ms > 100) {
      return 'node_local'
    }

    // Mild: Slightly elevated metrics
    if (packet_loss_rate > 1 || latency_ms > 150 || jitter_ms > 50) {
      return 'node_local'
    }

    return 'none'
  }

  const getConfidence = (): ConfidenceLevel => {
    if (!metrics || !nodeStatus) return 'low'

    const { packet_loss_rate, latency_ms, jitter_ms } = metrics
    const severityScore = Math.max(
      packet_loss_rate / 10,
      latency_ms / 200,
      jitter_ms / 50
    )

    if (
      nodeStatus.status === 'offline' ||
      packet_loss_rate > 10 ||
      latency_ms > 500
    ) {
      return 'high'
    }

    if (severityScore > 2) {
      return 'medium'
    }

    if (severityScore > 1) {
      return 'low'
    }

    return 'low'
  }

  // Format timestamp using i18n
  const formatTimestamp = (timestamp: string | undefined): string => {
    if (!timestamp) return 'N/A'

    try {
      const date = new Date(timestamp)
      const now = new Date()
      const diffMs = now.getTime() - date.getTime()
      const diffMins = Math.floor(diffMs / 60000)

      if (diffMins < 1) return t('time.justNow')
      if (diffMins < 60) return t('time.minutesAgo', { count: diffMins })

      const diffHours = Math.floor(diffMins / 60)
      if (diffHours < 24) return t('time.hoursAgo', { count: diffHours })

      return formatTime(date)
    } catch {
      return 'N/A'
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--color-bg-page)]">
        <div className="text-center">
          <div
            className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
            role="status"
            aria-label={t('common.loading')}
          />
          <p className="mt-4 text-[var(--color-text-secondary)]">
            {t('common.loading')}
          </p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--color-bg-page)]">
        <div className="rounded-lg shadow-md p-6 max-w-md bg-[var(--color-bg-surface)]">
          <div className="text-red-600 text-5xl mb-4" aria-hidden="true">
            ⚠️
          </div>
          <h2 className="text-xl font-semibold text-[var(--color-text-primary)] mb-2">
            {t('errors.failedToLoad')}
          </h2>
          <p className="text-[var(--color-text-secondary)] mb-4">{error.message}</p>
          <Link
            to="/dashboard"
            className="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            {t('common.back')}
          </Link>
        </div>
      </div>
    )
  }

  if (!node) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--color-bg-page)]">
        <div className="rounded-lg shadow-md p-6 max-w-md bg-[var(--color-bg-surface)]">
          <h2 className="text-xl font-semibold text-[var(--color-text-primary)] mb-2">
            {t('errors.nodeNotFound')}
          </h2>
          <p className="text-[var(--color-text-secondary)] mb-4">{t('errors.notFound')}</p>
          <Link
            to="/dashboard"
            className="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            {t('common.back')}
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[var(--color-bg-page)]">
      {/* Header */}
      <header className="border-b border-[var(--color-border)] bg-[var(--color-bg-surface)] shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="mb-3">
            <Breadcrumb />
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div>
                <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
                  {node.name}
                </h1>
                <p className="text-sm text-[var(--color-text-secondary)]">{node.ip}</p>
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <div
                className={`flex items-center space-x-2 px-3 py-1 rounded-full text-sm font-medium ${
                  nodeStatus?.status === 'online'
                    ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                    : nodeStatus?.status === 'connecting'
                    ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
                    : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                }`}
              >
                <span
                  className={`w-2 h-2 rounded-full ${
                    nodeStatus?.status === 'online'
                      ? 'bg-green-500'
                      : nodeStatus?.status === 'connecting'
                      ? 'bg-yellow-500'
                      : 'bg-red-500'
                  }`}
                  aria-hidden="true"
                />
                <span className="capitalize">{t(`status.${nodeStatus?.status || 'unknown'}`)}</span>
              </div>

              {isPolling && (
                <div
                  className="flex items-center space-x-1 text-sm text-[var(--color-text-secondary)]"
                  aria-label={t('nodes.live')}
                >
                  <span className="relative flex h-3 w-3">
                    <span className="relative inline-flex rounded-full h-3 w-3 bg-blue-500" />
                  </span>
                  <span>{t('nodes.live')}</span>
                </div>
              )}

              <button
                className="ml-2 px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors"
                onClick={() => {
                  // Export PDF functionality would be implemented here
                  // TODO: Implement PDF export
                }}
              >
                {t('nodes.exportPdf')}
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="space-y-6">
          {/* Node Information */}
          <div className="rounded-lg shadow-sm p-6 bg-[var(--color-bg-surface)]">
            <h2 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
              {t('nodes.nodeInfo')}
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div>
                <dt className="text-sm font-medium text-[var(--color-text-secondary)]">
                  {t('nodes.region')}
                </dt>
                <dd className="mt-1 text-lg font-semibold text-[var(--color-text-primary)]">
                  {node.region}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-[var(--color-text-secondary)]">
                  {t('nodes.ipAddress')}
                </dt>
                <dd className="mt-1 text-lg font-semibold text-[var(--color-text-primary)] font-mono">
                  {node.ip}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-[var(--color-text-secondary)]">
                  {t('nodes.lastHeartbeat')}
                </dt>
                <dd className="mt-1 text-lg font-semibold text-[var(--color-text-primary)]">
                  {formatTimestamp(nodeStatus?.last_heartbeat)}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-[var(--color-text-secondary)]">
                  {t('nodes.tags')}
                </dt>
                <dd className="mt-1 flex flex-wrap gap-2">
                  {node.tags && node.tags.length > 0 ? (
                    node.tags.map((tag, index) => (
                      <span
                        key={index}
                        className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
                      >
                        {tag}
                      </span>
                    ))
                  ) : (
                    <span className="text-[var(--color-text-muted)]">
                      {t('nodes.noTags')}
                    </span>
                  )}
                </dd>
              </div>
            </div>
          </div>

          {/* Metrics Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <MetricCard
              title={t('metrics.latency')}
              value={metrics?.latency_ms ?? 'N/A'}
              unit="ms"
              status={metrics && metrics.latency_ms < 100 ? 'good' : metrics && metrics.latency_ms < 200 ? 'warning' : 'critical'}
              icon={
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              }
            />

            <MetricCard
              title={t('metrics.packetLoss')}
              value={metrics?.packet_loss_rate ?? 'N/A'}
              unit="%"
              status={metrics && metrics.packet_loss_rate === 0 ? 'good' : metrics && metrics.packet_loss_rate < 2 ? 'warning' : 'critical'}
              icon={
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
              }
            />

            <MetricCard
              title={t('metrics.jitter')}
              value={metrics?.jitter_ms ?? 'N/A'}
              unit="ms"
              status={metrics && metrics.jitter_ms < 20 ? 'good' : metrics && metrics.jitter_ms < 50 ? 'warning' : 'critical'}
              icon={
                <svg
                  className="w-5 h-5"
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
              }
            />
          </div>

          {/* Historical Trend Charts */}
          <div className="space-y-6">
            {/* Error Message */}
            {historyError && (
              <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
                <div className="flex items-center">
                  <svg
                    className="w-5 h-5 text-red-600 dark:text-red-400 mr-2"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    aria-hidden="true"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  <p className="text-red-800 dark:text-red-300">{historyError}</p>
                  <button
                    onClick={() => {
                      setHistoryError(null)
                      window.location.reload()
                    }}
                    className="ml-auto px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700 transition-colors text-sm"
                  >
                    {t('common.retry')}
                  </button>
                </div>
              </div>
            )}

            {/* Latency Trend Chart */}
            <TrendChart
              data={historyData.latency_ms}
              metric="latency_ms"
              timeRange={timeRange}
              showBaseline={timeRange === '7d' || timeRange === '30d'}
              baselineValue={baselines.latency_ms}
              height="350px"
              onTimeRangeChange={handleTimeRangeChange}
              isLoading={isLoadingHistory}
            />

            {/* Packet Loss Rate Trend Chart */}
            <TrendChart
              data={historyData.packet_loss_rate}
              metric="packet_loss_rate"
              timeRange={timeRange}
              showBaseline={timeRange === '7d' || timeRange === '30d'}
              baselineValue={baselines.packet_loss_rate}
              height="350px"
              onTimeRangeChange={handleTimeRangeChange}
              isLoading={isLoadingHistory}
            />

            {/* Jitter Trend Chart */}
            <TrendChart
              data={historyData.jitter_ms}
              metric="jitter_ms"
              timeRange={timeRange}
              showBaseline={timeRange === '7d' || timeRange === '30d'}
              baselineValue={baselines.jitter_ms}
              height="350px"
              onTimeRangeChange={handleTimeRangeChange}
              isLoading={isLoadingHistory}
            />
          </div>

          {/* Problem Diagnosis */}
          <div className="rounded-lg shadow-sm p-6 bg-[var(--color-bg-surface)]">
            <h2 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
              {t('nodes.problemDiagnosis')}
            </h2>
            <ProblemDiagnosis
              problemType={getProblemType()}
              confidence={getConfidence()}
              details={
                metrics
                  ? `${t('metrics.latency')}: ${metrics.latency_ms}ms, ${t('metrics.packetLoss')}: ${metrics.packet_loss_rate}%, ${t('metrics.jitter')}: ${metrics.jitter_ms}ms`
                  : t('errors.notFound')
              }
              isExpanded={false}
            />
            <p className="mt-4 text-sm text-[var(--color-text-secondary)] italic">
              {t('nodes.diagnosisNote')}
            </p>
          </div>
        </div>
      </main>
    </div>
  )
}
