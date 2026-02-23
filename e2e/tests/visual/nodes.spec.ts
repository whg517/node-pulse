/**
 * Nodes Page Visual Regression Tests
 *
 * Visual tests for Nodes management page:
 * - Default view (empty state)
 * - With nodes data
 * - Create node dialog
 * - Delete confirmation
 */
import { test, expect } from '../../fixtures/auth.fixture'
import { NodesPage } from '../../pages'

test.describe('Nodes Visual Tests', () => {
  let nodesPage: NodesPage

  test.beforeEach(async ({ adminPage }) => {
    nodesPage = new NodesPage(adminPage)
    await nodesPage.goto()
  })

  test('nodes page default view', async ({ adminPage }) => {
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('nodes-default.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('nodes page empty state', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()
    
    // Check if empty state is shown
    if (await nodesPage.isEmptyStateVisible()) {
      await expect(adminPage).toHaveScreenshot('nodes-empty-state.png', {
        maxDiffPixels: 50,
      })
    }
  })

  test('nodes page create dialog', async ({ adminPage }) => {
    await nodesPage.expectCreateButtonVisible()
    await nodesPage.clickCreate()
    await nodesPage.waitForModalOpen()
    
    await expect(adminPage).toHaveScreenshot('nodes-create-dialog.png', {
      maxDiffPixels: 100,
    })
  })

  test('nodes page with validation errors', async ({ adminPage }) => {
    await nodesPage.clickCreate()
    await nodesPage.waitForModalOpen()
    
    // Try to submit empty form
    await nodesPage.submit()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('nodes-validation-errors.png', {
      maxDiffPixels: 100,
    })
  })

  test('nodes page table view', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()
    
    // Only capture if there's data
    if (await nodesPage.hasData()) {
      await expect(adminPage).toHaveScreenshot('nodes-table-view.png', {
        maxDiffPixels: 100,
      })
    }
  })

  test('nodes page delete confirmation', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()
    
    if (await nodesPage.hasData()) {
      await nodesPage.clickDelete(0)
      await adminPage.waitForTimeout(500)
      
      await expect(adminPage).toHaveScreenshot('nodes-delete-confirmation.png', {
        maxDiffPixels: 100,
      })
    }
  })
})
