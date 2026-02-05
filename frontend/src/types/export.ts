/**
 * Export API types
 *
 * TypeScript interfaces for export task management
 */

export type ExportStatus = 'pending' | 'processing' | 'completed' | 'failed'

export type ExportMetric = 'latency' | 'packet_loss_rate' | 'jitter'

export type ExportFormat = 'csv' | 'excel' // UI supports both, backend only CSV in MVP

/**
 * Export task representation from backend
 */
export interface ExportTask {
  id: string
  user_id: string
  node_ids: string[]
  start_time: string // ISO 8601
  end_time: string // ISO 8601
  metrics: ExportMetric[]
  format: ExportFormat
  status: ExportStatus
  file_path?: string
  file_size?: number
  record_count?: number
  error?: string
  created_at: string
  completed_at?: string
}

/**
 * Request payload for creating export
 */
export interface CreateExportRequest {
  node_ids: string[]
  start_time: string
  end_time: string
  metrics: ExportMetric[]
  format?: ExportFormat // defaults to 'csv'
}

/**
 * Response from create export API
 */
export interface CreateExportResponse {
  data: ExportTask
  message: string
  timestamp: string
}

/**
 * Response from get export status API
 */
export interface GetExportStatusResponse {
  data: ExportTask
  message: string
  timestamp: string
}

/**
 * Response from download export API
 * Note: The actual response is a binary file blob, not JSON
 */
