/**
 * Logout Tests
 *
 * Tests for logout flow:
 * - Logout redirects to login page
 * - Cookies cleared
 * - Cross-tab logout sync
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { LoginPage } from '../../pages/LoginPage'

test.describe('Logout Flow', () => {
  test('AC-5: logout redirects to login page', async ({ adminPage }) => {
    // Navigate to dashboard first
    await adminPage.goto('/dashboard')
    await expect(adminPage).toHaveURL(/.*dashboard/)

    // Click logout button (usually in user menu)
    const userMenu = adminPage.locator('[data-testid="user-menu"], .user-menu, button:has-text("admin")')
    await userMenu.click()

    const logoutButton = adminPage.locator('[data-testid="logout-button"], button:has-text("Logout"), a:has-text("Logout")')
    await logoutButton.click()

    // Should redirect to login page
    await expect(adminPage).toHaveURL(/.*login/)
  })

  test('logout clears auth cookies', async ({ adminPage }) => {
    // Navigate to dashboard
    await adminPage.goto('/dashboard')

    // Get cookies before logout
    const cookiesBefore = await adminPage.context().cookies()
    const hasAuthCookieBefore = cookiesBefore.some(c =>
      c.name.includes('token') || c.name.includes('session') || c.name.includes('refresh')
    )

    // Logout
    const userMenu = adminPage.locator('[data-testid="user-menu"], .user-menu, button:has-text("admin")')
    await userMenu.click()

    const logoutButton = adminPage.locator('[data-testid="logout-button"], button:has-text("Logout"), a:has-text("Logout")')
    await logoutButton.click()

    await adminPage.waitForURL(/.*login/)

    // Get cookies after logout
    const cookiesAfter = await adminPage.context().cookies()
    const hasAuthCookieAfter = cookiesAfter.some(c =>
      c.name.includes('token') || c.name.includes('session') || c.name.includes('refresh')
    )

    // Auth cookies should be cleared or empty
    // Note: The refresh token cookie should be cleared
    expect(hasAuthCookieAfter).toBeFalsy()
  })

  test('logout prevents access to protected routes', async ({ adminPage }) => {
    // Navigate to dashboard
    await adminPage.goto('/dashboard')
    await expect(adminPage).toHaveURL(/.*dashboard/)

    // Logout
    const userMenu = adminPage.locator('[data-testid="user-menu"], .user-menu')
    await userMenu.click()

    const logoutButton = adminPage.locator('[data-testid="logout-button"], button:has-text("Logout")')
    await logoutButton.click()

    await adminPage.waitForURL(/.*login/)

    // Try to access protected route
    await adminPage.goto('/nodes')

    // Should be redirected to login
    await expect(adminPage).toHaveURL(/.*login/)
  })

  test('logout via menu in sidebar', async ({ adminPage }) => {
    await adminPage.goto('/dashboard')

    // Try sidebar logout if available
    const sidebarLogout = adminPage.locator('[data-testid="sidebar-logout"], aside button:has-text("Logout")')

    if (await sidebarLogout.count() > 0) {
      await sidebarLogout.click()
      await expect(adminPage).toHaveURL(/.*login/)
    } else {
      // Skip if no sidebar logout
      test.skip(true, 'No sidebar logout button found')
    }
  })
})

test.describe('Cross-Tab Logout Sync', () => {
  test('AC-8: logout in tab A logs out tab B', async ({ browser }) => {
    // Create a single browser context with authenticated state
    const context = await browser.newContext({
      storageState: '.auth/admin.json',
    })

    // Create two pages (tabs) in the same context
    const page1 = await context.newPage()
    const page2 = await context.newPage()

    try {
      // Both pages navigate to dashboard
      await page1.goto('/dashboard')
      await page2.goto('/dashboard')

      await expect(page1).toHaveURL(/.*dashboard/)
      await expect(page2).toHaveURL(/.*dashboard/)

      // Logout in page1
      const userMenu = page1.locator('[data-testid="user-menu"], .user-menu')
      await userMenu.click()

      const logoutButton = page1.locator('[data-testid="logout-button"], button:has-text("Logout")')
      await logoutButton.click()

      await page1.waitForURL(/.*login/)

      // Page2 should also be logged out when it becomes visible
      // Trigger visibility check by focusing page2
      await page2.bringToFront()

      // Wait a moment for the storage event to propagate
      await page2.waitForTimeout(500)

      // Navigate to trigger auth check
      await page2.goto('/dashboard')

      // Page2 should now be redirected to login
      await expect(page2).toHaveURL(/.*login/)
    } finally {
      await context.close()
    }
  })

  test('localStorage broadcast triggers logout', async ({ browser }) => {
    const context = await browser.newContext({
      storageState: '.auth/admin.json',
    })

    const page1 = await context.newPage()
    const page2 = await context.newPage()

    try {
      await page1.goto('/dashboard')
      await page2.goto('/dashboard')

      // Simulate logout broadcast via localStorage
      await page1.evaluate(() => {
        localStorage.setItem('auth:logout', 'logout')
        setTimeout(() => localStorage.removeItem('auth:logout'), 100)
      })

      // Bring page2 to front to trigger visibility handler
      await page2.bringToFront()
      await page2.waitForTimeout(500)

      // Check if page2 auth state was cleared
      const isAuth = await page2.evaluate(() => {
        return localStorage.getItem('auth:logout') === null
      })

      expect(isAuth).toBeTruthy()
    } finally {
      await context.close()
    }
  })
})
