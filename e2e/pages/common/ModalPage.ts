/**
 * Modal Dialog Page Object Model
 *
 * Provides common functionality for modal dialogs:
 * - Open/close operations
 * - Form filling
 * - Submit/cancel actions
 * - Confirmation dialogs
 */
import { Page, Locator, expect } from '@playwright/test'
import { BasePage, PageSelectors, DEFAULT_SELECTORS } from './BasePage'

export interface ModalSelectors extends PageSelectors {
  modal?: string
  modalContent?: string
  modalTitle?: string
  closeButton?: string
  cancelButton?: string
  submitButton?: string
  confirmButton?: string
  overlay?: string
}

export const DEFAULT_MODAL_SELECTORS: ModalSelectors = {
  ...DEFAULT_SELECTORS,
  modal: '[data-testid="modal"], .fixed.inset-0, [role="dialog"], .modal',
  modalContent: '[data-testid="modal-content"], .modal-content',
  modalTitle: '[data-testid="modal-title"], .modal-title, h2:has-text("Add"), h2:has-text("Edit"), h2:has-text("Delete"), h3:has-text("Add"), h3:has-text("Edit"), h3:has-text("Delete")',
  closeButton: '[data-testid="modal-close"], button[aria-label="Close"], .close-button',
  cancelButton: '[data-testid="modal-cancel"], button:has-text("Cancel")',
  submitButton: '[data-testid="modal-submit"], button[type="submit"], button:has-text("Save"), button:has-text("Confirm"), button:has-text("OK")',
  confirmButton: '[data-testid="modal-confirm"], button:has-text("Confirm"), button:has-text("Delete"), button:has-text("Yes")',
  overlay: '[data-testid="modal-overlay"], .fixed.inset-0.bg-black',
}

export abstract class ModalPage extends BasePage {
  readonly modal: Locator
  readonly modalContent: Locator
  readonly modalTitle: Locator
  readonly closeButton: Locator
  readonly cancelButton: Locator
  readonly submitButton: Locator
  readonly confirmButton: Locator
  readonly overlay: Locator
  protected declare selectors: ModalSelectors

  constructor(page: Page, selectors: ModalSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_MODAL_SELECTORS, ...selectors }
    this.selectors = mergedSelectors

    this.modal = page.locator(mergedSelectors.modal!)
    this.modalContent = page.locator(mergedSelectors.modalContent!)
    this.modalTitle = page.locator(mergedSelectors.modalTitle!)
    this.closeButton = page.locator(mergedSelectors.closeButton!)
    this.cancelButton = page.locator(mergedSelectors.cancelButton!)
    this.submitButton = page.locator(mergedSelectors.submitButton!)
    this.confirmButton = page.locator(mergedSelectors.confirmButton!)
    this.overlay = page.locator(mergedSelectors.overlay!)
  }

  /**
   * Check if modal is visible
   */
  async isModalVisible(): Promise<boolean> {
    return await this.modal.first().isVisible()
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
  async waitForModalClose(timeout = 5000): Promise<void> {
    await this.modal.first().waitFor({ state: 'hidden', timeout })
  }

  /**
   * Close modal using close button
   */
  async close(): Promise<void> {
    if (await this.closeButton.count() > 0) {
      await this.closeButton.click()
    } else if (await this.cancelButton.count() > 0) {
      await this.cancelButton.click()
    } else {
      // Click overlay to close
      await this.overlay.first().click()
    }
    await this.waitForModalClose()
  }

  /**
   * Click cancel button
   */
  async cancel(): Promise<void> {
    await this.cancelButton.click()
    await this.waitForModalClose()
  }

  /**
   * Click submit button
   */
  async submit(): Promise<void> {
    await this.submitButton.click()
  }

  /**
   * Click confirm button (for confirmation dialogs)
   */
  async confirm(): Promise<void> {
    await this.confirmButton.click()
    await this.waitForModalClose()
  }

  /**
   * Get modal title text
   */
  async getTitle(): Promise<string | null> {
    if (await this.modalTitle.count() > 0) {
      return await this.modalTitle.first().textContent()
    }
    return null
  }

  /**
   * Assert modal is visible
   */
  async expectModalVisible(): Promise<void> {
    await expect(this.modal.first()).toBeVisible()
  }

  /**
   * Assert modal is hidden
   */
  async expectModalHidden(): Promise<void> {
    await expect(this.modal.first()).toBeHidden()
  }

  /**
   * Assert modal title
   */
  async expectTitleContains(text: string): Promise<void> {
    const title = await this.getTitle()
    expect(title).toContain(text)
  }

  /**
   * Fill form fields in modal
   */
  async fillForm(fields: Record<string, string>): Promise<void> {
    for (const [field, value] of Object.entries(fields)) {
      const input = this.modalContent.locator(
        `[name="${field}"], [data-testid="${field}-input"], #${field}, input[placeholder*="${field}" i], textarea[placeholder*="${field}" i]`
      )

      if (await input.count() > 0) {
        await input.fill(value)
      }
    }
  }

  /**
   * Select option in modal
   */
  async selectOption(fieldName: string, option: string): Promise<void> {
    const select = this.modalContent.locator(
      `select[name="${fieldName}"], [data-testid="${fieldName}-select"]`
    )

    if (await select.count() > 0) {
      await select.selectOption(option)
    }
  }

  /**
   * Check checkbox in modal
   */
  async checkCheckbox(fieldName: string): Promise<void> {
    const checkbox = this.modalContent.locator(
      `input[type="checkbox"][name="${fieldName}"], [data-testid="${fieldName}-checkbox"]`
    )

    if (await checkbox.count() > 0) {
      await checkbox.check()
    }
  }

  /**
   * Submit form and wait for modal to close
   */
  async submitAndWait(timeout = 5000): Promise<void> {
    await this.submit()
    await this.waitForModalClose(timeout)
  }

  /**
   * Fill and submit form
   */
  async fillAndSubmit(fields: Record<string, string>, waitForClose = true): Promise<void> {
    await this.fillForm(fields)
    await this.submit()

    if (waitForClose) {
      await this.waitForModalClose()
    }
  }

  /**
   * Get form validation errors
   */
  async getValidationErrors(): Promise<string[]> {
    const errors: string[] = []
    const errorElements = this.modalContent.locator('.text-red-600, [data-testid="field-error"], .error-message')
    const count = await errorElements.count()

    for (let i = 0; i < count; i++) {
      const errorText = await errorElements.nth(i).textContent()
      if (errorText) {
        errors.push(errorText.trim())
      }
    }

    return errors
  }

  /**
   * Assert form validation error
   */
  async expectValidationError(fieldName: string, message: string): Promise<void> {
    const errorLocator = this.modalContent.locator(
      `[name="${fieldName}"] + .text-red-600, [data-testid="${fieldName}-error"]`
    )
    await expect(errorLocator).toContainText(message)
  }

  /**
   * Click outside modal to dismiss
   */
  async clickOutside(): Promise<void> {
    await this.overlay.first().click()
  }

  /**
   * Press Escape to close modal
   */
  async pressEscape(): Promise<void> {
    await this.page.keyboard.press('Escape')
    await this.waitForModalClose()
  }
}
