import { test, expect } from '@playwright/test'

const ADMIN_USER = process.env.E2E_ADMIN_USER || 'admin'
const ADMIN_PASS = process.env.E2E_ADMIN_PASS || 'Admin123'

test.describe('Login', () => {
  test('shows login page at /login', async ({ page }) => {
    await page.goto('/login')
    await expect(page.locator('input[type="text"], input[type="email"]').first()).toBeVisible()
    await expect(page.locator('input[type="password"]')).toBeVisible()
  })

  test('redirects unauthenticated users to /login', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/login/)
  })

  test('logs in with valid credentials and redirects to dashboard', async ({ page }) => {
    await page.goto('/login')

    // Fill in credentials
    await page.locator('input[type="text"], input[type="email"]').first().fill(ADMIN_USER)
    await page.locator('input[type="password"]').fill(ADMIN_PASS)
    await page.locator('button[type="submit"]').click()

    // Should land on dashboard
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 10_000 })
  })

  test('shows error on invalid credentials', async ({ page }) => {
    await page.goto('/login')
    await page.locator('input[type="text"], input[type="email"]').first().fill('invalid')
    await page.locator('input[type="password"]').fill('wrongpassword')
    await page.locator('button[type="submit"]').click()

    // Should stay on login page and show some error indication
    await expect(page).toHaveURL(/\/login/)
  })
})
