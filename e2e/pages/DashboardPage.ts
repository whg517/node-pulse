/**
 * Dashboard Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class DashboardPage {
  readonly page: Page
  readonly metricsSection: Locator
  readonly nodeList: Locator
  readonly emptyNodesState: Locator
  readonly alertList: Locator
  readonly logoutButton: Locator
  readonly title: Locator
  readonly autoRefreshIndicator: Locator

  constructor(page: Page) {
    this.page = page
    // Metrics cards are rendered in a grid
    this.metricsSection = page.locator('.grid')
    // Node list table (or empty state)
    this.nodeList = page.locator('table')
    // Empty state when no nodes
    this.emptyNodesState = page.locator('text=/No nodes/i')
    // Top anomalies list
    this.alertList = page.locator('text=/anomaly/i, text=/alert/i')
    // Navigation logout button
    this.logoutButton = page.locator('button:has-text("Logout")')
    // Page title
    this.title = page.locator('h2:has-text("Dashboard")')
    // Auto-refresh indicator
    this.autoRefreshIndicator = page.locator('text=/auto.*refresh/i')
  }

  async goto() {
    await this.page.goto('/dashboard')
    await this.page.waitForLoadState('networkidle')
  }

  async expectMetricsVisible() {
    // Wait for metrics cards (grid) to be visible
    await this.metricsSection.first().waitFor({ state: 'visible', timeout: 10000 })
  }

  async expectNodesVisible() {
    // Wait for either table OR empty state message
    // Use locator with or() for complex selector logic
    const table = this.page.locator('table')
    const emptyState = this.page.getByText('No nodes')

    // Wait for either to be visible
    await table.or(emptyState).first().waitFor({ state: 'visible', timeout: 10000 })
  }

  async clickLogout() {
    await this.logoutButton.click()
  }

  async waitForAutoRefresh() {
    // Wait for API call to complete (auto-refresh is every 5 seconds)
    await this.page.waitForResponse(resp =>
      resp.url().includes('/api/v1/data/metrics') || resp.url().includes('/api/v1/nodes'),
      { timeout: 10000 }
    )
  }

  async getNodeCount(): Promise<number> {
    // Check if empty state is shown
    if (await this.emptyNodesState.isVisible()) {
      return 0
    }
    const rows = this.nodeList.locator('tbody tr')
    return await rows.count()
  }

  async hasNodes(): Promise<boolean> {
    return !(await this.emptyNodesState.isVisible())
  }
}
