import { useEffect, useState } from 'react'
import { fetchNodes } from '../api/nodes'
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
  // Store state and actions
  const {
    comparison: storeComparison,
    setComparisonNodeIds,
    setComparisonMetrics,
    setComparisonTimeRange,
    setComparisonCustomTimeRange,
    setComparisonGroupBy,
    resetComparison,
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
        setAvailableNodes(data)
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Failed to load nodes'
        setError(message)
        console.error('Error loading nodes:', err)
      } finally {
        setIsLoadingNodes(false)
      }
    }

    loadNodes()
  }, [])

  // Transform available nodes to selector format
  const nodeOptions = availableNodes.map((node) => ({
    node_id: node.id,
    name: node.name,
    region: node.region,
    isp: node.tags.find((tag) => ISP_TAGS.includes(tag as any)) || undefined,
    status: node.status as 'online' | 'offline' | 'connecting',
  }))

  // Handle node selection change
  const handleNodeSelectionChange = (nodeIds: string[]) => {
    // Validate min/max selection
    if (nodeIds.length > 5) {
      setError('Maximum 5 nodes can be selected')
      return
    }
    if (nodeIds.length < 2 && nodeIds.length > 0) {
      setError('At least 2 nodes must be selected for comparison')
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
      setError('Please select at least 2 nodes for comparison')
      return
    }
    if (selectedNodeIds.length > 5) {
      setError('Maximum 5 nodes can be compared')
      return
    }
    if (selectedMetrics.length === 0) {
      setError('Please select at least one metric')
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
      const message = err instanceof Error ? err.message : 'Failed to fetch comparison data'
      setError(message)
      console.error('Error fetching comparison data:', err)
    } finally {
      setIsLoadingComparison(false)
    }
  }

  // Metric options
  const metricOptions = [
    { key: 'latency_ms' as MetricType, label: 'Latency', unit: 'ms' },
    { key: 'packet_loss_rate' as MetricType, label: 'Packet Loss Rate', unit: '%' },
    { key: 'jitter_ms' as MetricType, label: 'Jitter', unit: 'ms' },
  ]

  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Node Comparison</h1>
          <p className="mt-2 text-gray-600">
            Compare network metrics across multiple nodes to identify performance differences and
            anomalies.
          </p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex">
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
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">Error</h3>
                <p className="mt-1 text-sm text-red-700">{error}</p>
              </div>
            </div>
          </div>
        )}

        {/* Node Selector */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Select Nodes (2-5)</h2>

          {isLoadingNodes ? (
            <div className="flex items-center justify-center py-8">
              <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
              <span className="ml-3 text-gray-600">Loading nodes...</span>
            </div>
          ) : (
            <>
              {/* Group By Selector */}
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Group By
                </label>
                <div className="flex space-x-2">
                  {(['none', 'region', 'isp'] as GroupByType[]).map((option) => (
                    <button
                      key={option}
                      onClick={() => handleGroupByChange(option)}
                      className={`px-4 py-2 rounded font-medium transition-colors ${
                        groupBy === option
                          ? 'bg-blue-600 text-white'
                          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                      }`}
                    >
                      {option === 'none' ? 'None' : option.charAt(0).toUpperCase() + option.slice(1)}
                    </button>
                  ))}
                </div>
              </div>

              {/* Node List */}
              <div className="space-y-2 max-h-64 overflow-y-auto border border-gray-200 rounded-lg p-4">
                {nodeOptions.map((node) => (
                  <label
                    key={node.node_id}
                    className="flex items-center p-3 hover:bg-gray-50 rounded cursor-pointer"
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
                        <span className="text-sm font-medium text-gray-900">{node.name}</span>
                        <div className="flex items-center space-x-2">
                          {node.region && (
                            <span className="text-xs text-gray-500">Region: {node.region}</span>
                          )}
                          {node.isp && <span className="text-xs text-gray-500">ISP: {node.isp}</span>}
                          <span
                            className={`text-xs font-medium ${
                              node.status === 'online'
                                ? 'text-green-600'
                                : node.status === 'offline'
                                ? 'text-red-600'
                                : 'text-yellow-600'
                            }`}
                          >
                            {node.status}
                          </span>
                        </div>
                      </div>
                    </div>
                  </label>
                ))}
              </div>

              {/* Selection Summary */}
              <div className="mt-4 text-sm text-gray-600">
                Selected: {selectedNodeIds.length} / 5 nodes
                {selectedNodeIds.length > 0 && selectedNodeIds.length < 2 && (
                  <span className="text-red-600 ml-2">(Select at least 2 nodes)</span>
                )}
              </div>
            </>
          )}
        </div>

        {/* Time Range Selector */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Time Range</h2>
          <div className="flex space-x-2">
            {(['24h', '7d', '30d'] as const).map((range) => (
              <button
                key={range}
                onClick={() => handleTimeRangeChange(range)}
                className={`px-4 py-2 rounded font-medium transition-colors ${
                  timeRange === range
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                {range === '24h' ? '24 Hours' : range === '7d' ? '7 Days' : '30 Days'}
              </button>
            ))}
            <button
              onClick={() => handleTimeRangeChange('custom')}
              className={`px-4 py-2 rounded font-medium transition-colors ${
                timeRange === 'custom'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Custom
            </button>
          </div>

          {/* Custom Time Range Inputs */}
          {timeRange === 'custom' && (
            <div className="mt-4 grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Start Time</label>
                <input
                  type="datetime-local"
                  value={customTimeRange?.start || ''}
                  onChange={(e) =>
                    handleCustomTimeRangeChange(e.target.value, customTimeRange?.end || '')
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">End Time</label>
                <input
                  type="datetime-local"
                  value={customTimeRange?.end || ''}
                  onChange={(e) =>
                    handleCustomTimeRangeChange(customTimeRange?.start || '', e.target.value)
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>
          )}
        </div>

        {/* Metric Selector */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Metrics</h2>
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
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
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
                : 'bg-gray-300 cursor-not-allowed'
            }`}
          >
            {isLoadingComparison ? 'Loading...' : 'Compare Nodes'}
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
          <div className="bg-white rounded-lg shadow-sm p-12">
            <div className="text-center">
              <svg
                className="mx-auto h-12 w-12 text-gray-400"
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
              <h3 className="mt-2 text-sm font-medium text-gray-900">No Comparison Data</h3>
              <p className="mt-1 text-sm text-gray-500">
                Select nodes, time range, and metrics, then click "Compare Nodes" to view the
                comparison chart.
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
