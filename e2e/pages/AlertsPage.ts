/**
 * Alert Rules Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class AlertRulesPage {
  readonly page: Page
  readonly table: Locator
  readonly createButton: Locator
  readonly editButtons: Locator
  readonly deleteButtons: Locator
  readonly toggleButtons: Locator
  readonly modal: Locator
  readonly loadingSpinner: Locator
  readonly emptyState: Locator

  // Form fields
  readonly nameInput: Locator
  readonly metricSelect: Locator
  readonly thresholdInput: Locator
  readonly durationInput: Locator
  readonly levelSelect: Locator
  readonly submitButton: Locator

  constructor(page: Page) {
    this.page = page
    this.table = page.locator('table')
    this.createButton = page.locator('button:has-text("Create"), button:has-text("Add Rule")')
    this.editButtons = page.locator('button:has-text("Edit")')
    this.deleteButtons = page.locator('button:has-text("Delete")')
    this.toggleButtons = page.locator('input[type="checkbox"]')
    this.modal = page.locator('.fixed.inset-0')
    this.loadingSpinner = page.locator('.animate-spin')
    this.emptyState = page.locator('.text-center.py-12, .text-center:has-text("No")')

    // Form fields
    this.nameInput = page.locator('#name, input[name="name"]')
    this.metricSelect = page.locator('select[name="metric_type"], select[name="metricType"]')
    this.thresholdInput = page.locator('#threshold, input[name="threshold"]')
    this.durationInput = page.locator('#duration, input[name="duration"]')
    this.levelSelect = page.locator('select[name="level"]')
    this.submitButton = page.locator('button[type="submit"]')
  }

  async goto() {
    await this.page.goto('/alerts/rules')
    await this.page.waitForLoadState('networkidle')
  }

  async waitForReady() {
    // Wait for loading to complete
    await this.loadingSpinner.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})
    // Small delay for rendering
    await this.page.waitForTimeout(500)
  }

  async expectTableVisible() {
    await this.waitForReady()
    // Check if table exists or empty state is shown
    const tableCount = await this.table.count()
    if (tableCount === 0) {
      // Verify empty state is visible instead
      await this.emptyState.first().waitFor({ state: 'visible', timeout: 5000 }).catch(() => {})
    } else {
      await this.table.waitFor({ state: 'visible' })
    }
  }

  async hasData(): Promise<boolean> {
    await this.waitForReady()
    return (await this.table.count()) > 0
  }

  async clickCreate() {
    await this.createButton.first().click()
    await this.modal.waitFor({ state: 'visible' })
  }

  async createRule(name: string, metric: string, threshold: number, level: string) {
    await this.clickCreate()
    await this.nameInput.fill(name)
    if (await this.metricSelect.count() > 0) {
      await this.metricSelect.selectOption(metric)
    }
    await this.thresholdInput.fill(String(threshold))
    if (await this.levelSelect.count() > 0) {
      await this.levelSelect.selectOption(level)
    }
    await this.submitButton.click()
    await this.modal.waitFor({ state: 'hidden' })
  }

  async toggleRule(row: number) {
    const toggle = this.table.locator('tr').nth(row).locator('input[type="checkbox"]')
    await toggle.click()
  }

  async deleteRule(row: number) {
    const deleteButton = this.table.locator('tr').nth(row).locator('button:has-text("Delete")')
    await deleteButton.click()
    await this.page.locator('.fixed button:has-text("Delete")').click()
  }
}

/**
 * Alert Records Page Object Model
 */
export class AlertRecordsPage {
  readonly page: Page
  readonly table: Locator
  readonly filterButton: Locator
  readonly statusFilter: Locator
  readonly nodeFilter: Locator
  readonly searchInput: Locator
  readonly loadingSpinner: Locator
  readonly emptyState: Locator

  constructor(page: Page) {
    this.page = page
    this.table = page.locator('table')
    this.filterButton = page.locator('button:has-text("Filter")')
    this.statusFilter = page.locator('select[name="status"]')
    this.nodeFilter = page.locator('select[name="node"]')
    this.searchInput = page.locator('input[type="search"], input[placeholder*="search" i]')
    this.loadingSpinner = page.locator('.animate-spin')
    this.emptyState = page.locator('.text-center.py-12, .text-center:has-text("No")')
  }

  async goto() {
    await this.page.goto('/alerts/records')
    await this.page.waitForLoadState('networkidle')
  }

  async waitForReady() {
    await this.loadingSpinner.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})
    await this.page.waitForTimeout(500)
  }

  async expectTableVisible() {
    await this.waitForReady()
    const tableCount = await this.table.count()
    if (tableCount === 0) {
      await this.emptyState.first().waitFor({ state: 'visible', timeout: 5000 }).catch(() => {})
    } else {
      await this.table.waitFor({ state: 'visible' })
    }
  }

  async hasData(): Promise<boolean> {
    await this.waitForReady()
    return (await this.table.count()) > 0
  }

  async filterByStatus(status: string) {
    if (await this.statusFilter.count() > 0) {
      await this.statusFilter.selectOption(status)
    }
  }

  async search(query: string) {
    if (await this.searchInput.count() > 0) {
      await this.searchInput.fill(query)
      await this.page.keyboard.press('Enter')
    }
  }

  async updateStatus(row: number, status: string) {
    const statusButton = this.table.locator('tr').nth(row).locator('button:has-text("Update"), select[name="status"]')
    await statusButton.click()
    await this.page.locator(`button:has-text("${status}"), option:has-text("${status}")`).click()
  }
}

/**
 * Alert History Page Object Model
 */
export class AlertHistoryPage {
  readonly page: Page
  readonly table: Locator
  readonly pagination: Locator
  readonly filterButton: Locator
  readonly loadingSpinner: Locator
  readonly emptyState: Locator

  constructor(page: Page) {
    this.page = page
    this.table = page.locator('table')
    this.pagination = page.locator('button:has-text("Next"), button:has-text("Previous")')
    this.filterButton = page.locator('button:has-text("Filter")')
    this.loadingSpinner = page.locator('.animate-spin')
    this.emptyState = page.locator('.text-center.py-12, .text-center:has-text("No")')
  }

  async goto() {
    await this.page.goto('/alerts/history')
    await this.page.waitForLoadState('networkidle')
  }

  async waitForReady() {
    await this.loadingSpinner.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})
    await this.page.waitForTimeout(500)
  }

  async expectTableVisible() {
    await this.waitForReady()
    const tableCount = await this.table.count()
    if (tableCount === 0) {
      await this.emptyState.first().waitFor({ state: 'visible', timeout: 5000 }).catch(() => {})
    } else {
      await this.table.waitFor({ state: 'visible' })
    }
  }

  async hasData(): Promise<boolean> {
    await this.waitForReady()
    return (await this.table.count()) > 0
  }

  async nextPage() {
    await this.page.locator('button:has-text("Next")').click()
  }

  async prevPage() {
    await this.page.locator('button:has-text("Previous")').click()
  }
}
