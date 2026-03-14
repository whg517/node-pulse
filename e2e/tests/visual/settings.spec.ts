/**
 * Settings Pages Visual Regression Tests
 *
 * Visual tests for Settings pages:
 * - Preferences
 * - Users (admin only)
 * - Sessions
 * - System Health
 */
import { test, expect } from '../../fixtures/auth.fixture'
import { PreferencesPage, UsersPage, SessionsPage, SystemHealthPage } from '../../pages'

test.describe('Settings Visual Tests', () => {
  test('preferences page default view', async ({ adminPage }) => {
    const preferencesPage = new PreferencesPage(adminPage)
    await preferencesPage.goto()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('settings-preferences-default.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('preferences page language selector', async ({ adminPage }) => {
    const preferencesPage = new PreferencesPage(adminPage)
    await preferencesPage.goto()

    const languageToggle = adminPage.locator(
      '[data-testid="language-select"], select[name="language"], select[name="lang"], button:has-text("English"), button:has-text("简体中文")'
    ).first()
    if (await languageToggle.isVisible().catch(() => false)) {
      await languageToggle.click()
      await adminPage.waitForTimeout(500)
    }
    
    await expect(adminPage).toHaveScreenshot('settings-preferences-language.png', {
      maxDiffPixels: 50,
    })
  })

  test('sessions page default view', async ({ adminPage }) => {
    const sessionsPage = new SessionsPage(adminPage)
    await sessionsPage.goto()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('settings-sessions-default.png', {
      maxDiffPixels: 12000,
      fullPage: true,
    })
  })

  test('sessions page current session marked', async ({ adminPage }) => {
    const sessionsPage = new SessionsPage(adminPage)
    await sessionsPage.goto()
    await sessionsPage.expectTableVisible()
    
    await expect(adminPage).toHaveScreenshot('settings-sessions-current.png', {
      maxDiffPixels: 12000,
    })
  })

  test('users page default view (admin)', async ({ adminPage }) => {
    const usersPage = new UsersPage(adminPage)
    await usersPage.goto()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('settings-users-default.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('users page create dialog', async ({ adminPage }) => {
    const usersPage = new UsersPage(adminPage)
    await usersPage.goto()
    const createButton = adminPage.locator(
      '[data-testid="create-user-button"], button:has-text("Create"), button:has-text("Add User"), button:has-text("Add User")'
    ).first()
    if (await createButton.isVisible().catch(() => false)) {
      await createButton.click()
      await adminPage.waitForTimeout(500)
    }
    
    await expect(adminPage).toHaveScreenshot('settings-users-create-dialog.png', {
      maxDiffPixels: 100,
    })
  })

  test('system health page default view', async ({ adminPage }) => {
    const healthPage = new SystemHealthPage(adminPage)
    await healthPage.goto()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('settings-system-health-default.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('system health page metrics', async ({ adminPage }) => {
    const healthPage = new SystemHealthPage(adminPage)
    await healthPage.goto()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('settings-system-health-metrics.png', {
      maxDiffPixels: 100,
    })
  })
})
