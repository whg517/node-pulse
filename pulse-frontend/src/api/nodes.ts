/**
 * Node API endpoints
 *
 * Provides typed functions for node-related API calls
 */

import { API_BASE_URL } from '../config/constants'

export interface NodeDTO {
  id: string
  name: string
  ip: string
  region: string
  tags: string[]
  status: 'online' | 'offline' | 'connecting'
}

/**
 * Fetch all nodes from the API
 */
export async function fetchNodes(): Promise<{ data: NodeDTO[] }> {
  const response = await fetch(`${API_BASE_URL}/api/v1/nodes`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error('Failed to fetch nodes')
  }

  return response.json()
}
