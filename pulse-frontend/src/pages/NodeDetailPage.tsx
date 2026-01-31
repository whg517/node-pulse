import { useParams, Link } from 'react-router-dom'
import { useNodeDetail } from '../hooks/useNodeDetail'
import MetricCard from '../components/dashboard/MetricCard'
import ProblemDiagnosis, {
  type ProblemType,
  type ConfidenceLevel,
} from '../components/dashboard/ProblemDiagnosis'

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
  const { id } = useParams<{ id: string }>()
  const { node, nodeStatus, metrics, isLoading, error, isPolling } = useNodeDetail(
    id || '',
    5000 // 5-second polling
  )

  // Determine problem type based on metrics (placeholder until Story 7.4)
  const getProblemType = (): ProblemType => {
    if (!metrics || !nodeStatus) return 'none'

    const { packet_loss_rate, latency_ms } = metrics

    if (packet_loss_rate > 5 || latency_ms > 500) {
      // This is a placeholder - real diagnosis will be implemented in Story 7.4
      return 'node_local'
    }

    return 'none'
  }

  const getConfidence = (): ConfidenceLevel => {
    if (!metrics || !nodeStatus) return 'low'
    return 'medium' // Placeholder - will be calculated in Story 7.4
  }

  // Format timestamp
  const formatTimestamp = (timestamp: string | undefined): string => {
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

      return date.toLocaleString()
    } catch {
      return 'Invalid timestamp'
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div
            className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
            role="status"
            aria-label="Loading"
          />
          <p className="mt-4 text-gray-600">Loading node details...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white rounded-lg shadow-md p-6 max-w-md">
          <div className="text-red-600 text-5xl mb-4" aria-hidden="true">
            ⚠️
          </div>
          <h2 className="text-xl font-semibold text-gray-900 mb-2">Error Loading Node</h2>
          <p className="text-gray-600 mb-4">{error.message}</p>
          <Link
            to="/dashboard"
            className="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Back to Dashboard
          </Link>
        </div>
      </div>
    )
  }

  if (!node) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white rounded-lg shadow-md p-6 max-w-md">
          <h2 className="text-xl font-semibold text-gray-900 mb-2">Node Not Found</h2>
          <p className="text-gray-600 mb-4">The requested node does not exist.</p>
          <Link
            to="/dashboard"
            className="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Back to Dashboard
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Link
                to="/dashboard"
                className="text-gray-600 hover:text-gray-900 transition-colors"
                aria-label="Back to dashboard"
              >
                <svg
                  className="w-6 h-6"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M10 19l-7-7m0 0l7-7m-7 7h18"
                  />
                </svg>
              </Link>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">{node.name}</h1>
                <p className="text-sm text-gray-600">{node.ip}</p>
              </div>
            </div>

            <div className="flex items-center space-x-2">
              <div
                className={`flex items-center space-x-2 px-3 py-1 rounded-full text-sm font-medium ${
                  nodeStatus?.status === 'online'
                    ? 'bg-green-100 text-green-800'
                    : nodeStatus?.status === 'connecting'
                    ? 'bg-yellow-100 text-yellow-800'
                    : 'bg-red-100 text-red-800'
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
                <span className="capitalize">{nodeStatus?.status || 'Unknown'}</span>
              </div>

              {isPolling && (
                <div
                  className="flex items-center space-x-1 text-sm text-gray-600"
                  aria-label="Data is polling in real-time"
                >
                  <span className="relative flex h-3 w-3">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75" />
                    <span className="relative inline-flex rounded-full h-3 w-3 bg-blue-500" />
                  </span>
                  <span>Live</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="space-y-6">
          {/* Node Information */}
          <div className="bg-white rounded-lg shadow-sm p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Node Information</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div>
                <dt className="text-sm font-medium text-gray-600">Region</dt>
                <dd className="mt-1 text-lg font-semibold text-gray-900">{node.region}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-600">IP Address</dt>
                <dd className="mt-1 text-lg font-semibold text-gray-900 font-mono">
                  {node.ip}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-600">Last Heartbeat</dt>
                <dd className="mt-1 text-lg font-semibold text-gray-900">
                  {formatTimestamp(nodeStatus?.last_heartbeat)}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-600">Tags</dt>
                <dd className="mt-1 flex flex-wrap gap-2">
                  {node.tags && node.tags.length > 0 ? (
                    node.tags.map((tag, index) => (
                      <span
                        key={index}
                        className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                      >
                        {tag}
                      </span>
                    ))
                  ) : (
                    <span className="text-gray-400">No tags</span>
                  )}
                </dd>
              </div>
            </div>
          </div>

          {/* Metrics Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <MetricCard
              title="Latency"
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
              title="Packet Loss Rate"
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
              title="Jitter"
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

          {/* Problem Diagnosis */}
          <div className="bg-white rounded-lg shadow-sm p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Problem Diagnosis</h2>
            <ProblemDiagnosis
              problemType={getProblemType()}
              confidence={getConfidence()}
              details={
                metrics
                  ? `Current metrics: Latency ${metrics.latency_ms}ms, Packet Loss ${metrics.packet_loss_rate}%, Jitter ${metrics.jitter_ms}ms`
                  : 'No metrics available'
              }
              isExpanded={false}
            />
            <p className="mt-4 text-sm text-gray-600 italic">
              Note: Automated problem diagnosis will be available in Story 7.4 (Problem Diagnosis Engine).
              Current assessment is based on simple threshold rules.
            </p>
          </div>
        </div>
      </main>
    </div>
  )
}
