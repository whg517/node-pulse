/**
 * Nodes Page Object Model (Node Management)
 */
import { Page, Locator } from '@playwright/test'

export class NodesPage {
  readonly page: Page
  readonly table: Locator
  readonly createButton: Locator
  readonly searchInput: Locator
  readonly editButtons: Locator
  readonly deleteButtons: Locator
  readonly loadingSpinner: Locator
  readonly emptyState: Locator

  // Form fields
  readonly nameInput: Locator
  readonly regionInput: Locator
  readonly submitButton: Locator
  readonly cancelButton: Locator

  // Dialog
  readonly confirmDeleteButton: Locator
  readonly modal: Locator

  constructor(page: Page) {
    this.page = page
    // Table uses standard table element with Tailwind classes
    this.table = page.locator('table')
    // Create button - could be "Create" or "Add" text
    this.createButton = page.locator('button:has-text("Create"), button:has-text("Add")')
    // Search input
    this.searchInput = page.locator('input[type="search"], input[placeholder*="search" i]')
    // Edit/Delete buttons in table rows
    this.editButtons = page.locator('button:has-text("Edit"), a:has-text("Edit")')
    this.deleteButtons = page.locator('button:has-text("Delete")')
    this.loadingSpinner = page.locator('.animate-spin')
    this.emptyState = page.locator('.text-center.py-12, .text-center:has-text("No"), .bg-white.rounded-lg.shadow-sm.p-6:has-text("No")')

    // Form fields - use ID selectors for reliability
    this.nameInput = page.locator('#name, input[name="name"]')
    this.regionInput = page.locator('#region, input[name="region"], select[name="region"]')
    this.submitButton = page.locator('button[type="submit"]')
    this.cancelButton = page.locator('button:has-text("Cancel")')

    // Modal dialog
    this.confirmDeleteButton = page.locator('.fixed button:has-text("Delete")')
    this.modal = page.locator('.fixed.inset-0')
  }

  async goto() {
    await this.page.goto('/nodes')
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

  async expectCreateButtonVisible() {
    await this.createButton.first().waitFor({ state: 'visible' })
  }

  async expectCreateButtonHidden() {
    await this.createButton.first().waitFor({ state: 'hidden' })
  }

  async clickCreate() {
    await this.createButton.first().click()
    await this.modal.waitFor({ state: 'visible' })
  }

  async createNode(name: string, region: string) {
    await this.clickCreate()
    await this.nameInput.fill(name)
    if (await this.regionInput.count() > 0) {
      await this.regionInput.fill(region)
    }
    await this.submitButton.click()
    await this.modal.waitFor({ state: 'hidden' })
  }

  async editNode(row: number, name: string) {
    const editButton = this.table.locator('tr').nth(row).locator('button:has-text("Edit")')
    await editButton.click()
    await this.modal.waitFor({ state: 'visible' })
    await this.nameInput.fill(name)
    await this.submitButton.click()
    await this.modal.waitFor({ state: 'hidden' })
  }

  async deleteNode(row: number) {
    const deleteButton = this.table.locator('tr').nth(row).locator('button:has-text("Delete")')
    await deleteButton.click()
    await this.confirmDeleteButton.click()
  }

  async search(query: string) {
    if (await this.searchInput.count() > 0) {
      await this.searchInput.fill(query)
      await this.page.keyboard.press('Enter')
    }
  }

  async getRowCount(): Promise<number> {
    return await this.table.locator('tbody tr').count()
  }

  async hasNode(name: string): Promise<boolean> {
    const cell = this.table.locator(`td:has-text("${name}")`)
    return (await cell.count()) > 0
  }
}
