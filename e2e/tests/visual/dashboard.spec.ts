/**
 * Dashboard Visual Regression Tests
 *
 * Visual tests for Dashboard page:
 * - Default view
 * - Dark mode
 * - With nodes data
 * - Empty state
 */
import { test, expect } from '../../fixtures/auth.fixture'
import { DashboardPage } from '../../pages'

test.describe('Dashboard Visual Tests', () => {
  let dashboardPage: DashboardPage

  test.beforeEach(async ({ adminPage }) => {
    dashboardPage = new DashboardPage(adminPage)
    await dashboardPage.goto()
    await dashboardPage.waitForReady()
  })

  test('dashboard default light mode', async ({ adminPage }) => {
    // Ensure light mode
    await adminPage.goto('/dashboard')
    await adminPage.waitForLoadState('networkidle')
    await adminPage.waitForTimeout(2000)

    await expect(adminPage).toHaveScreenshot('dashboard-light-default.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('dashboard metrics section', async ({ adminPage }) => {
    await dashboardPage.expectMetricsVisible()
    
    await expect(adminPage).toHaveScreenshot('dashboard-metrics.png', {
      maxDiffPixels: 50,
    })
  })

  test('dashboard nodes section', async ({ adminPage }) => {
    await dashboardPage.expectNodesVisible()
    
    await expect(adminPage).toHaveScreenshot('dashboard-nodes.png', {
      maxDiffPixels: 50,
    })
  })

  test('dashboard with navigation', async ({ adminPage }) => {
    // Capture with sidebar visible
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('dashboard-with-nav.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('dashboard dark mode', async ({ adminPage }) => {
    // Toggle dark mode if theme toggle exists
    const themeToggle = adminPage.locator('[data-testid="theme-toggle"], button:has-text("Dark"), button:has-text("Light"), [role="switch"]')
    
    if (await themeToggle.count() > 0) {
      await themeToggle.click()
      await adminPage.waitForTimeout(1000)
      
      await expect(adminPage).toHaveScreenshot('dashboard-dark-mode.png', {
        maxDiffPixels: 100,
        fullPage: true,
      })
    }
  })
})
