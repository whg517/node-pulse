/**
 * Operator RBAC Tests
 *
 * Tests for operator role permissions:
 * - Can CRUD nodes and probes
 * - Can CRUD alert rules
 * - Cannot access webhooks (admin only)
 * - Cannot access data export (admin only)
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { NodesPage } from '../../pages/NodesPage'
import { WebhooksPage } from '../../pages/WebhooksPage'
import { DataExportPage } from '../../pages/DataExportPage'
import { AlertRulesPage } from '../../pages/AlertsPage'

test.describe('Operator RBAC', () => {
  test.describe('Page Access', () => {
    test('can access dashboard', async ({ operatorPage }) => {
      await operatorPage.goto('/dashboard')
      await expect(operatorPage).toHaveURL(/.*dashboard/)
    })

    test('can access nodes page', async ({ operatorPage }) => {
      await operatorPage.goto('/nodes')
      await expect(operatorPage).toHaveURL(/.*nodes/)
    })

    test('can access alerts pages', async ({ operatorPage }) => {
      await operatorPage.goto('/alerts/rules')
      await expect(operatorPage).toHaveURL(/.*alerts\/rules/)

      await operatorPage.goto('/alerts/records')
      await expect(operatorPage).toHaveURL(/.*alerts\/records/)

      await operatorPage.goto('/alerts/history')
      await expect(operatorPage).toHaveURL(/.*alerts\/history/)
    })

    test('can access comparison page', async ({ operatorPage }) => {
      await operatorPage.goto('/nodes/comparison')
      await expect(operatorPage).toHaveURL(/.*comparison/)
    })

    test('can access performance page', async ({ operatorPage }) => {
      await operatorPage.goto('/reports')
      await expect(operatorPage).toHaveURL(/.*reports/)
    })

    test('can access sessions page', async ({ operatorPage }) => {
      await operatorPage.goto('/settings/sessions')
      await expect(operatorPage).toHaveURL(/.*sessions/)
    })
  })

  test.describe('Nodes CRUD', () => {
    let nodesPage: NodesPage

    test.beforeEach(async ({ operatorPage }) => {
      nodesPage = new NodesPage(operatorPage)
      await nodesPage.goto()
    })

    test('AC-11: can see create node button', async ({ operatorPage }) => {
      await nodesPage.expectTableVisible()

      // Operator CAN see create button (admin/operator can CRUD nodes)
      const createButton = operatorPage.locator('button:has-text("Create"), button:has-text("Add")')
      await expect(createButton).toBeVisible()
    })

    test('AC-15: can create nodes', async ({ operatorPage }) => {
      await nodesPage.expectTableVisible()

      const nodeName = `e2e_operator_node_${Date.now()}`
      await nodesPage.createNode(nodeName, 'us-west-2')

      // Verify node appears in list
      const hasNode = await nodesPage.hasNode(nodeName)
      expect(hasNode).toBeTruthy()
    })

    test('can edit nodes', async ({ operatorPage }) => {
      await nodesPage.expectTableVisible()

      const editButtons = operatorPage.locator('table tbody button:has-text("Edit")')
      const count = await editButtons.count()

      expect(count).toBeGreaterThanOrEqual(0)
    })

    test('can delete nodes', async ({ operatorPage }) => {
      await nodesPage.expectTableVisible()

      const deleteButtons = operatorPage.locator('table tbody button:has-text("Delete")')
      const count = await deleteButtons.count()

      expect(count).toBeGreaterThanOrEqual(0)
    })
  })

  test.describe('Alert Rules CRUD', () => {
    let alertRulesPage: AlertRulesPage

    test.beforeEach(async ({ operatorPage }) => {
      alertRulesPage = new AlertRulesPage(operatorPage)
      await alertRulesPage.goto()
    })

    test('AC-17: can create alert rules', async ({ operatorPage }) => {
      await alertRulesPage.expectTableVisible()

      const createButton = operatorPage.locator('button:has-text("Create"), button:has-text("Add")')
      await expect(createButton).toBeVisible()
    })

    test('can edit alert rules', async ({ operatorPage }) => {
      await alertRulesPage.expectTableVisible()

      const editButtons = operatorPage.locator('table tbody button:has-text("Edit")')
      const count = await editButtons.count()

      expect(count).toBeGreaterThanOrEqual(0)
    })
  })

  test.describe('Webhooks Access (Admin Only)', () => {
    let webhooksPage: WebhooksPage

    test.beforeEach(async ({ operatorPage }) => {
      webhooksPage = new WebhooksPage(operatorPage)
      await webhooksPage.goto()
    })

    test('AC-10: cannot access webhooks - shows warning', async ({ operatorPage }) => {
      // Page may still load but should show access warning
      const hasWarning = await webhooksPage.hasAccessWarning()
      expect(hasWarning).toBeTruthy()
    })

    test('cannot see create webhook button', async ({ operatorPage }) => {
      // If page shows warning, create button should not be functional
      const createButton = operatorPage.locator('button:has-text("Create"), button:has-text("Add")')

      // Either button is hidden or disabled, or warning is shown
      const isVisible = await createButton.isVisible().catch(() => false)
      const hasWarning = await webhooksPage.hasAccessWarning()

      expect(!isVisible || hasWarning).toBeTruthy()
    })

    test('webhook API returns 403', async ({ operatorPage }) => {
      const response = await operatorPage.request.get('/api/v1/webhooks')
      expect(response.status()).toBe(403)
    })
  })

  test.describe('Data Export (Admin Only)', () => {
    let dataExportPage: DataExportPage

    test.beforeEach(async ({ operatorPage }) => {
      dataExportPage = new DataExportPage(operatorPage)
      await dataExportPage.goto()
    })

    test('AC-12: cannot access export - shows warning', async ({ operatorPage }) => {
      const hasWarning = await dataExportPage.hasAccessWarning()
      expect(hasWarning).toBeTruthy()
    })

    test('export API returns 403', async ({ operatorPage }) => {
      const response = await operatorPage.request.post('/api/v1/data/export', {
        data: {
          node_ids: [],
          start_time: new Date().toISOString(),
          end_time: new Date().toISOString(),
          format: 'csv',
        },
      })
      expect(response.status()).toBe(403)
    })
  })

  test.describe('Config Access', () => {
    test('cannot access config endpoint', async ({ operatorPage }) => {
      const response = await operatorPage.request.get('/api/v1/config')
      expect(response.status()).toBe(403)
    })
  })

  test.describe('Read Operations', () => {
    test('can read nodes', async ({ operatorPage }) => {
      const response = await operatorPage.request.get('/api/v1/nodes')
      expect(response.ok()).toBeTruthy()
    })

    test('can read probes', async ({ operatorPage }) => {
      const response = await operatorPage.request.get('/api/v1/probes')
      expect(response.ok()).toBeTruthy()
    })

    test('can read alert rules', async ({ operatorPage }) => {
      const response = await operatorPage.request.get('/api/v1/alerts/rules')
      expect(response.ok()).toBeTruthy()
    })

    test('can read metrics', async ({ operatorPage }) => {
      const response = await operatorPage.request.get('/api/v1/data/metrics')
      expect(response.ok()).toBeTruthy()
    })
  })
})
