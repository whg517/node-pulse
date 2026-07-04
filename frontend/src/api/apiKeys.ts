/**
 * Admin API Key Management endpoints
 *
 * Backend routes: /api/v1/admin/apikeys (admin only). Used to provision and
 * rotate API keys for Beacon authentication (see J12 in docs/user-journey.md).
 */

import { apiClient } from './client'

/** A stored API key. The full secret is NEVER returned after creation; only the prefix. */
export interface ApiKeyDTO {
  id: number
  key_id: string
  key_prefix: string
  user_id?: string | null
  service_account_id?: string | null
  name: string
  is_active: boolean
  expires_at?: string | null
  created_at: string
  last_used_at?: string | null
}

export interface CreateApiKeyRequest {
  name: string
  /** Optional owner user id. Omit for a general-purpose key. */
  user_id?: string | null
  service_account_id?: string | null
}

export interface CreateApiKeyResponse {
  data: {
    /** The stored key record (without the secret). */
    key: ApiKeyDTO
    /** The full key string, shown ONLY on creation. Must be saved by the caller. */
    full_key: string
  }
  message: string
  timestamp: string
}

export interface RotateApiKeyResponse {
  data: {
    key: ApiKeyDTO
    /** The new full key string, shown ONLY on rotation. */
    full_key: string
  }
  message: string
  timestamp: string
}

/** List all API keys. */
export async function getApiKeys(): Promise<{ data: { keys: ApiKeyDTO[] }; message: string; timestamp: string }> {
  return apiClient('/api/v1/admin/apikeys')
}

/** Get a single API key by ID. */
export async function getApiKey(id: number): Promise<{ data: { key: ApiKeyDTO }; message: string; timestamp: string }> {
  return apiClient(`/api/v1/admin/apikeys/${id}`)
}

/**
 * Create a new API key. The full key is returned ONCE in `data.full_key` and
 * must be saved immediately — it cannot be retrieved again.
 */
export async function createApiKey(
  request: CreateApiKeyRequest
): Promise<CreateApiKeyResponse> {
  return apiClient('/api/v1/admin/apikeys', {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

/** Rotate an API key. The old key stays valid for a grace period (24h backend-side). */
export async function rotateApiKey(id: number): Promise<RotateApiKeyResponse> {
  return apiClient(`/api/v1/admin/apikeys/${id}/rotate`, {
    method: 'POST',
  })
}

/** Revoke (permanently deactivate) an API key. */
export async function revokeApiKey(id: number): Promise<{ message: string; timestamp: string }> {
  return apiClient(`/api/v1/admin/apikeys/${id}`, {
    method: 'DELETE',
  })
}
