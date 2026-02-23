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
    
    await preferencesPage.languageSelect.click()
    await adminPage.waitForTimeout(500)
    
    await expect(adminPage).toHaveScreenshot('settings-preferences-language.png', {
      maxDiffPixels: 50,
    })
  })

  test('sessions page default view', async ({ adminPage }) => {
    const sessionsPage = new SessionsPage(adminPage)
    await sessionsPage.goto()
    await adminPage.waitForTimeout(1000)
    
    await expect(adminPage).toHaveScreenshot('settings-sessions-default.png', {
      maxDiffPixels: 100,
      fullPage: true,
    })
  })

  test('sessions page current session marked', async ({ adminPage }) => {
    const sessionsPage = new SessionsPage(adminPage)
    await sessionsPage.goto()
    await sessionsPage.expectTableVisible()
    
    await expect(adminPage).toHaveScreenshot('settings-sessions-current.png', {
      maxDiffPixels: 100,
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
    await usersPage.expectCreateButtonVisible()
    await usersPage.clickCreate()
    await usersPage.waitForModalOpen()
    
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
    await healthPage.waitForHealthData()
    
    await expect(adminPage).toHaveScreenshot('settings-system-health-metrics.png', {
      maxDiffPixels: 100,
    })
  })
})
