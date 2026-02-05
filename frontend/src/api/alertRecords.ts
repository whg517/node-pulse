/**
 * Alert Records API endpoints
 *
 * Provides typed functions for alert record queries and
 * status update operations.
 */

import { apiClient } from './client'

// ============================================================================
// Types
// ============================================================================

/**
 * Alert record status values
 */
export type AlertRecordStatus = 'pending' | 'in_progress' | 'resolved'

/**
 * Alert level values
 */
export type AlertLevel = 'P0' | 'P1' | 'P2'

/**
 * Alert record data transfer object
 */
export interface AlertRecordDTO {
  id: string
  alert_event_id: string
  node_id: string
  metric: 'latency' | 'packet_loss_rate' | 'jitter'
  level: AlertLevel
  status: AlertRecordStatus
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
  level?: AlertLevel
  status?: AlertRecordStatus
  limit?: number
  offset?: number
}

/**
 * Request to update alert record status
 */
export interface UpdateAlertRecordStatusRequest {
  status: AlertRecordStatus
}

// ============================================================================
// Validation Functions
// ============================================================================

/**
 * Validates if a status transition is allowed
 *
 * Status flow: pending → in_progress → resolved
 * Allowed transitions:
 * - pending → in_progress
 * - pending → resolved
 * - in_progress → resolved
 *
 * @param currentStatus - Current alert record status
 * @param newStatus - Desired new status
 * @returns true if transition is valid, false otherwise
 */
export function isValidStatusTransition(
  currentStatus: AlertRecordStatus,
  newStatus: AlertRecordStatus
): boolean {
  // No change needed
  if (currentStatus === newStatus) return true

  // Define allowed transitions
  const transitions: Record<AlertRecordStatus, AlertRecordStatus[]> = {
    pending: ['in_progress', 'resolved'],
    in_progress: ['resolved'],
    resolved: [], // Cannot reopen in MVP
  }

  return transitions[currentStatus]?.includes(newStatus) ?? false
}

// ============================================================================
// API Functions
// ============================================================================

/**
 * Paginated response wrapper for alert records
 */
export interface AlertRecordsPaginatedResponse {
  data: AlertRecordDTO[]
  total?: number
  message?: string
  timestamp?: string
}

/**
 * Fetch alert records with optional filters
 *
 * @param filters - Optional filters for alert records
 * @returns Array of alert records matching filters
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * // Fetch all alert records
 * const { data } = await getAlertRecords()
 *
 * // Fetch with filters
 * const { data: filtered } = await getAlertRecords({
 *   node_id: 'node-id',
 *   level: 'P0',
 *   status: 'pending',
 *   limit: 20,
 *   offset: 0
 * })
 */
export async function getAlertRecords(
  filters?: AlertRecordFilters
): Promise<AlertRecordsPaginatedResponse> {
  if (!filters || Object.keys(filters).length === 0) {
    // No filters, fetch all records
    return apiClient('/api/v1/alerts/records')
  }

  // Build query parameters from filters
  const params = new URLSearchParams()
  if (filters.node_id) params.append('node_id', filters.node_id)
  if (filters.start_time) params.append('start_time', filters.start_time)
  if (filters.end_time) params.append('end_time', filters.end_time)
  if (filters.level) params.append('level', filters.level)
  if (filters.status) params.append('status', filters.status)
  if (filters.limit) params.append('limit', filters.limit.toString())
  if (filters.offset) params.append('offset', filters.offset.toString())

  return apiClient(`/api/v1/alerts/records?${params}`)
}

/**
 * Update alert record status
 *
 * @param id - Alert record ID to update
 * @param status - New status value
 * @returns Updated alert record data
 * @throws ValidationError if status transition is invalid
 * @throws AuthenticationError if user is not authenticated
 * @throws NotFoundError if alert record does not exist
 *
 * @example
 * const { data } = await updateAlertRecordStatus('record-id', 'in_progress')
 */
export async function updateAlertRecordStatus(
  id: string,
  status: AlertRecordStatus
): Promise<{ data: AlertRecordDTO; message: string; timestamp: string }> {
  return apiClient(`/api/v1/alerts/records/${id}/status`, {
    method: 'PUT',
    body: JSON.stringify({ status }),
  })
}
