/**
 * Global Teardown for E2E Tests
 *
 * This runs once after all tests:
 * 1. Cleans up test data (nodes, alerts, etc.)
 * 2. Closes database connections
 */

import { Pool } from 'pg'
import * as fs from 'fs'

// Test database connection
const TEST_DB_URL = process.env.TEST_DB_URL || 'postgresql://testuser:testpass123@localhost:5432/nodepulse_test'

/**
 * Clean up test data from database
 */
async function cleanupTestData(pool: Pool): Promise<void> {
  console.log('[Global Teardown] Cleaning up test data...')

  try {
    // Delete auth audit logs first (foreign key constraint from users)
    await pool.query(`
      DELETE FROM auth_audit_logs WHERE user_id IN (
        SELECT user_id FROM users WHERE username IN ('e2e_operator', 'e2e_viewer')
      )
    `)

    // Delete test users (operator, viewer) - created for RBAC tests
    await pool.query(`
      DELETE FROM users WHERE username IN ('e2e_operator', 'e2e_viewer')
    `)

    // TODO: Remove this cleanup when shard-specific user workaround is removed
    // SECURITY: Only cleans up users with the test-only prefix
    // First, delete auth_audit_logs for shard-specific admin users (foreign key constraint)
    await pool.query(`
      DELETE FROM auth_audit_logs WHERE user_id IN (
        SELECT user_id FROM users WHERE username LIKE 'e2e_test_shard_%_admin'
      )
    `)
    // Then delete the shard-specific admin users (matches e2e_test_shard_%_admin pattern)
    const shardUserResult = await pool.query(`
      DELETE FROM users WHERE username LIKE 'e2e_test_shard_%_admin'
    `)
    if (shardUserResult.rowCount && shardUserResult.rowCount > 0) {
      console.log(`[Global Teardown] Cleaned up ${shardUserResult.rowCount} shard-specific test users`)
    }

    // Delete test alert records first (foreign key constraint)
    await pool.query(`
      DELETE FROM alert_records WHERE node_id IN (
        SELECT id FROM nodes WHERE name LIKE 'e2e_test_%'
      )
    `)

    // Delete test alert rules
    await pool.query(`
      DELETE FROM alerts WHERE node_id IN (
        SELECT id FROM nodes WHERE name LIKE 'e2e_test_%'
      )
    `)

    // Delete test probes
    await pool.query(`
      DELETE FROM probes WHERE node_id IN (
        SELECT id FROM nodes WHERE name LIKE 'e2e_test_%'
      )
    `)

    // Delete test metrics
    await pool.query(`
      DELETE FROM metrics WHERE node_id IN (
        SELECT id FROM nodes WHERE name LIKE 'e2e_test_%'
      )
    `)

    // Delete test nodes
    await pool.query(`
      DELETE FROM nodes WHERE name LIKE 'e2e_test_%'
    `)

    console.log('[Global Teardown] Test data cleanup complete!')
  } catch (error) {
    console.error('[Global Teardown] Error cleaning up test data:', error)
    // Don't throw - we want to continue with teardown
  }
}

/**
 * Clean up auth state files (including worker-specific directories)
 */
function cleanupAuthStates(): void {
  console.log('[Global Teardown] Cleaning up auth state files...')

  const authDir = '.auth'
  if (fs.existsSync(authDir)) {
    const entries = fs.readdirSync(authDir, { withFileTypes: true })
    for (const entry of entries) {
      const fullPath = `${authDir}/${entry.name}`
      if (entry.isDirectory()) {
        // Remove worker-specific directories recursively
        if (entry.name.startsWith('worker-')) {
          fs.rmSync(fullPath, { recursive: true, force: true })
        }
      } else if (entry.name.endsWith('.json')) {
        fs.unlinkSync(fullPath)
      }
    }
    // Remove directory if empty
    try {
      fs.rmdirSync(authDir)
    } catch {
      // Directory not empty, leave it
    }
  }

  console.log('[Global Teardown] Auth state files cleaned up!')
}

/**
 * Main global teardown function
 */
export default async function globalTeardown() {
  console.log('[Global Teardown] Starting e2e test teardown...')

  // Connect to test database
  const pool = new Pool({ connectionString: TEST_DB_URL })

  try {
    // Clean up test data
    await cleanupTestData(pool)
  } finally {
    await pool.end()
  }

  // Clean up auth state files
  cleanupAuthStates()

  console.log('[Global Teardown] Teardown complete!')
}
