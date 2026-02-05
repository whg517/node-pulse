import { apiClient } from './client'

export interface PerformanceMetric {
  metric_name: string
  display_name: string
  current_p99: number
  current_p95: number
  target_p99: number
  target_p95: number
  unit: string
  status: 'healthy' | 'unhealthy'
  anomaly?: string
}

export interface TrendDataPoint {
  timestamp: string
  p99: number
  p95: number
}

export interface MetricTrendData {
  metric_name: string
  data_points: TrendDataPoint[]
}

export interface Anomaly {
  metric_name: string
  severity: 'P0' | 'P1'
  message: string
}

export interface PerformanceSummary {
  total_requests: number
  avg_response_time: number
  max_response_time: number
}

export interface PerformanceDataResponse {
  metrics: PerformanceMetric[]
  trend_data: MetricTrendData[]
  system_health: 'healthy' | 'unhealthy'
  anomalies: Anomaly[]
  summary: PerformanceSummary
}

export interface PerformanceAPIResponse {
  data: PerformanceDataResponse
  message: string
  timestamp: string
}

/**
 * Fetches performance metrics data from the backend
 * @param timeRange - Time range for data (e.g., "24h", "7d"). Defaults to "24h"
 * @returns Performance data with metrics, trends, and health status
 */
export async function fetchPerformanceData(
  timeRange: string = '24h'
): Promise<PerformanceAPIResponse> {
  const response = await apiClient<PerformanceAPIResponse>(
    `/api/v1/data/performance?time_range=${timeRange}`
  )
  return response
}
