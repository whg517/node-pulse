/**
 * Auth Fixture for E2E Tests
 *
 * Provides authenticated page fixtures for different roles.
 * Uses pre-saved storage states from globalSetup for efficiency.
 * Falls back to fresh login if storage state doesn't exist.
 */

import { test as base, Page, BrowserContext } from '@playwright/test'
import * as fs from 'fs'

// Test database connection (consistent with global-setup.ts)
const TEST_DB_URL = process.env.TEST_DB_URL || 'postgresql://testuser:testpass123@localhost:5432/nodepulse_test'

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
    const pool = new Pool({ connectionString: TEST_DB_URL })
    await pool.query('DELETE FROM rate_limits')
    await pool.end()
    console.log('[auth.fixture] Rate limits cleared')
  } catch (error) {
    console.error('[auth.fixture] Failed to clear rate limits:', error)
  }
}

/**
 * Check if storage state file exists and is valid
 */
function hasValidStorageState(role: Role): boolean {
  const statePath = AUTH_STATES[role]
  if (!fs.existsSync(statePath)) {
    return false
  }
  try {
    const content = fs.readFileSync(statePath, 'utf-8')
    const state = JSON.parse(content)
    // Check if it has cookies or origins with localStorage
    return (state.cookies && state.cookies.length > 0) ||
           (state.origins && state.origins.length > 0)
  } catch {
    return false
  }
}

/**
 * Perform login via UI (fallback when storage state is not available)
 */
async function performLogin(page: Page, username: string, password: string): Promise<void> {
  console.log(`[auth.fixture] Starting login for ${username}`)
  await page.goto('/login')
  await page.waitForSelector('#username')
  await page.fill('#username', username)
  await page.fill('#password', password)
  await page.click('button[type="submit"]')
  console.log('[auth.fixture] Login form submitted, waiting for redirect...')

  // Wait for successful login - handle both full page load and SPA navigation
  try {
    await page.waitForURL('**/dashboard**', { timeout: 10000 })
  } catch {
    // Fallback for SPA navigation - wait for dashboard content
    await page.waitForSelector('text=/Welcome|Dashboard|Nodes/i', { timeout: 5000 })
  }
  console.log('[auth.fixture] Login successful, redirected to dashboard')
}

/**
 * Create authenticated context using saved storage state or fresh login
 */
async function createAuthenticatedContext(
  browser: import('@playwright/test').Browser,
  role: Role
): Promise<{ context: BrowserContext; page: Page }> {
  const credentials = TEST_CREDENTIALS[role]
  const statePath = AUTH_STATES[role]

  let context: BrowserContext
  let page: Page

  // Try to use saved storage state first
  if (hasValidStorageState(role)) {
    console.log(`[auth.fixture] Using saved storage state for ${role}`)
    context = await browser.newContext({ storageState: statePath })
    page = await context.newPage()

    // Verify the session is still valid by checking a protected route
    try {
      await page.goto('/dashboard')
      // Wait a bit for potential redirect
      await page.waitForLoadState('networkidle')

      // If redirected to login, session is expired
      if (page.url().includes('login')) {
        console.log(`[auth.fixture] Saved state expired for ${role}, performing fresh login`)
        await context.close()
        await clearRateLimits()
        context = await browser.newContext()
        page = await context.newPage()
        await performLogin(page, credentials.username, credentials.password)
      }
    } catch {
      // If verification fails, fall back to fresh login
      console.log(`[auth.fixture] State verification failed for ${role}, performing fresh login`)
      await context.close()
      await clearRateLimits()
      context = await browser.newContext()
      page = await context.newPage()
      await performLogin(page, credentials.username, credentials.password)
    }
  } else {
    // No saved state, perform fresh login
    console.log(`[auth.fixture] No saved state for ${role}, performing fresh login`)
    await clearRateLimits()
    context = await browser.newContext()
    page = await context.newPage()
    await performLogin(page, credentials.username, credentials.password)
  }

  return { context, page }
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
    const { context, page } = await createAuthenticatedContext(browser, 'admin')
    await use(page)
    await context.close()
  },

  // Operator authenticated page
  operatorPage: async ({ browser }, use) => {
    const { context, page } = await createAuthenticatedContext(browser, 'operator')
    await use(page)
    await context.close()
  },

  // Viewer authenticated page
  viewerPage: async ({ browser }, use) => {
    const { context, page } = await createAuthenticatedContext(browser, 'viewer')
    await use(page)
    await context.close()
  },

  // Login helper for dynamic authentication
  loginAs: async ({ page: _page }, use) => {
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
