/**
 * Nodes Page Object Model (Node Management)
 *
 * Handles node management functionality:
 * - View nodes list
 * - Create, edit, delete nodes
 * - Search and filter nodes
 */
import { Page, Locator, expect } from '@playwright/test'
import { TablePage, TableSelectors, DEFAULT_TABLE_SELECTORS } from './common/TablePage'
import { ModalPage, ModalSelectors, DEFAULT_MODAL_SELECTORS } from './common/ModalPage'

export interface NodesSelectors extends TableSelectors, ModalSelectors {
  createButton?: string
  nameInput?: string
  regionInput?: string
  ipInput?: string
  confirmDeleteButton?: string
  cancelButton?: string
}

export const DEFAULT_NODES_SELECTORS: NodesSelectors = {
  ...DEFAULT_TABLE_SELECTORS,
  ...DEFAULT_MODAL_SELECTORS,
  createButton: '[data-testid="create-button"], button:has-text("Add New Node"), button:has-text("Create"), button:has-text("Add")',
  nameInput: '[data-testid="name-input"], #name, input[name="name"]',
  regionInput: '[data-testid="region-input"], #region, input[name="region"], select[name="region"]',
  ipInput: '[data-testid="ip-input"], #ip, input[name="ip"]',
  confirmDeleteButton: '[data-testid="confirm-delete-button"], .fixed button:has-text("Delete")',
  cancelButton: '[data-testid="cancel-button"], button:has-text("Cancel")',
}

export class NodesPage extends TablePage {
  readonly createButton: Locator
  readonly nameInput: Locator
  readonly regionInput: Locator
  readonly ipInput: Locator
  readonly confirmDeleteButton: Locator
  readonly modal: Locator
  readonly submitButton: Locator
  readonly cancelButton: Locator

  constructor(page: Page, selectors: NodesSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_NODES_SELECTORS, ...selectors }

    this.createButton = page.locator(mergedSelectors.createButton!)
    this.nameInput = page.locator(mergedSelectors.nameInput!)
    this.regionInput = page.locator(mergedSelectors.regionInput!)
    this.ipInput = page.locator(mergedSelectors.ipInput!)
    this.confirmDeleteButton = page.locator(mergedSelectors.confirmDeleteButton!)
    this.modal = page.locator(mergedSelectors.modal!)
    this.submitButton = page.locator(mergedSelectors.submitButton!)
    this.cancelButton = page.locator(mergedSelectors.cancelButton!)
  }

  /**
   * Wait for modal to open
   */
  async waitForModalOpen(timeout = 5000): Promise<void> {
    await this.modal.first().waitFor({ state: 'visible', timeout })
  }

  /**
   * Wait for modal to close
   */
  async waitForModalClose(timeout = 10000): Promise<void> {
    try {
      await this.modal.first().waitFor({ state: 'hidden', timeout })
    } catch {
      // Modal might already be closed or have different structure
      // Just wait a bit for the page to stabilize
      await this.page.waitForTimeout(500)
    }
  }

  /**
   * Submit form
   */
  async submit(): Promise<void> {
    await this.submitButton.click()
  }

  /**
   * Navigate to nodes page
   */
  async goto(): Promise<void> {
    await super.goto('/nodes')
    await this.waitForReady()
  }

  /**
   * Expect create button to be visible
   */
  async expectCreateButtonVisible(): Promise<void> {
    await this.createButton.first().waitFor({ state: 'visible' })
  }

  /**
   * Expect create button to be hidden
   */
  async expectCreateButtonHidden(): Promise<void> {
    await this.createButton.first().waitFor({ state: 'hidden' })
  }

  /**
   * Click create button
   */
  async clickCreate(): Promise<void> {
    // Wait for the create button to be visible and click it
    await this.createButton.first().waitFor({ state: 'visible', timeout: 10000 })
    await this.createButton.first().click()
    await this.waitForModalOpen()
  }

  /**
   * Create a new node
   */
  async createNode(name: string, region: string): Promise<void> {
    await this.clickCreate()

    // Wait for form to be interactive and fill fields within the modal
    const modalForm = this.modal.first()

    // Fill name - wait for it to be visible first
    const nameField = modalForm.locator('#name, input[name="name"]')
    await nameField.waitFor({ state: 'visible', timeout: 5000 })
    await nameField.fill(name)

    // Fill region
    const regionField = modalForm.locator('#region, input[name="region"]')
    await regionField.waitFor({ state: 'visible', timeout: 5000 })
    await regionField.fill(region)

    await this.submit()
    await this.waitForModalClose()
  }

  /**
   * Create node with IP address
   */
  async createNodeWithIp(name: string, region: string, ip: string): Promise<void> {
    await this.clickCreate()

    // Wait for form to be interactive and fill fields within the modal
    const modalForm = this.modal.first()

    // Fill name - wait for it to be visible first
    const nameField = modalForm.locator('#name, input[name="name"]')
    await nameField.waitFor({ state: 'visible', timeout: 5000 })
    await nameField.fill(name)

    // Fill IP
    const ipField = modalForm.locator('#ip, input[name="ip"]')
    await ipField.waitFor({ state: 'visible', timeout: 5000 })
    await ipField.fill(ip)

    // Fill region
    const regionField = modalForm.locator('#region, input[name="region"]')
    await regionField.waitFor({ state: 'visible', timeout: 5000 })
    await regionField.fill(region)

    await this.submit()
    await this.waitForModalClose()
  }

  /**
   * Edit node by row index
   */
  async editNode(rowIndex: number, name: string): Promise<void> {
    await this.clickEdit(rowIndex)
    await this.waitForModalOpen()
    await this.nameInput.fill(name)
    await this.submit()
    await this.waitForModalClose()
  }

  /**
   * Delete node by row index
   */
  async deleteNode(rowIndex: number): Promise<void> {
    await this.clickDelete(rowIndex)
    await this.confirmDeleteButton.click()
    await this.waitForReady()
  }

  /**
   * Search for nodes
   */
  async search(query: string): Promise<void> {
    await super.search(query)
  }

  /**
   * Check if node exists by name
   */
  async hasNode(name: string): Promise<boolean> {
    return await this.hasRowWithText(name)
  }

  /**
   * Get node row index by name
   */
  async getNodeRowIndex(name: string): Promise<number> {
    const rows = this.table.locator('tbody tr')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i)
      const text = await row.textContent()
      if (text?.includes(name)) {
        return i
      }
    }
    return -1
  }

  /**
   * Click on node row to view details
   */
  async clickNodeRow(rowIndex: number): Promise<void> {
    const row = this.getRow(rowIndex)
    await row.click()
  }

  /**
   * Click on node name link
   */
  async clickNodeLink(name: string): Promise<void> {
    const link = this.table.locator(`a:has-text("${name}"), td:has-text("${name}")`)
    await link.first().click()
  }

  /**
   * Get node status by row
   */
  async getNodeStatus(rowIndex: number): Promise<string | null> {
    const statusCell = this.getRow(rowIndex).locator('td').nth(1)
    return await statusCell.textContent()
  }

  /**
   * Check if node is online
   */
  async isNodeOnline(rowIndex: number): Promise<boolean> {
    const status = await this.getNodeStatus(rowIndex)
    return status?.toLowerCase().includes('online') ?? false
  }

  /**
   * Wait for node to appear in list
   */
  async waitForNodeToAppear(name: string, timeout = 10000): Promise<void> {
    await this.waitForRow(name, timeout)
  }

  /**
   * Wait for node to disappear from list
   */
  async waitForNodeToDisappear(name: string, timeout = 10000): Promise<void> {
    await this.waitForRowToDisappear(name, timeout)
  }

  /**
   * Create node and wait for it to appear
   */
  async createNodeAndWait(name: string, region: string, timeout = 10000): Promise<void> {
    await this.createNode(name, region)
    await this.waitForNodeToAppear(name, timeout)
  }

  /**
   * Delete node and wait for it to disappear
   */
  async deleteNodeAndWait(rowIndex: number, nodeName: string, timeout = 10000): Promise<void> {
    await this.deleteNode(rowIndex)
    await this.waitForNodeToDisappear(nodeName, timeout)
  }

  /**
   * Assert node count
   */
  async expectNodeCount(expected: number): Promise<void> {
    const count = await this.getRowCount()
    expect(count).toBe(expected)
  }

  /**
   * Assert node exists
   */
  async expectNodeExists(name: string): Promise<void> {
    const exists = await this.hasNode(name)
    expect(exists).toBeTruthy()
  }

  /**
   * Assert node does not exist
   */
  async expectNodeNotExists(name: string): Promise<void> {
    const exists = await this.hasNode(name)
    expect(exists).toBeFalsy()
  }
}
