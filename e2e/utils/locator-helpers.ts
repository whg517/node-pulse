/**
 * Locator Helpers
 *
 * Provides helper functions for creating resilient locators
 * following Playwright best practices.
 * 
 * Priority order for locators:
 * 1. getByRole - Most resilient, uses ARIA roles
 * 2. getByTestId - Developer-controlled, stable
 * 3. getByText / getByLabel - User-visible text
 * 4. getByPlaceholder - Form fields
 * 5. getByAltText - Images
 * 6. CSS/XPath selectors - Last resort
 */

import { Page, Locator } from '@playwright/test'

/**
 * Get button locator by text
 * Priority: data-testid > role=button > text match
 */
export function getButtonLocator(page: Page, text: string): Locator {
  return page.locator(
    `[data-testid="${text.toLowerCase().replace(/\s+/g, '-')}-button"], button:has-text("${text}")`
  )
}

/**
 * Get input field locator by label/name
 * Priority: data-testid > label > name > placeholder
 */
export function getInputLocator(page: Page, fieldName: string): Locator {
  return page.locator(
    `[data-testid="${fieldName}-input"], label:has-text("${fieldName}") input, input[name="${fieldName}"], input[placeholder*="${fieldName}" i]`
  )
}

/**
 * Get select dropdown locator
 * Priority: data-testid > label > name
 */
export function getSelectLocator(page: Page, fieldName: string): Locator {
  return page.locator(
    `[data-testid="${fieldName}-select"], label:has-text("${fieldName}") select, select[name="${fieldName}"]`
  )
}

/**
 * Get table row by index with resilient selector
 */
export function getTableRowLocator(table: Locator, index: number): Locator {
  return table.locator('tbody tr').nth(index)
}

/**
 * Get cell in row by column index or header text
 */
export function getCellLocator(
  row: Locator,
  columnIndex: number,
  columnHeader?: string
): Locator {
  if (columnHeader) {
    // Try to find cell by header text match
    return row.locator('td').filter({ hasText: new RegExp(columnHeader, 'i') }).first()
  }
  return row.locator('td').nth(columnIndex)
}

/**
 * Get action button in row
 * Priority: data-testid > text match
 */
export function getRowActionLocator(row: Locator, action: string): Locator {
  return row.locator(
    `[data-testid="${action}-button"], button:has-text("${action}")`
  )
}

/**
 * Get modal/dialog locator
 * Priority: role=dialog > data-testid > CSS
 */
export function getModalLocator(page: Page): Locator {
  return page.locator(
    '[role="dialog"], [data-testid="modal"], .fixed.inset-0, [data-testid="dialog"]'
  )
}

/**
 * Get loading indicator locator
 * Priority: data-testid > role=status > CSS
 */
export function getLoadingLocator(page: Page): Locator {
  return page.locator(
    '[data-testid="loading-spinner"], [role="status"], .animate-spin, [aria-busy="true"]'
  )
}

/**
 * Get error message locator
 * Priority: role=alert > data-testid > CSS
 */
export function getErrorLocator(page: Page): Locator {
  return page.locator(
    '[role="alert"], [data-testid="error-message"], .text-red-600, .bg-red-50, .error-message'
  )
}

/**
 * Get success/toast message locator
 * Priority: data-testid > CSS
 */
export function getToastLocator(page: Page): Locator {
  return page.locator(
    '[data-testid="toast"], .toast, [role="status"]:has-text("success"), .notification'
  )
}

/**
 * Get empty state locator
 * Priority: data-testid > text patterns
 */
export function getEmptyStateLocator(page: Page): Locator {
  return page.locator(
    '[data-testid="empty-state"], .text-center:has-text("No"), .text-center:has-text("no"), :text-is("Nothing to show")'
  )
}

/**
 * Create chained locator for nested elements
 */
export function chainLocator(
  parent: Locator,
  childSelector: string
): Locator {
  return parent.locator(childSelector)
}

/**
 * Get element within table by row text content
 */
export function getRowByText(table: Locator, text: string): Locator {
  return table.locator(`tr:has-text("${text}")`)
}

/**
 * Get checkbox by state (checked/unchecked)
 */
export function getCheckboxByState(
  container: Locator,
  state: 'checked' | 'unchecked'
): Locator {
  return state === 'checked'
    ? container.locator('input[type="checkbox"]:checked')
    : container.locator('input[type="checkbox"]:not(:checked)')
}

/**
 * Wait for locator to be visible with timeout
 */
export async function waitForLocatorVisible(
  locator: Locator,
  timeout: number = 5000
): Promise<void> {
  await locator.waitFor({ state: 'visible', timeout })
}

/**
 * Wait for locator to be hidden with timeout
 */
export async function waitForLocatorHidden(
  locator: Locator,
  timeout: number = 5000
): Promise<void> {
  await locator.waitFor({ state: 'hidden', timeout })
}
