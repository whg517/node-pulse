import { test, expect } from '@playwright/test'

const ADMIN_USER = process.env.E2E_ADMIN_USER || 'admin'
const ADMIN_PASS = process.env.E2E_ADMIN_PASS || 'Admin123'

test.describe('Node List', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.locator('input[type="text"], input[type="email"]').first().fill(ADMIN_USER)
    await page.locator('input[type="password"]').fill(ADMIN_PASS)
    await page.locator('button[type="submit"]').click()
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 10_000 })
  })

  test('navigates to /nodes via sidebar', async ({ page }) => {
    // Click the Nodes nav link in sidebar
    await page.locator('a[href="/nodes"]').first().click()
    await expect(page).toHaveURL(/\/nodes/)
  })

  test('node list page loads without error', async ({ page }) => {
    await page.goto('/nodes')
    // Should not show a generic crash error
    await expect(page.locator('body')).not.toContainText('Something went wrong', { timeout: 5_000 })
    // Page should render within 5 seconds
    await expect(page.locator('aside, nav').first()).toBeVisible()
  })

  test('node list shows table or empty state', async ({ page }) => {
    await page.goto('/nodes')
    // Wait for loading to settle
    await page.waitForTimeout(2000)

    const hasTable = await page.locator('table, [role="table"]').count() > 0
    const hasEmptyState = await page.getByText(/no nodes/i).count() > 0
    const hasLoading = await page.locator('[class*="spin"]').count() > 0
    expect(hasTable || hasEmptyState || hasLoading).toBe(true)
  })
})
