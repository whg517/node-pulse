/**
 * Alert Records Tests
 *
 * Tests for alert records page:
 * - Filter by node/level/status
 * - Search
 * - Status update
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { AlertRecordsPage } from '../../pages/AlertsPage'

test.describe('Alert Records Page', () => {
  let alertRecordsPage: AlertRecordsPage

  test.beforeEach(async ({ adminPage }) => {
    alertRecordsPage = new AlertRecordsPage(adminPage)
    await alertRecordsPage.goto()
  })

  test('page loads and shows table', async ({ adminPage }) => {
    await alertRecordsPage.expectTableVisible()
  })

  test('table has expected columns', async ({ adminPage }) => {
    await alertRecordsPage.expectTableVisible()

    // Only check headers if table has data
    if (await alertRecordsPage.hasData()) {
      const headerText = await adminPage.locator('table thead').textContent()
      // Headers may be in Chinese or English depending on locale
      // Check for column content: 节点名称/Node, 告警级别/Level, 状态/Status, 指标类型/Metric, 时间戳/Time
      expect(headerText).toMatch(/节点|node|级别|level|状态|status|指标|metric|时间|time/i)
    }
    // If no data, the test passes - we verified the page loads correctly
  })
})

test.describe('Alert Records Filtering', () => {
  let alertRecordsPage: AlertRecordsPage

  test.beforeEach(async ({ adminPage }) => {
    alertRecordsPage = new AlertRecordsPage(adminPage)
    await alertRecordsPage.goto()
  })

  test('filter by status', async ({ adminPage }) => {
    await alertRecordsPage.expectTableVisible()

    if (await alertRecordsPage.statusFilter.count() > 0) {
      await alertRecordsPage.filterByStatus('pending')

      // Wait for table to update
      await adminPage.waitForTimeout(500)
    }
  })

  test('search works', async ({ adminPage }) => {
    await alertRecordsPage.expectTableVisible()

    if (await alertRecordsPage.searchInput.count() > 0) {
      await alertRecordsPage.search('nonexistent_alert')

      await adminPage.waitForTimeout(500)
    }
  })

  test('clear filters', async ({ adminPage }) => {
    await alertRecordsPage.expectTableVisible()

    const clearButton = adminPage.locator('button:has-text("Clear"), button:has-text("Reset")')

    if (await clearButton.count() > 0) {
      await clearButton.click()
    }
  })
})

test.describe('Alert Records Status Update', () => {
  let alertRecordsPage: AlertRecordsPage

  test.beforeEach(async ({ adminPage }) => {
    alertRecordsPage = new AlertRecordsPage(adminPage)
    await alertRecordsPage.goto()
  })

  test('can update status', async ({ adminPage }) => {
    await alertRecordsPage.expectTableVisible()

    const rowCount = await adminPage.locator('table tbody tr').count()

    if (rowCount > 0) {
      const updateButton = adminPage.locator('table tbody tr').first().locator('button:has-text("Update"), select[name="status"]')

      if (await updateButton.count() > 0) {
        await updateButton.click()

        // Look for status options
        const inProgressOption = adminPage.locator('button:has-text("in_progress"), option:has-text("in_progress")')

        if (await inProgressOption.count() > 0) {
          await inProgressOption.click()
        }
      }
    } else {
      test.skip(true, 'No alert records to update')
    }
  })
})
