import { test, expect } from '../fixtures/test'
import { expectPageTitle } from '../support/selectors'

test.describe('settings', () => {
  test('shows preferences and can save them', async ({ authenticatedPage: page }) => {
    await page.goto('/settings/preferences')

    await expectPageTitle(page, 'Preferences')
    await page.getByRole('button', { name: 'Save Preferences' }).click()
    await expect(page.getByText('Saved')).toBeVisible()
  })

  test('shows active sessions', async ({ authenticatedPage: page }) => {
    await page.goto('/settings/sessions')

    await expectPageTitle(page, 'Session Management')
    await expect(page.getByText('Security Tips')).toBeVisible()
  })

  test('shows admin user management', async ({ authenticatedPage: page }) => {
    await page.goto('/settings/users')

    await expectPageTitle(page, 'Users')
    await expect(page.getByRole('button', { name: 'Add User' })).toBeVisible()
  })
})
