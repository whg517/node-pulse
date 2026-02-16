/**
 * Login Page Object Model
 */
import { Page, Locator } from '@playwright/test'

export class LoginPage {
  readonly page: Page
  readonly usernameInput: Locator
  readonly passwordInput: Locator
  readonly submitButton: Locator
  readonly errorMessage: Locator
  readonly form: Locator
  readonly validationErrors: Locator

  constructor(page: Page) {
    this.page = page
    // Use ID selectors as the frontend uses id="username" and id="password"
    this.usernameInput = page.locator('#username')
    this.passwordInput = page.locator('#password')
    // Submit button with text "Sign in"
    this.submitButton = page.locator('button[type="submit"]')
    // API errors shown in red alert box
    this.errorMessage = page.locator('.bg-red-50 .text-red-700, .bg-red-50 p')
    this.form = page.locator('form')
    // Field validation errors
    this.validationErrors = page.locator('.text-red-600')
  }

  async goto() {
    await this.page.goto('/login')
    await this.page.waitForLoadState('networkidle')
  }

  async login(username: string, password: string) {
    await this.usernameInput.fill(username)
    await this.passwordInput.fill(password)
    await this.submitButton.click()
  }

  async expectError(message: string) {
    await this.errorMessage.waitFor({ state: 'visible' })
    const text = await this.errorMessage.textContent()
    return text?.includes(message)
  }

  async expectRedirectToDashboard() {
    await this.page.waitForURL('**/dashboard**', { timeout: 15000 })
  }

  async expectRedirectToLogin() {
    await this.page.waitForURL('**/login**', { timeout: 10000 })
  }

  async isSubmitButtonDisabled(): Promise<boolean> {
    return await this.submitButton.isDisabled()
  }
}
