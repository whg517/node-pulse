/**
 * Node Comparison Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class NodeComparisonPage {
  readonly page: Page
  readonly nodeSelector: Locator
  readonly metricsSelector: Locator
  readonly comparisonChart: Locator
  readonly selectedNodes: Locator
  readonly compareButton: Locator

  constructor(page: Page) {
    this.page = page
    this.nodeSelector = page.locator('[data-testid="node-selector"]')
    this.metricsSelector = page.locator('[data-testid="metrics-selector"], button[data-metric-key], button')
    this.comparisonChart = page.locator('[data-testid="comparison-chart"], canvas, .chart')
    this.selectedNodes = page.locator('[data-testid="node-selector"] input[type="checkbox"]:checked')
    this.compareButton = page.locator('[data-testid="compare-button"], button:has-text("Compare")')
  }

  async goto() {
    await this.page.goto('/nodes/comparison', { waitUntil: 'domcontentloaded' })
    await this.page.waitForLoadState('domcontentloaded')
  }

  async selectNodes(nodeNames: string[]) {
    for (const name of nodeNames) {
      const checkbox = this.page.locator(`[data-testid="node-selector"] label:has-text("${name}") input[type="checkbox"]`).first()
      if (!(await checkbox.isChecked())) {
        await checkbox.check()
      }
    }
  }

  async selectMetrics(metrics: string[]) {
    for (const metric of metrics) {
      const metricButton = this.page.locator(`button:has-text("${metric}")`).first()
      await metricButton.click()
    }
  }

  async clickCompare() {
    await this.compareButton.first().click()
  }

  async expectChartVisible() {
    await this.comparisonChart.first().waitFor({ state: 'visible' })
  }

  async getSelectedNodeCount(): Promise<number> {
    return await this.selectedNodes.count()
  }
}
