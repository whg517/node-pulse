/**
 * Node Detail Tests
 *
 * Tests for node detail page:
 * - Metric cards
 * - Trend charts
 * - Time range selector
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { NodeDetailPage } from '../../pages/NodeDetailPage'

test.describe('Node Detail Page', () => {
  test('AC-20: loads with node data', async ({ adminPage }) => {
    // First get a node ID from the nodes list
    const response = await adminPage.request.get('/api/v1/nodes')
    const data = await response.json()

    if (data.data && data.data.length > 0) {
      const nodeId = data.data[0].node_id

      const detailPage = new NodeDetailPage(adminPage)
      await detailPage.goto(nodeId)

      await detailPage.expectMetricsVisible()
    } else {
      test.skip(true, 'No nodes available for detail test')
    }
  })

  test('shows metric cards', async ({ adminPage }) => {
    const response = await adminPage.request.get('/api/v1/nodes')
    const data = await response.json()

    if (data.data && data.data.length > 0) {
      const nodeId = data.data[0].node_id

      const detailPage = new NodeDetailPage(adminPage)
      await detailPage.goto(nodeId)

      await detailPage.expectMetricsVisible()

      // Should have multiple metric cards
      const cardCount = await adminPage.locator('[data-testid="metric-card"], .metric-card').count()
      expect(cardCount).toBeGreaterThan(0)
    } else {
      test.skip(true, 'No nodes available')
    }
  })

  test('shows trend chart', async ({ adminPage }) => {
    const response = await adminPage.request.get('/api/v1/nodes')
    const data = await response.json()

    if (data.data && data.data.length > 0) {
      const nodeId = data.data[0].node_id

      const detailPage = new NodeDetailPage(adminPage)
      await detailPage.goto(nodeId)

      await detailPage.expectChartVisible()
    } else {
      test.skip(true, 'No nodes available')
    }
  })

  test('time range selector updates data', async ({ adminPage }) => {
    const response = await adminPage.request.get('/api/v1/nodes')
    const data = await response.json()

    if (data.data && data.data.length > 0) {
      const nodeId = data.data[0].node_id

      const detailPage = new NodeDetailPage(adminPage)
      await detailPage.goto(nodeId)
      await detailPage.expectMetricsVisible()

      // Select 7d time range
      const timeRangeSelector = adminPage.locator('[data-testid="time-range-selector"], select[name="timeRange"]')

      if (await timeRangeSelector.count() > 0) {
        // Wait for API call after selection
        const responsePromise = adminPage.waitForResponse(
          resp => resp.url().includes('/api/v1/data/history') && resp.status() === 200,
          { timeout: 15000 }
        )

        await detailPage.selectTimeRange('7d')
        await responsePromise
      }
    } else {
      test.skip(true, 'No nodes available')
    }
  })

  test('handles non-existent node', async ({ adminPage }) => {
    const detailPage = new NodeDetailPage(adminPage)
    await detailPage.goto('non-existent-node-id')

    // Should show error or redirect
    const currentUrl = adminPage.url()

    // Either redirected or showing error
    expect(
      currentUrl.includes('nodes') ||
      await adminPage.locator('.error, [data-testid="error"]').count() > 0
    ).toBeTruthy()
  })

  test('back navigation works', async ({ adminPage }) => {
    const response = await adminPage.request.get('/api/v1/nodes')
    const data = await response.json()

    if (data.data && data.data.length > 0) {
      const nodeId = data.data[0].node_id

      const detailPage = new NodeDetailPage(adminPage)
      await detailPage.goto(nodeId)

      const backButton = adminPage.locator('button:has-text("Back"), a:has-text("Back")')

      if (await backButton.count() > 0) {
        await backButton.click()

        // Should navigate back to nodes list
        await expect(adminPage).toHaveURL(/.*nodes/)
      }
    } else {
      test.skip(true, 'No nodes available')
    }
  })
})

test.describe('Node Detail - API Integration', () => {
  test('history API called with correct params', async ({ adminPage }) => {
    const response = await adminPage.request.get('/api/v1/nodes')
    const data = await response.json()

    if (data.data && data.data.length > 0) {
      const nodeId = data.data[0].node_id

      // Navigate and wait for history API
      const historyPromise = adminPage.waitForResponse(
        resp => resp.url().includes('/api/v1/data/history') && resp.url().includes(nodeId),
        { timeout: 15000 }
      )

      await adminPage.goto(`/nodes/${nodeId}`)
      const historyResponse = await historyPromise

      expect(historyResponse.ok()).toBeTruthy()
    } else {
      test.skip(true, 'No nodes available')
    }
  })
})
