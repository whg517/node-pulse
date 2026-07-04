/**
 * Beacon Config Templates API (ADR-003).
 * Server-owned reusable probe-config templates. Backend: /api/v1/beacon-config-templates.
 */
import { apiClient } from './client'

export interface TemplateProbeDTO {
  type?: string
  target?: string
  port?: number
  interval_seconds?: number
  timeout_seconds?: number
  count?: number
  [k: string]: unknown
}

export interface BeaconConfigTemplateDTO {
  id: string
  owner_user_id: string
  name: string
  description?: string
  probes: TemplateProbeDTO[]
  interval_seconds: number
  timeout_seconds: number
  created_at: string
  updated_at: string
}

export interface CreateTemplateRequest {
  name: string
  description?: string
  probes: TemplateProbeDTO[]
  interval_seconds: number
  timeout_seconds: number
}

export async function listBeaconConfigTemplates(): Promise<{ data: { templates: BeaconConfigTemplateDTO[] }; message: string; timestamp: string }> {
  return apiClient('/api/v1/beacon-config-templates')
}

export async function createBeaconConfigTemplate(req: CreateTemplateRequest): Promise<{ data: BeaconConfigTemplateDTO; message: string; timestamp: string }> {
  return apiClient('/api/v1/beacon-config-templates', { method: 'POST', body: JSON.stringify(req) })
}

export async function updateBeaconConfigTemplate(id: string, req: CreateTemplateRequest): Promise<{ data: BeaconConfigTemplateDTO; message: string; timestamp: string }> {
  return apiClient(`/api/v1/beacon-config-templates/${id}`, { method: 'PUT', body: JSON.stringify(req) })
}

export async function deleteBeaconConfigTemplate(id: string): Promise<{ message: string; timestamp: string }> {
  return apiClient(`/api/v1/beacon-config-templates/${id}`, { method: 'DELETE' })
}
