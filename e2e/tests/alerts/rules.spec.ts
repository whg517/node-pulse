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
    const createButton = adminPage.locator('button:has-text("Create"), button:has-text("Add")')
    await expect(createButton).toBeVisible()
  })

  test('table has expected columns', async ({ adminPage }) => {
    await alertRulesPage.expectTableVisible()

    const headerText = await adminPage.locator('table thead').textContent()

    expect(headerText).toMatch(/name|rule/i)
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

    const ruleName = `e2e_test_rule_${Date.now()}`
    await alertRulesPage.createRule(ruleName, 'latency', 100, 'warning')

    // Verify rule appears in list
    const tableText = await adminPage.locator('table').textContent()
    expect(tableText).toContain(ruleName)
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
    // Create a rule to delete
    const ruleName = `e2e_delete_rule_${Date.now()}`
    await alertRulesPage.createRule(ruleName, 'latency', 100, 'warning')

    // Find and delete
    const rows = adminPage.locator('table tbody tr')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const rowText = await rows.nth(i).textContent()
      if (rowText?.includes(ruleName)) {
        const deleteButton = rows.nth(i).locator('button:has-text("Delete")')
        await deleteButton.click()

        const confirmButton = adminPage.locator('button:has-text("Confirm")')
        await confirmButton.click()

        await adminPage.waitForTimeout(500)
        break
      }
    }

    // Verify rule is gone
    const tableText = await adminPage.locator('table').textContent()
    expect(tableText).not.toContain(ruleName)
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
