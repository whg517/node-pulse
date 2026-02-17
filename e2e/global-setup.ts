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

// Test user credentials (using default admin for now)
const TEST_USERS = {
  admin: {
    username: 'admin',
    password: 'Admin123',
    role: 'admin',
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
 * Authenticate user and save storage state
 */
async function authenticateAndSaveState(
  username: string,
  password: string,
  statePath: string
): Promise<void> {
  console.log(`[Global Setup] Authenticating ${username}...`)

  const browser = await chromium.launch()
  const context = await browser.newContext()
  const page = await context.newPage()

  try {
    // Navigate to login page
    await page.goto(`${FRONTEND_BASE_URL}/login`)

    // Wait for the form to be ready (using id selectors)
    await page.waitForSelector('#username', { timeout: 10000 })

    // Fill in credentials
    await page.fill('#username', username)
    await page.fill('#password', password)

    // Submit form
    await page.click('button[type="submit"]')

    // Wait for successful login - either dashboard URL or a dashboard element
    // React Router SPA navigation might not trigger URL change detection reliably
    try {
      // First try waiting for URL change
      await page.waitForURL('**/dashboard**', { timeout: 10000 })
    } catch {
      // Fallback: wait for dashboard content to appear (SPA navigation)
      await page.waitForSelector('text=/Welcome|Dashboard|Nodes/i', { timeout: 5000 })
    }

    // Save storage state (includes cookies and localStorage)
    await context.storageState({ path: statePath })

    console.log(`[Global Setup] Saved auth state for ${username} to ${statePath}`)
  } catch (error) {
    console.error(`[Global Setup] Failed to authenticate ${username}:`, error)
    throw error
  } finally {
    await browser.close()
  }
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
