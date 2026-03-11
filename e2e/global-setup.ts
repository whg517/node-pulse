/**
 * Global Setup for E2E Tests
 *
 * This runs once before all tests:
 * 1. Waits for backend to be healthy
 * 2. Seeds test nodes
 * 3. Seeds test alerts
 * 4. Authenticates each role and saves storage state
 */

import { Pool } from 'pg'
import { chromium, FullConfig } from '@playwright/test'
import * as fs from 'fs'
import bcrypt from 'bcryptjs'

// Test database connection
const TEST_DB_URL = process.env.TEST_DB_URL || 'postgresql://testuser:testpass123@localhost:5432/nodepulse_test'
const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:6532'
const FRONTEND_BASE_URL = process.env.FRONTEND_BASE_URL || 'http://localhost:5173'

// Test user credentials - use environment variables for security
// Falls back to defaults only for local development
const TEST_USERS = {
  admin: {
    username: process.env.TEST_ADMIN_USER || 'admin',
    password: process.env.TEST_ADMIN_PASS || 'Admin123',
    role: 'admin' as const,
  },
}

/**
 * Wait for backend to be healthy
 */
async function waitForBackend(maxAttempts = 30, intervalMs = 1000): Promise<void> {
  console.log('[Global Setup] Waiting for backend to be healthy...')

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/health`)
      if (response.ok) {
        console.log('[Global Setup] Backend is healthy!')
        return
      }
    } catch {
      // Backend not ready yet
    }

    if (attempt < maxAttempts) {
      await new Promise(resolve => setTimeout(resolve, intervalMs))
    }
  }

  throw new Error('[Global Setup] Backend did not become healthy in time')
}

/**
 * Seed test nodes
 */
async function seedTestNodes(pool: Pool): Promise<void> {
  console.log('[Global Setup] Seeding test nodes...')

  // Delete existing test nodes first
  await pool.query(`
    DELETE FROM nodes WHERE name LIKE 'e2e_test_%'
  `)

  // Insert test nodes (using actual schema: id, name, ip, region, status)
  await pool.query(`
    INSERT INTO nodes (id, name, ip, region, status, created_at, updated_at)
    VALUES
      (gen_random_uuid(), 'e2e_test_node_1', '192.168.1.101', 'us-east-1', 'online', NOW(), NOW()),
      (gen_random_uuid(), 'e2e_test_node_2', '192.168.1.102', 'us-west-2', 'online', NOW(), NOW()),
      (gen_random_uuid(), 'e2e_test_node_3', '192.168.1.103', 'eu-west-1', 'offline', NOW(), NOW())
    ON CONFLICT DO NOTHING
  `)
}

/**
 * Clear rate limits to allow test logins
 */
async function clearRateLimits(pool: Pool): Promise<void> {
  console.log('[Global Setup] Clearing rate limits...')
  await pool.query(`DELETE FROM rate_limits`)
}

/**
 * Seed test users for different roles
 */
async function seedTestUsers(pool: Pool): Promise<void> {
  console.log('[Global Setup] Seeding test users...')

  const testUsers = [
    { username: 'e2e_operator', password: 'E2eOperator123!', role: 'operator' },
    { username: 'e2e_viewer', password: 'E2eViewer123!', role: 'viewer' },
  ]

  for (const user of testUsers) {
    // Check if user exists
    const existingUser = await pool.query(
      'SELECT user_id FROM users WHERE username = $1',
      [user.username]
    )

    if (existingUser.rows.length === 0) {
      // Hash password with bcrypt (cost factor 12 to match backend)
      const passwordHash = await bcrypt.hash(user.password, 12)

      // Insert user
      await pool.query(`
        INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
        ON CONFLICT (username) DO NOTHING
      `, [user.username, passwordHash, user.role])

      console.log(`[Global Setup] Created test user: ${user.username} (${user.role})`)
    } else {
      console.log(`[Global Setup] Test user already exists: ${user.username}`)
    }
  }
}

/**
 * Seed test alert rules
 */
async function seedTestAlertRules(pool: Pool): Promise<void> {
  console.log('[Global Setup] Seeding test alert rules...')

  // Delete existing test alert rules
  await pool.query(`
    DELETE FROM alerts WHERE id IN (
      SELECT id FROM alerts WHERE node_id IN (
        SELECT id FROM nodes WHERE name LIKE 'e2e_test_%'
      )
    )
  `)

  // Get a node ID for the alert rules
  const nodeResult = await pool.query(`
    SELECT id FROM nodes WHERE name = 'e2e_test_node_1' LIMIT 1
  `)

  if (nodeResult.rows.length > 0) {
    const nodeId = nodeResult.rows[0].id

    // Insert alert rules (using actual schema: metric, threshold, level)
    await pool.query(`
      INSERT INTO alerts (id, metric, threshold, level, node_id, enabled, created_at)
      VALUES
        (gen_random_uuid(), 'latency', 100, 'P2', $1, true, NOW()),
        (gen_random_uuid(), 'packet_loss_rate', 5, 'P1', $1, true, NOW())
      ON CONFLICT DO NOTHING
    `, [nodeId])
  }
}

/**
 * Authenticate user and save storage state with retry logic
 */
async function authenticateAndSaveState(
  username: string,
  password: string,
  statePath: string,
  maxRetries = 3
): Promise<void> {
  console.log(`[Global Setup] Authenticating ${username}...`)

  let lastError: Error | null = null

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    const browser = await chromium.launch()
    const context = await browser.newContext()
    const page = await context.newPage()

    try {
      console.log(`[Global Setup] Login attempt ${attempt} for ${username}`)

      // Clear rate limits before login attempt (except first try)
      if (attempt > 1) {
        const pool = new Pool({ connectionString: TEST_DB_URL })
        await clearRateLimits(pool)
        await pool.end()
        await new Promise(resolve => setTimeout(resolve, 500))
      }

      // Navigate to login page - use domcontentloaded to avoid timeout
      await page.goto(`${FRONTEND_BASE_URL}/login`, { waitUntil: 'domcontentloaded' })

      // Wait for form elements to be visible
      await page.waitForSelector('#username', { state: 'visible', timeout: 10000 })
      await page.waitForSelector('#password', { state: 'visible', timeout: 5000 })
      // Wait for submit button to be enabled (not disabled by isLoading state)
      await page.waitForSelector('button[type="submit"]:not([disabled])', { state: 'visible', timeout: 10000 })

      // Fill in credentials
      await page.fill('#username', username)
      await page.fill('#password', password)

      // Small delay to ensure form state is updated
      await page.waitForTimeout(100)

      // Submit form
      await page.click('button[type="submit"]')
      console.log(`[Global Setup] Login form submitted for ${username}, waiting for redirect...`)

      // Wait for successful login - use Promise.race for reliability
      await Promise.race([
        page.waitForURL('**/dashboard**', { timeout: 25000, waitUntil: 'domcontentloaded' }),
        page.waitForSelector('nav, [data-testid="sidebar"], .sidebar', { timeout: 25000 }),
      ])

      // Save storage state
      await context.storageState({ path: statePath })
      console.log(`[Global Setup] Saved auth state for ${username} to ${statePath}`)

      await browser.close()
      return // Success!
    } catch (error) {
      lastError = error as Error
      console.error(`[Global Setup] Login attempt ${attempt} failed for ${username}:`, error)

      // Take screenshot for debugging
      try {
        await page.screenshot({ path: `.auth/debug-${username}-attempt-${attempt}.png`, fullPage: true })
        console.log(`[Global Setup] Current URL: ${page.url()}`)
      } catch {
        // Ignore screenshot errors
      }

      await browser.close()

      if (attempt < maxRetries) {
        console.log(`[Global Setup] Retrying login for ${username}...`)
        await new Promise(resolve => setTimeout(resolve, 1000))
      }
    }
  }

  throw lastError || new Error(`[Global Setup] Failed to authenticate ${username} after ${maxRetries} attempts`)
}

/**
 * Main global setup function
 */
export default async function globalSetup(_config: FullConfig) {
  console.log('[Global Setup] Starting e2e test setup...')

  // Create .auth directory for storage states
  const authDir = '.auth'
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true })
  }

  // Wait for backend to be healthy
  await waitForBackend()

  // Connect to test database
  const pool = new Pool({ connectionString: TEST_DB_URL })

  try {
    // Clear rate limits first to allow logins
    await clearRateLimits(pool)

    // Seed test users (operator, viewer) for RBAC tests
    await seedTestUsers(pool)

    // Seed test data
    await seedTestNodes(pool)
    await seedTestAlertRules(pool)

    // Authenticate and save state for all roles
    try {
      await authenticateAndSaveState(
        TEST_USERS.admin.username,
        TEST_USERS.admin.password,
        '.auth/admin.json'
      )

      await authenticateAndSaveState(
        'e2e_operator',
        'E2eOperator123!',
        '.auth/operator.json'
      )

      await authenticateAndSaveState(
        'e2e_viewer',
        'E2eViewer123!',
        '.auth/viewer.json'
      )

      console.log('[Global Setup] Setup complete!')
    } catch (error) {
      console.error('[Global Setup] Authentication failed:', error)
      throw error
    }
  } finally {
    await pool.end()
  }
}
