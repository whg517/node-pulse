/**
 * Test Data Management Utilities
 *
 * Provides helper functions for creating, managing, and cleaning up test data.
 * Uses Playwright's API testing capabilities for fast and isolated test data setup.
 */
import { APIRequestContext, request } from '@playwright/test'

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:6532'

/**
 * Create API context for test data operations
 */
async function createApiContext(authToken?: string): Promise<APIRequestContext> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (authToken) {
    headers['Authorization'] = `Bearer ${authToken}`
  }

  return await request.newContext({
    baseURL: API_BASE_URL,
    extraHTTPHeaders: headers,
  })
}

/**
 * Node data interface
 */
export interface NodeData {
  name: string
  region?: string
  ip?: string
  tags?: string[]
}

/**
 * Create a test node via API
 */
export async function createTestNode(
  nodeData: NodeData,
  authToken?: string
): Promise<{ id: string; name: string }> {
  const apiContext = await createApiContext(authToken)

  try {
    const response = await apiContext.post('/api/v1/nodes', {
      data: nodeData,
    })

    if (!response.ok()) {
      throw new Error(`Failed to create node: ${response.status()} ${await response.text()}`)
    }

    const result = await response.json()
    return {
      id: result.data.id,
      name: result.data.name,
    }
  } finally {
    await apiContext.dispose()
  }
}

/**
 * Delete a test node via API
 */
export async function deleteTestNode(
  nodeId: string,
  authToken?: string
): Promise<void> {
  const apiContext = await createApiContext(authToken)

  try {
    const response = await apiContext.delete(`/api/v1/nodes/${nodeId}`)

    if (!response.ok() && response.status() !== 404) {
      throw new Error(`Failed to delete node: ${response.status()} ${await response.text()}`)
    }
  } finally {
    await apiContext.dispose()
  }
}

/**
 * Probe data interface
 */
export interface ProbeData {
  nodeId: string
  name: string
  type: 'tcp' | 'udp' | 'http' | 'https' | 'ping'
  target: string
  port?: number
  interval?: number
}

/**
 * Create a test probe via API
 */
export async function createTestProbe(
  probeData: ProbeData,
  authToken?: string
): Promise<{ id: string; name: string }> {
  const apiContext = await createApiContext(authToken)

  try {
    const response = await apiContext.post('/api/v1/probes', {
      data: probeData,
    })

    if (!response.ok()) {
      throw new Error(`Failed to create probe: ${response.status()} ${await response.text()}`)
    }

    const result = await response.json()
    return {
      id: result.data.id,
      name: result.data.name,
    }
  } finally {
    await apiContext.dispose()
  }
}

/**
 * Delete a test probe via API
 */
export async function deleteTestProbe(
  probeId: string,
  authToken?: string
): Promise<void> {
  const apiContext = await createApiContext(authToken)

  try {
    const response = await apiContext.delete(`/api/v1/probes/${probeId}`)

    if (!response.ok() && response.status() !== 404) {
      throw new Error(`Failed to delete probe: ${response.status()} ${await response.text()}`)
    }
  } finally {
    await apiContext.dispose()
  }
}

/**
 * Alert rule data interface
 */
export interface AlertRuleData {
  name: string
  metricType: string
  threshold: number
  operator: 'gt' | 'lt' | 'eq' | 'gte' | 'lte'
  duration?: number
  level?: 'info' | 'warning' | 'critical'
  enabled?: boolean
}

/**
 * Create a test alert rule via API
 */
export async function createTestAlertRule(
  ruleData: AlertRuleData,
  authToken?: string
): Promise<{ id: string; name: string }> {
  const apiContext = await createApiContext(authToken)

  try {
    const response = await apiContext.post('/api/v1/alerts/rules', {
      data: ruleData,
    })

    if (!response.ok()) {
      throw new Error(`Failed to create alert rule: ${response.status()} ${await response.text()}`)
    }

    const result = await response.json()
    return {
      id: result.data.id,
      name: result.data.name,
    }
  } finally {
    await apiContext.dispose()
  }
}

/**
 * Delete a test alert rule via API
 */
export async function deleteTestAlertRule(
  ruleId: string,
  authToken?: string
): Promise<void> {
  const apiContext = await createApiContext(authToken)

  try {
    const response = await apiContext.delete(`/api/v1/alerts/rules/${ruleId}`)

    if (!response.ok() && response.status() !== 404) {
      throw new Error(`Failed to delete alert rule: ${response.status()} ${await response.text()}`)
    }
  } finally {
    await apiContext.dispose()
  }
}

/**
 * Webhook data interface
 */
export interface WebhookData {
  name: string
  url: string
  secret?: string
  events?: string[]
  enabled?: boolean
}

/**
 * Create a test webhook via API
 */
export async function createTestWebhook(
  webhookData: WebhookData,
  authToken?: string
): Promise<{ id: string; name: string }> {
  const apiContext = await createApiContext(authToken)

  try {
    const response = await apiContext.post('/api/v1/webhooks', {
      data: webhookData,
    })

    if (!response.ok()) {
      throw new Error(`Failed to create webhook: ${response.status()} ${await response.text()}`)
    }

    const result = await response.json()
    return {
      id: result.data.id,
      name: result.data.name,
    }
  } finally {
    await apiContext.dispose()
  }
}

/**
 * Delete a test webhook via API
 */
export async function deleteTestWebhook(
  webhookId: string,
  authToken?: string
): Promise<void> {
  const apiContext = await createApiContext(authToken)

  try {
    const response = await apiContext.delete(`/api/v1/webhooks/${webhookId}`)

    if (!response.ok() && response.status() !== 404) {
      throw new Error(`Failed to delete webhook: ${response.status()} ${await response.text()}`)
    }
  } finally {
    await apiContext.dispose()
  }
}

/**
 * Bulk cleanup utility - delete multiple test entities
 */
export async function cleanupTestData(
  entities: Array<{ type: 'node' | 'probe' | 'alert' | 'webhook'; id: string }>,
  authToken?: string
): Promise<void> {
  const apiContext = await createApiContext(authToken)

  try {
    for (const entity of entities) {
      try {
        let endpoint = ''
        switch (entity.type) {
          case 'node':
            endpoint = `/api/v1/nodes/${entity.id}`
            break
          case 'probe':
            endpoint = `/api/v1/probes/${entity.id}`
            break
          case 'alert':
            endpoint = `/api/v1/alerts/rules/${entity.id}`
            break
          case 'webhook':
            endpoint = `/api/v1/webhooks/${entity.id}`
            break
        }

        if (endpoint) {
          await apiContext.delete(endpoint)
        }
      } catch (error) {
        // Ignore errors during cleanup
        console.warn(`Failed to cleanup ${entity.type} ${entity.id}:`, error)
      }
    }
  } finally {
    await apiContext.dispose()
  }
}

/**
 * Generate unique test identifier
 */
export function generateTestId(prefix: string = 'test'): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`
}

/**
 * Wait for async operation with timeout
 */
export async function waitForCondition<T>(
  condition: () => Promise<T>,
  timeout: number = 10000,
  interval: number = 500
): Promise<T> {
  const startTime = Date.now()

  while (Date.now() - startTime < timeout) {
    const result = await condition()
    if (result) {
      return result
    }
    await new Promise(resolve => setTimeout(resolve, interval))
  }

  throw new Error(`Condition not met within ${timeout}ms`)
}
