/**
 * Webhooks Page Object Model
 *
 * Handles webhook management:
 * - View webhooks
 * - Create, edit, delete webhooks
 * - Enable/disable webhooks
 */
import { Page, Locator, expect } from '@playwright/test'
import { TablePage, TableSelectors, DEFAULT_TABLE_SELECTORS } from './common/TablePage'
import { ModalPage, ModalSelectors, DEFAULT_MODAL_SELECTORS } from './common/ModalPage'

export interface WebhooksSelectors extends TableSelectors, ModalSelectors {
  createButton?: string
  nameInput?: string
  urlInput?: string
  secretInput?: string
  eventsSelect?: string
  checkboxToggle?: string
  confirmDeleteButton?: string
  accessWarning?: string
}

export const DEFAULT_WEBHOOKS_SELECTORS: WebhooksSelectors = {
  ...DEFAULT_TABLE_SELECTORS,
  ...DEFAULT_MODAL_SELECTORS,
  createButton: '[data-testid="create-webhook-button"], button:has-text("Add Webhook"), button:has-text("Create")',
  nameInput: '[data-testid="webhook-name-input"], #name, input[name="name"]',
  urlInput: '[data-testid="webhook-url-input"], #url, input[name="url"], input[type="url"]',
  secretInput: '[data-testid="webhook-secret-input"], #secret, input[name="secret"], input[type="password"]',
  eventsSelect: '[data-testid="webhook-events-select"], select[name="events"], select[name="eventTypes"]',
  checkboxToggle: '[data-testid="webhook-toggle"], input[type="checkbox"]',
  confirmDeleteButton: '[data-testid="confirm-delete-button"], .fixed button:has-text("Delete")',
  accessWarning: '[data-testid="access-warning"], .bg-amber-50:has-text("Admin"), .access-warning',
}

export class WebhooksPage extends TablePage {
  readonly createButton: Locator
  readonly nameInput: Locator
  readonly urlInput: Locator
  readonly secretInput: Locator
  readonly eventsSelect: Locator
  readonly checkboxToggle: Locator
  readonly modal: Locator
  readonly submitButton: Locator
  readonly confirmDeleteButton: Locator
  readonly accessWarning: Locator

  constructor(page: Page, selectors: WebhooksSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_WEBHOOKS_SELECTORS, ...selectors }

    this.createButton = page.locator(mergedSelectors.createButton!)
    this.nameInput = page.locator(mergedSelectors.nameInput!)
    this.urlInput = page.locator(mergedSelectors.urlInput!)
    this.secretInput = page.locator(mergedSelectors.secretInput!)
    this.eventsSelect = page.locator(mergedSelectors.eventsSelect!)
    this.checkboxToggle = page.locator(mergedSelectors.checkboxToggle!)
    this.modal = page.locator(mergedSelectors.modal!)
    this.submitButton = page.locator(mergedSelectors.submitButton!)
    this.confirmDeleteButton = page.locator(mergedSelectors.confirmDeleteButton!)
    this.accessWarning = page.locator(mergedSelectors.accessWarning!)
  }

  async hasAccessWarning(): Promise<boolean> {
    return (await this.accessWarning.count()) > 0
  }

  async waitForModalOpen(timeout = 5000): Promise<void> {
    await this.modal.first().waitFor({ state: 'visible', timeout })
  }

  async waitForModalClose(timeout = 5000): Promise<void> {
    await this.modal.first().waitFor({ state: 'hidden', timeout })
  }

  async submit(): Promise<void> {
    await this.submitButton.click()
  }

  /**
   * Navigate to webhooks page
   */
  async goto(): Promise<void> {
    await super.goto('/integrations/webhooks')
    await this.waitForReady()
  }

  /**
   * Click create button
   */
  async clickCreate(): Promise<void> {
    await this.createButton.first().click()
    await this.waitForModalOpen()
  }

  /**
   * Create webhook
   */
  async createWebhook(name: string, url: string, events?: string[]): Promise<void> {
    await this.clickCreate()
    await this.nameInput.fill(name)
    await this.urlInput.fill(url)

    if (events && events.length > 0 && await this.eventsSelect.count() > 0) {
      for (const event of events) {
        await this.eventsSelect.selectOption(event)
      }
    }

    await this.submit()
    await this.waitForModalClose()
  }

  /**
   * Toggle webhook by row index
   */
  async toggleWebhook(rowIndex: number): Promise<void> {
    const toggle = this.getRow(rowIndex).locator('[data-testid="webhook-toggle"], input[type="checkbox"]')
    await toggle.click()
  }

  /**
   * Enable webhook by row index
   */
  async enableWebhook(rowIndex: number): Promise<void> {
    const toggle = this.getRow(rowIndex).locator('[data-testid="webhook-toggle"], input[type="checkbox"]')
    const isChecked = await toggle.isChecked()
    if (!isChecked) {
      await toggle.click()
    }
  }

  /**
   * Disable webhook by row index
   */
  async disableWebhook(rowIndex: number): Promise<void> {
    const toggle = this.getRow(rowIndex).locator('[data-testid="webhook-toggle"], input[type="checkbox"]')
    const isChecked = await toggle.isChecked()
    if (isChecked) {
      await toggle.click()
    }
  }

  /**
   * Check if webhook is enabled
   */
  async isWebhookEnabled(rowIndex: number): Promise<boolean> {
    const toggle = this.getRow(rowIndex).locator('[data-testid="webhook-toggle"], input[type="checkbox"]')
    return await toggle.isChecked()
  }

  /**
   * Get webhook URL from row
   */
  async getWebhookUrl(rowIndex: number): Promise<string | null> {
    const urlCell = this.getRow(rowIndex).locator('td').nth(1)
    return await urlCell.textContent()
  }

  /**
   * Get webhook status from row
   */
  async getWebhookStatus(rowIndex: number): Promise<string | null> {
    const statusCell = this.getRow(rowIndex).locator('td').nth(2)
    return await statusCell.textContent()
  }

  /**
   * Check if webhook is active by status text
   */
  async isWebhookActive(rowIndex: number): Promise<boolean> {
    const status = await this.getWebhookStatus(rowIndex)
    return status?.toLowerCase().includes('active') ?? false
  }

  /**
   * Expect create button visible
   */
  async expectCreateButtonVisible(): Promise<void> {
    await expect(this.createButton.first()).toBeVisible()
  }

  /**
   * Expect webhook exists by name
   */
  async expectWebhookExists(name: string): Promise<void> {
    const exists = await this.hasRowWithText(name)
    expect(exists).toBeTruthy()
  }

  /**
   * Expect webhook does not exist
   */
  async expectWebhookNotExists(name: string): Promise<void> {
    const exists = await this.hasRowWithText(name)
    expect(exists).toBeFalsy()
  }

  /**
   * Wait for webhook to appear
   */
  async waitForWebhookToAppear(name: string, timeout = 10000): Promise<void> {
    await this.waitForRow(name, timeout)
  }

  /**
   * Delete webhook and wait for it to disappear
   */
  async deleteWebhookAndWait(rowIndex: number, webhookName: string): Promise<void> {
    await this.clickDelete(rowIndex)
    await this.confirmDeleteButton.click()
    await this.waitForReady()
    await this.waitForRowToDisappear(webhookName)
  }

  /**
   * Test webhook connection
   */
  async testWebhook(rowIndex: number): Promise<void> {
    const testButton = this.getRow(rowIndex).locator('button:has-text("Test"), button:has-text("Ping")')
    if (await testButton.count() > 0) {
      await testButton.click()
      await this.waitForReady()
    }
  }

  /**
   * Get webhook row index by URL
   */
  async getWebhookRowIndexByUrl(url: string): Promise<number> {
    const rows = this.table.locator('tbody tr')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const rowUrl = await this.getWebhookUrl(i)
      if (rowUrl?.includes(url)) {
        return i
      }
    }
    return -1
  }
}
