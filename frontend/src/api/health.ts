import { apiClient } from './client'

export interface HealthCheckResponse {
  status: 'healthy' | 'degraded' | 'unhealthy'
  checks: {
    database: string
    alert_engine: string
    webhook_delivery: string
    alert_suppression: string
  }
  scheduler?: {
    running: boolean
    tasks: Record<
      string,
      {
        is_running: boolean
        last_run: string
        run_count: number
        last_error: string
      }
    >
  }
  alert_system?: {
    alert_engine: {
      status: string
      cached_rules: number
      rule_cache_last_refresh: string
      metric_channel_depth: number
      metric_channel_capacity: number
    }
    webhook_delivery: {
      status: string
      success_rate: number
      total_count: number
      success_count: number
    }
    alert_suppression: {
      status: string
      active_suppression_count: number
    }
  }
  timestamp: string
}

export async function fetchSystemHealth(): Promise<HealthCheckResponse> {
  return apiClient<HealthCheckResponse>('/api/v1/health')
}
