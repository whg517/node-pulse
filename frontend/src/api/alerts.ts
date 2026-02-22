/**
 * Alert Management API endpoints
 *
 * Provides typed functions for alert rule CRUD operations and
 * alert record queries with filtering support.
 */

import { apiClient } from './client'
import type {
  AlertRuleDTO,
  CreateAlertRuleRequest,
  UpdateAlertRuleRequest,
  AlertRecordDTO,
  AlertRecordFilters
} from './types'

// Export types for use in components and stores
export type {
  AlertRuleDTO,
  CreateAlertRuleRequest,
  UpdateAlertRuleRequest,
  AlertRecordDTO,
  AlertRecordFilters
}

/**
 * Fetch all alert rules from the API
 *
 * @returns Object containing alerts array
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const { data } = await fetchAlertRules()
 * data.alerts.forEach(rule => console.log(rule.metric, rule.threshold))
 */
export async function fetchAlertRules(): Promise<{ data: { alerts: AlertRuleDTO[] } }> {
  return apiClient<{ data: { alerts: AlertRuleDTO[] } }>('/api/v1/alerts/rules')
}

/**
 * Create a new alert rule
 *
 * @param request - Alert rule creation parameters
 * @returns Created alert rule data
 * @throws ValidationError if request parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const { data } = await createAlertRule({
 *   metric: 'latency',
 *   threshold: 100,
 *   level: 'P1',
 *   node_id: null
 * })
 */
export async function createAlertRule(
  request: CreateAlertRuleRequest
): Promise<{ data: AlertRuleDTO }> {
  return apiClient<{ data: AlertRuleDTO }>('/api/v1/alerts/rules', {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

/**
 * Update an existing alert rule
 *
 * @param id - Alert rule ID to update
 * @param request - Alert rule update parameters (all optional)
 * @returns Updated alert rule data
 * @throws ValidationError if request parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 * @throws NotFoundError if alert rule does not exist
 *
 * @example
 * const { data } = await updateAlertRule('rule-id', {
 *   threshold: 150,
 *   enabled: false
 * })
 */
export async function updateAlertRule(
  id: string,
  request: UpdateAlertRuleRequest
): Promise<{ data: AlertRuleDTO }> {
  return apiClient<{ data: AlertRuleDTO }>(`/api/v1/alerts/rules/${id}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  })
}

/**
 * Delete an alert rule
 *
 * @param id - Alert rule ID to delete
 * @returns Success message
 * @throws AuthenticationError if user is not authenticated
 * @throws NotFoundError if alert rule does not exist
 *
 * @example
 * const { message } = await deleteAlertRule('rule-id')
 * console.log(message)
 */
export async function deleteAlertRule(
  id: string
): Promise<{ message: string }> {
  return apiClient<{ message: string }>(`/api/v1/alerts/rules/${id}`, {
    method: 'DELETE',
  })
}

/**
 * Fetch all alert records from the API
 *
 * @param filters - Optional filters for alert records
 * @returns Array of alert records matching filters
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * // Fetch all alert records
 * const { data } = await fetchAlertRecords()
 *
 * // Fetch with filters
 * const { data: filtered } = await fetchAlertRecords({
 *   node_id: 'node-id',
 *   level: 'P0',
 *   status: 'pending'
 * })
 */
export async function fetchAlertRecords(
  filters?: AlertRecordFilters
): Promise<{ data: AlertRecordDTO[] }> {
  if (!filters || Object.keys(filters).length === 0) {
    // No filters, fetch all records
    return apiClient<{ data: AlertRecordDTO[] }>('/api/v1/alerts/records')
  }

  // Build query parameters from filters
  const params = new URLSearchParams()
  if (filters.node_id) params.append('node_id', filters.node_id)
  if (filters.start_time) params.append('start_time', filters.start_time)
  if (filters.end_time) params.append('end_time', filters.end_time)
  if (filters.level) params.append('level', filters.level)
  if (filters.status) params.append('status', filters.status)

  return apiClient<{ data: AlertRecordDTO[] }>(
    `/api/v1/alerts/records?${params}`
  )
}
