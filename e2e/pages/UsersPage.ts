/**
 * Users Page Object Model
 *
 * Handles user management (Admin only):
 * - View users list
 * - Create, edit, delete users
 * - Manage user roles and status
 */
import { Page, Locator, expect } from '@playwright/test'
import { TablePage, TableSelectors, DEFAULT_TABLE_SELECTORS } from './common/TablePage'
import { ModalPage, ModalSelectors, DEFAULT_MODAL_SELECTORS } from './common/ModalPage'

export interface UsersSelectors extends TableSelectors, ModalSelectors {
  createButton?: string
  usernameInput?: string
  emailInput?: string
  passwordInput?: string
  roleSelect?: string
  statusSelect?: string
  confirmDeleteButton?: string
}

export const DEFAULT_USERS_SELECTORS: UsersSelectors = {
  ...DEFAULT_TABLE_SELECTORS,
  ...DEFAULT_MODAL_SELECTORS,
  createButton: '[data-testid="create-user-button"], button:has-text("Create"), button:has-text("Add User")',
  usernameInput: '[data-testid="username-input"], #username, input[name="username"]',
  emailInput: '[data-testid="email-input"], #email, input[name="email"], input[type="email"]',
  passwordInput: '[data-testid="password-input"], #password, input[name="password"], input[type="password"]',
  roleSelect: '[data-testid="role-select"], select[name="role"]',
  statusSelect: '[data-testid="status-select"], select[name="status"]',
  confirmDeleteButton: '[data-testid="confirm-delete-button"], .fixed button:has-text("Delete")',
}

export class UsersPage extends TablePage {
  readonly createButton: Locator
  readonly usernameInput: Locator
  readonly emailInput: Locator
  readonly passwordInput: Locator
  readonly roleSelect: Locator
  readonly statusSelect: Locator
  readonly confirmDeleteButton: Locator
  readonly modal: Locator
  readonly submitButton: Locator

  constructor(page: Page, selectors: UsersSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_USERS_SELECTORS, ...selectors }

    this.createButton = page.locator(mergedSelectors.createButton!)
    this.usernameInput = page.locator(mergedSelectors.usernameInput!)
    this.emailInput = page.locator(mergedSelectors.emailInput!)
    this.passwordInput = page.locator(mergedSelectors.passwordInput!)
    this.roleSelect = page.locator(mergedSelectors.roleSelect!)
    this.statusSelect = page.locator(mergedSelectors.statusSelect!)
    this.confirmDeleteButton = page.locator(mergedSelectors.confirmDeleteButton!)
    this.modal = page.locator(mergedSelectors.modal!)
    this.submitButton = page.locator(mergedSelectors.submitButton!)
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
   * Navigate to users page
   */
  async goto(): Promise<void> {
    await super.goto('/settings/users')
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
   * Create user
   */
  async createUser(
    username: string,
    email: string,
    password: string,
    role: string
  ): Promise<void> {
    await this.clickCreate()
    await this.usernameInput.fill(username)
    await this.emailInput.fill(email)
    await this.passwordInput.fill(password)
    if (await this.roleSelect.count() > 0) {
      await this.roleSelect.selectOption(role)
    }
    await this.submit()
    await this.waitForModalClose()
  }

  /**
   * Get user row index by username
   */
  async getUserRowIndexByUsername(username: string): Promise<number> {
    return await this.hasRowWithText(username) ? 0 : -1
  }

  /**
   * Get user role by row
   */
  async getUserRole(rowIndex: number): Promise<string | null> {
    const roleCell = this.getRow(rowIndex).locator('td').nth(2)
    return await roleCell.textContent()
  }

  /**
   * Get user status by row
   */
  async getUserStatus(rowIndex: number): Promise<string | null> {
    const statusCell = this.getRow(rowIndex).locator('td').nth(3)
    return await statusCell.textContent()
  }

  /**
   * Update user role
   */
  async updateUserRole(rowIndex: number, role: string): Promise<void> {
    const roleSelect = this.getRow(rowIndex).locator('select[name="role"]')
    if (await roleSelect.count() > 0) {
      await roleSelect.selectOption(role)
    }
  }

  /**
   * Update user status
   */
  async updateUserStatus(rowIndex: number, status: string): Promise<void> {
    const statusSelect = this.getRow(rowIndex).locator('select[name="status"]')
    if (await statusSelect.count() > 0) {
      await statusSelect.selectOption(status)
    }
  }

  /**
   * Delete user
   */
  async deleteUser(rowIndex: number): Promise<void> {
    await this.clickDelete(rowIndex)
    await this.confirmDeleteButton.click()
    await this.waitForReady()
  }

  /**
   * Delete user and wait for disappearance
   */
  async deleteUserAndWait(rowIndex: number, username: string): Promise<void> {
    await this.deleteUser(rowIndex)
    await this.waitForRowToDisappear(username)
  }

  /**
   * Expect user exists
   */
  async expectUserExists(username: string): Promise<void> {
    const exists = await this.hasRowWithText(username)
    expect(exists).toBeTruthy()
  }

  /**
   * Expect user does not exist
   */
  async expectUserNotExists(username: string): Promise<void> {
    const exists = await this.hasRowWithText(username)
    expect(exists).toBeFalsy()
  }

  /**
   * Expect create button visible
   */
  async expectCreateButtonVisible(): Promise<void> {
    await expect(this.createButton.first()).toBeVisible()
  }

  /**
   * Expect users table has expected columns
   */
  async expectTableColumns(expectedColumns: string[]): Promise<void> {
    await this.expectHeaders(expectedColumns)
  }

  /**
   * Check if user is active
   */
  async isUserActive(rowIndex: number): Promise<boolean> {
    const status = await this.getUserStatus(rowIndex)
    return status?.toLowerCase().includes('active') ?? false
  }

  /**
   * Get user count by role
   */
  async getUserCountByRole(role: string): Promise<number> {
    let count = 0
    const rowCount = await this.getRowCount()

    for (let i = 0; i < rowCount; i++) {
      const userRole = await this.getUserRole(i)
      if (userRole?.toLowerCase().includes(role.toLowerCase())) {
        count++
      }
    }

    return count
  }
}
