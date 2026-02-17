/**
 * Data Export Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class DataExportPage {
  readonly page: Page
  readonly form: Locator
  readonly nodeSelect: Locator
  readonly timeRangeSelect: Locator
  readonly formatSelect: Locator
  readonly submitButton: Locator
  readonly progressBar: Locator
  readonly downloadButton: Locator
  readonly accessWarning: Locator

  constructor(page: Page) {
    this.page = page
    this.form = page.locator('form')
    this.nodeSelect = page.locator('select[name="node"], [data-testid="node-select"]')
    this.timeRangeSelect = page.locator('select[name="timeRange"], select[name="time_range"]')
    this.formatSelect = page.locator('select[name="format"]')
    this.submitButton = page.locator('button[type="submit"], button:has-text("Export")')
    this.progressBar = page.locator('[data-testid="progress-bar"], .progress-bar')
    this.downloadButton = page.locator('button:has-text("Download"), a:has-text("Download")')
    this.accessWarning = page.locator('.access-warning, :text-is("Access Denied"), :text-is("Admin Only")')
  }

  async goto() {
    await this.page.goto('/export')
    await this.page.waitForLoadState('networkidle')
  }

  async expectFormVisible() {
    await this.form.waitFor({ state: 'visible' })
  }

  async expectAccessWarning() {
    await this.accessWarning.waitFor({ state: 'visible' })
  }

  async hasAccessWarning(): Promise<boolean> {
    return (await this.accessWarning.count()) > 0
  }

  async selectNode(nodeName: string) {
    await this.nodeSelect.selectOption({ label: nodeName })
  }

  async selectTimeRange(range: string) {
    await this.timeRangeSelect.selectOption(range)
  }

  async selectFormat(format: string) {
    await this.formatSelect.selectOption(format)
  }

  async submitExport() {
    await this.submitButton.click()
  }

  async waitForCompletion(timeout = 60000) {
    await this.downloadButton.waitFor({ state: 'visible', timeout })
  }

  async download() {
    const [download] = await Promise.all([
      this.page.waitForEvent('download'),
      this.downloadButton.click(),
    ])
    return download
  }
}
