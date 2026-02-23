/**
 * Login Page Object Model
 *
 * Handles user login functionality:
 * - Login form filling
 * - Error message handling
 * - Authentication validation
 */
import { Page, Locator, expect } from '@playwright/test'
import { BasePage, PageSelectors, DEFAULT_SELECTORS } from './common/BasePage'

export interface LoginSelectors extends PageSelectors {
  form?: string
  usernameInput?: string
  passwordInput?: string
  submitButton?: string
  errorMessage?: string
  validationErrors?: string
  showPasswordButton?: string
}

export const DEFAULT_LOGIN_SELECTORS: LoginSelectors = {
  ...DEFAULT_SELECTORS,
  form: '[data-testid="login-form"], form',
  usernameInput: '[data-testid="username-input"], #username, input[name="username"], input[type="text"]',
  passwordInput: '[data-testid="password-input"], #password, input[name="password"], input[type="password"]',
  submitButton: '[data-testid="login-button"], button[type="submit"], button:has-text("Sign in"), button:has-text("Login")',
  errorMessage: '[data-testid="login-error"], .bg-red-50 .text-red-700, .bg-red-50 p, [role="alert"]',
  validationErrors: '[data-testid="field-error"], .text-red-600',
  showPasswordButton: '[data-testid="show-password"], button:has-text("Show")',
}

export class LoginPage extends BasePage {
  readonly form: Locator
  readonly usernameInput: Locator
  readonly passwordInput: Locator
  readonly submitButton: Locator
  readonly errorMessage: Locator
  readonly validationErrors: Locator
  readonly showPasswordButton: Locator

  constructor(page: Page, selectors: LoginSelectors = {}) {
    super(page, selectors)
    const mergedSelectors = { ...DEFAULT_LOGIN_SELECTORS, ...selectors }

    this.form = page.locator(mergedSelectors.form!)
    this.usernameInput = page.locator(mergedSelectors.usernameInput!)
    this.passwordInput = page.locator(mergedSelectors.passwordInput!)
    this.submitButton = page.locator(mergedSelectors.submitButton!)
    this.errorMessage = page.locator(mergedSelectors.errorMessage!)
    this.validationErrors = page.locator(mergedSelectors.validationErrors!)
    this.showPasswordButton = page.locator(mergedSelectors.showPasswordButton!)
  }

  /**
   * Navigate to login page
   */
  async goto(): Promise<void> {
    await super.goto('/login')
    await this.waitForReady()
  }

  /**
   * Fill login form
   */
  async fillForm(username: string, password: string): Promise<void> {
    await this.usernameInput.fill(username)
    await this.passwordInput.fill(password)
  }

  /**
   * Submit login form
   */
  async submit(): Promise<void> {
    await this.submitButton.click()
  }

  /**
   * Complete login process
   */
  async login(username: string, password: string): Promise<void> {
    await this.fillForm(username, password)
    await this.submit()
  }

  /**
   * Login and wait for redirect to dashboard
   */
  async loginAndWait(username: string, password: string, timeout = 15000): Promise<void> {
    await this.login(username, password)
    await this.page.waitForURL('**/dashboard**', { timeout })
  }

  /**
   * Get error message text
   */
  async getErrorMessage(): Promise<string | null> {
    if (await this.errorMessage.first().isVisible()) {
      return await this.errorMessage.first().textContent()
    }
    return null
  }

  /**
   * Check if error message is visible
   */
  async hasError(): Promise<boolean> {
    return await this.errorMessage.first().isVisible()
  }

  /**
   * Get validation errors
   */
  async getValidationErrors(): Promise<string[]> {
    const errors: string[] = []
    const count = await this.validationErrors.count()

    for (let i = 0; i < count; i++) {
      const errorText = await this.validationErrors.nth(i).textContent()
      if (errorText) {
        errors.push(errorText.trim())
      }
    }

    return errors
  }

  /**
   * Expect error message to contain text
   */
  async expectError(message: string): Promise<boolean> {
    await this.errorMessage.waitFor({ state: 'visible', timeout: 5000 })
    const text = await this.errorMessage.first().textContent()
    return text?.includes(message) ?? false
  }

  /**
   * Expect validation error for field
   */
  async expectFieldError(fieldName: string, message: string): Promise<void> {
    const fieldError = this.form.locator(
      `[name="${fieldName}"] + .text-red-600, [data-testid="${fieldName}-error"]`
    )
    await expect(fieldError).toContainText(message)
  }

  /**
   * Expect redirect to dashboard after login
   */
  async expectRedirectToDashboard(): Promise<void> {
    await this.page.waitForURL('**/dashboard**', { timeout: 15000 })
  }

  /**
   * Expect redirect to login page
   */
  async expectRedirectToLogin(): Promise<void> {
    await this.page.waitForURL('**/login**', { timeout: 10000 })
  }

  /**
   * Check if submit button is disabled
   */
  async isSubmitButtonDisabled(): Promise<boolean> {
    return await this.submitButton.isDisabled()
  }

  /**
   * Check if submit button is enabled
   */
  async isSubmitButtonEnabled(): Promise<boolean> {
    return await this.submitButton.isEnabled()
  }

  /**
   * Toggle password visibility
   */
  async togglePasswordVisibility(): Promise<void> {
    if (await this.showPasswordButton.count() > 0) {
      await this.showPasswordButton.click()
    }
  }

  /**
   * Check if password is visible (text type)
   */
  async isPasswordVisible(): Promise<boolean> {
    const type = await this.passwordInput.getAttribute('type')
    return type === 'text'
  }

  /**
   * Assert login form is visible
   */
  async expectFormVisible(): Promise<void> {
    await expect(this.form).toBeVisible()
    await expect(this.usernameInput).toBeVisible()
    await expect(this.passwordInput).toBeVisible()
    await expect(this.submitButton).toBeVisible()
  }

  /**
   * Assert form fields are enabled
   */
  async expectFieldsEnabled(): Promise<void> {
    await expect(this.usernameInput).toBeEnabled()
    await expect(this.passwordInput).toBeEnabled()
    await expect(this.submitButton).toBeEnabled()
  }

  /**
   * Clear form fields
   */
  async clearForm(): Promise<void> {
    await this.usernameInput.clear()
    await this.passwordInput.clear()
  }

  /**
   * Submit form with keyboard (Enter key)
   */
  async submitWithKeyboard(): Promise<void> {
    await this.page.keyboard.press('Enter')
  }

  /**
   * Wait for login to complete (API response)
   */
  async waitForLoginComplete(timeout = 15000): Promise<void> {
    await this.page.waitForResponse(
      (response) => response.url().includes('/api/v1/auth/login') && response.status() === 200,
      { timeout }
    )
  }

  /**
   * Wait for login error (API response)
   */
  async waitForLoginError(timeout = 15000): Promise<void> {
    await this.page.waitForResponse(
      (response) => response.url().includes('/api/v1/auth/login') && response.status() >= 400,
      { timeout }
    )
  }
}
