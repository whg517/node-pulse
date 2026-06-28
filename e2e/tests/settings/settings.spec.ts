import { test, expect } from '../fixtures/test'
import { expectPageTitle, gotoRoute } from '../support/selectors'

test.describe('settings', () => {
  test('saves preferences and confirms', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/settings/preferences')
    await expectPageTitle(page, 'Preferences')
    await page.getByRole('button', { name: 'Save Preferences' }).click()
    await expect(page.getByText(/saved/i)).toBeVisible({ timeout: 5_000 })
  })

  test('toggles dark mode', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/settings/preferences')
    const toggle = page.getByRole('switch', { name: /dark mode|light mode/i })
    const before = await toggle.getAttribute('aria-checked')
    await toggle.click()
    await expect(toggle).not.toHaveAttribute('aria-checked', before ?? '', { timeout: 5_000 })
  })

  test('lists the current session', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/settings/sessions')
    await expectPageTitle(page, 'Session Management')
    // The active session is marked "Current".
    await expect(page.getByText('Current').first()).toBeVisible()
  })

  test('shows admin user management', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/settings/users')
    await expectPageTitle(page, 'Users')
    await expect(page.getByRole('button', { name: 'Add User' })).toBeVisible()
  })
})
