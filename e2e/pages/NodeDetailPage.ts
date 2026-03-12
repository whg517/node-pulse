/**
 * Node Detail Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class NodeDetailPage {
  readonly page: Page
  readonly metricCards: Locator
  readonly trendChart: Locator
  readonly timeRangeSelector: Locator
  readonly backButton: Locator

  constructor(page: Page) {
    this.page = page
    this.metricCards = page.locator('[data-testid="metric-card"], .metric-card')
    this.trendChart = page.locator('[data-testid="trend-chart"], canvas, .chart')
    this.timeRangeSelector = page.locator('[data-testid="time-range-selector"], select[name="timeRange"], .time-range-selector')
    this.backButton = page.locator('button:has-text("Back"), a:has-text("Back")')
  }

  async goto(nodeId: string) {
    await this.page.goto(`/nodes/${nodeId}`)
    await this.page.waitForLoadState('domcontentloaded')
  }

  async expectMetricsVisible() {
    await this.metricCards.first().waitFor({ state: 'visible' })
  }

  async expectChartVisible() {
    await this.trendChart.waitFor({ state: 'visible' })
  }

  async selectTimeRange(range: '24h' | '7d' | '30d') {
    await this.timeRangeSelector.click()
    await this.page.locator(`option:has-text("${range}"), [data-value="${range}"]`).click()
  }

  async waitForDataLoad() {
    await this.page.waitForResponse(resp =>
      resp.url().includes('/api/v1/data/history') && resp.status() === 200,
      { timeout: 15000 }
    )
  }
}
