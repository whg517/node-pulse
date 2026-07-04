/**
 * Admin Audit Log endpoints
 *
 * Backend route: GET /api/v1/admin/audit/logs (admin only). Security event
 * audit trail (login/logout/refresh/password changes/admin actions).
 */

import { apiClient } from './client'

/** A single audit log entry. Mirrors internal/auth AuditLog model. */
export interface AuditLogDTO {
  id: number
  event_type: string
  user_id?: string | null
  service_account_id?: string | null
  session_id?: string | null
  ip_address?: string | null
  user_agent?: string | null
  details?: Record<string, unknown> | null
  created_at: string
}

export interface AuditLogQuery {
  event_type?: string
  user_id?: string
  /** RFC3339 */
  start_time?: string
  /** RFC3339 */
  end_time?: string
  limit?: number
  offset?: number
}

export interface AuditLogsResponse {
  logs: AuditLogDTO[]
  total_count: number
  limit: number
  offset: number
}

/** Query security audit logs with optional filters. */
export async function getAuditLogs(query: AuditLogQuery = {}): Promise<AuditLogsResponse> {
  const search = new URLSearchParams()
  if (query.event_type) search.set('event_type', query.event_type)
  if (query.user_id) search.set('user_id', query.user_id)
  if (query.start_time) search.set('start_time', query.start_time)
  if (query.end_time) search.set('end_time', query.end_time)
  if (query.limit !== undefined) search.set('limit', String(query.limit))
  if (query.offset !== undefined) search.set('offset', String(query.offset))
  const qs = search.toString() ? `?${search.toString()}` : ''
  return apiClient<AuditLogsResponse>(`/api/v1/admin/audit/logs${qs}`)
}

/** Get a single audit log entry by ID. */
export async function getAuditLog(id: number): Promise<AuditLogDTO> {
  return apiClient<AuditLogDTO>(`/api/v1/admin/audit/logs/${id}`)
}
