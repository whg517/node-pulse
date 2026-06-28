import { expect, type Page } from '@playwright/test'

export const ADMIN_USER = process.env.E2E_ADMIN_USER || 'admin'
export const ADMIN_PASS = process.env.E2E_ADMIN_PASS || 'Admin123'

export async function signIn(
  page: Page,
  username = ADMIN_USER,
  password = ADMIN_PASS
) {
  await page.goto('/login')
  await expect(page.getByText('NodePulse').first()).toBeVisible()
  await page.getByLabel('Username').fill(username)
  await page.getByLabel('Password').fill(password)
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page).toHaveURL(/\/dashboard(?:$|[?#])/, { timeout: 15_000 })
}
