/**
 * Auth Fixture for E2E Tests
 *
 * Provides authenticated page fixtures for different roles.
 * Uses pre-saved storage states from globalSetup for efficiency.
 * Falls back to fresh login if storage state doesn't exist.
 */

import { test as base, Page, BrowserContext } from '@playwright/test'
import * as fs from 'fs'
import * as path from 'path'

// Test database connection (consistent with global-setup.ts)
const TEST_DB_URL = process.env.TEST_DB_URL || 'postgresql://testuser:testpass123@localhost:5432/nodepulse_test'

// Worker-specific auth state directory
const WORKER_INDEX = process.env.TEST_WORKER_INDEX || '0'
const AUTH_DIR = `.auth/worker-${WORKER_INDEX}`

// Auth state file paths (worker-isolated)
export const AUTH_STATES = {
  admin: `${AUTH_DIR}/admin.json`,
  operator: `${AUTH_DIR}/operator.json`,
  viewer: `${AUTH_DIR}/viewer.json`,
}

// Test user credentials
export const TEST_CREDENTIALS = {
  admin: {
    username: process.env.TEST_ADMIN_USER || 'admin',
    password: process.env.TEST_ADMIN_PASS || 'Admin123',
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

// Cache for verified auth contexts (prevents repeated verification)
const verifiedContexts = new Set<string>()

/**
 * Clear rate limits via direct DB connection with retry
 */
async function clearRateLimits(): Promise<void> {
  const maxRetries = 3
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      const { Pool } = await import('pg')
      const pool = new Pool({ connectionString: TEST_DB_URL })
      await pool.query('DELETE FROM rate_limits')
      await pool.query(
        `UPDATE users
         SET failed_login_attempts = 0, locked_until = NULL
         WHERE username = ANY($1::text[])`,
        [[
          TEST_CREDENTIALS.admin.username,
          TEST_CREDENTIALS.operator.username,
          TEST_CREDENTIALS.viewer.username,
        ]]
      )
      await pool.end()
      console.log('[auth.fixture] Rate limits cleared')
      // Small delay to ensure DB propagation
      await new Promise(resolve => setTimeout(resolve, 100))
      return
    } catch (error) {
      const pgError = error as { code?: string }
      if (pgError.code === '42P01') {
        console.warn('[auth.fixture] Skipping rate limit cleanup: schema not found for TEST_DB_URL')
        return
      }
      console.error(`[auth.fixture] Failed to clear rate limits (attempt ${attempt}):`, error)
      if (attempt < maxRetries) {
        await new Promise(resolve => setTimeout(resolve, 500))
      }
    }
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
 * Copy global auth state to worker-specific location
 */
function copyGlobalAuthState(role: Role): boolean {
  const globalStatePath = `.auth/${role}.json`
  const workerStatePath = AUTH_STATES[role]

  // Create worker auth directory if needed
  const authDir = path.dirname(workerStatePath)
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true })
  }

  // Copy from global state if it exists and worker state doesn't
  if (fs.existsSync(globalStatePath) && !fs.existsSync(workerStatePath)) {
    try {
      fs.copyFileSync(globalStatePath, workerStatePath)
      console.log(`[auth.fixture] Copied global auth state for ${role} to worker ${WORKER_INDEX}`)
      return true
    } catch (error) {
      console.error(`[auth.fixture] Failed to copy auth state:`, error)
    }
  }
  return false
}

/**
 * Perform login via UI with retry logic
 */
async function performLogin(page: Page, username: string, password: string, maxRetries = 3): Promise<void> {
  let lastError: Error | null = null

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      console.log(`[auth.fixture] Starting login for ${username} (attempt ${attempt})`)

      // Clear rate limits before login attempt
      if (attempt > 1) {
        await clearRateLimits()
        await new Promise(resolve => setTimeout(resolve, 500))
      }

      await page.goto('/login', { waitUntil: 'networkidle' })
      await page.waitForSelector('#username', { timeout: 15000 })
      await page.fill('#username', username)
      await page.fill('#password', password)
      await page.click('button[type="submit"]')
      console.log('[auth.fixture] Login form submitted, waiting for redirect...')

      // Wait for successful login (URL first, then SPA selector fallback)
      try {
        await page.waitForURL('**/dashboard**', { timeout: 12000 })
      } catch {
        if (page.url().includes('/dashboard')) {
          // URL already updated but waitForURL missed SPA transition
        } else {
          await page.waitForFunction(
            () => !window.location.pathname.includes('/login'),
            undefined,
            { timeout: 12000 }
          )
        }
      }

      // Wait for page to be interactive (SPA hydration)
      await page.waitForLoadState('domcontentloaded')
      await page.waitForTimeout(1000)

      console.log('[auth.fixture] Login successful, redirected to dashboard')
      return
    } catch (error) {
      lastError = error as Error
      console.error(`[auth.fixture] Login attempt ${attempt} failed:`, error)

      if (attempt < maxRetries) {
        // Wait before retry
        await new Promise(resolve => setTimeout(resolve, 1000))
        // Try to navigate away and back
        try {
          await page.goto('/', { timeout: 5000 })
        } catch {
          // Ignore navigation errors
        }
      }
    }
  }

  throw lastError || new Error(`Login failed after ${maxRetries} attempts`)
}

/**
 * Verify auth state is still valid
 */
async function verifyAuthState(page: Page, role: Role): Promise<boolean> {
  // Check cache first
  const cacheKey = `${WORKER_INDEX}-${role}`
  if (verifiedContexts.has(cacheKey)) {
    return true
  }

  try {
    await page.goto('/dashboard', { waitUntil: 'networkidle', timeout: 25000 })

    // Wait for SPA to settle
    await page.waitForTimeout(1000)

    // If still on dashboard (not redirected to login), state is valid
    const currentUrl = page.url()
    if (!currentUrl.includes('login')) {
      verifiedContexts.add(cacheKey)
      return true
    }
  } catch {
    // Verification failed
  }

  return false
}

/**
 * Create authenticated context using saved storage state or fresh login
 */
async function createAuthenticatedContext(
  browser: import('@playwright/test').Browser,
  role: Role
): Promise<{ context: BrowserContext; page: Page }> {
  const credentials = TEST_CREDENTIALS[role]
  const context = await browser.newContext()
  const page = await context.newPage()
  await clearRateLimits()
  await performLogin(page, credentials.username, credentials.password)

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
  adminPage: [async ({ browser }, use) => {
    const { context, page } = await createAuthenticatedContext(browser, 'admin')
    await use(page)
    await context.close()
  }, { scope: 'worker' }],

  // Operator authenticated page
  operatorPage: [async ({ browser }, use) => {
    const { context, page } = await createAuthenticatedContext(browser, 'operator')
    await use(page)
    await context.close()
  }, { scope: 'worker' }],

  // Viewer authenticated page
  viewerPage: [async ({ browser }, use) => {
    const { context, page } = await createAuthenticatedContext(browser, 'viewer')
    await use(page)
    await context.close()
  }, { scope: 'worker' }],

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
