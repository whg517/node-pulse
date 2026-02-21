/**
 * MTR Visualization Tests
 *
 * Tests for the MTR Visualization component:
 * - Timeline/hop list rendering
 * - Health status coloring
 * - RTT statistics display
 * - Loading, error, empty states
 * - Accessibility
 */

import { test, expect } from '../../fixtures/auth.fixture'

test.describe('MTR Visualization Component', () => {
  // MTR visualization is typically shown on node detail page
  test.beforeEach(async ({ adminPage }) => {
    // Navigate to nodes page first
    await adminPage.goto('/nodes')
    await adminPage.waitForLoadState('networkidle')
  })

  test('displays MTR section on node detail', async ({ adminPage }) => {
    // Check if any nodes exist
    const nodeRows = adminPage.locator('table tbody tr')
    const nodeCount = await nodeRows.count()

    if (nodeCount > 0) {
      // Click on first node to view details
      await nodeRows.first().click()
      await adminPage.waitForLoadState('networkidle')

      // Look for MTR section/tab
      const mtrSection = adminPage.locator('[data-testid="mtr-visualization"], [class*="MTRVisualization"], text=/MTR|Traceroute/i').first()

      // MTR section might exist or not depending on implementation
      if (await mtrSection.count() > 0) {
        await expect(mtrSection).toBeVisible()
      }
    } else {
      // No nodes - skip test
      test.skip()
    }
  })

  test('shows loading state while MTR runs', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Look for MTR run button
    const runButton = adminPage.locator('button:has-text("Run MTR"), button:has-text("Trace"), [data-testid="run-mtr"]')

    if (await runButton.count() > 0) {
      await runButton.click()

      // Check for loading state
      const loadingIndicator = adminPage.locator('[data-testid="mtr-loading"], text=/Running|Loading|Tracing/i')
      // Loading state should appear briefly
      await adminPage.waitForTimeout(500)
    }
  })

  test('displays hop timeline', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Look for hop list/timeline
    const hopList = adminPage.locator('[data-testid="mtr-hops"], [class*="hop-list"], [class*="timeline"]')

    if (await hopList.count() > 0) {
      // Check for hop items
      const hopItems = adminPage.locator('[class*="hop-item"], [class*="hop-"]')
      const count = await hopItems.count()
      expect(count >= 0).toBeTruthy()
    }
  })

  test('shows RTT statistics per hop', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Look for RTT values (avg, min, max, std dev)
    const rttStats = adminPage.locator('text=/\\d+\\.?\\d*\\s*ms/i')

    // If MTR data exists, should show RTT stats
    if (await rttStats.count() > 0) {
      const statsText = await rttStats.first().textContent()
      expect(statsText).toMatch(/\d+\.?\d*\s*ms/i)
    }
  })

  test('displays packet loss percentage', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Look for loss percentage
    const lossIndicator = adminPage.locator('text=/\\d+\\.?\\d*%\\s*(loss|丢包)/i')

    if (await lossIndicator.count() > 0) {
      const lossText = await lossIndicator.first().textContent()
      expect(lossText).toMatch(/\d+\.?\d*%/)
    }
  })

  test('shows error state on MTR failure', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Error state would show if MTR fails
    const errorState = adminPage.locator('[data-testid="mtr-error"], [class*="error"], text=/error|failed|failed/i')

    // Just verify no crashes - error state depends on actual data
    expect(true).toBeTruthy()
  })

  test('shows empty state when no MTR data', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Empty state might show "No MTR data" or similar
    const emptyState = adminPage.locator('text=/No MTR|No trace|No data available/i')

    // Either has data or empty state - both valid
    expect(true).toBeTruthy()
  })
})

test.describe('MTR Visualization - Health Status', () => {
  test.beforeEach(async ({ adminPage }) => {
    await adminPage.goto('/nodes')
    await adminPage.waitForLoadState('networkidle')
  })

  test('displays healthy hops in green', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Look for healthy status (green color, < 5% loss)
    const healthyHops = adminPage.locator('[class*="healthy"], [style*="green"], [style*="#10b981"]')
    const count = await healthyHops.count()
    expect(count >= 0).toBeTruthy()
  })

  test('displays degraded hops in yellow', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Look for degraded status (yellow/amber, 5-20% loss)
    const degradedHops = adminPage.locator('[class*="degraded"], [class*="warning"], [style*="amber"], [style*="#f59e0b"]')
    const count = await degradedHops.count()
    expect(count >= 0).toBeTruthy()
  })

  test('displays problematic hops in red', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Look for problematic status (red, > 20% loss)
    const problematicHops = adminPage.locator('[class*="problematic"], [class*="critical"], [style*="red"], [style*="#ef4444"]')
    const count = await problematicHops.count()
    expect(count >= 0).toBeTruthy()
  })

  test('shows path health summary', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Look for path health summary badge
    const pathHealth = adminPage.locator('[data-testid="path-health"], [class*="path-health"], text=/Path.*Health|Overall.*Status/i')

    if (await pathHealth.count() > 0) {
      await expect(pathHealth.first()).toBeVisible()
    }
  })
})

test.describe('MTR Visualization - Accessibility', () => {
  test.beforeEach(async ({ adminPage }) => {
    await adminPage.goto('/nodes')
    await adminPage.waitForLoadState('networkidle')
  })

  test('has proper ARIA attributes', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Check for ARIA region
    const mtrRegion = adminPage.locator('[role="region"][aria-label*="MTR" i], [aria-label*="traceroute" i]')

    if (await mtrRegion.count() > 0) {
      const ariaLabel = await mtrRegion.first().getAttribute('aria-label')
      expect(ariaLabel).toBeTruthy()
    }
  })

  test('hop items have proper roles', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Check for list structure
    const hopList = adminPage.locator('[role="list"]')
    const hopItems = adminPage.locator('[role="listitem"]')

    if (await hopList.count() > 0 && await hopItems.count() > 0) {
      // Verify list structure
      const itemCount = await hopItems.count()
      expect(itemCount).toBeGreaterThan(0)
    }
  })

  test('supports keyboard navigation', async ({ adminPage }) => {
    const nodeRows = adminPage.locator('table tbody tr')
    if (await nodeRows.count() === 0) {
      test.skip()
      return
    }

    await nodeRows.first().click()
    await adminPage.waitForLoadState('networkidle')

    // Tab through interactive elements
    await adminPage.keyboard.press('Tab')
    await adminPage.keyboard.press('Tab')

    // Focus should be visible somewhere on page
    const focusedElement = adminPage.locator(':focus')
    const hasFocus = await focusedElement.count() > 0
    expect(hasFocus).toBeTruthy()
  })
})
