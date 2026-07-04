/**
 * Alert Routing Rules API (ADR-002).
 * Per-webhook routing rules. Backend: /api/v1/alerts/routing-rules.
 */
import { apiClient } from './client'

export interface AlertRoutingRuleDTO {
  id: string
  owner_user_id: string
  name: string
  enabled: boolean
  metric?: string
  severities?: string[]
  node_id?: string
  webhook_id: string
  created_at: string
  updated_at: string
}

export interface CreateRoutingRuleRequest {
  name: string
  enabled: boolean
  metric?: string
  severities?: string[]
  node_id?: string
  webhook_id: string
}

export async function listRoutingRules(): Promise<{ data: { rules: AlertRoutingRuleDTO[] }; message: string; timestamp: string }> {
  return apiClient('/api/v1/alerts/routing-rules')
}

export async function createRoutingRule(req: CreateRoutingRuleRequest): Promise<{ data: AlertRoutingRuleDTO; message: string; timestamp: string }> {
  return apiClient('/api/v1/alerts/routing-rules', { method: 'POST', body: JSON.stringify(req) })
}

export async function updateRoutingRule(id: string, req: CreateRoutingRuleRequest): Promise<{ data: AlertRoutingRuleDTO; message: string; timestamp: string }> {
  return apiClient(`/api/v1/alerts/routing-rules/${id}`, { method: 'PUT', body: JSON.stringify(req) })
}

export async function deleteRoutingRule(id: string): Promise<{ message: string; timestamp: string }> {
  return apiClient(`/api/v1/alerts/routing-rules/${id}`, { method: 'DELETE' })
}
