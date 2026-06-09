import { apiClient } from './client'

export interface ProbeConfigDTO {
  id: string
  type: 'TCP' | 'UDP'
  target: string
  port: number
  interval_seconds: number
  timeout_seconds: number
  count: number
}

export interface BeaconConfigDTO {
  probes: ProbeConfigDTO[]
  interval_seconds: number
  timeout_seconds: number
  updated_at: string
  version: number
}

export interface ConfigHistoryEntry {
  version: number
  config: BeaconConfigDTO
  changed_at: string
  changed_by: string
}

export interface ConfigPreviewResult {
  valid: boolean
  warnings: string[]
  conflicts: string[]
}

export interface BeaconConfigUpdateRequest {
  probes?: ProbeConfigDTO[]
  interval_seconds?: number
  timeout_seconds?: number
}

export interface BatchConfigUpdateRequest {
  beacon_ids: string[]
  config: BeaconConfigUpdateRequest
}

export async function fetchBeaconConfig(
  beaconId: string
): Promise<{ data: BeaconConfigDTO; message: string; timestamp: string }> {
  return apiClient(`/api/v1/beacons/${beaconId}/config`)
}

export async function updateBeaconConfig(
  beaconId: string,
  request: BeaconConfigUpdateRequest
): Promise<{ data: BeaconConfigDTO; message: string; timestamp: string }> {
  return apiClient(`/api/v1/beacons/${beaconId}/config`, {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

export async function fetchConfigHistory(
  beaconId: string
): Promise<{ data: ConfigHistoryEntry[]; message: string; timestamp: string }> {
  return apiClient(`/api/v1/beacons/${beaconId}/config/history`)
}

export async function previewConfig(
  beaconId: string,
  request: BeaconConfigUpdateRequest
): Promise<{ data: ConfigPreviewResult; message: string; timestamp: string }> {
  return apiClient(`/api/v1/beacons/${beaconId}/config/preview`, {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

export async function batchUpdateConfig(
  groupId: string,
  request: BatchConfigUpdateRequest
): Promise<{
  data: { success_count: number; failed_count: number; failed_ids?: string[]; errors?: string[] }
  message: string
  timestamp: string
}> {
  return apiClient(`/api/v1/beacon-groups/${groupId}/config`, {
    method: 'POST',
    body: JSON.stringify(request),
  })
}
