/**
 * Performance Dashboard Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class PerformancePage {
  readonly page: Page
  readonly metricCards: Locator
  readonly trendChart: Locator
  readonly refreshButton: Locator
  readonly lastUpdated: Locator

  constructor(page: Page) {
    this.page = page
    // Metric cards in grid layout
    this.metricCards = page.locator('.grid > div, [class*="card"]')
    // Chart is typically canvas or svg
    this.trendChart = page.locator('canvas, svg')
    // Refresh button
    this.refreshButton = page.locator('button:has-text("Refresh")')
    // Last updated text
    this.lastUpdated = page.locator('text=/last updated/i, text=/updated/i')
  }

  async goto() {
    await this.page.goto('/performance')
    await this.page.waitForLoadState('networkidle')
  }

  async expectMetricsVisible() {
    await this.metricCards.first().waitFor({ state: 'visible', timeout: 10000 })
  }

  async expectChartVisible() {
    await this.trendChart.waitFor({ state: 'visible', timeout: 10000 })
  }

  async clickRefresh() {
    if (await this.refreshButton.count() > 0) {
      await this.refreshButton.click()
    }
  }

  async waitForAutoRefresh() {
    // Wait for auto-poll (60 seconds by default)
    await this.page.waitForResponse(resp =>
      resp.url().includes('/api/v1/data/performance') || resp.url().includes('/api/v1/metrics'),
      { timeout: 65000 }
    )
  }
}
