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
    this.nodeSelector = page.locator('[data-testid="node-selector"], select[name="nodes"], .node-selector')
    this.metricsSelector = page.locator('[data-testid="metrics-selector"], select[name="metrics"], .metrics-selector')
    this.comparisonChart = page.locator('[data-testid="comparison-chart"], canvas, .chart')
    this.selectedNodes = page.locator('[data-testid="selected-nodes"], .selected-nodes')
    this.compareButton = page.locator('button:has-text("Compare")')
  }

  async goto() {
    await this.page.goto('/comparison')
    await this.page.waitForLoadState('networkidle')
  }

  async selectNodes(nodeNames: string[]) {
    for (const name of nodeNames) {
      await this.nodeSelector.click()
      await this.page.locator(`option:has-text("${name}"), [data-value="${name}"]`).click()
    }
  }

  async selectMetrics(metrics: string[]) {
    for (const metric of metrics) {
      await this.metricsSelector.click()
      await this.page.locator(`option:has-text("${metric}"), [data-value="${metric}"]`).click()
    }
  }

  async clickCompare() {
    await this.compareButton.click()
  }

  async expectChartVisible() {
    await this.comparisonChart.waitFor({ state: 'visible' })
  }

  async getSelectedNodeCount(): Promise<number> {
    return await this.selectedNodes.locator('[data-testid="selected-node"], .selected-node').count()
  }
}
