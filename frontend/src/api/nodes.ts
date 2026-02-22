/**
 * Node Management API endpoints
 *
 * Provides typed functions for node CRUD operations including
 * fetching, creating, updating, and deleting nodes.
 */

import { apiClient } from './client'
import type {
  NodeDTO,
  CreateNodeRequest,
  UpdateNodeRequest
} from './types'

// Export types for use in components and stores
export type { NodeDTO, CreateNodeRequest, UpdateNodeRequest }

/**
 * Fetch all nodes from the API
 *
 * @returns Object containing nodes array
 * @throws AuthenticationError if user is not authenticated
 * @throws ApiError on other HTTP errors
 *
 * @example
 * const { data } = await fetchNodes()
 * console.log('Total nodes:', data.nodes.length)
 */
export async function fetchNodes(): Promise<{ data: { nodes: NodeDTO[] } }> {
  return apiClient<{ data: { nodes: NodeDTO[] } }>('/api/v1/nodes')
}

/**
 * Fetch a single node by ID
 *
 * @param id - Node ID to fetch
 * @returns Node data
 * @throws AuthenticationError if user is not authenticated
 * @throws NotFoundError if node does not exist
 *
 * @example
 * const { data } = await fetchNode('node-id')
 * console.log('Node:', data.name)
 */
export async function fetchNode(id: string): Promise<{ data: NodeDTO }> {
  return apiClient<{ data: NodeDTO }>(`/api/v1/nodes/${id}`)
}

/**
 * Create a new node
 *
 * @param request - Node creation parameters
 * @returns Created node data
 * @throws ValidationError if request parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const { data } = await createNode({
 *   name: 'New Node',
 *   ip: '192.168.1.100',
 *   region: 'us-east',
 *   tags: ['production']
 * })
 */
export async function createNode(
  request: CreateNodeRequest
): Promise<{ data: NodeDTO }> {
  return apiClient<{ data: NodeDTO }>('/api/v1/nodes', {
    method: 'POST',
    body: JSON.stringify(request),
  })
}

/**
 * Update an existing node
 *
 * @param id - Node ID to update
 * @param request - Node update parameters (all optional)
 * @returns Updated node data
 * @throws ValidationError if request parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 * @throws NotFoundError if node does not exist
 *
 * @example
 * const { data } = await updateNode('node-id', {
 *   name: 'Updated Name',
 *   tags: ['production', 'critical']
 * })
 */
export async function updateNode(
  id: string,
  request: UpdateNodeRequest
): Promise<{ data: NodeDTO }> {
  return apiClient<{ data: NodeDTO }>(`/api/v1/nodes/${id}`, {
    method: 'PUT',
    body: JSON.stringify(request),
  })
}

/**
 * Delete a node
 *
 * @param id - Node ID to delete
 * @returns Success message
 * @throws AuthenticationError if user is not authenticated
 * @throws NotFoundError if node does not exist
 *
 * @example
 * const { message } = await deleteNode('node-id')
 * console.log(message)
 */
export async function deleteNode(
  id: string
): Promise<{ message: string }> {
  return apiClient<{ message: string }>(`/api/v1/nodes/${id}`, {
    method: 'DELETE',
  })
}

/**
 * Fetch node status
 *
 * @param id - Node ID to query
 * @returns Node status information
 * @throws AuthenticationError if user is not authenticated
 * @throws NotFoundError if node does not exist
 *
 * @example
 * const { data } = await fetchNodeStatus('node-id')
 * console.log('Status:', data.status)
 * console.log('Last heartbeat:', data.last_heartbeat)
 */
export async function fetchNodeStatus(
  id: string
): Promise<{ data: { status: string; last_heartbeat: string } }> {
  return apiClient<{ data: { status: string; last_heartbeat: string } }>(
    `/api/v1/nodes/${id}/status`
  )
}
