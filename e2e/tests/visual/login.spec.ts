/**
 * Login Page Visual Regression Tests
 *
 * Visual tests for Login page:
 * - Default view
 * - With validation errors
 * - With API errors
 */
import { test, expect } from '../../fixtures/auth.fixture'
import { LoginPage } from '../../pages'

test.describe('Login Visual Tests', () => {
  let loginPage: LoginPage

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page)
    await loginPage.goto()
  })

  test('login page default view', async ({ page }) => {
    await page.waitForTimeout(1000)
    
    await expect(page).toHaveScreenshot('login-default.png', {
      maxDiffPixels: 50,
      fullPage: true,
    })
  })

  test('login page form focused', async ({ page }) => {
    await loginPage.usernameInput.focus()
    await page.waitForTimeout(500)
    
    await expect(page).toHaveScreenshot('login-form-focused.png', {
      maxDiffPixels: 50,
    })
  })

  test('login page validation errors', async ({ page }) => {
    // Try to submit empty form
    await loginPage.submit()
    await page.waitForTimeout(1000)
    
    await expect(page).toHaveScreenshot('login-validation-errors.png', {
      maxDiffPixels: 100,
    })
  })

  test('login page invalid credentials', async ({ page }) => {
    await loginPage.login('invaliduser', 'Invalidpassword1')
    await page.waitForTimeout(2000)
    
    await expect(page).toHaveScreenshot('login-invalid-credentials.png', {
      maxDiffPixels: 100,
    })
  })

  test('login page password visible', async ({ page }) => {
    // If show password button exists
    const showPasswordBtn = page.locator('[data-testid="show-password"], button:has-text("Show")')
    
    if (await showPasswordBtn.count() > 0) {
      await loginPage.fillForm('testuser', 'Testpassword1')
      await showPasswordBtn.click()
      await page.waitForTimeout(500)
      
      await expect(page).toHaveScreenshot('login-password-visible.png', {
        maxDiffPixels: 50,
      })
    }
  })
})
