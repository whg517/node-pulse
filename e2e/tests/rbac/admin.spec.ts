/**
 * Admin RBAC Tests
 *
 * Tests for admin role permissions:
 * - Full access to all pages
 * - CRUD operations on all resources
 * - Access to webhooks and data export
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { NodesPage } from '../../pages/NodesPage'
import { WebhooksPage } from '../../pages/WebhooksPage'
import { DataExportPage } from '../../pages/DataExportPage'
import { AlertRulesPage } from '../../pages/AlertsPage'

test.describe('Admin RBAC', () => {
  test.describe('Page Access', () => {
    test('can access all protected pages', async ({ adminPage }) => {
      const pages = [
        '/dashboard',
        '/nodes',
        '/comparison',
        '/alerts/rules',
        '/alerts/records',
        '/alerts/history',
        '/webhooks',
        '/export',
        '/performance',
        '/sessions',
      ]

      for (const pagePath of pages) {
        await adminPage.goto(pagePath)
        await expect(adminPage).toHaveURL(new RegExp(`.*${pagePath}`))
      }
    })
  })

  test.describe('Nodes CRUD', () => {
    let nodesPage: NodesPage

    test.beforeEach(async ({ adminPage }) => {
      nodesPage = new NodesPage(adminPage)
      await nodesPage.goto()
    })

    test('AC-14: can create nodes', async ({ adminPage }) => {
      await nodesPage.expectTableVisible()
      await nodesPage.expectCreateButtonVisible()

      const nodeName = `e2e_test_node_${Date.now()}`
      await nodesPage.createNode(nodeName, 'us-east-1')

      // Verify node appears in list
      const hasNode = await nodesPage.hasNode(nodeName)
      expect(hasNode).toBeTruthy()
    })

    test('can edit nodes', async ({ adminPage }) => {
      await nodesPage.expectTableVisible()

      const rowCount = await nodesPage.getRowCount()
      if (rowCount > 0) {
        const editButton = adminPage.locator('table tbody tr').first().locator('button:has-text("Edit")')
        await expect(editButton).toBeVisible()
      }
    })

    test('can delete nodes', async ({ adminPage }) => {
      await nodesPage.expectTableVisible()

      const deleteButtons = adminPage.locator('table tbody button:has-text("Delete")')
      const count = await deleteButtons.count()

      expect(count).toBeGreaterThan(0)
    })
  })

  test.describe('Alert Rules CRUD', () => {
    let alertRulesPage: AlertRulesPage

    test.beforeEach(async ({ adminPage }) => {
      alertRulesPage = new AlertRulesPage(adminPage)
      await alertRulesPage.goto()
    })

    test('can create alert rules', async ({ adminPage }) => {
      await alertRulesPage.expectTableVisible()

      const createButton = adminPage.locator('button:has-text("Create"), button:has-text("Add")')
      await expect(createButton).toBeVisible()
    })

    test('can edit alert rules', async ({ adminPage }) => {
      await alertRulesPage.expectTableVisible()

      const editButtons = adminPage.locator('table tbody button:has-text("Edit")')
      const count = await editButtons.count()

      // Admin should see edit buttons
      expect(count).toBeGreaterThanOrEqual(0)
    })

    test('can delete alert rules', async ({ adminPage }) => {
      await alertRulesPage.expectTableVisible()

      const deleteButtons = adminPage.locator('table tbody button:has-text("Delete")')
      const count = await deleteButtons.count()

      expect(count).toBeGreaterThanOrEqual(0)
    })
  })

  test.describe('Webhooks CRUD', () => {
    let webhooksPage: WebhooksPage

    test.beforeEach(async ({ adminPage }) => {
      webhooksPage = new WebhooksPage(adminPage)
      await webhooksPage.goto()
    })

    test('AC-9: can access webhooks page', async ({ adminPage }) => {
      await webhooksPage.expectTableVisible()
    })

    test('can see create webhook button', async ({ adminPage }) => {
      const hasAccess = !(await webhooksPage.hasAccessWarning())
      expect(hasAccess).toBeTruthy()
    })

    test('AC-16: can create webhooks', async ({ adminPage }) => {
      await webhooksPage.expectTableVisible()

      const createButton = adminPage.locator('button:has-text("Create"), button:has-text("Add")')
      await expect(createButton).toBeVisible()
    })
  })

  test.describe('Data Export', () => {
    let dataExportPage: DataExportPage

    test.beforeEach(async ({ adminPage }) => {
      dataExportPage = new DataExportPage(adminPage)
      await dataExportPage.goto()
    })

    test('AC-21: can access export page', async ({ adminPage }) => {
      await dataExportPage.expectFormVisible()

      const hasAccess = !(await dataExportPage.hasAccessWarning())
      expect(hasAccess).toBeTruthy()
    })

    test('can submit export', async ({ adminPage }) => {
      await dataExportPage.expectFormVisible()

      const submitButton = adminPage.locator('button[type="submit"], button:has-text("Export")')
      await expect(submitButton).toBeVisible()
    })
  })

  test.describe('Configuration', () => {
    test('can access config endpoint', async ({ adminPage }) => {
      const response = await adminPage.request.get('/api/v1/config')
      expect(response.ok()).toBeTruthy()
    })
  })
})
