import { test, expect } from '../fixtures/test'
import { signIn, signOut } from '../support/auth'
import { gotoRoute } from '../support/selectors'

test.describe('authentication', () => {
  test('shows the login form', async ({ page }) => {
    await gotoRoute(page, '/login')

    await expect(page.getByLabel('Username')).toBeVisible({ timeout: 15_000 })
    await expect(page.getByLabel('Password')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Sign in' })).toBeEnabled()
  })

  test('redirects protected routes to login', async ({ page }) => {
    await gotoRoute(page, '/dashboard')

    await expect(page).toHaveURL(/\/login/)
    await expect(page.getByLabel('Username')).toBeVisible({ timeout: 15_000 })
  })

  test('keeps users on login after invalid credentials', async ({ page }) => {
    await gotoRoute(page, '/login')
    await expect(page.getByLabel('Username')).toBeVisible({ timeout: 15_000 })
    await page.getByLabel('Username').fill('invalid')
    await page.getByLabel('Password').fill('wrong-password')
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page).toHaveURL(/\/login/)
    await expect(page.getByText(/Invalid username or password|Connection failed/)).toBeVisible()
  })

  test('signs in and lands on the dashboard', async ({ page }) => {
    await signIn(page)

    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
  })

  test('returns to the originally requested route after login', async ({ page }) => {
    await gotoRoute(page, '/nodes')
    await expect(page).toHaveURL(/\/login/)
    await expect(page.getByLabel('Username')).toBeVisible({ timeout: 15_000 })

    await page.getByLabel('Username').fill('admin')
    await page.getByLabel('Password').fill('Admin123')
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page).toHaveURL(/\/nodes(?:$|[?#])/, { timeout: 20_000 })
  })

  test('opens the user menu and shows the logout action', async ({ authenticatedPage: page }) => {
    await page.getByTestId('user-menu-button').click()
    await expect(page.getByRole('menuitem', { name: 'Logout' })).toBeVisible()
  })

  test('logs out and returns to the login page', async ({ authenticatedPage: page }) => {
    await signOut(page)
    await expect(page.getByLabel('Username')).toBeVisible({ timeout: 15_000 })
  })

  test('protects routes again after logout', async ({ authenticatedPage: page }) => {
    await signOut(page)
    // After logout, visiting a protected route must redirect back to login.
    // A full reload forces the SPA to re-run its auth guard from a clean state.
    await gotoRoute(page, '/dashboard')
    await page.reload({ waitUntil: 'domcontentloaded' })
    await expect(page).toHaveURL(/\/login/, { timeout: 15_000 })
  })
})
