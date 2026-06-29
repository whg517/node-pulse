import { expect, type Page } from '@playwright/test'

export const ADMIN_USER = process.env.E2E_ADMIN_USER || 'admin'
export const ADMIN_PASS = process.env.E2E_ADMIN_PASS || 'Admin123'

export async function signIn(
  page: Page,
  username = ADMIN_USER,
  password = ADMIN_PASS
) {
  await page.goto('/login', { waitUntil: 'domcontentloaded' })
  // Wait for the SPA to actually render the form. The shell HTML loads before
  // JS hydrates, and on a containerized stack the first cold parse of the JS
  // bundle can be slow, so allow generous headroom here.
  const usernameInput = page.getByLabel('Username')
  await expect(usernameInput).toBeVisible({ timeout: 30_000 })
  await usernameInput.fill(username)
  await page.getByLabel('Password').fill(password)
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page).toHaveURL(/\/dashboard(?:$|[?#])/, { timeout: 20_000 })
}

// signOut triggers the logout flow via the Header user-menu dropdown and
// waits until the SPA returns to /login.
export async function signOut(page: Page) {
  await page.getByTestId('user-menu-button').click()
  await page.getByRole('menuitem', { name: 'Logout' }).click()
  await expect(page).toHaveURL(/\/login/, { timeout: 15_000 })
}
