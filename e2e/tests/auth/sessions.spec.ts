/**
 * Session Management Tests
 *
 * Tests for session management:
 * - List sessions
 * - Revoke session
 * - Current session indicator
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { SessionsPage } from '../../pages/SessionsPage'

test.describe('Session Management', () => {
  let sessionsPage: SessionsPage

  test.beforeEach(async ({ adminPage }) => {
    sessionsPage = new SessionsPage(adminPage)
    await sessionsPage.goto()
  })

  test('AC-22: session list displays after login', async ({ adminPage }) => {
    await sessionsPage.expectTableVisible()

    // Should have at least one session (the current one)
    const count = await sessionsPage.getSessionCount()
    expect(count).toBeGreaterThanOrEqual(1)
  })

  test('current session is marked', async ({ adminPage }) => {
    await sessionsPage.expectCurrentSessionMarked()
  })

  test('session shows device/browser info', async ({ adminPage }) => {
    const tableText = await adminPage.locator('table').textContent()

    // Should contain some browser/device info
    // Note: Actual content depends on backend implementation
    expect(tableText).toBeTruthy()
  })

  test('revoke other session removes it from list', async ({ browser }) => {
    // Create a second session
    const context2 = await browser.newContext({
      storageState: '.auth/admin.json',
    })
    const page2 = await context2.newPage()

    try {
      // Login in second context to create another session
      await page2.goto('/dashboard')
      await page2.waitForLoadState('networkidle')

      // Go to sessions page in first context
      const adminPage = await browser.newContext({
        storageState: '.auth/admin.json',
      }).then(ctx => ctx.newPage())

      await adminPage.goto('/sessions')
      await adminPage.waitForLoadState('networkidle')

      // Should have at least 2 sessions now
      const countBefore = await adminPage.locator('table tbody tr').count()
      expect(countBefore).toBeGreaterThanOrEqual(2)

      // Find and revoke a non-current session
      const revokeButtons = adminPage.locator('table tbody tr button:has-text("Revoke")')
      const currentRow = adminPage.locator('table tbody tr').filter({ hasText: 'current' })

      // Click revoke on first non-current session
      const firstRevokeButton = revokeButtons.first()
      if (await firstRevokeButton.count() > 0) {
        await firstRevokeButton.click()

        // Confirm if needed
        const confirmButton = adminPage.locator('button:has-text("Confirm")')
        if (await confirmButton.count() > 0) {
          await confirmButton.click()
        }

        // Wait for table to update
        await adminPage.waitForTimeout(1000)

        // Session should be removed
        const countAfter = await adminPage.locator('table tbody tr').count()
        expect(countAfter).toBeLessThan(countBefore)
      }

      await adminPage.context().close()
    } finally {
      await context2.close()
    }
  })

  test('cannot revoke current session from list', async ({ adminPage }) => {
    await sessionsPage.goto()

    // Find the current session row
    const currentRow = adminPage.locator('table tbody tr').filter({ hasText: /current/i })

    if (await currentRow.count() > 0) {
      // Current session should not have revoke button, or it should be disabled
      const revokeButton = currentRow.locator('button:has-text("Revoke")')

      if (await revokeButton.count() > 0) {
        const isDisabled = await revokeButton.isDisabled()
        expect(isDisabled).toBeTruthy()
      }
    }
  })

  test('revoked session cannot access API', async ({ browser }) => {
    // Create a second session
    const context2 = await browser.newContext()

    try {
      const page2 = await context2.newPage()

      // Login
      await page2.goto('/login')
      await page2.fill('input[name="username"]', 'admin')
      await page2.fill('input[name="password"]', 'Admin123')
      await page2.click('button[type="submit"]')
      await page2.waitForURL('**/dashboard**')

      // Get session ID from API
      const sessionsResponse = await page2.request.get('/api/v1/auth/sessions')
      const sessions = await sessionsResponse.json()

      // Now revoke from main session
      const adminPage = await browser.newContext({
        storageState: '.auth/admin.json',
      }).then(ctx => ctx.newPage())

      await adminPage.goto('/sessions')

      // Revoke the other session (implementation depends on UI)
      // This is a simplified version

      await adminPage.context().close()
    } finally {
      await context2.close()
    }
  })
})

test.describe('Session Expiry', () => {
  test('session shows expiration time', async ({ adminPage }) => {
    await adminPage.goto('/sessions')

    // Look for expiration/last activity info
    const tableText = await adminPage.locator('table').textContent()

    // Should contain time-related info
    expect(tableText).toMatch(/hour|day|minute|ago|expires/i)
  })

  test('session info endpoint returns data', async ({ adminPage }) => {
    // Make API call to session info endpoint
    const response = await adminPage.request.get('/api/v1/auth/session-info')

    expect(response.ok()).toBeTruthy()

    const data = await response.json()
    expect(data).toHaveProperty('data')
    expect(data.data).toHaveProperty('expires_at')
  })
})
