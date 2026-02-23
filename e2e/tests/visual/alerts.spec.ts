/**
 * Alerts Pages Visual Regression Tests
 *
 * Visual tests for Alerts pages:
 * - Alert Rules
 * - Alert Records
 * - Alert History
 */
import { test, expect } from '../../fixtures/auth.fixture'
import { AlertRulesPage, AlertRecordsPage, AlertHistoryPage } from '../../pages/AlertsPage'

test.describe('Alerts Visual Tests', () => {
  test('alert rules page default view', async ({ adminPage }) => {
    const alertRulesPage = new AlertRulesPage(adminPage)
    await alertRulesPage.goto()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('alerts-rules-default.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('alert rules page create dialog', async ({ adminPage }) => {
    const alertRulesPage = new AlertRulesPage(adminPage)
    await alertRulesPage.goto()
    await alertRulesPage.expectTableVisible()
    
    const createButton = adminPage.locator('button:has-text("Create"), button:has-text("Add Rule")')
    if (await createButton.count() > 0) {
      await createButton.click()
      await adminPage.waitForTimeout(500)
      
      await expect(adminPage).toHaveScreenshot('alerts-rules-create-dialog.png', {
        maxDiffPixels: 100,
      })
    }
  })

  test('alert records page default view', async ({ adminPage }) => {
    const alertRecordsPage = new AlertRecordsPage(adminPage)
    await alertRecordsPage.goto()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('alerts-records-default.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('alert records page with filters', async ({ adminPage }) => {
    const alertRecordsPage = new AlertRecordsPage(adminPage)
    await alertRecordsPage.goto()
    await alertRecordsPage.expectTableVisible()
    
    // Check if filters are visible
    const filterButton = adminPage.locator('button:has-text("Filter")')
    if (await filterButton.count() > 0) {
      await filterButton.click()
      await adminPage.waitForTimeout(500)
      
      await expect(adminPage).toHaveScreenshot('alerts-records-filters.png', {
        maxDiffPixels: 50,
      })
    }
  })

  test('alert history page default view', async ({ adminPage }) => {
    const alertHistoryPage = new AlertHistoryPage(adminPage)
    await alertHistoryPage.goto()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('alerts-history-default.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('alert history page pagination', async ({ adminPage }) => {
    const alertHistoryPage = new AlertHistoryPage(adminPage)
    await alertHistoryPage.goto()
    await alertHistoryPage.expectTableVisible()
    
    // Check if pagination exists
    if (await alertHistoryPage.pagination.count() > 0) {
      await expect(adminPage).toHaveScreenshot('alerts-history-pagination.png', {
        maxDiffPixels: 50,
      })
    }
  })
})
