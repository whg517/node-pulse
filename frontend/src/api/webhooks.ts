/**
 * Webhook Management API endpoints
 *
 * Provides typed functions for webhook CRUD operations.
 */

import { apiClient } from './client'

export interface WebhookDTO {
  id: string
  url: string
  event_format: Record<string, any>
  enabled: boolean
  created_at: string
}

export interface CreateWebhookRequest {
  url: string
  event_format?: Record<string, any>
  enabled?: boolean
}

export interface UpdateWebhookRequest {
  url?: string
  event_format?: Record<string, any>
  enabled?: boolean
}

/**
 * Fetch all webhooks from the API
 *
 * @returns Object containing webhooks array
 * @throws AuthenticationError if user is not authenticated
 */
export async function fetchWebhooks(): Promise<{ data: { webhooks: WebhookDTO[] } }> {
  return apiClient<{ data: { webhooks: WebhookDTO[] } }>('/api/v1/webhooks')
}

/**
 * Create a new webhook
 *
 * @param request - Webhook creation parameters
 * @returns Created webhook data
 * @throws ValidationError if request parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 */
export async function createWebhook(
  request: CreateWebhookRequest
): Promise<{ data: WebhookDTO }> {
  return apiClient<{ data: WebhookDTO }>('/api/v1/webhooks', {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

/**
 * Update an existing webhook
 *
 * @param id - Webhook ID to update
 * @param request - Webhook update parameters (all optional)
 * @returns Updated webhook data
 * @throws ValidationError if request parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 * @throws NotFoundError if webhook does not exist
 */
export async function updateWebhook(
  id: string,
  request: UpdateWebhookRequest
): Promise<{ data: WebhookDTO }> {
  return apiClient<{ data: WebhookDTO }>(`/api/v1/webhooks/${id}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  })
}

/**
 * Delete a webhook
 *
 * @param id - Webhook ID to delete
 * @returns Success message
 * @throws AuthenticationError if user is not authenticated
 * @throws NotFoundError if webhook does not exist
 */
export async function deleteWebhook(
  id: string
): Promise<{ message: string }> {
  return apiClient<{ message: string }>(`/api/v1/webhooks/${id}`, {
    method: 'DELETE',
  })
}
