/**
 * Form Page Object Model
 *
 * Extends ModalPage with form-specific functionality:
 * - Field filling
 * - Form validation
 * - File upload
 * - Form submission
 */
import { Page, Locator, expect } from '@playwright/test'
import { ModalPage, ModalSelectors, DEFAULT_MODAL_SELECTORS } from './ModalPage'

export interface FormSelectors extends ModalSelectors {
  form?: string
  fields?: Record<string, string>
  submitButton?: string
  cancelButton?: string
  resetButton?: string
  errorSummary?: string
}

export const DEFAULT_FORM_SELECTORS: FormSelectors = {
  ...DEFAULT_MODAL_SELECTORS,
  form: '[data-testid="form"], form',
  submitButton: '[data-testid="form-submit"], button[type="submit"], button:has-text("Submit"), button:has-text("Save"), button:has-text("Create"), button:has-text("Update")',
  cancelButton: '[data-testid="form-cancel"], button[type="button"]:has-text("Cancel")',
  resetButton: '[data-testid="form-reset"], button[type="reset"]',
  errorSummary: '[data-testid="error-summary"], .error-summary, .form-errors',
}

export interface FormField {
  name: string
  value: string
  type?: 'text' | 'password' | 'email' | 'number' | 'select' | 'checkbox' | 'radio' | 'textarea' | 'file'
}

export abstract class FormPage extends ModalPage {
  readonly form: Locator
  readonly submitButton: Locator
  readonly cancelButton: Locator
  readonly resetButton: Locator
  readonly errorSummary: Locator
  protected declare selectors: FormSelectors

  constructor(page: Page, selectors: FormSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_FORM_SELECTORS, ...selectors }
    this.selectors = mergedSelectors

    this.form = page.locator(mergedSelectors.form!)
    this.submitButton = page.locator(mergedSelectors.submitButton!)
    this.cancelButton = page.locator(mergedSelectors.cancelButton!)
    this.resetButton = page.locator(mergedSelectors.resetButton!)
    this.errorSummary = page.locator(mergedSelectors.errorSummary!)
  }

  /**
   * Get field locator by name
   */
  getFieldLocator(fieldName: string): Locator {
    return this.form.locator(
      `[name="${fieldName}"], [data-testid="${fieldName}-input"], #${fieldName}, input[placeholder*="${fieldName}" i], textarea[placeholder*="${fieldName}" i]`
    )
  }

  /**
   * Fill text field
   */
  async fillField(fieldName: string, value: string): Promise<void> {
    const field = this.getFieldLocator(fieldName)
    await field.fill(value)
  }

  /**
   * Fill multiple fields
   */
  async fillFields(fields: Record<string, string>): Promise<void> {
    for (const [name, value] of Object.entries(fields)) {
      await this.fillField(name, value)
    }
  }

  /**
   * Clear field
   */
  async clearField(fieldName: string): Promise<void> {
    const field = this.getFieldLocator(fieldName)
    await field.clear()
  }

  /**
   * Select option from dropdown
   */
  async selectOption(fieldName: string, option: string | number): Promise<void> {
    const select = this.form.locator(
      `select[name="${fieldName}"], [data-testid="${fieldName}-select"]`
    )
    await select.selectOption(option.toString())
  }

  /**
   * Select option by label
   */
  async selectOptionByLabel(fieldName: string, label: string): Promise<void> {
    const select = this.form.locator(
      `select[name="${fieldName}"], [data-testid="${fieldName}-select"]`
    )
    await select.selectOption({ label })
  }

  /**
   * Check checkbox
   */
  async checkCheckbox(fieldName: string): Promise<void> {
    const checkbox = this.form.locator(
      `input[type="checkbox"][name="${fieldName}"], [data-testid="${fieldName}-checkbox"]`
    )
    await checkbox.check()
  }

  /**
   * Uncheck checkbox
   */
  async uncheckCheckbox(fieldName: string): Promise<void> {
    const checkbox = this.form.locator(
      `input[type="checkbox"][name="${fieldName}"], [data-testid="${fieldName}-checkbox"]`
    )
    await checkbox.uncheck()
  }

  /**
   * Check if checkbox is checked
   */
  async isCheckboxChecked(fieldName: string): Promise<boolean> {
    const checkbox = this.form.locator(
      `input[type="checkbox"][name="${fieldName}"], [data-testid="${fieldName}-checkbox"]`
    )
    return await checkbox.isChecked()
  }

  /**
   * Select radio option
   */
  async selectRadio(fieldName: string, value: string): Promise<void> {
    const radio = this.form.locator(
      `input[type="radio"][name="${fieldName}"][value="${value}"], [data-testid="${fieldName}-radio-${value}"]`
    )
    await radio.check()
  }

  /**
   * Upload file
   */
  async uploadFile(fieldName: string, filePath: string): Promise<void> {
    const fileInput = this.form.locator(
      `input[type="file"][name="${fieldName}"], [data-testid="${fieldName}-file"]`
    )
    await fileInput.setInputFiles(filePath)
  }

  /**
   * Upload multiple files
   */
  async uploadFiles(fieldName: string, filePaths: string[]): Promise<void> {
    const fileInput = this.form.locator(
      `input[type="file"][name="${fieldName}"], [data-testid="${fieldName}-file"]`
    )
    await fileInput.setInputFiles(filePaths)
  }

  /**
   * Fill text area
   */
  async fillTextArea(fieldName: string, value: string): Promise<void> {
    const textarea = this.form.locator(
      `textarea[name="${fieldName}"], [data-testid="${fieldName}-textarea"]`
    )
    await textarea.fill(value)
  }

  /**
   * Fill password field
   */
  async fillPasswordField(fieldName: string, value: string): Promise<void> {
    const passwordInput = this.form.locator(
      `input[type="password"][name="${fieldName}"], [data-testid="${fieldName}-password"]`
    )
    await passwordInput.fill(value)
  }

  /**
   * Fill email field
   */
  async fillEmailField(fieldName: string, value: string): Promise<void> {
    const emailInput = this.form.locator(
      `input[type="email"][name="${fieldName}"], [data-testid="${fieldName}-email"]`
    )
    await emailInput.fill(value)
  }

  /**
   * Fill number field
   */
  async fillNumberField(fieldName: string, value: number): Promise<void> {
    const numberInput = this.form.locator(
      `input[type="number"][name="${fieldName}"], [data-testid="${fieldName}-number"]`
    )
    await numberInput.fill(value.toString())
  }

  /**
   * Submit form
   */
  async submit(): Promise<void> {
    await this.submitButton.click()
  }

  /**
   * Submit form and wait for success
   */
  async submitAndWaitForSuccess(successMessage?: string, timeout = 10000): Promise<void> {
    await this.submit()

    if (successMessage) {
      await this.waitForToast(successMessage, timeout)
    } else {
      await this.waitForModalClose(timeout)
    }
  }

  /**
   * Submit form and wait for error
   */
  async submitAndWaitForError(errorMessage?: string, timeout = 10000): Promise<void> {
    await this.submit()

    if (errorMessage) {
      await this.page.locator(`:has-text("${errorMessage}")`).waitFor({ state: 'visible', timeout })
    } else {
      await this.errorMessage.first().waitFor({ state: 'visible', timeout })
    }
  }

  /**
   * Cancel form
   */
  async cancel(): Promise<void> {
    await this.cancelButton.click()
    await this.waitForModalClose()
  }

  /**
   * Reset form
   */
  async reset(): Promise<void> {
    if (await this.resetButton.count() > 0) {
      await this.resetButton.click()
    }
  }

  /**
   * Get field validation error
   */
  async getFieldError(fieldName: string): Promise<string | null> {
    const errorLocator = this.form.locator(
      `[name="${fieldName}"] + .text-red-600, [data-testid="${fieldName}-error"], #${fieldName}-error`
    )

    if (await errorLocator.count() > 0) {
      return await errorLocator.first().textContent()
    }
    return null
  }

  /**
   * Get all form validation errors
   */
  async getValidationErrors(): Promise<string[]> {
    const errors: string[] = []

    // Check error summary first
    if (await this.errorSummary.count() > 0) {
      const summaryText = await this.errorSummary.first().textContent()
      if (summaryText) {
        errors.push(summaryText.trim())
      }
    }

    // Check individual field errors
    const errorElements = this.form.locator('.text-red-600, [data-testid="field-error"], .error-message')
    const count = await errorElements.count()

    for (let i = 0; i < count; i++) {
      const errorText = await errorElements.nth(i).textContent()
      if (errorText && !errors.includes(errorText.trim())) {
        errors.push(errorText.trim())
      }
    }

    return errors
  }

  /**
   * Assert field has error
   */
  async expectFieldError(fieldName: string, message: string): Promise<void> {
    const errorLocator = this.form.locator(
      `[name="${fieldName}"] + .text-red-600, [data-testid="${fieldName}-error"]`
    )
    await expect(errorLocator).toContainText(message)
  }

  /**
   * Assert field has no error
   */
  async expectFieldNoError(fieldName: string): Promise<void> {
    const errorLocator = this.form.locator(
      `[name="${fieldName}"] + .text-red-600, [data-testid="${fieldName}-error"]`
    )
    await expect(errorLocator).toBeHidden()
  }

  /**
   * Assert form is valid (no errors)
   */
  async expectFormValid(): Promise<void> {
    await expect(this.errorSummary).toBeHidden()
    const errorElements = this.form.locator('.text-red-600')
    await expect(errorElements.first()).toBeHidden()
  }

  /**
   * Assert form is invalid
   */
  async expectFormInvalid(): Promise<void> {
    const hasErrorSummary = await this.errorSummary.count() > 0
    const hasFieldErrors = await this.form.locator('.text-red-600').count() > 0
    expect(hasErrorSummary || hasFieldErrors).toBeTruthy()
  }

  /**
   * Assert submit button is disabled
   */
  async expectSubmitDisabled(): Promise<void> {
    await expect(this.submitButton).toBeDisabled()
  }

  /**
   * Assert submit button is enabled
   */
  async expectSubmitEnabled(): Promise<void> {
    await expect(this.submitButton).toBeEnabled()
  }

  /**
   * Assert field is required
   */
  async expectFieldRequired(fieldName: string): Promise<void> {
    const field = this.getFieldLocator(fieldName)
    const required = await field.getAttribute('required')
    const ariaRequired = await field.getAttribute('aria-required')
    expect(required === '' || ariaRequired === 'true').toBeTruthy()
  }

  /**
   * Get form data
   */
  async getFormData(): Promise<Record<string, string>> {
    const formData: Record<string, string> = {}

    // Get all inputs
    const inputs = this.form.locator('input, select, textarea')
    const count = await inputs.count()

    for (let i = 0; i < count; i++) {
      const input = inputs.nth(i)
      const name = await input.getAttribute('name')
      if (name) {
        const value = await input.inputValue()
        formData[name] = value
      }
    }

    return formData
  }

  /**
   * Assert field value
   */
  async expectFieldValue(fieldName: string, expectedValue: string): Promise<void> {
    const field = this.getFieldLocator(fieldName)
    await expect(field).toHaveValue(expectedValue)
  }
}
