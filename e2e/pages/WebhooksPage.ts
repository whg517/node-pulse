/**
 * Webhooks Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class WebhooksPage {
  readonly page: Page
  readonly table: Locator
  readonly createButton: Locator
  readonly editButtons: Locator
  readonly deleteButtons: Locator
  readonly toggleButtons: Locator
  readonly modal: Locator
  readonly accessWarning: Locator

  // Form fields
  readonly nameInput: Locator
  readonly urlInput: Locator
  readonly secretInput: Locator
  readonly eventsSelect: Locator
  readonly submitButton: Locator

  constructor(page: Page) {
    this.page = page
    this.table = page.locator('table')
    // Create button with "Add Webhook" text
    this.createButton = page.locator('button:has-text("Add Webhook"), button:has-text("Create")')
    this.editButtons = page.locator('button:has-text("Edit")')
    this.deleteButtons = page.locator('button:has-text("Delete")')
    this.toggleButtons = page.locator('input[type="checkbox"]')
    // Modal uses fixed positioning
    this.modal = page.locator('.fixed.inset-0')
    // Access warning uses yellow background with "Admin-only" text
    this.accessWarning = page.locator('.bg-yellow-50:has-text("Admin-only"), .bg-yellow-50:has-text("admin")')

    // Form fields
    this.nameInput = page.locator('#name, input[name="name"]')
    this.urlInput = page.locator('#url, input[name="url"]')
    this.secretInput = page.locator('#secret, input[name="secret"]')
    this.eventsSelect = page.locator('select[name="events"]')
    this.submitButton = page.locator('button[type="submit"]')
  }

  async goto() {
    await this.page.goto('/webhooks')
    await this.page.waitForLoadState('networkidle')
  }

  async expectTableVisible() {
    await this.table.waitFor({ state: 'visible' })
  }

  async expectAccessWarning() {
    await this.accessWarning.waitFor({ state: 'visible' })
  }

  async hasAccessWarning(): Promise<boolean> {
    // Look for the yellow warning box with admin-only text
    const warning = this.page.locator('.bg-yellow-50')
    const count = await warning.count()
    if (count === 0) return false
    const text = await warning.first().textContent()
    return text?.toLowerCase().includes('admin') ?? false
  }

  async clickCreate() {
    await this.createButton.click()
    await this.modal.waitFor({ state: 'visible' })
  }

  async createWebhook(name: string, url: string) {
    await this.clickCreate()
    await this.nameInput.fill(name)
    await this.urlInput.fill(url)
    await this.submitButton.click()
    await this.modal.waitFor({ state: 'hidden' })
  }

  async deleteWebhook(row: number) {
    const deleteButton = this.table.locator('tr').nth(row).locator('button:has-text("Delete")')
    await deleteButton.click()
    await this.page.locator('.fixed button:has-text("Delete")').click()
  }

  async toggleWebhook(row: number) {
    const toggle = this.table.locator('tr').nth(row).locator('input[type="checkbox"]')
    await toggle.click()
  }
}
