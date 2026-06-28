import { test, expect } from '../fixtures/test'
import { expectPageTitle, gotoRoute } from '../support/selectors'

const stamp = () => Date.now().toString().slice(-6)

test.describe('users CRUD', () => {
  test('creates a viewer user', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/settings/users')
    await expectPageTitle(page, 'Users')

    const username = `e2e-user-${stamp()}`
    await page.getByRole('button', { name: 'Add User' }).click()
    const dialog = page.getByRole('dialog')
    await expect(dialog).toBeVisible()
    await dialog.locator('#user-username').fill(username)
    await dialog.locator('#user-password').fill('TempPass123')
    await dialog.locator('#user-role').selectOption('Viewer')
    await dialog.getByRole('button', { name: 'Save' }).click()

    await expect(dialog).toBeHidden()
    await expect(page.getByText(username).first()).toBeVisible({ timeout: 10_000 })
  })

  test('requires username and password', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/settings/users')
    await page.getByRole('button', { name: 'Add User' }).click()

    const dialog = page.getByRole('dialog')
    await dialog.getByRole('button', { name: 'Save' }).click()

    await expect(dialog).toBeVisible()
    await expect(dialog.getByText(/required/i)).toBeVisible()
  })

  test('deletes a user after confirming', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/settings/users')

    // Create a throwaway user.
    const username = `e2e-del-${stamp()}`
    await page.getByRole('button', { name: 'Add User' }).click()
    const createDialog = page.getByRole('dialog')
    await createDialog.locator('#user-username').fill(username)
    await createDialog.locator('#user-password').fill('TempPass123')
    await createDialog.getByRole('button', { name: 'Save' }).click()
    await expect(createDialog).toBeHidden()
    await expect(page.getByText(username).first()).toBeVisible({ timeout: 10_000 })

    const row = page.getByRole('row').filter({ hasText: username }).first()
    await row.getByRole('button', { name: /delete/i }).click()

    const confirm = page.getByRole('dialog').filter({ hasText: /delete this user/i })
    await expect(confirm).toBeVisible()
    await confirm.getByRole('button', { name: 'Delete' }).click()

    await expect(confirm).toBeHidden()
    await expect(page.getByText(username)).toHaveCount(0, { timeout: 10_000 })
  })
})
