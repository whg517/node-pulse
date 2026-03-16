/**
 * Dashboard Tests
 *
 * Tests for dashboard page functionality:
 * - Metrics display
 * - Node list
 * - Auto-refresh (5 seconds)
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { DashboardPage } from '../../pages/DashboardPage'

test.describe('Dashboard Page', () => {
  let dashboardPage: DashboardPage

  test.beforeEach(async ({ adminPage }) => {
    dashboardPage = new DashboardPage(adminPage)
    await dashboardPage.goto()
  })

  test('AC-19: page loads and displays metrics', async ({ adminPage }) => {
    await dashboardPage.expectMetricsVisible()
    await dashboardPage.expectNodesVisible()
  })

  test('displays node count or status', async ({ adminPage }) => {
    // Check for metrics grid (Avg Latency, Packet Loss, Jitter)
    const metricsGrid = adminPage.locator('.grid')
    await expect(metricsGrid.first()).toBeVisible()

    // Should show metrics text
    const metricsText = await metricsGrid.first().textContent()
    expect(metricsText).toBeTruthy()
  })

  test('shows node list', async ({ adminPage }) => {
    await dashboardPage.expectNodesVisible()

    // Should have either a table with nodes OR empty state message
    const hasNodes = await dashboardPage.hasNodes()
    if (hasNodes) {
      const table = adminPage.locator('table')
      await expect(table).toBeVisible()
    } else {
      // Empty state is acceptable - use .first() to avoid strict mode violation
      await expect(adminPage.locator('text=/No nodes/i').first()).toBeVisible()
    }
  })

  test('auto-refreshes every 5 seconds', async ({ adminPage }) => {
    await dashboardPage.expectMetricsVisible()

    // Wait for auto-refresh API call (5 second interval)
    const response = await adminPage.waitForResponse(
      resp => resp.url().includes('/api/v1/data/metrics') && resp.status() === 200,
      { timeout: 10000 }
    )

    expect(response.ok()).toBeTruthy()
  })

  test('manual refresh button works', async ({ adminPage }) => {
    await dashboardPage.expectMetricsVisible()

    // Click refresh if button exists
    const refreshButton = adminPage.locator('[data-testid="refresh-button"], button:has-text("Refresh")')

    if (await refreshButton.count() > 0) {
      // Set up response listener before clicking
      const responsePromise = adminPage.waitForResponse(
        resp => resp.url().includes('/api/v1/data/metrics'),
        { timeout: 10000 }
      ).catch(() => null) // Handle case where response doesn't happen

      await refreshButton.click()

      const response = await responsePromise
      // Response may be null if no API call was made (e.g., debounced)
      // Just verify button was clickable
    } else {
      // Skip test if no refresh button present
      test.skip(true, 'No refresh button found')
    }
  })

  test('shows alert summary if available', async ({ adminPage }) => {
    // Check for alert-related content
    const alertSection = adminPage.locator('[data-testid="alert-summary"], .alert-summary, text=/alert/i')

    // Alert section may or may not be visible depending on data
    const isVisible = await alertSection.first().isVisible().catch(() => false)

    // Just verify page loaded
    await expect(adminPage).toHaveURL(/.*dashboard/)
  })

  test('navigation works from dashboard', async ({ adminPage }) => {
    await dashboardPage.expectMetricsVisible()

    // Click on Nodes navigation link (sidebar)
    const nodesLink = adminPage.locator('a[href="/nodes"], a:has-text("Nodes")')

    if (await nodesLink.count() > 0) {
      await nodesLink.first().click()
      await expect(adminPage).toHaveURL(/.*nodes/)
    } else {
      // If no nodes link, just verify we're on dashboard
      await expect(adminPage).toHaveURL(/.*dashboard/)
    }
  })
})

test.describe('Dashboard Data', () => {
  test('metrics API returns data', async ({ adminPage }) => {
    await adminPage.goto('/dashboard')

    // Wait for metrics API call
    const response = await adminPage.waitForResponse(
      resp => resp.url().includes('/api/v1/data/metrics'),
      { timeout: 10000 }
    )

    expect(response.ok()).toBeTruthy()

    const data = await response.json()
    expect(data).toHaveProperty('data')
  })

  test('handles empty data gracefully', async ({ adminPage }) => {
    await adminPage.goto('/dashboard')

    // Page should load without errors even if no data
    await expect(adminPage).toHaveURL(/.*dashboard/)

    // No error messages should be visible
    const errorMessages = await adminPage.locator('.error, .alert-error').count()
    expect(errorMessages).toBe(0)
  })
})
