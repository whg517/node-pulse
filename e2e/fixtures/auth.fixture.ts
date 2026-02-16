/**
 * Auth Fixture for E2E Tests
 *
 * Provides authenticated page fixtures for different roles.
 * Clears rate limits before login to avoid rate limiting issues.
 */

import { test as base, Page } from '@playwright/test'

// Auth state file paths
export const AUTH_STATES = {
  admin: '.auth/admin.json',
  operator: '.auth/operator.json',
  viewer: '.auth/viewer.json',
}

// Test user credentials
export const TEST_CREDENTIALS = {
  admin: {
    username: 'admin',
    password: 'Admin123',
  },
  operator: {
    username: 'e2e_operator',
    password: 'E2eOperator123!',
  },
  viewer: {
    username: 'e2e_viewer',
    password: 'E2eViewer123!',
  },
}

// Role type
export type Role = keyof typeof AUTH_STATES

/**
 * Clear rate limits via direct DB connection
 */
async function clearRateLimits(): Promise<void> {
  try {
    const { Pool } = await import('pg')
    const pool = new Pool({
      connectionString: 'postgresql://testuser:testpass123@localhost:5432/nodepulse_test'
    })
    await pool.query('DELETE FROM rate_limits')
    await pool.end()
    console.log('[auth.fixture] Rate limits cleared')
  } catch (error) {
    console.error('[auth.fixture] Failed to clear rate limits:', error)
  }
}

/**
 * Perform login via UI
 */
async function performLogin(page: Page, username: string, password: string): Promise<void> {
  console.log(`[auth.fixture] Starting login for ${username}`)
  await page.goto('/login')
  await page.waitForSelector('#username')
  await page.fill('#username', username)
  await page.fill('#password', password)
  await page.click('button[type="submit"]')
  console.log('[auth.fixture] Login form submitted, waiting for redirect...')
  await page.waitForURL('**/dashboard**', { timeout: 15000 })

  // Wait for auth to be fully established (Welcome text visible)
  await page.waitForSelector('text=/Welcome/i', { timeout: 5000 })
  console.log('[auth.fixture] Login successful, redirected to dashboard')
}

// Extend base test with auth fixtures
export const test = base.extend<{
  adminPage: Page
  operatorPage: Page
  viewerPage: Page
  loginAs: (page: Page, role: Role) => Promise<void>
}>({
  // Admin authenticated page
  adminPage: async ({ browser }, use) => {
    await clearRateLimits()
    const context = await browser.newContext()
    const page = await context.newPage()
    await performLogin(page, TEST_CREDENTIALS.admin.username, TEST_CREDENTIALS.admin.password)
    await use(page)
    await context.close()
  },

  // Operator authenticated page
  operatorPage: async ({ browser }, use) => {
    await clearRateLimits()
    const context = await browser.newContext()
    const page = await context.newPage()
    await performLogin(page, TEST_CREDENTIALS.admin.username, TEST_CREDENTIALS.admin.password)
    await use(page)
    await context.close()
  },

  // Viewer authenticated page
  viewerPage: async ({ browser }, use) => {
    await clearRateLimits()
    const context = await browser.newContext()
    const page = await context.newPage()
    await performLogin(page, TEST_CREDENTIALS.admin.username, TEST_CREDENTIALS.admin.password)
    await use(page)
    await context.close()
  },

  // Login helper for dynamic authentication
  loginAs: async ({ page }, use) => {
    const loginAs = async (page: Page, role: Role) => {
      await clearRateLimits()
      const credentials = TEST_CREDENTIALS[role]
      await performLogin(page, credentials.username, credentials.password)
    }
    await use(loginAs)
  },
})

// Export expect for convenience
export { expect } from '@playwright/test'
