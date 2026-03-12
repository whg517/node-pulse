/**
 * Comparison Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class ComparisonPage {
  readonly page: Page
  readonly nodeSelector: Locator
  readonly compareButton: Locator
  readonly chart: Locator
  readonly clearButton: Locator
  readonly selectedNodes: Locator

  constructor(page: Page) {
    this.page = page
    // Node selector - could be multi-select or checkboxes
    this.nodeSelector = page.locator('select[multiple], input[type="checkbox"]')
    // Compare button
    this.compareButton = page.locator('button:has-text("Compare")')
    // Chart for comparison results
    this.chart = page.locator('canvas, svg, [class*="chart"]')
    // Clear/reset button
    this.clearButton = page.locator('button:has-text("Clear"), button:has-text("Reset")')
    // Selected nodes list
    this.selectedNodes = page.locator('[class*="selected"]')
  }

  async goto() {
    await this.page.goto('/comparison')
    await this.page.waitForLoadState('domcontentloaded')
  }

  async selectNode(nodeName: string) {
    // Try checkbox first
    const checkbox = this.page.locator(`input[type="checkbox"][value*="${nodeName}"], label:has-text("${nodeName}") input`)
    if (await checkbox.count() > 0) {
      await checkbox.check()
    } else {
      // Try multi-select
      const select = this.page.locator('select[multiple]')
      if (await select.count() > 0) {
        await select.selectOption({ label: nodeName })
      }
    }
  }

  async clickCompare() {
    await this.compareButton.click()
  }

  async clickClear() {
    if (await this.clearButton.count() > 0) {
      await this.clearButton.click()
    }
  }

  async expectChartVisible() {
    await this.chart.waitFor({ state: 'visible', timeout: 10000 })
  }
}
