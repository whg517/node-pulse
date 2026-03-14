/**
 * Data Export Page Object Model
 *
 * Handles data export functionality:
 * - Export form filling
 * - Export status monitoring
 * - Download exported files
 */
import { Page, Locator, expect, Download } from '@playwright/test'
import { FormPage, FormSelectors, DEFAULT_FORM_SELECTORS } from './common/FormPage'

export interface DataExportSelectors extends FormSelectors {
  nodeSelect?: string
  timeRangeSelect?: string
  formatSelect?: string
  progressBar?: string
  downloadButton?: string
  accessWarning?: string
  exportHistoryTable?: string
}

export const DEFAULT_DATA_EXPORT_SELECTORS: DataExportSelectors = {
  ...DEFAULT_FORM_SELECTORS,
  nodeSelect: '[data-testid="node-select"], select[name="node"], select[name="nodeId"]',
  timeRangeSelect: '[data-testid="time-range-select"], select[name="timeRange"], select[name="time_range"]',
  formatSelect: '[data-testid="format-select"], select[name="format"]',
  progressBar: '[data-testid="progress-bar"], .progress-bar, [role="progressbar"]',
  downloadButton: '[data-testid="download-button"], button:has-text("Download"), a:has-text("Download")',
  accessWarning: '[data-testid="access-warning"], .bg-yellow-50:has-text("Admin"), .access-warning',
  exportHistoryTable: '[data-testid="export-history-table"], table:has-text("Export History")',
}

export class DataExportPage extends FormPage {
  readonly nodeSelect: Locator
  readonly timeRangeSelect: Locator
  readonly formatSelect: Locator
  readonly progressBar: Locator
  readonly downloadButton: Locator
  readonly accessWarning: Locator
  readonly exportHistoryTable: Locator

  constructor(page: Page, selectors: DataExportSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_DATA_EXPORT_SELECTORS, ...selectors }

    this.nodeSelect = page.locator(mergedSelectors.nodeSelect!)
    this.timeRangeSelect = page.locator(mergedSelectors.timeRangeSelect!)
    this.formatSelect = page.locator(mergedSelectors.formatSelect!)
    this.progressBar = page.locator(mergedSelectors.progressBar!)
    this.downloadButton = page.locator(mergedSelectors.downloadButton!)
    this.accessWarning = page.locator(mergedSelectors.accessWarning!)
    this.exportHistoryTable = page.locator(mergedSelectors.exportHistoryTable!)
  }

  /**
   * Navigate to export page
   */
  async goto(): Promise<void> {
    await super.goto('/reports/history')
    await this.waitForReady()
  }

  /**
   * Expect access warning (for non-admin users)
   */
  async expectAccessWarning(): Promise<void> {
    await this.accessWarning.first().waitFor({ state: 'visible' })
  }

  /**
   * Check if access warning is visible
   */
  async hasAccessWarning(): Promise<boolean> {
    return (await this.accessWarning.count()) > 0
  }

  /**
   * Select node for export
   */
  async selectNode(nodeName: string): Promise<void> {
    await this.selectOption('node', nodeName)
  }

  /**
   * Select node by value
   */
  async selectNodeByValue(nodeValue: string): Promise<void> {
    await this.nodeSelect.selectOption({ value: nodeValue })
  }

  /**
   * Select time range
   */
  async selectTimeRange(range: string): Promise<void> {
    await this.selectOption('timeRange', range)
  }

  /**
   * Select export format
   */
  async selectFormat(format: string): Promise<void> {
    await this.selectOption('format', format)
  }

  /**
   * Submit export request
   */
  async submitExport(): Promise<void> {
    await this.submit()
  }

  /**
   * Create export and wait for completion
   */
  async createExport(nodeName: string, timeRange: string, format: string, timeout = 60000): Promise<void> {
    await this.selectNode(nodeName)
    await this.selectTimeRange(timeRange)
    await this.selectFormat(format)
    await this.submit()
    await this.waitForExportComplete(timeout)
  }

  /**
   * Wait for export to complete
   */
  async waitForExportComplete(timeout = 60000): Promise<void> {
    // Wait for progress bar to appear and disappear
    if (await this.progressBar.count() > 0) {
      await this.progressBar.first().waitFor({ state: 'visible', timeout })
      await this.progressBar.first().waitFor({ state: 'hidden', timeout })
    }

    // Wait for download button or success message
    await this.downloadButton.first().waitFor({ state: 'visible', timeout }).catch(() => {})
  }

  /**
   * Download exported file
   */
  async download(): Promise<Download> {
    const [download] = await Promise.all([
      this.page.waitForEvent('download'),
      this.downloadButton.first().click(),
    ])
    return download
  }

  /**
   * Download and save file
   */
  async downloadAndSave(savePath: string): Promise<void> {
    const download = await this.download()
    await download.saveAs(savePath)
  }

  /**
   * Check if download button is visible
   */
  async isDownloadReady(): Promise<boolean> {
    return await this.downloadButton.first().isVisible()
  }

  /**
   * Wait for download button to appear
   */
  async waitForDownloadReady(timeout = 60000): Promise<void> {
    await this.downloadButton.first().waitFor({ state: 'visible', timeout })
  }

  /**
   * Get export history row count
   */
  async getExportHistoryCount(): Promise<number> {
    if (await this.exportHistoryTable.count() === 0) {
      return 0
    }
    return await this.exportHistoryTable.locator('tbody tr').count()
  }

  /**
   * Get export status from history
   */
  async getExportStatus(rowIndex: number): Promise<string | null> {
    const statusCell = this.exportHistoryTable.locator('tbody tr').nth(rowIndex).locator('td').nth(2)
    return await statusCell.textContent()
  }

  /**
   * Check if export is complete in history
   */
  async isExportComplete(rowIndex: number): Promise<boolean> {
    const status = await this.getExportStatus(rowIndex)
    return status?.toLowerCase().includes('complete') ?? false
  }

  /**
   * Download from export history
   */
  async downloadFromHistory(rowIndex: number): Promise<Download> {
    const downloadButton = this.exportHistoryTable.locator('tbody tr').nth(rowIndex).locator('button:has-text("Download")')
    const [download] = await Promise.all([
      this.page.waitForEvent('download'),
      downloadButton.click(),
    ])
    return download
  }

  /**
   * Refresh export status
   */
  async refreshStatus(): Promise<void> {
    const refreshButton = this.page.locator('[data-testid="refresh-button"], button:has-text("Refresh")')
    if (await refreshButton.count() > 0) {
      await refreshButton.click()
      await this.waitForReady()
    }
  }

  /**
   * Assert export form is visible
   */
  async expectFormVisible(): Promise<void> {
    await expect(this.form).toBeVisible()
    await expect(this.nodeSelect).toBeVisible()
    await expect(this.timeRangeSelect).toBeVisible()
    await expect(this.formatSelect).toBeVisible()
    await expect(this.submitButton).toBeVisible()
  }

  /**
   * Assert export success
   */
  async expectExportSuccess(): Promise<void> {
    await this.waitForDownloadReady()
  }

  /**
   * Assert export history is visible
   */
  async expectExportHistoryVisible(): Promise<void> {
    await expect(this.exportHistoryTable).toBeVisible()
  }
}
