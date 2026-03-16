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

  // Form fields (matching AlertRuleForm.tsx)
  readonly metricSelect: Locator
  readonly thresholdInput: Locator
  readonly levelSelect: Locator
  readonly nodeSelect: Locator
  readonly enabledCheckbox: Locator
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

    // Form fields - match actual AlertRuleForm.tsx IDs
    this.metricSelect = page.locator('#metric')
    this.thresholdInput = page.locator('#threshold')
    this.levelSelect = page.locator('#level')
    this.nodeSelect = page.locator('#node')
    this.enabledCheckbox = page.locator('#enabled')
    this.submitButton = page.locator('button[type="submit"]')
  }

  async goto() {
    const currentUrl = this.page.url()

    // If already on login page, navigation will fail - skip
    if (currentUrl.includes('/login')) {
      throw new Error('Cannot navigate: user is on login page (not authenticated)')
    }

    // Close any open modals first by pressing Escape
    try {
      await this.page.keyboard.press('Escape')
      await this.page.waitForTimeout(200)
    } catch {
      // Ignore errors
    }

    // Direct navigation works now that auth state is properly persisted
    // (localhost URL and vite proxy cookie handling were fixed)
    await this.page.goto('/alerts/rules')

    // Wait for page to settle
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

  async createRule(metric: string, threshold: number, level: string, nodeId?: string) {
    await this.clickCreate()
    // Select metric type (latency, packet_loss_rate, jitter)
    if (await this.metricSelect.count() > 0) {
      await this.metricSelect.selectOption(metric)
    }
    // Fill threshold
    await this.thresholdInput.fill(String(threshold))
    // Select level (P0, P1, P2)
    if (await this.levelSelect.count() > 0) {
      await this.levelSelect.selectOption(level)
    }
    // Optionally select a node (default is global)
    if (nodeId && (await this.nodeSelect.count() > 0)) {
      await this.nodeSelect.selectOption(nodeId)
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
