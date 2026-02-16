/**
 * Sessions Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class SessionsPage {
  readonly page: Page
  readonly table: Locator
  readonly currentSessionIndicator: Locator
  readonly revokeButtons: Locator

  constructor(page: Page) {
    this.page = page
    this.table = page.locator('table')
    // Current session is typically indicated by text like "current" or "(this device)"
    this.currentSessionIndicator = page.locator('text=/current/i, text=/this device/i, text=/this session/i')
    this.revokeButtons = page.locator('button:has-text("Revoke")')
  }

  async goto() {
    await this.page.goto('/sessions')
    await this.page.waitForLoadState('networkidle')
  }

  async expectTableVisible() {
    await this.table.waitFor({ state: 'visible' })
  }

  async expectCurrentSessionMarked() {
    await this.currentSessionIndicator.first().waitFor({ state: 'visible' })
  }

  async getSessionCount(): Promise<number> {
    return await this.table.locator('tbody tr').count()
  }

  async revokeSession(row: number) {
    const revokeButton = this.table.locator('tr').nth(row).locator('button:has-text("Revoke")')
    await revokeButton.click()
    // Confirm if there's a confirmation dialog
    const confirmButton = this.page.locator('.fixed button:has-text("Confirm"), .fixed button:has-text("Revoke")')
    if (await confirmButton.count() > 0) {
      await confirmButton.click()
    }
  }

  async revokeOtherSessions() {
    // Revoke all sessions except the current one
    const count = await this.revokeButtons.count()
    for (let i = 0; i < count; i++) {
      const button = this.revokeButtons.nth(0) // Always get first as list shrinks
      if (await button.isVisible()) {
        await button.click()
        const confirmButton = this.page.locator('.fixed button:has-text("Confirm"), .fixed button:has-text("Revoke")')
        if (await confirmButton.count() > 0) {
          await confirmButton.click()
        }
      }
    }
  }
}
