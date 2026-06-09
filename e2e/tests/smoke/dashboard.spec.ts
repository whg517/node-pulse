import { test, expect } from '@playwright/test'

const ADMIN_USER = process.env.E2E_ADMIN_USER || 'admin'
const ADMIN_PASS = process.env.E2E_ADMIN_PASS || 'Admin123'

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.locator('input[type="text"], input[type="email"]').first().fill(ADMIN_USER)
    await page.locator('input[type="password"]').fill(ADMIN_PASS)
    await page.locator('button[type="submit"]').click()
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 10_000 })
  })

  test('dashboard page loads and shows sidebar', async ({ page }) => {
    await expect(page.locator('aside, nav').first()).toBeVisible()
  })

  test('dashboard shows NodePulse branding', async ({ page }) => {
    await expect(page.getByText('NodePulse').first()).toBeVisible()
  })

  test('metric cards or loading state is visible', async ({ page }) => {
    // Either metric cards are visible, or a loading spinner is present
    const hasMetricCards = await page.locator('[class*="metric"], [data-testid*="metric"]').count() > 0
    const hasLoading = await page.locator('[role="status"], [class*="loading"], [class*="spin"]').count() > 0
    expect(hasMetricCards || hasLoading).toBe(true)
  })
})
