/**
 * Data Fixtures
 *
 * Provides test data management fixtures for E2E tests.
 * Use these fixtures to create, manage, and clean up test data.
 */
import { test as base } from '@playwright/test'
import {
  createTestNode,
  deleteTestNode,
  createTestProbe,
  deleteTestProbe,
  createTestAlertRule,
  deleteTestAlertRule,
  createTestWebhook,
  deleteTestWebhook,
  generateTestId,
  type NodeData,
  type ProbeData,
  type AlertRuleData,
  type WebhookData,
} from '../utils/test-data'

/**
 * Test data context for tracking created entities
 */
interface TestDataContext {
  nodes: Array<{ id: string; name: string }>
  probes: Array<{ id: string; name: string }>
  alerts: Array<{ id: string; name: string }>
  webhooks: Array<{ id: string; name: string }>
}

/**
 * Data fixtures type
 */
type DataFixtures = {
  testData: TestDataContext
  createNode: (data?: Partial<NodeData>) => Promise<{ id: string; name: string }>
  createProbe: (data?: Partial<ProbeData>) => Promise<{ id: string; name: string }>
  createAlertRule: (data?: Partial<AlertRuleData>) => Promise<{ id: string; name: string }>
  createWebhook: (data?: Partial<WebhookData>) => Promise<{ id: string; name: string }>
  cleanup: () => Promise<void>
}

/**
 * Extend base test with data fixtures
 */
export const test = base.extend<DataFixtures>({
  testData: async ({}, use) => {
    const context: TestDataContext = {
      nodes: [],
      probes: [],
      alerts: [],
      webhooks: [],
    }
    await use(context)
  },

  createNode: async ({ testData, request }, use) => {
    const createdNodes: Array<{ id: string; name: string }> = []

    const createNode = async (data?: Partial<NodeData>) => {
      const nodeData: NodeData = {
        name: generateTestId('node'),
        region: 'us-east-1',
        ...data,
      }

      const node = await createTestNode(nodeData)
      testData.nodes.push(node)
      createdNodes.push(node)
      return node
    }

    await use(createNode)

    // Cleanup
    for (const node of createdNodes) {
      try {
        await deleteTestNode(node.id)
      } catch (error) {
        console.warn(`Failed to delete node ${node.id}:`, error)
      }
    }
  },

  createProbe: async ({ testData, request }, use) => {
    const createdProbes: Array<{ id: string; name: string }> = []

    const createProbe = async (data?: Partial<ProbeData>) => {
      const probeData: ProbeData = {
        name: generateTestId('probe'),
        nodeId: data?.nodeId || (testData.nodes.length > 0 ? testData.nodes[testData.nodes.length - 1].id : ''),
        type: 'tcp',
        target: 'localhost',
        ...data,
      }

      const probe = await createTestProbe(probeData)
      testData.probes.push(probe)
      createdProbes.push(probe)
      return probe
    }

    await use(createProbe)

    // Cleanup
    for (const probe of createdProbes) {
      try {
        await deleteTestProbe(probe.id)
      } catch (error) {
        console.warn(`Failed to delete probe ${probe.id}:`, error)
      }
    }
  },

  createAlertRule: async ({ testData, request }, use) => {
    const createdAlerts: Array<{ id: string; name: string }> = []

    const createAlertRule = async (data?: Partial<AlertRuleData>) => {
      const alertData: AlertRuleData = {
        name: generateTestId('alert'),
        metricType: 'latency_avg',
        threshold: 100,
        operator: 'gt',
        ...data,
      }

      const alert = await createTestAlertRule(alertData)
      testData.alerts.push(alert)
      createdAlerts.push(alert)
      return alert
    }

    await use(createAlertRule)

    // Cleanup
    for (const alert of createdAlerts) {
      try {
        await deleteTestAlertRule(alert.id)
      } catch (error) {
        console.warn(`Failed to delete alert ${alert.id}:`, error)
      }
    }
  },

  createWebhook: async ({ testData, request }, use) => {
    const createdWebhooks: Array<{ id: string; name: string }> = []

    const createWebhook = async (data?: Partial<WebhookData>) => {
      const webhookData: WebhookData = {
        name: generateTestId('webhook'),
        url: 'https://webhook.site/test',
        ...data,
      }

      const webhook = await createTestWebhook(webhookData)
      testData.webhooks.push(webhook)
      createdWebhooks.push(webhook)
      return webhook
    }

    await use(createWebhook)

    // Cleanup
    for (const webhook of createdWebhooks) {
      try {
        await deleteTestWebhook(webhook.id)
      } catch (error) {
        console.warn(`Failed to delete webhook ${webhook.id}:`, error)
      }
    }
  },

  cleanup: async ({ testData }, use) => {
    await use(async () => {
      // Cleanup all nodes
      for (const node of testData.nodes) {
        try {
          await deleteTestNode(node.id)
        } catch (error) {
          console.warn(`Failed to delete node ${node.id}:`, error)
        }
      }

      // Cleanup all probes
      for (const probe of testData.probes) {
        try {
          await deleteTestProbe(probe.id)
        } catch (error) {
          console.warn(`Failed to delete probe ${probe.id}:`, error)
        }
      }

      // Cleanup all alerts
      for (const alert of testData.alerts) {
        try {
          await deleteTestAlertRule(alert.id)
        } catch (error) {
          console.warn(`Failed to delete alert ${alert.id}:`, error)
        }
      }

      // Cleanup all webhooks
      for (const webhook of testData.webhooks) {
        try {
          await deleteTestWebhook(webhook.id)
        } catch (error) {
          console.warn(`Failed to delete webhook ${webhook.id}:`, error)
        }
      }

      // Reset context
      testData.nodes = []
      testData.probes = []
      testData.alerts = []
      testData.webhooks = []
    })
  },
})

export { expect } from '@playwright/test'
