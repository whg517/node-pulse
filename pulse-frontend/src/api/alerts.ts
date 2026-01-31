/**
 * Alerts API endpoints
 *
 * Provides typed functions for alert-related API calls
 */

import { API_BASE_URL } from '../config/constants'

export interface AlertRuleDTO {
  id: string
  metric: 'latency' | 'packet_loss_rate' | 'jitter'
  threshold: number
  level: 'P0' | 'P1' | 'P2'
  node_id: string | null
  enabled: boolean
}

export interface AlertRecordDTO {
  id: string
  node_id: string
  metric: string
  level: string
  status: 'pending' | 'processing' | 'resolved'
  timestamp: string
}

/**
 * Fetch all alert rules from the API
 */
export async function fetchAlertRules(): Promise<{ data: AlertRuleDTO[] }> {
  const response = await fetch(`${API_BASE_URL}/api/v1/alerts/rules`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error('Failed to fetch alert rules')
  }

  return response.json()
}

/**
 * Fetch all alert records from the API
 */
export async function fetchAlertRecords(): Promise<{ data: AlertRecordDTO[] }> {
  const response = await fetch(`${API_BASE_URL}/api/v1/alerts/records`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error('Failed to fetch alert records')
  }

  return response.json()
}
