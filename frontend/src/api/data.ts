/**
 * Data Query API endpoints
 *
 * Provides typed functions for querying real-time metrics,
 * historical data, and exporting data from the Pulse backend.
 */

import { apiClient, NotFoundError } from './client'
import type {
  MetricsDTO,
  HistoryQueryDTO,
  HistoryDataDTO,
  ExportQueryDTO
} from './types'

// Export types for use in components and stores
export type {
  MetricsDTO,
  HistoryQueryDTO,
  HistoryDataDTO,
  ExportQueryDTO
}

/**
 * Fetch real-time metrics for nodes
 *
 * Retrieves current metrics from the in-memory cache (< 1 hour old).
 * This is the fastest query method and should be used for dashboard
 * real-time updates.
 *
 * @param nodeIds - Array of node IDs to fetch metrics for
 * @returns Real-time metrics data for specified nodes
 * @throws ValidationError if node IDs are invalid
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const { data } = await fetchMetrics(['node-1', 'node-2'])
 * data.forEach(metric => {
 *   console.log(`Node ${metric.node_id}:`, metric.latency_ms, 'ms')
 * })
 */
export async function fetchMetrics(
  nodeIds: string[]
): Promise<{ data: MetricsDTO[] }> {
  const params = new URLSearchParams()
  nodeIds.forEach(id => params.append('node_id', id))

  const queryString = params.toString()
  const endpoint = queryString
    ? `/api/v1/data/metrics?${queryString}`
    : '/api/v1/data/metrics'

  return apiClient<{ data: MetricsDTO[] }>(endpoint)
}

/**
 * Fetch historical data
 *
 * Retrieves historical metrics from the PostgreSQL metrics table.
 * Supports time range filtering and data aggregation.
 * Data is available for the past 7 days.
 *
 * @param query - Historical data query parameters
 * @returns Historical data points grouped by node and metric
 * @throws ValidationError if query parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const { data } = await fetchHistory({
 *   node_ids: ['node-1'],
 *   start_time: '2024-01-01T00:00:00Z',
 *   end_time: '2024-01-02T00:00:00Z',
 *   metrics: ['latency', 'packet_loss_rate'],
 *   aggregation: '5m'
 * })
 *
 * data.forEach(series => {
 *   console.log(`Node ${series.node_id} - ${series.metric}`)
 *   series.data_points.forEach(point => {
 *     console.log(`  ${point.timestamp}: ${point.value}`)
 *   })
 * })
 */
export async function fetchHistory(
  query: HistoryQueryDTO
): Promise<{ data: HistoryDataDTO[] }> {
  const params = new URLSearchParams()

  // Add node IDs
  query.node_ids.forEach(id => params.append('node_id', id))

  // Add time range
  params.append('start_time', query.start_time)
  params.append('end_time', query.end_time)

  // Add metrics
  query.metrics.forEach(m => params.append('metric', m))

  // Add optional aggregation
  if (query.aggregation) {
    params.append('aggregation', query.aggregation)
  }

  return apiClient<{ data: HistoryDataDTO[] }>(
    `/api/v1/data/history?${params}`
  )
}

/**
 * Export data as CSV or Excel
 *
 * Initiates an async export job for the specified data range.
 * Returns a download URL that can be used to retrieve the exported file.
 *
 * @param query - Export query parameters
 * @returns Download URL for the exported file
 * @throws ValidationError if export parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const { data } = await exportData({
 *   node_ids: ['node-1', 'node-2'],
 *   start_time: '2024-01-01T00:00:00Z',
 *   end_time: '2024-01-07T23:59:59Z',
 *   format: 'csv'
 * })
 *
 * console.log('Download URL:', data.download_url)
 * // Use this URL to download the file
 * window.open(data.download_url, '_blank')
 */
export async function exportData(
  query: ExportQueryDTO
): Promise<{ data: { download_url: string } }> {
  const params = new URLSearchParams()

  // Add node IDs
  query.node_ids.forEach(id => params.append('node_id', id))

  // Add time range
  params.append('start_time', query.start_time)
  params.append('end_time', query.end_time)

  // Add format
  params.append('format', query.format)

  return apiClient<{ data: { download_url: string } }>(
    `/api/v1/data/export?${params}`
  )
}

/**
 * Comparison Query Types
 */

export interface ComparisonQueryDTO {
  node_ids: string[]
  start_time: string
  end_time: string
  metrics: string[]
}

export interface NodeComparisonMetrics {
  [metric: string]: {
    data_points: Array<{ timestamp: string; value: number }>
    avg: number
    max: number
    min: number
  }
}

export interface ComparisonNodeData {
  node_id: string
  name: string
  region?: string
  isp?: string
  metrics: NodeComparisonMetrics
}

export interface ComparisonStatistics {
  [metric: string]: {
    overall_avg: number
    overall_max: number
    overall_min: number
    differences: Array<{
      node_id: string
      diff_from_avg: number
    }>
  }
}

export interface ComparisonResponseDTO {
  data: {
    time_range: {
      start: string
      end: string
    }
    nodes: ComparisonNodeData[]
    statistics: ComparisonStatistics
  }
  message: string
  timestamp: string
}

/**
 * Fetch comparison data for multiple nodes
 *
 * Retrieves aggregated metrics data for comparing performance across multiple nodes.
 * Returns pre-calculated statistics including averages, max/min values, and differences.
 *
 * @param query - Comparison query parameters
 * @returns Comparison data with statistics
 * @throws ValidationError if query parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const { data } = await getComparisonData({
 *   node_ids: ['node-1', 'node-2', 'node-3'],
 *   start_time: '2024-01-01T00:00:00Z',
 *   end_time: '2024-01-02T00:00:00Z',
 *   metrics: ['latency_ms', 'packet_loss_rate']
 * })
 *
 * console.log('Time range:', data.time_range)
 * console.log('Nodes:', data.nodes.length)
 * console.log('Statistics:', data.statistics)
 */
export async function getComparisonData(
  query: ComparisonQueryDTO
): Promise<ComparisonResponseDTO> {
  const params = new URLSearchParams()

  // Add node IDs (comma-separated)
  params.append('node_ids', query.node_ids.join(','))

  // Add time range
  params.append('start_time', query.start_time)
  params.append('end_time', query.end_time)

  // Add metrics (comma-separated)
  params.append('metrics', query.metrics.join(','))

  return apiClient<ComparisonResponseDTO>(
    `/api/v1/data/comparison?${params}`
  )
}

/**
 * Pulse `/data/diagnosis` — multi-node statistical diagnosis (requires ≥3 nodes with cache/DB data).
 */
export interface DiagnosisResultDTO {
  problem_type: string
  confidence: string
  analysis: {
    nodes_analyzed: number
    affected_nodes?: string[]
    regions_analyzed?: string[]
    metrics?: Record<string, unknown>
    regional_comparison?: Record<string, unknown>
  }
  recommendation: string
  timestamp: string
}

export interface DiagnosisApiResponse {
  data: DiagnosisResultDTO
  message: string
  timestamp: string
}

/**
 * Fetch cross-node diagnosis for the given node IDs (minimum 3).
 */
export async function fetchDiagnosis(
  nodeIds: string[]
): Promise<DiagnosisApiResponse> {
  if (nodeIds.length < 3) {
    throw new Error('Diagnosis requires at least 3 node IDs')
  }
  const params = new URLSearchParams()
  nodeIds.forEach((id) => params.append('node_ids', id))
  return apiClient<DiagnosisApiResponse>(
    `/api/v1/data/diagnosis?${params}`
  )
}

export interface MTRHopDTO {
  hop_number: number
  ip: string
  hostname?: string
  as_number?: string
  sent: number
  received: number
  loss_rate: number
  last_rtt_ms: number
  avg_rtt_ms: number
  best_rtt_ms: number
  worst_rtt_ms: number
  std_dev_ms: number
  location?: string
}

export interface MTRResultDTO {
  id: string
  node_id: string
  probe_id?: string
  target: string
  success: boolean
  total_hops: number
  hops: MTRHopDTO[]
  completed_at: string
  error_message?: string
  created_at: string
}

export interface MTRHopData {
  hopNumber: number
  ip: string
  hostname?: string
  asNumber?: string
  sent: number
  received: number
  lossRate: number
  lastRTTMs: number
  avgRTTMs: number
  bestRTTMs: number
  worstRTTMs: number
  stdDevMs: number
  location?: string
}

export interface MTRResultData {
  id: string
  nodeId: string
  probeId?: string
  target: string
  totalHops: number
  hops: MTRHopData[]
  completedAt: string
  success: boolean
  errorMessage?: string
  createdAt: string
}

interface MTRApiResponse {
  data: MTRResultDTO
  message: string
  timestamp: string
}

interface MTRHistoryApiResponse {
  data: MTRResultDTO[]
  message: string
  timestamp: string
}

function mapMTRResult(result: MTRResultDTO): MTRResultData {
  return {
    id: result.id,
    nodeId: result.node_id,
    probeId: result.probe_id,
    target: result.target,
    totalHops: result.total_hops,
    completedAt: result.completed_at,
    success: result.success,
    errorMessage: result.error_message,
    createdAt: result.created_at,
    hops: (result.hops || []).map((hop) => ({
      hopNumber: hop.hop_number,
      ip: hop.ip,
      hostname: hop.hostname,
      asNumber: hop.as_number,
      sent: hop.sent,
      received: hop.received,
      lossRate: hop.loss_rate,
      lastRTTMs: hop.last_rtt_ms,
      avgRTTMs: hop.avg_rtt_ms,
      bestRTTMs: hop.best_rtt_ms,
      worstRTTMs: hop.worst_rtt_ms,
      stdDevMs: hop.std_dev_ms,
      location: hop.location,
    })),
  }
}

export async function fetchLatestMTR(nodeId: string): Promise<MTRResultData | null> {
  const params = new URLSearchParams({ node_id: nodeId })

  try {
    const response = await apiClient<MTRApiResponse>(`/api/v1/data/mtr?${params}`)
    return mapMTRResult(response.data)
  } catch (err) {
    if (err instanceof NotFoundError) {
      return null
    }
    throw err
  }
}

export async function fetchMTRHistory(
  nodeId: string,
  limit = 10
): Promise<MTRResultData[]> {
  const params = new URLSearchParams({
    node_id: nodeId,
    limit: String(limit),
  })
  const response = await apiClient<MTRHistoryApiResponse>(`/api/v1/data/mtr/history?${params}`)
  return response.data.map(mapMTRResult)
}
