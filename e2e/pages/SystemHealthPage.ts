/**
 * System Health Page Object Model
 *
 * Handles system health monitoring:
 * - View system metrics
 * - Database health status
 * - API health status
 * - Performance metrics
 */
import { Page, Locator, expect } from '@playwright/test'
import { BasePage, PageSelectors, DEFAULT_SELECTORS } from './common/BasePage'

export interface SystemHealthSelectors extends PageSelectors {
  refreshButton?: string
  databaseStatus?: string
  apiStatus?: string
  metricsSection?: string
  healthCards?: string
  uptimeDisplay?: string
  versionDisplay?: string
}

export const DEFAULT_SYSTEM_HEALTH_SELECTORS: SystemHealthSelectors = {
  ...DEFAULT_SELECTORS,
  refreshButton: '[data-testid="refresh-button"], button:has-text("Refresh")',
  databaseStatus: '[data-testid="database-status"], .database-status',
  apiStatus: '[data-testid="api-status"], .api-status',
  metricsSection: '[data-testid="metrics-section"], .metrics-section',
  healthCards: '[data-testid="health-card"], .health-card',
  uptimeDisplay: '[data-testid="uptime-display"], .uptime',
  versionDisplay: '[data-testid="version-display"], .version',
}

export class SystemHealthPage extends BasePage {
  readonly refreshButton: Locator
  readonly databaseStatus: Locator
  readonly apiStatus: Locator
  readonly metricsSection: Locator
  readonly healthCards: Locator
  readonly uptimeDisplay: Locator
  readonly versionDisplay: Locator

  constructor(page: Page, selectors: SystemHealthSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_SYSTEM_HEALTH_SELECTORS, ...selectors }

    this.refreshButton = page.locator(mergedSelectors.refreshButton!)
    this.databaseStatus = page.locator(mergedSelectors.databaseStatus!)
    this.apiStatus = page.locator(mergedSelectors.apiStatus!)
    this.metricsSection = page.locator(mergedSelectors.metricsSection!)
    this.healthCards = page.locator(mergedSelectors.healthCards!)
    this.uptimeDisplay = page.locator(mergedSelectors.uptimeDisplay!)
    this.versionDisplay = page.locator(mergedSelectors.versionDisplay!)
  }

  /**
   * Navigate to system health page
   */
  async goto(): Promise<void> {
    await super.goto('/integrations/health')
    await this.waitForReady()
  }

  /**
   * Refresh health data
   */
  async refresh(): Promise<void> {
    if (await this.refreshButton.count() > 0) {
      await this.refreshButton.click()
      await this.waitForReady()
    }
  }

  /**
   * Get database status text
   */
  async getDatabaseStatus(): Promise<string | null> {
    if (await this.databaseStatus.count() > 0) {
      return await this.databaseStatus.first().textContent()
    }
    return null
  }

  /**
   * Get API status text
   */
  async getApiStatus(): Promise<string | null> {
    if (await this.apiStatus.count() > 0) {
      return await this.apiStatus.first().textContent()
    }
    return null
  }

  /**
   * Check if database is healthy
   */
  async isDatabaseHealthy(): Promise<boolean> {
    const status = await this.getDatabaseStatus()
    return status?.toLowerCase().includes('healthy') ??
           status?.toLowerCase().includes('connected') ??
           status?.toLowerCase().includes('ok') ??
           false
  }

  /**
   * Check if API is healthy
   */
  async isApiHealthy(): Promise<boolean> {
    const status = await this.getApiStatus()
    return status?.toLowerCase().includes('healthy') ??
           status?.toLowerCase().includes('ok') ??
           status?.toLowerCase().includes('running') ??
           false
  }

  /**
   * Get uptime text
   */
  async getUptime(): Promise<string | null> {
    if (await this.uptimeDisplay.count() > 0) {
      return await this.uptimeDisplay.first().textContent()
    }
    return null
  }

  /**
   * Get version text
   */
  async getVersion(): Promise<string | null> {
    if (await this.versionDisplay.count() > 0) {
      return await this.versionDisplay.first().textContent()
    }
    return null
  }

  /**
   * Get health card count
   */
  async getHealthCardCount(): Promise<number> {
    return await this.healthCards.count()
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
   * Expect database healthy
   */
  async expectDatabaseHealthy(): Promise<void> {
    const healthy = await this.isDatabaseHealthy()
    expect(healthy).toBeTruthy()
  }

  /**
   * Expect API healthy
   */
  async expectApiHealthy(): Promise<void> {
    const healthy = await this.isApiHealthy()
    expect(healthy).toBeTruthy()
  }

  /**
   * Expect health cards visible
   */
  async expectHealthCardsVisible(): Promise<void> {
    await expect(this.healthCards.first()).toBeVisible()
  }

  /**
   * Expect metrics section visible
   */
  async expectMetricsVisible(): Promise<void> {
    await expect(this.metricsSection.first()).toBeVisible()
  }

  /**
   * Expect uptime displayed
   */
  async expectUptimeDisplayed(): Promise<void> {
    await expect(this.uptimeDisplay.first()).toBeVisible()
  }

  /**
   * Expect version displayed
   */
  async expectVersionDisplayed(): Promise<void> {
    await expect(this.versionDisplay.first()).toBeVisible()
  }

  /**
   * Wait for health data to load
   */
  async waitForHealthData(timeout = 10000): Promise<void> {
    await this.databaseStatus.first().waitFor({ state: 'visible', timeout })
    await this.apiStatus.first().waitFor({ state: 'visible', timeout })
  }

  /**
   * Get all system metrics
   */
  async getAllMetrics(): Promise<Record<string, string>> {
    const metrics: Record<string, string> = {}
    const metricCards = this.page.locator('[data-testid="metric-card"], .metric-card')
    const count = await metricCards.count()

    for (let i = 0; i < count; i++) {
      const card = metricCards.nth(i)
      const nameLocator = card.locator('[data-testid="metric-name"], .metric-name, h3, .card-title')
      const valueLocator = card.locator('[data-testid="metric-value"], .metric-value, .card-value')

      const name = await nameLocator.first().textContent()
      const value = await valueLocator.first().textContent()

      if (name && value) {
        metrics[name.trim()] = value.trim()
      }
    }

    return metrics
  }
}
