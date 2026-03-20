import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { fetchNodes } from '../api/nodes'
import { PageContainer, ErrorBanner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { getComparisonData } from '../api/data'
import type { NodeDTO } from '../api/types'
import ComparisonChart from '../components/dashboard/ComparisonChart'
import type {
  NodeComparisonData,
  MetricType,
} from '../components/dashboard/ComparisonChart'
import { useDashboardStore } from '../stores/dashboardStore'
import type { ExtendedTimeRange, GroupByType } from '../stores/types'

// ISP tags for grouping
const ISP_TAGS = [
  'AWS',
  'GCP',
  'Azure',
  'Alibaba',
  'Tencent',
  'DigitalOcean',
  'Linode',
  'Vultr',
  'OVH',
  'Hetzner',
] as const

/**
 * NodeComparison Page
 *
 * Multi-node comparison page allowing users to:
 * - Select 2-5 nodes for comparison
 * - Group nodes by region or ISP
 * - Select time range (24h/7d/30d/custom)
 * - Select metrics (latency/packet loss/jitter)
 * - View comparison charts with statistics
 * - Highlight differences between nodes
 *
 * @returns NodeComparison page component
 */
export default function NodeComparisonPage() {
  const { t } = useTranslation()

  // Store state and actions
  const {
    comparison: storeComparison,
    setComparisonNodeIds,
    setComparisonMetrics,
    setComparisonTimeRange,
    setComparisonCustomTimeRange,
    setComparisonGroupBy,
  } = useDashboardStore()

  // Local state
  const [availableNodes, setAvailableNodes] = useState<NodeDTO[]>([])
  const [comparisonData, setComparisonData] = useState<NodeComparisonData[] | null>(null)
  const [isLoadingNodes, setIsLoadingNodes] = useState(true)
  const [isLoadingComparison, setIsLoadingComparison] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Destructure store state for cleaner access
  const { selectedNodeIds, selectedMetrics, timeRange, customTimeRange, groupBy } = storeComparison

  // Load available nodes on mount
  useEffect(() => {
    async function loadNodes() {
      try {
        setIsLoadingNodes(true)
        setError(null)
        const { data } = await fetchNodes()
        setAvailableNodes(data.nodes)
      } catch (err) {
        const message = err instanceof Error ? err.message : t('errors.failedToLoad')
        setError(message)
        console.error('Error loading nodes:', err)
      } finally {
        setIsLoadingNodes(false)
      }
    }

    loadNodes()
  }, [t])

  // Transform available nodes to selector format
  const nodeOptions = availableNodes.map((node) => ({
    node_id: node.id,
    name: node.name,
    region: node.region,
    isp: node.tags.find((tag): tag is typeof ISP_TAGS[number] =>
      ISP_TAGS.includes(tag as typeof ISP_TAGS[number])) || undefined,
    status: node.status as 'online' | 'offline' | 'connecting',
  }))

  // Handle node selection change
  const handleNodeSelectionChange = (nodeIds: string[]) => {
    // Validate min/max selection
    if (nodeIds.length > 5) {
      setError(t('nodes.maxNodesSelected', { max: 5 }))
      return
    }
    if (nodeIds.length < 2 && nodeIds.length > 0) {
      setError(t('nodes.selectAtLeast', { count: 2 }))
    }
    setComparisonNodeIds(nodeIds)
    setError(null)
  }

  // Handle time range change
  const handleTimeRangeChange = (range: ExtendedTimeRange) => {
    setComparisonTimeRange(range)
  }

  // Handle custom time range change
  const handleCustomTimeRangeChange = (start: string, end: string) => {
    setComparisonCustomTimeRange({ start, end })
  }

  // Handle metric selection change
  const handleMetricSelectionChange = (metrics: MetricType[]) => {
    setComparisonMetrics(metrics)
  }

  // Handle group by change
  const handleGroupByChange = (newGroupBy: GroupByType) => {
    setComparisonGroupBy(newGroupBy)
  }

  // Calculate time range parameters
  const getTimeRangeParams = (): { start_time: string; end_time: string } => {
    const end = new Date()
    let start: Date

    if (timeRange === 'custom' && customTimeRange) {
      return {
        start_time: customTimeRange.start,
        end_time: customTimeRange.end,
      }
    }

    switch (timeRange) {
      case '24h':
        start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
        break
      case '7d':
        start = new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000)
        break
      case '30d':
        start = new Date(end.getTime() - 30 * 24 * 60 * 60 * 1000)
        break
      default:
        start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
    }

    return {
      start_time: start.toISOString(),
      end_time: end.toISOString(),
    }
  }

  // Handle comparison button click
  const handleCompare = async () => {
    if (selectedNodeIds.length < 2) {
      setError(t('nodes.selectAtLeast', { count: 2 }))
      return
    }
    if (selectedNodeIds.length > 5) {
      setError(t('nodes.maxNodesSelected', { max: 5 }))
      return
    }
    if (selectedMetrics.length === 0) {
      setError(t('reports.selectMetrics'))
      return
    }

    try {
      setIsLoadingComparison(true)
      setError(null)

      const { start_time, end_time } = getTimeRangeParams()

      // Call comparison API using centralized API function
      const apiResponse = await getComparisonData({
        node_ids: selectedNodeIds,
        start_time,
        end_time,
        metrics: selectedMetrics,
      })

      // Transform API response to ComparisonChart format
      const transformedData: NodeComparisonData[] = apiResponse.data.nodes.map(
        (node) => ({
          node_id: node.node_id,
          node_name: node.name,
          region: node.region,
          isp: node.isp,
          data: node.metrics[selectedMetrics[0]]?.data_points || [],
        })
      )

      setComparisonData(transformedData)
    } catch (err) {
      const message = err instanceof Error ? err.message : t('errors.failedToLoad')
      setError(message)
      console.error('Error fetching comparison data:', err)
    } finally {
      setIsLoadingComparison(false)
    }
  }

  // Metric options
  const metricOptions = [
    { key: 'latency_ms' as MetricType, label: t('metrics.latency'), unit: 'ms' },
    { key: 'packet_loss_rate' as MetricType, label: t('metrics.packetLoss'), unit: '%' },
    { key: 'jitter_ms' as MetricType, label: t('metrics.jitter'), unit: 'ms' },
  ]

  return (
    <PageContainer>
      <PageHeader
        title={t('nodes.comparison')}
        subtitle={t('nodes.comparisonDescription')}
      />

        {/* Error Message */}
        {error && (
          <ErrorBanner error={error} className="mb-6" />
        )}

        {/* Node Selector */}
        <div className="rounded-lg shadow-sm p-6 mb-6 bg-[var(--color-bg-surface)]">
          <h2 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
            {t('nodes.selectNodes')} (2-5)
          </h2>

          {isLoadingNodes ? (
            <div className="flex items-center justify-center py-8">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
              <span className="ml-3 text-[var(--color-text-secondary)]">
                {t('common.loading')}
              </span>
            </div>
          ) : (
            <>
              {/* Group By Selector */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
                  {t('nodes.groupBy')}
                </label>
                <div className="flex space-x-2">
                  {(['none', 'region', 'isp'] as GroupByType[]).map((option) => (
                    <button
                      key={option}
                      onClick={() => handleGroupByChange(option)}
                      className={`px-4 py-2 rounded font-medium transition-colors ${
                        groupBy === option
                          ? 'bg-blue-600 text-white'
                          : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-subtle)]'
                      }`}
                    >
                      {option === 'none' ? t('nodes.none') : option.charAt(0).toUpperCase() + option.slice(1)}
                    </button>
                  ))}
                </div>
              </div>

              {/* Node List */}
              <div data-testid="node-selector" className="space-y-2 max-h-64 overflow-y-auto border border-[var(--color-border)] rounded-lg p-4">
                {nodeOptions.map((node) => (
                  <label
                    key={node.node_id}
                    className="flex items-center p-3 hover:bg-[var(--color-hover-overlay)] rounded cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={selectedNodeIds.includes(node.node_id)}
                      onChange={(e) => {
                        const newSelected = e.target.checked
                          ? [...selectedNodeIds, node.node_id]
                          : selectedNodeIds.filter((id) => id !== node.node_id)
                        handleNodeSelectionChange(newSelected)
                      }}
                      disabled={
                        !selectedNodeIds.includes(node.node_id) && selectedNodeIds.length >= 5
                      }
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                    <div className="ml-3 flex-1">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium text-[var(--color-text-primary)]">
                          {node.name}
                        </span>
                        <div className="flex items-center space-x-2">
                          {node.region && (
                            <span className="text-xs text-[var(--color-text-muted)]">
                              {t('nodes.region')}: {node.region}
                            </span>
                          )}
                          {node.isp && (
                            <span className="text-xs text-[var(--color-text-muted)]">
                              ISP: {node.isp}
                            </span>
                          )}
                          <span
                            className={`text-xs font-medium ${
                              node.status === 'online'
                                ? 'text-green-600'
                                : node.status === 'offline'
                                ? 'text-red-600'
                                : 'text-yellow-600'
                            }`}
                          >
                            {t(`status.${node.status}`)}
                          </span>
                        </div>
                      </div>
                    </div>
                  </label>
                ))}
              </div>

              {/* Selection Summary */}
              <div className="mt-4 text-sm text-[var(--color-text-secondary)]">
                {t('nodes.selectedCount', { count: selectedNodeIds.length, max: 5 })}
                {selectedNodeIds.length > 0 && selectedNodeIds.length < 2 && (
                  <span className="text-red-600 ml-2">({t('nodes.selectAtLeast', { count: 2 })})</span>
                )}
              </div>
            </>
          )}
        </div>

        {/* Time Range Selector */}
        <div className="rounded-lg shadow-sm p-6 mb-6 bg-[var(--color-bg-surface)]">
          <h2 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
            {t('nodes.timeRange')}
          </h2>
          <div className="flex space-x-2">
            {(['24h', '7d', '30d'] as const).map((range) => (
              <button
                key={range}
                onClick={() => handleTimeRangeChange(range)}
                className={`px-4 py-2 rounded font-medium transition-colors ${
                  timeRange === range
                    ? 'bg-blue-600 text-white'
                    : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-subtle)]'
                }`}
              >
                {range === '24h' ? t('nodes.hours24') : range === '7d' ? t('nodes.days7') : t('nodes.days30')}
              </button>
            ))}
            <button
              onClick={() => handleTimeRangeChange('custom')}
              className={`px-4 py-2 rounded font-medium transition-colors ${
                timeRange === 'custom'
                  ? 'bg-blue-600 text-white'
                  : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-subtle)]'
              }`}
            >
              {t('nodes.custom')}
            </button>
          </div>

          {/* Custom Time Range Inputs */}
          {timeRange === 'custom' && (
            <div className="mt-4 grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                  {t('nodes.startTime')}
                </label>
                <input
                  type="datetime-local"
                  value={customTimeRange?.start || ''}
                  onChange={(e) =>
                    handleCustomTimeRangeChange(e.target.value, customTimeRange?.end || '')
                  }
                  className="w-full px-3 py-2 border border-[var(--color-input-border)] bg-[var(--color-input-bg)] text-[var(--color-text-primary)] rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                  {t('nodes.endTime')}
                </label>
                <input
                  type="datetime-local"
                  value={customTimeRange?.end || ''}
                  onChange={(e) =>
                    handleCustomTimeRangeChange(customTimeRange?.start || '', e.target.value)
                  }
                  className="w-full px-3 py-2 border border-[var(--color-input-border)] bg-[var(--color-input-bg)] text-[var(--color-text-primary)] rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>
          )}
        </div>

        {/* Metric Selector */}
        <div className="rounded-lg shadow-sm p-6 mb-6 bg-[var(--color-bg-surface)]">
          <h2 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
            {t('nodes.metricsSelector')}
          </h2>
          <div className="flex space-x-2">
            {metricOptions.map((metric) => (
              <button
                key={metric.key}
                onClick={() => {
                  const newMetrics = selectedMetrics.includes(metric.key)
                    ? selectedMetrics.filter((m) => m !== metric.key)
                    : [...selectedMetrics, metric.key]
                  handleMetricSelectionChange(newMetrics)
                }}
                className={`px-4 py-2 rounded font-medium transition-colors ${
                  selectedMetrics.includes(metric.key)
                    ? 'bg-blue-600 text-white'
                    : 'bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-subtle)]'
                }`}
              >
                {metric.label} ({metric.unit})
              </button>
            ))}
          </div>
        </div>

        {/* Compare Button */}
        <div className="mb-6">
          <button
            data-testid="compare-button"
            onClick={handleCompare}
            disabled={
              selectedNodeIds.length < 2 ||
              selectedNodeIds.length > 5 ||
              selectedMetrics.length === 0 ||
              isLoadingComparison
            }
            className={`w-full py-3 px-4 rounded-lg font-semibold text-white transition-colors ${
              selectedNodeIds.length >= 2 &&
              selectedNodeIds.length <= 5 &&
              selectedMetrics.length > 0 &&
              !isLoadingComparison
                ? 'bg-blue-600 hover:bg-blue-700'
                : 'bg-gray-300 dark:bg-gray-700 cursor-not-allowed'
            }`}
          >
            {isLoadingComparison ? t('common.loading') : t('nodes.compareNodes')}
          </button>
        </div>

        {/* Comparison Chart */}
        {comparisonData && comparisonData.length > 0 && (
          <ComparisonChart
            nodes={comparisonData}
            metric={selectedMetrics[0]}
            timeRange={timeRange === 'custom' ? '24h' : timeRange}
            showStatistics={true}
            highlightDifferences={true}
            groupBy={groupBy}
            isLoading={isLoadingComparison}
          />
        )}

        {/* Empty State */}
        {!comparisonData && !isLoadingComparison && (
          <div className="rounded-lg shadow-sm p-12 bg-[var(--color-bg-surface)]">
            <div className="text-center">
              <svg
                className="mx-auto h-12 w-12 text-[var(--color-text-muted)]"
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
              <h3 className="mt-2 text-sm font-medium text-[var(--color-text-primary)]">
                {t('nodes.noComparisonData')}
              </h3>
              <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
                {t('nodes.noComparisonDataDescription')}
              </p>
            </div>
          </div>
        )}
    </PageContainer>
  )
}
