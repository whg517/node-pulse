/**
 * Shared API Type Definitions
 *
 * Common Data Transfer Objects (DTOs) and request/response types
 * used across the API layer. Exported for use in stores and components.
 */

// ============================================================================
// Node DTOs
// ============================================================================

/**
 * Node data transfer object
 * Represents a monitoring node in the system
 */
export interface NodeDTO {
  id: string
  name: string
  ip: string
  region: string
  tags: string[]
  status: 'online' | 'offline' | 'connecting'
  created_at: string
  updated_at: string
}

/**
 * Request to create a new node
 */
export interface CreateNodeRequest {
  name: string
  ip: string
  region: string
  tags: string[]
}

/**
 * Request to update an existing node
 * All fields are optional
 */
export interface UpdateNodeRequest {
  name?: string
  ip?: string
  region?: string
  tags?: string[]
}

// ============================================================================
// Data Query DTOs
// ============================================================================

/**
 * Real-time metrics data point
 */
export interface MetricsDTO {
  node_id: string
  latency_ms: number
  packet_loss_rate: number
  jitter_ms: number
  timestamp: string
}

/**
 * Query parameters for historical data request
 */
export interface HistoryQueryDTO {
  node_ids: string[]
  start_time: string
  end_time: string
  metrics: ('latency' | 'packet_loss_rate' | 'jitter')[]
  aggregation?: '1m' | '5m'
}

/**
 * Historical data response
 */
export interface HistoryDataDTO {
  node_id: string
  metric: string
  data_points: Array<{
    timestamp: string
    value: number
  }>
}

/**
 * Query parameters for data export request
 */
export interface ExportQueryDTO {
  node_ids: string[]
  start_time: string
  end_time: string
  format: 'csv' | 'excel'
}

// ============================================================================
// Alert DTOs
// ============================================================================

/**
 * Alert rule data transfer object
 */
export interface AlertRuleDTO {
  id: string
  metric: 'latency' | 'packet_loss_rate' | 'jitter'
  threshold: number
  level: 'P0' | 'P1' | 'P2'
  node_id: string | null
  enabled: boolean
  created_at: string
}

/**
 * Request to create a new alert rule
 */
export interface CreateAlertRuleRequest {
  metric: 'latency' | 'packet_loss_rate' | 'jitter'
  threshold: number
  level: 'P0' | 'P1' | 'P2'
  node_id: string | null
  enabled?: boolean
}

/**
 * Request to update an existing alert rule
 * All fields are optional
 */
export interface UpdateAlertRuleRequest {
  metric?: 'latency' | 'packet_loss_rate' | 'jitter'
  threshold?: number
  level?: 'P0' | 'P1' | 'P2'
  node_id?: string | null
  enabled?: boolean
}

/**
 * Alert record data transfer object
 */
export interface AlertRecordDTO {
  id: string
  node_id: string
  metric: string
  level: string
  status: 'pending' | 'in_progress' | 'resolved'
  created_at: string
  updated_at: string
}

/**
 * Filter parameters for alert records query
 * All fields are optional
 */
export interface AlertRecordFilters {
  node_id?: string
  start_time?: string
  end_time?: string
  level?: string
  status?: string
}

// ============================================================================
// Response Wrapper Types
// ============================================================================

/**
 * Standard API response wrapper for successful responses
 */
export interface ApiResponse<T> {
  data: T
  message?: string
  timestamp?: string
}

/**
 * Standard API error response wrapper
 */
export interface ApiErrorResponse {
  code: string
  message: string
  details?: unknown
}
