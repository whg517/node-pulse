/**
 * Table Page Object Model
 *
 * Extends BasePage with table-specific functionality:
 * - Table row operations
 * - Search and filtering
 * - Pagination
 * - Sorting
 * - Row actions (edit, delete)
 */
import { Page, Locator, expect } from '@playwright/test'
import { BasePage, PageSelectors, DEFAULT_SELECTORS } from './BasePage'

export interface TableSelectors extends PageSelectors {
  table?: string
  row?: string
  cell?: string
  search?: string
  pagination?: string
  editButton?: string
  deleteButton?: string
  actionButton?: string
  checkbox?: string
  header?: string
}

export const DEFAULT_TABLE_SELECTORS: TableSelectors = {
  ...DEFAULT_SELECTORS,
  table: '[data-testid="table"], table',
  row: 'tbody tr',
  cell: 'td',
  search: '[data-testid="search-input"], input[type="search"], input[placeholder*="search" i]',
  pagination: '[data-testid="pagination"], .pagination',
  editButton: '[data-testid="edit-button"], button:has-text("Edit")',
  deleteButton: '[data-testid="delete-button"], button:has-text("Delete")',
  actionButton: '[data-testid="action-button"]',
  checkbox: 'input[type="checkbox"]',
  header: 'thead th',
}

export abstract class TablePage extends BasePage {
  readonly table: Locator
  readonly tableBody: Locator
  readonly tableHead: Locator
  readonly searchInput: Locator
  readonly pagination: Locator
  readonly editButtons: Locator
  readonly deleteButtons: Locator
  readonly checkboxes: Locator
  protected selectors!: TableSelectors

  constructor(page: Page, selectors: TableSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_TABLE_SELECTORS, ...selectors }
    this.selectors = mergedSelectors

    this.table = page.locator(mergedSelectors.table!)
    this.tableBody = page.locator(`${mergedSelectors.table!} tbody`)
    this.tableHead = page.locator(`${mergedSelectors.table!} thead`)
    this.searchInput = page.locator(mergedSelectors.search!)
    this.pagination = page.locator(mergedSelectors.pagination!)
    this.editButtons = page.locator(mergedSelectors.editButton!)
    this.deleteButtons = page.locator(mergedSelectors.deleteButton!)
    this.checkboxes = page.locator(mergedSelectors.checkbox!)
  }

  /**
   * Get table row count
   */
  async getRowCount(): Promise<number> {
    if (await this.isEmptyStateVisible()) {
      return 0
    }
    return await this.table.locator(this.selectors.row!).count()
  }

  /**
   * Check if table has data
   */
  async hasData(): Promise<boolean> {
    const count = await this.getRowCount()
    return count > 0
  }

  /**
   * Get row by index
   */
  getRow(index: number): Locator {
    return this.table.locator(this.selectors.row!).nth(index)
  }

  /**
   * Get cell value by row and column
   */
  async getCellValue(rowIndex: number, columnIndex: number): Promise<string | null> {
    const cell = this.getRow(rowIndex).locator('td').nth(columnIndex)
    return await cell.textContent()
  }

  /**
   * Get cell by row and column header text
   */
  async getCellByColumnHeader(rowIndex: number, columnHeader: string): Promise<string | null> {
    const headerIndex = await this.getColumnIndex(columnHeader)
    if (headerIndex === -1) return null
    return await this.getCellValue(rowIndex, headerIndex)
  }

  /**
   * Get column index by header text
   */
  async getColumnIndex(headerText: string | RegExp): Promise<number> {
    const headers = this.table.locator('thead th')
    const count = await headers.count()

    for (let i = 0; i < count; i++) {
      const header = await headers.nth(i).textContent()
      if (header && (typeof headerText === 'string' ? header.includes(headerText) : headerText.test(header))) {
        return i
      }
    }
    return -1
  }

  /**
   * Search in table
   */
  async search(query: string): Promise<void> {
    if (await this.searchInput.count() > 0) {
      await this.searchInput.fill(query)
      await this.page.keyboard.press('Enter')
      await this.waitForReady()
    }
  }

  /**
   * Clear search
   */
  async clearSearch(): Promise<void> {
    if (await this.searchInput.count() > 0) {
      await this.searchInput.clear()
      await this.page.keyboard.press('Enter')
      await this.waitForReady()
    }
  }

  /**
   * Click edit button on row
   */
  async clickEdit(rowIndex: number): Promise<void> {
    const editButton = this.getRow(rowIndex).locator(this.selectors.editButton!)
    await editButton.click()
  }

  /**
   * Click delete button on row
   */
  async clickDelete(rowIndex: number): Promise<void> {
    const deleteButton = this.getRow(rowIndex).locator(this.selectors.deleteButton!)
    await deleteButton.click()
  }

  /**
   * Click action button on row
   */
  async clickAction(rowIndex: number, actionText: string): Promise<void> {
    const actionButton = this.getRow(rowIndex).locator(`button:has-text("${actionText}")`)
    await actionButton.click()
  }

  /**
   * Toggle checkbox on row
   */
  async toggleCheckbox(rowIndex: number): Promise<void> {
    const checkbox = this.getRow(rowIndex).locator(this.selectors.checkbox!)
    await checkbox.click()
  }

  /**
   * Toggle all checkboxes
   */
  async toggleAll(): Promise<void> {
    const selectAll = this.tableHead.locator(this.selectors.checkbox!)
    if (await selectAll.count() > 0) {
      await selectAll.click()
    }
  }

  /**
   * Check if row exists by cell text
   */
  async hasRowWithText(text: string, columnIndex?: number): Promise<boolean> {
    if (columnIndex !== undefined) {
      const cells = this.table.locator(`td:nth-child(${columnIndex + 1}):has-text("${text}")`)
      return (await cells.count()) > 0
    }

    const cells = this.table.locator(`td:has-text("${text}")`)
    return (await cells.count()) > 0
  }

  /**
   * Get all row texts in specific column
   */
  async getColumnTexts(columnIndex: number): Promise<string[]> {
    const cells = this.table.locator('tbody tr td').filter({ hasText: /.+/ })
    const texts: string[] = []
    const count = await cells.count()

    for (let i = 0; i < count; i++) {
      const cell = cells.nth(i)
      const text = await cell.textContent()
      if (text) {
        texts.push(text.trim())
      }
    }

    return texts
  }

  /**
   * Sort by column
   */
  async sortByColumn(columnHeader: string): Promise<void> {
    const header = this.table.locator('thead th').getByText(columnHeader)
    await header.click()
    await this.waitForReady()
  }

  /**
   * Get pagination info
   */
  async getPaginationInfo(): Promise<{ currentPage: number; totalPages: number; totalItems: number } | null> {
    if (await this.pagination.count() === 0) {
      return null
    }

    const paginationText = await this.pagination.textContent()
    const match = paginationText?.match(/Page\s+(\d+)\s+of\s+(\d+)/i)

    if (match) {
      return {
        currentPage: parseInt(match[1], 10),
        totalPages: parseInt(match[2], 10),
        totalItems: 0,
      }
    }

    return null
  }

  /**
   * Go to next page
   */
  async nextPage(): Promise<void> {
    const nextButton = this.pagination.locator('button:has-text("Next"), [aria-label="Next page"]')
    if (await nextButton.count() > 0) {
      await nextButton.click()
      await this.waitForReady()
    }
  }

  /**
   * Go to previous page
   */
  async prevPage(): Promise<void> {
    const prevButton = this.pagination.locator('button:has-text("Previous"), [aria-label="Previous page"]')
    if (await prevButton.count() > 0) {
      await prevButton.click()
      await this.waitForReady()
    }
  }

  /**
   * Go to specific page
   */
  async goToPage(pageNumber: number): Promise<void> {
    const pageButton = this.pagination.locator(`button:has-text("${pageNumber}"), [aria-label="Page ${pageNumber}"]`)
    if (await pageButton.count() > 0) {
      await pageButton.click()
      await this.waitForReady()
    }
  }

  /**
   * Assert table is visible
   */
  async expectTableVisible(): Promise<void> {
    await this.waitForReady()
    const tableCount = await this.table.count()
    if (tableCount === 0) {
      await this.emptyState.first().waitFor({ state: 'visible', timeout: 5000 }).catch(() => {})
    } else {
      await this.table.waitFor({ state: 'visible' })
    }
  }

  /**
   * Assert table headers
   */
  async expectHeaders(expectedHeaders: string[]): Promise<void> {
    const headers = this.table.locator('thead th')
    for (let i = 0; i < expectedHeaders.length; i++) {
      const header = await headers.nth(i).textContent()
      expect(header?.toLowerCase().trim()).toBe(expectedHeaders[i].toLowerCase().trim())
    }
  }

  /**
   * Wait for row to appear
   */
  async waitForRow(text: string, timeout = 10000): Promise<void> {
    const row = this.table.locator(`tr:has-text("${text}")`)
    await row.waitFor({ state: 'visible', timeout })
  }

  /**
   * Wait for row to disappear
   */
  async waitForRowToDisappear(text: string, timeout = 10000): Promise<void> {
    const row = this.table.locator(`tr:has-text("${text}")`)
    await row.waitFor({ state: 'hidden', timeout })
  }

  /**
   * Get row index for a locator
   */
  protected async getRowIndex(rowLocator: Locator): Promise<number> {
    const rows = this.table.locator('tbody tr')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i)
      const rowElement = await row.elementHandle()
      const locatorElement = await rowLocator.elementHandle()

      if (rowElement && locatorElement && rowElement === locatorElement) {
        return i
      }
    }
    return -1
  }
}
