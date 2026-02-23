/**
 * Reports Page Object Model
 *
 * Handles report generation and viewing:
 * - Generate health/performance/comparison reports
 * - View export history
 * - Download reports
 */
import { Page, Locator, expect, Download } from '@playwright/test'
import { FormPage, FormSelectors, DEFAULT_FORM_SELECTORS } from './common/FormPage'

export interface ReportsSelectors extends FormSelectors {
  reportTypeSelect?: string
  timeRangeSelect?: string
  nodeSelect?: string
  generateButton?: string
  exportHistoryTable?: string
  downloadButton?: string
  previewSection?: string
}

export const DEFAULT_REPORTS_SELECTORS: ReportsSelectors = {
  ...DEFAULT_FORM_SELECTORS,
  reportTypeSelect: '[data-testid="report-type-select"], select[name="reportType"], select[name="report_type"]',
  timeRangeSelect: '[data-testid="time-range-select"], select[name="timeRange"], select[name="time_range"]',
  nodeSelect: '[data-testid="node-select"], select[name="node"], select[name="nodeId"]',
  generateButton: '[data-testid="generate-report-button"], button:has-text("Generate"), button:has-text("Create Report")',
  exportHistoryTable: '[data-testid="export-history-table"], table:has-text("Export History")',
  downloadButton: '[data-testid="download-button"], button:has-text("Download")',
  previewSection: '[data-testid="report-preview"], .report-preview',
}

export class ReportsPage extends FormPage {
  readonly reportTypeSelect: Locator
  readonly timeRangeSelect: Locator
  readonly nodeSelect: Locator
  readonly generateButton: Locator
  readonly exportHistoryTable: Locator
  readonly downloadButton: Locator
  readonly previewSection: Locator

  constructor(page: Page, selectors: ReportsSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_REPORTS_SELECTORS, ...selectors }

    this.reportTypeSelect = page.locator(mergedSelectors.reportTypeSelect!)
    this.timeRangeSelect = page.locator(mergedSelectors.timeRangeSelect!)
    this.nodeSelect = page.locator(mergedSelectors.nodeSelect!)
    this.generateButton = page.locator(mergedSelectors.generateButton!)
    this.exportHistoryTable = page.locator(mergedSelectors.exportHistoryTable!)
    this.downloadButton = page.locator(mergedSelectors.downloadButton!)
    this.previewSection = page.locator(mergedSelectors.previewSection!)
  }

  /**
   * Navigate to reports page
   */
  async goto(): Promise<void> {
    await super.goto('/reports')
    await this.waitForReady()
  }

  /**
   * Select report type
   */
  async selectReportType(type: string): Promise<void> {
    await this.selectOption('reportType', type)
  }

  /**
   * Select time range
   */
  async selectTimeRange(range: string): Promise<void> {
    await this.selectOption('timeRange', range)
  }

  /**
   * Select nodes for report
   */
  async selectNodes(nodeNames: string[]): Promise<void> {
    for (const nodeName of nodeNames) {
      await this.selectOption('node', nodeName)
    }
  }

  /**
   * Generate report
   */
  async generateReport(
    reportType: string,
    timeRange: string,
    nodeNames?: string[]
  ): Promise<void> {
    await this.selectReportType(reportType)
    await this.selectTimeRange(timeRange)

    if (nodeNames && nodeNames.length > 0) {
      await this.selectNodes(nodeNames)
    }

    await this.generateButton.click()
    await this.waitForReady()
  }

  /**
   * Generate and download report
   */
  async generateAndDownload(
    reportType: string,
    timeRange: string,
    nodeNames?: string[]
  ): Promise<Download> {
    await this.generateReport(reportType, timeRange, nodeNames)
    await this.waitForDownloadReady()
    return await this.download()
  }

  /**
   * Wait for download button to appear
   */
  async waitForDownloadReady(timeout = 30000): Promise<void> {
    await this.downloadButton.first().waitFor({ state: 'visible', timeout })
  }

  /**
   * Download report
   */
  async download(): Promise<Download> {
    const [download] = await Promise.all([
      this.page.waitForEvent('download'),
      this.downloadButton.first().click(),
    ])
    return download
  }

  /**
   * Check if report preview is visible
   */
  async isPreviewVisible(): Promise<boolean> {
    return await this.previewSection.first().isVisible()
  }

  /**
   * Expect report preview visible
   */
  async expectPreviewVisible(): Promise<void> {
    await expect(this.previewSection.first()).toBeVisible()
  }

  /**
   * Get export history count
   */
  async getExportHistoryCount(): Promise<number> {
    if (await this.exportHistoryTable.count() === 0) {
      return 0
    }
    return await this.exportHistoryTable.locator('tbody tr').count()
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
   * Expect generate button visible
   */
  async expectGenerateButtonVisible(): Promise<void> {
    await expect(this.generateButton).toBeVisible()
  }

  /**
   * Expect report type options available
   */
  async expectReportTypeOptions(options: string[]): Promise<void> {
    for (const option of options) {
      const optionLocator = this.reportTypeSelect.locator(`option:has-text("${option}")`)
      await expect(optionLocator).toBeVisible()
    }
  }
}
