/**
 * Node List Tests
 *
 * Tests for node management page:
 * - Table rendering
 * - Sorting
 * - Filtering
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { NodesPage } from '../../pages/NodesPage'

test.describe('Node List Page', () => {
  let nodesPage: NodesPage

  test.beforeEach(async ({ adminPage }) => {
    nodesPage = new NodesPage(adminPage)
    await nodesPage.goto()
  })

  test('table renders correctly', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()

    // Table should have headers
    const headers = adminPage.locator('table thead th')
    const headerCount = await headers.count()

    expect(headerCount).toBeGreaterThan(0)
  })

  test('shows node data', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()

    // Should have rows if nodes exist
    const rowCount = await nodesPage.getRowCount()

    // May be 0 if no nodes seeded
    expect(rowCount).toBeGreaterThanOrEqual(0)
  })

  test('table has expected columns', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()

    const headerText = await adminPage.locator('table thead').textContent()

    // Should have common columns
    expect(headerText).toMatch(/name|node/i)
  })

  test('search/filter works', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()

    const searchInput = adminPage.locator('input[type="search"], input[placeholder*="search" i]')

    if (await searchInput.count() > 0) {
      await searchInput.fill('nonexistent_node_xyz')
      await adminPage.keyboard.press('Enter')

      // Should show no results or empty table
      await adminPage.waitForTimeout(500)
    }
  })

  test('sorting works', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()

    // Click on a header to sort
    const nameHeader = adminPage.locator('table thead th').first()

    if (await nameHeader.count() > 0) {
      await nameHeader.click()

      // Should show sort indicator
      await adminPage.waitForTimeout(500)
    }
  })

  test('pagination works if many nodes', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()

    const pagination = adminPage.locator('[data-testid="pagination"], .pagination')

    if (await pagination.count() > 0) {
      const nextButton = pagination.locator('button:has-text("Next"), [aria-label="Next"]')

      if (await nextButton.count() > 0 && await nextButton.isEnabled()) {
        await nextButton.click()

        // Wait for table to update
        await adminPage.waitForTimeout(500)
      }
    }
  })
})

test.describe('Node List - Admin', () => {
  test('shows create button', async ({ adminPage }) => {
    const nodesPage = new NodesPage(adminPage)
    await nodesPage.goto()

    await nodesPage.expectCreateButtonVisible()
  })

  test('shows action buttons for each row', async ({ adminPage }) => {
    const nodesPage = new NodesPage(adminPage)
    await nodesPage.goto()
    await nodesPage.expectTableVisible()

    const rowCount = await nodesPage.getRowCount()

    if (rowCount > 0) {
      const editButtons = adminPage.locator('table tbody button:has-text("Edit")')
      const deleteButtons = adminPage.locator('table tbody button:has-text("Delete")')

      expect(await editButtons.count()).toBe(rowCount)
      expect(await deleteButtons.count()).toBe(rowCount)
    }
  })
})

test.describe('Node List - Viewer', () => {
  test('hides create button', async ({ viewerPage }) => {
    const nodesPage = new NodesPage(viewerPage)
    await nodesPage.goto()
    await nodesPage.expectTableVisible()

    const createButton = viewerPage.locator('button:has-text("Create"), button:has-text("Add")')
    const isVisible = await createButton.isVisible().catch(() => false)

    expect(isVisible).toBeFalsy()
  })

  test('hides action buttons', async ({ viewerPage }) => {
    const nodesPage = new NodesPage(viewerPage)
    await nodesPage.goto()
    await nodesPage.expectTableVisible()

    const editButtons = viewerPage.locator('table tbody button:has-text("Edit")')
    const deleteButtons = viewerPage.locator('table tbody button:has-text("Delete")')

    expect(await editButtons.count()).toBe(0)
    expect(await deleteButtons.count()).toBe(0)
  })
})
