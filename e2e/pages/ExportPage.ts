/**
 * Export Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class ExportPage {
  readonly page: Page
  readonly form: Locator
  readonly startDateInput: Locator
  readonly endDateInput: Locator
  readonly formatSelect: Locator
  readonly submitButton: Locator
  readonly accessWarning: Locator
  readonly progressIndicator: Locator

  constructor(page: Page) {
    this.page = page
    this.form = page.locator('form')
    this.startDateInput = page.locator('input[name="start_date"], input[type="date"]').first()
    this.endDateInput = page.locator('input[name="end_date"], input[type="date"]').last()
    this.formatSelect = page.locator('select[name="format"]')
    this.submitButton = page.locator('button[type="submit"], button:has-text("Export")')
    // Access warning uses yellow background
    this.accessWarning = page.locator('.bg-yellow-50:has-text("Admin"), .bg-yellow-50:has-text("admin")')
    this.progressIndicator = page.locator('.animate-spin, text=/exporting/i')
  }

  async goto() {
    await this.page.goto('/reports/history')
    await this.page.waitForLoadState('domcontentloaded')
  }

  async expectFormVisible() {
    await this.form.waitFor({ state: 'visible' })
  }

  async hasAccessWarning(): Promise<boolean> {
    const warning = this.page.locator('.bg-yellow-50')
    const count = await warning.count()
    if (count === 0) return false
    const text = await warning.first().textContent()
    return text?.toLowerCase().includes('admin') ?? false
  }

  async submitExport(startDate: string, endDate: string, format: string = 'csv') {
    if (await this.startDateInput.count() > 0) {
      await this.startDateInput.fill(startDate)
    }
    if (await this.endDateInput.count() > 0) {
      await this.endDateInput.fill(endDate)
    }
    if (await this.formatSelect.count() > 0) {
      await this.formatSelect.selectOption(format)
    }
    await this.submitButton.click()
  }
}
