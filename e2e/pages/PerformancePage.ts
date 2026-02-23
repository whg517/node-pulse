/**
 * Performance Dashboard Page Object Model
 *
 * Handles performance metrics viewing:
 * - Performance metric cards
 * - Trend charts
 * - Auto-refresh
 */
import { Page, Locator, expect } from '@playwright/test'
import { BasePage, PageSelectors, DEFAULT_SELECTORS } from './common/BasePage'

export interface PerformanceSelectors extends PageSelectors {
  metricCards?: string
  trendChart?: string
  refreshButton?: string
  lastUpdated?: string
  performanceSection?: string
}

export const DEFAULT_PERFORMANCE_SELECTORS: PerformanceSelectors = {
  ...DEFAULT_SELECTORS,
  metricCards: '[data-testid="performance-metrics"], .grid > div, [class*="card"]',
  trendChart: '[data-testid="trend-chart"], canvas, svg',
  refreshButton: '[data-testid="refresh-button"], button:has-text("Refresh")',
  lastUpdated: '[data-testid="last-updated"], text=/last updated/i, text=/updated/i',
  performanceSection: '[data-testid="performance-section"], .performance-section',
}

export class PerformancePage extends BasePage {
  readonly metricCards: Locator
  readonly trendChart: Locator
  readonly refreshButton: Locator
  readonly lastUpdated: Locator
  readonly performanceSection: Locator

  constructor(page: Page, selectors: PerformanceSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_PERFORMANCE_SELECTORS, ...selectors }

    this.metricCards = page.locator(mergedSelectors.metricCards!)
    this.trendChart = page.locator(mergedSelectors.trendChart!)
    this.refreshButton = page.locator(mergedSelectors.refreshButton!)
    this.lastUpdated = page.locator(mergedSelectors.lastUpdated!)
    this.performanceSection = page.locator(mergedSelectors.performanceSection!)
  }

  /**
   * Navigate to performance page
   */
  async goto(): Promise<void> {
    await super.goto('/performance')
    await this.waitForReady()
  }

  /**
   * Expect metrics visible
   */
  async expectMetricsVisible(): Promise<void> {
    await this.metricCards.first().waitFor({ state: 'visible', timeout: 10000 })
  }

  /**
   * Expect chart visible
   */
  async expectChartVisible(): Promise<void> {
    await this.trendChart.waitFor({ state: 'visible', timeout: 10000 })
  }

  /**
   * Click refresh button
   */
  async clickRefresh(): Promise<void> {
    if (await this.refreshButton.count() > 0) {
      await this.refreshButton.click()
      await this.waitForReady()
    }
  }

  /**
   * Wait for auto-refresh
   */
  async waitForAutoRefresh(): Promise<void> {
    await this.page.waitForResponse(
      (resp) => resp.url().includes('/api/v1/data/performance') || resp.url().includes('/api/v1/metrics'),
      { timeout: 65000 }
    )
  }

  /**
   * Get metric card count
   */
  async getMetricCardCount(): Promise<number> {
    return await this.metricCards.count()
  }

  /**
   * Get performance metric value by name
   */
  async getMetricValue(metricName: string): Promise<string | null> {
    const metricCard = this.page.locator(`[data-testid="performance-${metricName}"], .metric-card:has-text("${metricName}")`)
    if (await metricCard.count() > 0) {
      const valueLocator = metricCard.locator('[data-testid="metric-value"], .metric-value')
      if (await valueLocator.count() > 0) {
        return await valueLocator.first().textContent()
      }
    }
    return null
  }

  /**
   * Get metric target by name
   */
  async getMetricTarget(metricName: string): Promise<string | null> {
    const metricCard = this.page.locator(`[data-testid="performance-${metricName}"], .metric-card:has-text("${metricName}")`)
    if (await metricCard.count() > 0) {
      const targetLocator = metricCard.locator('[data-testid="metric-target"], .metric-target')
      if (await targetLocator.count() > 0) {
        return await targetLocator.first().textContent()
      }
    }
    return null
  }

  /**
   * Check if metric meets target
   */
  async isMetricOnTarget(metricName: string): Promise<boolean> {
    const metricCard = this.page.locator(`[data-testid="performance-${metricName}"], .metric-card:has-text("${metricName}")`)
    if (await metricCard.count() > 0) {
      const statusLocator = metricCard.locator('[data-testid="metric-status"], .on-target, .status-good')
      return await statusLocator.count() > 0
    }
    return false
  }

  /**
   * Get last updated text
   */
  async getLastUpdatedText(): Promise<string | null> {
    if (await this.lastUpdated.count() > 0) {
      return await this.lastUpdated.first().textContent()
    }
    return null
  }

  /**
   * Expect performance section visible
   */
  async expectPerformanceSectionVisible(): Promise<void> {
    await expect(this.performanceSection.first()).toBeVisible()
  }

  /**
   * Expect all metrics visible
   */
  async expectAllMetricsVisible(expectedMetrics: string[]): Promise<void> {
    for (const metric of expectedMetrics) {
      const metricLocator = this.page.locator(`[data-testid="performance-${metric}"], .metric-card:has-text("${metric}")`)
      await expect(metricLocator.first()).toBeVisible()
    }
  }
}
