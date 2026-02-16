/**
 * Viewer RBAC Tests
 *
 * Tests for viewer role permissions:
 * - Read-only access to all pages
 * - No CRUD buttons visible
 * - API returns 403 for write operations
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { NodesPage } from '../../pages/NodesPage'
import { WebhooksPage } from '../../pages/WebhooksPage'
import { DataExportPage } from '../../pages/DataExportPage'

test.describe('Viewer RBAC', () => {
  test.describe('Page Access', () => {
    test('can access dashboard', async ({ viewerPage }) => {
      await viewerPage.goto('/dashboard')
      await expect(viewerPage).toHaveURL(/.*dashboard/)
    })

    test('can access nodes page', async ({ viewerPage }) => {
      await viewerPage.goto('/nodes')
      await expect(viewerPage).toHaveURL(/.*nodes/)
    })

    test('can access alerts pages', async ({ viewerPage }) => {
      await viewerPage.goto('/alerts/rules')
      await expect(viewerPage).toHaveURL(/.*alerts\/rules/)

      await viewerPage.goto('/alerts/records')
      await expect(viewerPage).toHaveURL(/.*alerts\/records/)
    })

    test('can access comparison page', async ({ viewerPage }) => {
      await viewerPage.goto('/comparison')
      await expect(viewerPage).toHaveURL(/.*comparison/)
    })

    test('can access performance page', async ({ viewerPage }) => {
      await viewerPage.goto('/performance')
      await expect(viewerPage).toHaveURL(/.*performance/)
    })
  })

  test.describe('Nodes Page - Read Only', () => {
    let nodesPage: NodesPage

    test.beforeEach(async ({ viewerPage }) => {
      nodesPage = new NodesPage(viewerPage)
      await nodesPage.goto()
    })

    test('AC-13: cannot see create node button', async ({ viewerPage }) => {
      await nodesPage.expectTableVisible()

      const createButton = viewerPage.locator('button:has-text("Create"), button:has-text("Add")')
      const isVisible = await createButton.isVisible().catch(() => false)

      expect(isVisible).toBeFalsy()
    })

    test('cannot see edit buttons', async ({ viewerPage }) => {
      await nodesPage.expectTableVisible()

      const editButtons = viewerPage.locator('table tbody button:has-text("Edit")')
      const count = await editButtons.count()

      expect(count).toBe(0)
    })

    test('cannot see delete buttons', async ({ viewerPage }) => {
      await nodesPage.expectTableVisible()

      const deleteButtons = viewerPage.locator('table tbody button:has-text("Delete")')
      const count = await deleteButtons.count()

      expect(count).toBe(0)
    })

    test('can view node list', async ({ viewerPage }) => {
      await nodesPage.expectTableVisible()

      const rowCount = await nodesPage.getRowCount()
      // Table should show nodes (read access)
      expect(rowCount).toBeGreaterThanOrEqual(0)
    })
  })

  test.describe('Alert Rules Page - Read Only', () => {
    test.beforeEach(async ({ viewerPage }) => {
      await viewerPage.goto('/alerts/rules')
    })

    test('cannot see create button', async ({ viewerPage }) => {
      const createButton = viewerPage.locator('button:has-text("Create"), button:has-text("Add Rule")')
      const isVisible = await createButton.isVisible().catch(() => false)

      expect(isVisible).toBeFalsy()
    })

    test('cannot see edit buttons', async ({ viewerPage }) => {
      await viewerPage.waitForSelector('table')

      const editButtons = viewerPage.locator('table tbody button:has-text("Edit")')
      const count = await editButtons.count()

      expect(count).toBe(0)
    })

    test('can view alert rules list', async ({ viewerPage }) => {
      await viewerPage.waitForSelector('table')
      const table = viewerPage.locator('table')
      await expect(table).toBeVisible()
    })
  })

  test.describe('Webhooks Access (Admin Only)', () => {
    let webhooksPage: WebhooksPage

    test.beforeEach(async ({ viewerPage }) => {
      webhooksPage = new WebhooksPage(viewerPage)
      await webhooksPage.goto()
    })

    test('cannot access webhooks - shows warning', async ({ viewerPage }) => {
      const hasWarning = await webhooksPage.hasAccessWarning()
      expect(hasWarning).toBeTruthy()
    })

    test('webhook API returns 403', async ({ viewerPage }) => {
      const response = await viewerPage.request.get('/api/v1/webhooks')
      expect(response.status()).toBe(403)
    })
  })

  test.describe('Data Export (Admin Only)', () => {
    let dataExportPage: DataExportPage

    test.beforeEach(async ({ viewerPage }) => {
      dataExportPage = new DataExportPage(viewerPage)
      await dataExportPage.goto()
    })

    test('cannot access export - shows warning', async ({ viewerPage }) => {
      const hasWarning = await dataExportPage.hasAccessWarning()
      expect(hasWarning).toBeTruthy()
    })

    test('export API returns 403', async ({ viewerPage }) => {
      const response = await viewerPage.request.post('/api/v1/data/export', {
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

  test.describe('API Write Operations - All Return 403', () => {
    test('AC-18: cannot create node via API', async ({ viewerPage }) => {
      const response = await viewerPage.request.post('/api/v1/nodes', {
        data: {
          name: 'test_node',
          region: 'us-east-1',
        },
      })
      expect(response.status()).toBe(403)
    })

    test('cannot update node via API', async ({ viewerPage }) => {
      const response = await viewerPage.request.put('/api/v1/nodes/test-id', {
        data: {
          name: 'updated_node',
        },
      })
      expect(response.status()).toBe(403)
    })

    test('cannot delete node via API', async ({ viewerPage }) => {
      const response = await viewerPage.request.delete('/api/v1/nodes/test-id')
      expect(response.status()).toBe(403)
    })

    test('cannot create alert rule via API', async ({ viewerPage }) => {
      const response = await viewerPage.request.post('/api/v1/alerts/rules', {
        data: {
          name: 'test_rule',
          metric_type: 'latency',
          condition_type: 'greater_than',
          threshold: 100,
          level: 'warning',
        },
      })
      expect(response.status()).toBe(403)
    })

    test('cannot create webhook via API', async ({ viewerPage }) => {
      const response = await viewerPage.request.post('/api/v1/webhooks', {
        data: {
          name: 'test_webhook',
          url: 'https://example.com/webhook',
        },
      })
      expect(response.status()).toBe(403)
    })
  })

  test.describe('Read Operations - All Succeed', () => {
    test('can read nodes', async ({ viewerPage }) => {
      const response = await viewerPage.request.get('/api/v1/nodes')
      expect(response.ok()).toBeTruthy()
    })

    test('can read probes', async ({ viewerPage }) => {
      const response = await viewerPage.request.get('/api/v1/probes')
      expect(response.ok()).toBeTruthy()
    })

    test('can read alert rules', async ({ viewerPage }) => {
      const response = await viewerPage.request.get('/api/v1/alerts/rules')
      expect(response.ok()).toBeTruthy()
    })

    test('can read alert records', async ({ viewerPage }) => {
      const response = await viewerPage.request.get('/api/v1/alerts/records')
      expect(response.ok()).toBeTruthy()
    })

    test('can read metrics', async ({ viewerPage }) => {
      const response = await viewerPage.request.get('/api/v1/data/metrics')
      expect(response.ok()).toBeTruthy()
    })

    test('can read history', async ({ viewerPage }) => {
      const response = await viewerPage.request.get('/api/v1/data/history')
      expect(response.ok()).toBeTruthy()
    })
  })
})
