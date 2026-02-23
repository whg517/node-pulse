/**
 * Base Page Object Model
 *
 * Provides common functionality for all page objects:
 * - Navigation
 * - Loading state handling
 * - Empty state detection
 * - Common assertions
 */
import { Page, Locator, expect } from '@playwright/test'

export interface PageSelectors {
  loading?: string
  empty?: string
  error?: string
  success?: string
}

export const DEFAULT_SELECTORS: PageSelectors = {
  loading: '[data-testid="loading-spinner"], .animate-spin, [role="status"]',
  empty: '[data-testid="empty-state"], .text-center:has-text("No"), .text-center:has-text("no")',
  error: '[data-testid="error-message"], .text-red-600, .bg-red-50',
  success: '[data-testid="success-message"], .text-green-600, .bg-green-50',
}

export abstract class BasePage {
  readonly page: Page
  protected selectors: PageSelectors

  // Common locators
  readonly loadingSpinner: Locator
  readonly emptyState: Locator
  readonly errorMessage: Locator
  readonly successMessage: Locator

  constructor(page: Page, selectors: PageSelectors = {}) {
    this.page = page
    this.selectors = { ...DEFAULT_SELECTORS, ...selectors }

    this.loadingSpinner = page.locator(this.selectors.loading!)
    this.emptyState = page.locator(this.selectors.empty!)
    this.errorMessage = page.locator(this.selectors.error!)
    this.successMessage = page.locator(this.selectors.success!)
  }

  /**
   * Navigate to page path
   */
  async goto(path: string): Promise<void> {
    await this.page.goto(path, { waitUntil: 'networkidle' })
    await this.page.waitForLoadState('domcontentloaded')
  }

  /**
   * Wait for page to be ready (loading complete)
   */
  async waitForReady(timeout = 10000): Promise<void> {
    await this.loadingSpinner.waitFor({ state: 'hidden', timeout }).catch(() => {})
    await this.page.waitForTimeout(300)
  }

  /**
   * Wait for loading state to appear
   */
  async waitForLoading(timeout = 5000): Promise<void> {
    await this.loadingSpinner.waitFor({ state: 'visible', timeout })
  }

  /**
   * Wait for loading state to disappear
   */
  async waitForLoadingComplete(timeout = 10000): Promise<void> {
    await this.loadingSpinner.waitFor({ state: 'hidden', timeout })
  }

  /**
   * Check if empty state is visible
   */
  async isEmptyStateVisible(): Promise<boolean> {
    return await this.emptyState.first().isVisible().catch(() => false)
  }

  /**
   * Check if error message is visible
   */
  async isErrorMessageVisible(): Promise<boolean> {
    return await this.errorMessage.first().isVisible().catch(() => false)
  }

  /**
   * Get error message text
   */
  async getErrorMessage(): Promise<string | null> {
    const visible = await this.isErrorMessageVisible()
    if (!visible) return null
    return await this.errorMessage.first().textContent()
  }

  /**
   * Assert page title
   */
  async expectTitle(title: string): Promise<void> {
    await expect(this.page).toHaveTitle(new RegExp(title, 'i'))
  }

  /**
   * Assert page URL contains path
   */
  async expectUrlContains(path: string): Promise<void> {
    await expect(this.page).toHaveURL(new RegExp(path))
  }

  /**
   * Assert element is visible
   */
  async expectVisible(locator: Locator): Promise<void> {
    await expect(locator).toBeVisible()
  }

  /**
   * Assert element is hidden
   */
  async expectHidden(locator: Locator): Promise<void> {
    await expect(locator).toBeHidden()
  }

  /**
   * Take screenshot for visual regression
   */
  async takeScreenshot(name: string, options?: { fullPage?: boolean; maxDiffPixels?: number }): Promise<void> {
    await expect(this.page).toHaveScreenshot(name, {
      fullPage: options?.fullPage ?? false,
      maxDiffPixels: options?.maxDiffPixels ?? 100,
    })
  }

  /**
   * Wait for network to be idle
   */
  async waitForNetworkIdle(timeout = 30000): Promise<void> {
    await this.page.waitForLoadState('networkidle', { timeout })
  }

  /**
   * Wait for toast notification
   */
  async waitForToast(message: string, timeout = 5000): Promise<void> {
    const toast = this.page.locator(`[data-testid="toast"], .toast:has-text("${message}")`)
    await toast.waitFor({ state: 'visible', timeout })
  }

  /**
   * Dismiss toast notification
   */
  async dismissToast(): Promise<void> {
    const closeButton = this.page.locator('[data-testid="toast-close"], .toast-close')
    if (await closeButton.count() > 0) {
      await closeButton.click()
    }
  }
}
