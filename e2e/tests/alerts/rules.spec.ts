/**
 * Alert Rules Tests
 *
 * Tests for alert rules page:
 * - List rules
 * - Create rule form
 * - Edit rule
 * - Delete rule
 * - Toggle enable/disable
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { AlertRulesPage } from '../../pages/AlertsPage'

test.describe('Alert Rules Page', () => {
  let alertRulesPage: AlertRulesPage

  test.beforeEach(async ({ adminPage }) => {
    alertRulesPage = new AlertRulesPage(adminPage)
    await alertRulesPage.goto()
  })

  test('page loads and shows table', async ({ adminPage }) => {
    await alertRulesPage.expectTableVisible()
  })

  test('shows create button for admin/operator', async ({ adminPage }) => {
    // Wait for page to be ready
    await alertRulesPage.expectTableVisible()

    // Button text is "Create Alert Rule" or similar
    const createButton = adminPage.locator('button:has-text("Create")')

    // Button might be in page header or empty state
    await expect(createButton.first()).toBeVisible({ timeout: 15000 })
  })

  test('table has expected columns', async ({ adminPage }) => {
    await alertRulesPage.expectTableVisible()

    // Check if we have data or empty state
    const hasTable = (await adminPage.locator('table').count()) > 0

    if (hasTable) {
      const headerText = await adminPage.locator('table thead').textContent()
      // Check for actual column headers: Type/Metric, Threshold, Severity, Node, Status
      expect(headerText).toMatch(/type|threshold|severity|node|status/i)
    } else {
      // Empty state is valid - just verify it's showing
      await expect(alertRulesPage.emptyState.first()).toBeVisible()
    }
  })
})

test.describe('Alert Rules CRUD', () => {
  let alertRulesPage: AlertRulesPage

  test.beforeEach(async ({ adminPage }) => {
    alertRulesPage = new AlertRulesPage(adminPage)
    await alertRulesPage.goto()
  })

  test('create alert rule modal opens', async ({ adminPage }) => {
    await alertRulesPage.expectTableVisible()
    await alertRulesPage.clickCreate()

    await expect(alertRulesPage.modal).toBeVisible()
  })

  test('create rule with valid data', async ({ adminPage }) => {
    await alertRulesPage.expectTableVisible()

    // Create a rule with: metric=latency, threshold=100, level=P1
    await alertRulesPage.createRule('latency', 100, 'P1')

    // Verify rule appears in list (check for latency or threshold)
    const tableText = await adminPage.locator('table').textContent()
    expect(tableText).toMatch(/latency|100/i)
  })

  test('edit rule modal opens with data', async ({ adminPage }) => {
    await alertRulesPage.expectTableVisible()

    const rowCount = await adminPage.locator('table tbody tr').count()

    if (rowCount > 0) {
      const editButton = adminPage.locator('table tbody tr').first().locator('button:has-text("Edit")')

      if (await editButton.count() > 0) {
        await editButton.click()
        await expect(alertRulesPage.modal).toBeVisible()
      }
    } else {
      test.skip(true, 'No alert rules to edit')
    }
  })

  test('delete rule removes from list', async ({ adminPage }) => {
    // Create a rule to delete with unique threshold for identification
    const uniqueThreshold = Date.now() % 10000 + 500 // e.g., 5234
    await alertRulesPage.createRule('jitter', uniqueThreshold, 'P2')

    // Wait for table to update
    await adminPage.waitForTimeout(500)

    // Find the row with our unique threshold and delete it
    const rows = adminPage.locator('table tbody tr')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const rowText = await rows.nth(i).textContent()
      if (rowText?.includes(String(uniqueThreshold))) {
        const deleteButton = rows.nth(i).locator('button:has-text("Delete")')
        await deleteButton.click()

        // ConfirmDialog uses "Delete" as confirm button text
        const confirmButton = adminPage.locator('.fixed button:has-text("Delete")').last()
        await confirmButton.click()

        await adminPage.waitForTimeout(500)
        break
      }
    }

    // Verify rule is gone
    const tableText = await adminPage.locator('table').textContent()
    expect(tableText).not.toContain(String(uniqueThreshold))
  })

  test('toggle rule enable/disable', async ({ adminPage }) => {
    await alertRulesPage.expectTableVisible()

    const toggle = adminPage.locator('table tbody tr').first().locator('input[type="checkbox"]')

    if (await toggle.count() > 0) {
      const isCheckedBefore = await toggle.isChecked()
      await toggle.click()

      await adminPage.waitForTimeout(500)

      const isCheckedAfter = await toggle.isChecked()
      expect(isCheckedAfter).toBe(!isCheckedBefore)
    } else {
      test.skip(true, 'No toggle available')
    }
  })
})

test.describe('Alert Rules - Viewer', () => {
  test('viewer cannot see create button', async ({ viewerPage }) => {
    const alertRulesPage = new AlertRulesPage(viewerPage)
    await alertRulesPage.goto()
    await alertRulesPage.expectTableVisible()

    const createButton = viewerPage.locator('button:has-text("Create"), button:has-text("Add")')
    const isVisible = await createButton.isVisible().catch(() => false)

    expect(isVisible).toBeFalsy()
  })
})
