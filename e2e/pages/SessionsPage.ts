/**
 * Sessions Page Object Model
 *
 * Handles session management functionality:
 * - View active sessions
 * - Revoke sessions
 * - Current session indicator
 */
import { Page, Locator, expect } from '@playwright/test'
import { TablePage, TableSelectors, DEFAULT_TABLE_SELECTORS } from './common/TablePage'

export interface SessionsSelectors extends TableSelectors {
  currentSessionIndicator?: string
  revokeButton?: string
  deviceColumn?: string
  browserColumn?: string
  ipColumn?: string
  lastActiveColumn?: string
}

export const DEFAULT_SESSIONS_SELECTORS: SessionsSelectors = {
  ...DEFAULT_TABLE_SELECTORS,
  currentSessionIndicator: '[data-testid="current-session"], .current-session, text=/current/i, text=/this device/i, text=/this session/i',
  revokeButton: '[data-testid="revoke-button"], button:has-text("Revoke")',
  deviceColumn: 'text=/Device|Browser|Platform/i',
  browserColumn: 'text=/Browser|User Agent/i',
  ipColumn: 'text=/IP|Address/i',
  lastActiveColumn: 'text=/Last Active|Last seen|Time/i',
}

export class SessionsPage extends TablePage {
  readonly currentSessionIndicator: Locator
  readonly revokeButtons: Locator

  constructor(page: Page, selectors: SessionsSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_SESSIONS_SELECTORS, ...selectors }

    this.currentSessionIndicator = page.locator(mergedSelectors.currentSessionIndicator!)
    this.revokeButtons = page.locator(mergedSelectors.revokeButton!)
  }

  /**
   * Navigate to sessions page
   */
  async goto(): Promise<void> {
    await super.goto('/sessions')
    await this.waitForReady()
  }

  /**
   * Get session count from table
   */
  async getSessionCount(): Promise<number> {
    return await this.getRowCount()
  }

  /**
   * Check if current session is marked
   */
  async expectCurrentSessionMarked(): Promise<void> {
    await this.currentSessionIndicator.first().waitFor({ state: 'visible', timeout: 5000 })
  }

  /**
   * Find current session row index
   */
  async getCurrentSessionRowIndex(): Promise<number> {
    const rows = this.table.locator('tbody tr')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i)
      const text = await row.textContent()
      if (text?.toLowerCase().includes('current') || text?.toLowerCase().includes('this device')) {
        return i
      }
    }
    return -1
  }

  /**
   * Revoke session by row index
   */
  async revokeSession(rowIndex: number): Promise<void> {
    const revokeButton = this.getRow(rowIndex).locator('[data-testid="revoke-button"], button:has-text("Revoke")')
    await revokeButton.click()

    // Look for confirmation dialog
    const confirmButton = this.page.locator(
      '[data-testid="confirm-button"], .fixed button:has-text("Confirm"), .fixed button:has-text("Revoke")'
    )
    if (await confirmButton.count() > 0) {
      await confirmButton.click()
    }

    await this.waitForReady()
  }

  /**
   * Revoke all other sessions (not current)
   */
  async revokeOtherSessions(): Promise<void> {
    const currentRowIndex = await this.getCurrentSessionRowIndex()

    // Get all revoke buttons
    const revokeButtons = this.table.locator('[data-testid="revoke-button"], button:has-text("Revoke")')
    const count = await revokeButtons.count()

    for (let i = count - 1; i >= 0; i--) {
      const button = revokeButtons.nth(i)
      if (await button.isVisible()) {
        await button.click()

        // Confirm if needed
        const confirmButton = this.page.locator(
          '[data-testid="confirm-button"], .fixed button:has-text("Confirm"), .fixed button:has-text("Revoke")'
        )
        if (await confirmButton.count() > 0) {
          await confirmButton.click()
        }

        await this.waitForReady()
      }
    }
  }

  /**
   * Check if revoke button is disabled for current session
   */
  async isCurrentSessionRevokeDisabled(): Promise<boolean> {
    const currentRowIndex = await this.getCurrentSessionRowIndex()
    if (currentRowIndex === -1) return false

    const revokeButton = this.getRow(currentRowIndex).locator('[data-testid="revoke-button"], button:has-text("Revoke")')
    if (await revokeButton.count() === 0) return true

    return await revokeButton.isDisabled()
  }

  /**
   * Expect cannot revoke current session
   */
  async expectCannotRevokeCurrentSession(): Promise<void> {
    const currentRowIndex = await this.getCurrentSessionRowIndex()
    if (currentRowIndex === -1) return

    const revokeButton = this.getRow(currentRowIndex).locator('[data-testid="revoke-button"], button:has-text("Revoke")')
    if (await revokeButton.count() > 0) {
      await expect(revokeButton).toBeDisabled()
    }
  }

  /**
   * Get session info by row
   */
  async getSessionInfo(rowIndex: number): Promise<{
    device: string
    browser: string
    ip: string
    lastActive: string
    isCurrent: boolean
  } | null> {
    const row = this.getRow(rowIndex)
    const text = await row.textContent()

    if (!text) return null

    return {
      device: text.trim(),
      browser: '',
      ip: '',
      lastActive: '',
      isCurrent: text.toLowerCase().includes('current') || text.toLowerCase().includes('this device'),
    }
  }

  /**
   * Wait for session to be revoked (row to disappear)
   */
  async waitForSessionRevoked(rowIndex: number, timeout = 5000): Promise<void> {
    const row = this.getRow(rowIndex)
    await row.waitFor({ state: 'hidden', timeout })
  }

  /**
   * Assert sessions table has expected columns
   */
  async expectHasColumns(expectedColumns: string[]): Promise<void> {
    const headerText = await this.tableHead.textContent()
    for (const column of expectedColumns) {
      expect(headerText).toMatch(new RegExp(column, 'i'))
    }
  }
}
