/**
 * Webhook Management API endpoints
 *
 * Provides typed functions for webhook CRUD operations.
 */

import { apiClient } from './client'

export type WebhookEventFormat = Record<string, unknown>

export interface WebhookDTO {
  id: string
  url: string
  event_format: WebhookEventFormat
  enabled: boolean
  created_at: string
}

export interface CreateWebhookRequest {
  url: string
  event_format?: WebhookEventFormat
  enabled?: boolean
}

export interface UpdateWebhookRequest {
  url?: string
  event_format?: WebhookEventFormat
  enabled?: boolean
}

export interface PreviewWebhookPayloadRequest {
  event_format?: WebhookEventFormat
}

export interface PreviewWebhookPayloadResponse {
  data: {
    payload: WebhookEventFormat
  }
}

export interface TestWebhookResponse {
  data: {
    webhook_id: string
    status: 'success' | 'failure'
    error?: string
  }
  message: string
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

/**
 * Render a webhook payload preview using a sample alert event.
 *
 * @param request - Event format to render
 * @returns Rendered payload preview
 * @throws ValidationError if the event format is invalid
 */
export async function previewWebhookPayload(
  request: PreviewWebhookPayloadRequest
): Promise<PreviewWebhookPayloadResponse> {
  return apiClient<PreviewWebhookPayloadResponse>('/api/v1/webhooks/preview', {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

/**
 * Send a sample alert payload to a configured webhook.
 *
 * @param id - Webhook ID to test
 * @returns Delivery result from the backend
 */
export async function testWebhook(id: string): Promise<TestWebhookResponse> {
  return apiClient<TestWebhookResponse>(`/api/v1/webhooks/${id}/test`, {
    method: 'POST',
  })
}
