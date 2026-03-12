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

// TODO: Remove shard-specific users once backend supports test-mode rate limiting bypass
// Current workaround: rate limiter is IP-scoped (auth_handler.go:114), CI shards share IP
// SECURITY: These users are for E2E testing ONLY. Never use in production.
const TEST_USER_PREFIX = 'e2e_test_'  // Makes it obvious these are test artifacts

/**
 * Get shard-specific username for a given role (runtime access, not module-load time)
 * IMPORTANT: This function MUST be called at runtime, not captured at module scope
 */
const getShardUsername = (role: 'admin' | 'operator' | 'viewer'): string => {
  const shardId = process.env.SHARD_ID
  if (shardId) {
    // Parse shard number from "1/3" format if needed
    const shardNum = shardId.split('/')[0]
    const username = `${TEST_USER_PREFIX}shard_${shardNum}_${role}`
    console.log(`[Global Setup] Using shard ${role}: ${username}`)
    return username
  }
  // Local dev fallback - use shared users
  const fallbacks = {
    admin: process.env.TEST_ADMIN_USER || 'admin',
    operator: 'e2e_operator',
    viewer: 'e2e_viewer',
  }
  return fallbacks[role]
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
 * Seed shard-specific admin user for CI parallel execution
 * SECURITY: These users are for E2E testing ONLY. Never use in production.
 */
async function seedShardUser(pool: Pool, username: string, password: string, role: 'admin' | 'operator' | 'viewer'): Promise<void> {
  console.log(`[Global Setup] Checking shard ${role} user: ${username}...`)

  const existingUser = await pool.query(
    'SELECT user_id FROM users WHERE username = $1',
    [username]
  )

  if (existingUser.rows.length === 0) {
    const passwordHash = await bcrypt.hash(password, 12)
    await pool.query(`
      INSERT INTO users (user_id, username, password_hash, role, created_at, updated_at)
      VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
      ON CONFLICT (username) DO NOTHING
    `, [username, passwordHash, role])
    console.log(`[Global Setup] Created shard ${role} user: ${username}`)
  } else {
    console.log(`[Global Setup] Shard ${role} user already exists: ${username}`)
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

    // Capture browser console messages for debugging
    const consoleMessages: string[] = []
    page.on('console', msg => {
      const text = `[Browser Console] ${msg.type()}: ${msg.text()}`
      consoleMessages.push(text)
      console.log(text)
    })

    // Capture network failures for debugging
    page.on('requestfailed', request => {
      console.log(`[Network] Failed: ${request.method()} ${request.url()} - ${request.failure()?.errorText}`)
    })

    // Capture API responses for debugging
    page.on('response', response => {
      if (response.url().includes('/api/')) {
        console.log(`[Network] ${response.status()} ${response.request().method()} ${response.url()}`)
      }
    })

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

    // Determine credentials based on SHARD_ID
    const adminUsername = getShardUsername('admin')
    const adminPassword = 'Admin123'
    const operatorUsername = getShardUsername('operator')
    const operatorPassword = 'E2eOperator123!'
    const viewerUsername = getShardUsername('viewer')
    const viewerPassword = 'E2eViewer123!'

    // Create shard-specific users if SHARD_ID is set
    if (process.env.SHARD_ID) {
      await seedShardUser(pool, adminUsername, adminPassword, 'admin')
      await seedShardUser(pool, operatorUsername, operatorPassword, 'operator')
      await seedShardUser(pool, viewerUsername, viewerPassword, 'viewer')
    }

    // Authenticate and save state for all roles
    try {
      await authenticateAndSaveState(
        adminUsername,
        adminPassword,
        '.auth/admin.json'
      )

      await authenticateAndSaveState(
        operatorUsername,
        operatorPassword,
        '.auth/operator.json'
      )

      await authenticateAndSaveState(
        viewerUsername,
        viewerPassword,
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
