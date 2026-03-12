/**
 * Dashboard Page Object Model
 *
 * Handles dashboard viewing:
 * - Metrics display
 * - Nodes overview
 * - Alerts/anomalies
 * - Auto-refresh
 */
import { Page, Locator, expect } from '@playwright/test'
import { BasePage, PageSelectors, DEFAULT_SELECTORS } from './common/BasePage'

export interface DashboardSelectors extends PageSelectors {
  metricsSection?: string
  nodeList?: string
  alertList?: string
  logoutButton?: string
  title?: string
  autoRefreshIndicator?: string
  welcomeMessage?: string
}

export const DEFAULT_DASHBOARD_SELECTORS: DashboardSelectors = {
  ...DEFAULT_SELECTORS,
  metricsSection: '[data-testid="metrics-section"], .grid:has(.metric-card), .grid:has(.rounded-lg)',
  nodeList: '[data-testid="node-list"], table',
  alertList: '[data-testid="alert-list"], .alert-list, text=/anomaly/i',
  logoutButton: '[data-testid="logout-button"], button:has-text("Logout"), button:has-text("logout"), button:has-text("登出"), button:has-text("退出")',
  title: '[data-testid="dashboard-title"], h1:has-text("Dashboard"), h2:has-text("Dashboard")',
  autoRefreshIndicator: '[data-testid="auto-refresh"], text=/auto.*refresh/i, text=/refreshing/i',
  welcomeMessage: '[data-testid="welcome-message"], text=/welcome/i',
}

export class DashboardPage extends BasePage {
  readonly metricsSection: Locator
  readonly nodeList: Locator
  readonly alertList: Locator
  readonly logoutButton: Locator
  readonly title: Locator
  readonly autoRefreshIndicator: Locator
  readonly welcomeMessage: Locator

  constructor(page: Page, selectors: DashboardSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_DASHBOARD_SELECTORS, ...selectors }

    this.metricsSection = page.locator(mergedSelectors.metricsSection!)
    this.nodeList = page.locator(mergedSelectors.nodeList!)
    this.alertList = page.locator(mergedSelectors.alertList!)
    this.logoutButton = page.locator(mergedSelectors.logoutButton!)
    this.title = page.locator(mergedSelectors.title!)
    this.autoRefreshIndicator = page.locator(mergedSelectors.autoRefreshIndicator!)
    this.welcomeMessage = page.locator(mergedSelectors.welcomeMessage!)
  }

  /**
   * Navigate to dashboard
   */
  async goto(): Promise<void> {
    await super.goto('/dashboard')
    await this.waitForReady()
  }

  /**
   * Expect metrics visible
   */
  async expectMetricsVisible(): Promise<void> {
    await this.metricsSection.first().waitFor({ state: 'visible', timeout: 10000 })
  }

  /**
   * Expect nodes visible
   */
  async expectNodesVisible(): Promise<void> {
    const table = this.page.locator('table')
    const emptyState = this.page.getByText('No nodes')
    await table.or(emptyState).first().waitFor({ state: 'visible', timeout: 10000 })
  }

  /**
   * Expect alerts list visible
   */
  async expectAlertsVisible(): Promise<void> {
    await this.alertList.first().waitFor({ state: 'visible', timeout: 10000 })
  }

  /**
   * Click logout button
   * Note: Logout is in a dropdown menu, so we need to open the user menu first
   */
  async clickLogout(): Promise<void> {
    // First, click the user menu button to open the dropdown
    const userMenuButton = this.page.locator('button:has(.rounded-full), [aria-haspopup="menu"]').first()
    await userMenuButton.click()

    // Wait for dropdown to appear and click logout
    await this.page.waitForTimeout(300)
    await this.logoutButton.click()
  }

  /**
   * Wait for auto-refresh to complete
   */
  async waitForAutoRefresh(): Promise<void> {
    await this.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/data/metrics') || resp.url().includes('/api/v1/nodes'),
      { timeout: 10000 }
    )
  }

  /**
   * Get node count
   */
  async getNodeCount(): Promise<number> {
    if (await this.isEmptyStateVisible()) {
      return 0
    }
    const rows = this.nodeList.locator('tbody tr')
    return await rows.count()
  }

  /**
   * Check if dashboard has nodes
   */
  async hasNodes(): Promise<boolean> {
    return !(await this.isEmptyStateVisible())
  }

  /**
   * Get metrics count
   */
  async getMetricsCount(): Promise<number> {
    const metricCards = this.metricsSection.locator('[data-testid="metric-card"], .metric-card')
    return await metricCards.count()
  }

  /**
   * Get metric value by name
   */
  async getMetricValue(metricName: string): Promise<string | null> {
    const metricCard = this.page.locator(`[data-testid="metric-${metricName}"], .metric-card:has-text("${metricName}")`)
    if (await metricCard.count() > 0) {
      const valueLocator = metricCard.locator('[data-testid="metric-value"], .metric-value')
      if (await valueLocator.count() > 0) {
        return await valueLocator.first().textContent()
      }
    }
    return null
  }

  /**
   * Expect dashboard title
   */
  async expectTitle(): Promise<void> {
    await expect(this.title.first()).toBeVisible()
  }

  /**
   * Expect welcome message
   */
  async expectWelcomeMessage(): Promise<void> {
    await expect(this.welcomeMessage.first()).toBeVisible()
  }

  /**
   * Check if auto-refresh indicator is visible
   */
  async isAutoRefreshIndicatorVisible(): Promise<boolean> {
    return await this.autoRefreshIndicator.first().isVisible().catch(() => false)
  }

  /**
   * Logout and wait for redirect
   */
  async logoutAndWait(): Promise<void> {
    await this.clickLogout()
    await this.page.waitForURL('**/login**', { timeout: 10000 })
  }
}
