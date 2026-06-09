import { apiClient } from './client'

export interface ProbeDTO {
  id: string
  node_id: string
  type: 'TCP' | 'UDP'
  target: string
  port: number
  interval_seconds: number
  count: number
  timeout_seconds: number
  created_at: string
  updated_at: string
}

export interface CreateProbeRequest {
  node_id: string
  type: 'TCP' | 'UDP'
  target: string
  port: number
  interval_seconds: number
  count: number
  timeout_seconds: number
}

export interface UpdateProbeRequest {
  type?: 'TCP' | 'UDP'
  target?: string
  port?: number
  interval_seconds?: number
  count?: number
  timeout_seconds?: number
}

export async function fetchProbes(
  nodeId?: string
): Promise<{ data: { probes: ProbeDTO[] }; message: string; timestamp: string }> {
  const params = nodeId ? `?node_id=${nodeId}` : ''
  return apiClient(`/api/v1/probes${params}`)
}

export async function fetchProbe(
  id: string
): Promise<{ data: { probe: ProbeDTO }; message: string; timestamp: string }> {
  return apiClient(`/api/v1/probes/${id}`)
}

export async function createProbe(
  request: CreateProbeRequest
): Promise<{ data: { probe: ProbeDTO }; message: string; timestamp: string }> {
  return apiClient('/api/v1/probes', {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

export async function updateProbe(
  id: string,
  request: UpdateProbeRequest
): Promise<{ data: { probe: ProbeDTO }; message: string; timestamp: string }> {
  return apiClient(`/api/v1/probes/${id}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  })
}

export async function deleteProbe(
  id: string
): Promise<{ message: string; timestamp: string }> {
  return apiClient(`/api/v1/probes/${id}?confirm=true`, {
    method: 'DELETE',
  })
}
