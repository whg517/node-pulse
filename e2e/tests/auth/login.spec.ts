import { test, expect } from '../fixtures/test'
import { signIn } from '../support/auth'

test.describe('authentication', () => {
  test('shows the login form', async ({ page }) => {
    await page.goto('/login')

    await expect(page.getByText('NodePulse').first()).toBeVisible()
    await expect(page.getByLabel('Username')).toBeVisible()
    await expect(page.getByLabel('Password')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Sign in' })).toBeEnabled()
  })

  test('redirects protected routes to login', async ({ page }) => {
    await page.goto('/dashboard')

    await expect(page).toHaveURL(/\/login/)
    await expect(page.getByText('NodePulse').first()).toBeVisible()
  })

  test('keeps users on login after invalid credentials', async ({ page }) => {
    await page.goto('/login')
    await page.getByLabel('Username').fill('invalid')
    await page.getByLabel('Password').fill('wrong-password')
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page).toHaveURL(/\/login/)
    await expect(page.getByText(/Invalid username or password|Connection failed/)).toBeVisible()
  })

  test('signs in and lands on the dashboard', async ({ page }) => {
    await signIn(page)

    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
    await expect(page.getByText('NodePulse').first()).toBeVisible()
  })

  test('returns to the originally requested route after login', async ({ page }) => {
    await page.goto('/nodes')
    await expect(page).toHaveURL(/\/login/)

    await page.getByLabel('Username').fill('admin')
    await page.getByLabel('Password').fill('Admin123')
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page).toHaveURL(/\/nodes(?:$|[?#])/, { timeout: 15_000 })
  })
})
