import { test, expect } from '../fixtures/test'
import { expectPageTitle, gotoRoute } from '../support/selectors'

// Unique suffix so repeated runs don't collide on uniqueness constraints.
const stamp = () => Date.now().toString().slice(-6)

test.describe('nodes', () => {
  test('shows node management controls', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/nodes')
    await expectPageTitle(page, 'Node Management')
    await expect(page.getByRole('button', { name: 'Add New Node' })).toBeVisible()
  })

  test('shows probe management', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/nodes/probes')
    await expectPageTitle(page, 'Probe Management')
    await expect(page.getByRole('button', { name: 'Add Probe' })).toBeVisible()
  })

  test('shows beacon configuration', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/beacons/config')
    await expectPageTitle(page, 'Beacon Configuration')
    await expect(page.getByText('Select Node')).toBeVisible()
  })
})

test.describe('nodes CRUD', () => {
  test('creates a node via the form', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/nodes')
    await expectPageTitle(page, 'Node Management')

    const name = `e2e-node-${stamp()}`
    await page.getByRole('button', { name: 'Add New Node' }).click()

    const dialog = page.getByRole('dialog')
    await expect(dialog).toBeVisible()
    await dialog.locator('#name').fill(name)
    await dialog.locator('#ip').fill('10.10.0.1')
    await dialog.locator('#region').fill('e2e-test')
    await dialog.locator('#tags').fill('e2e, smoke')
    await dialog.getByRole('button', { name: 'Create Node' }).click()

    await expect(dialog).toBeHidden()
    await expect(page.getByText(name).first()).toBeVisible({ timeout: 10_000 })
  })

  test('validates required node fields', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/nodes')
    await page.getByRole('button', { name: 'Add New Node' }).click()

    const dialog = page.getByRole('dialog')
    await dialog.getByRole('button', { name: 'Create Node' }).click()

    await expect(dialog).toBeVisible()
    await expect(dialog.locator('p').filter({ hasText: /required|invalid/i }).first()).toBeVisible()
  })

  test('rejects an invalid IP address', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/nodes')
    await page.getByRole('button', { name: 'Add New Node' }).click()

    const dialog = page.getByRole('dialog')
    await dialog.locator('#name').fill(`e2e-badip-${stamp()}`)
    await dialog.locator('#ip').fill('not-an-ip')
    await dialog.locator('#region').fill('e2e-test')
    await dialog.getByRole('button', { name: 'Create Node' }).click()

    await expect(dialog).toBeVisible()
    await expect(dialog.getByText(/invalid ip/i)).toBeVisible()
  })

  test('edits a node name', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/nodes')

    const name = `e2e-edit-${stamp()}`
    await page.getByRole('button', { name: 'Add New Node' }).click()
    const createDialog = page.getByRole('dialog')
    await createDialog.locator('#name').fill(name)
    await createDialog.locator('#ip').fill('10.10.0.2')
    await createDialog.locator('#region').fill('e2e-test')
    await createDialog.getByRole('button', { name: 'Create Node' }).click()
    await expect(createDialog).toBeHidden()

    const row = page.getByRole('row').filter({ hasText: name }).first()
    await row.getByRole('button', { name: /edit/i }).click()

    const editDialog = page.getByRole('dialog')
    await expect(editDialog.getByText('Edit Node')).toBeVisible()
    const renamed = `${name}-renamed`
    await editDialog.locator('#name').fill('')
    await editDialog.locator('#name').fill(renamed)
    await editDialog.getByRole('button', { name: 'Save Changes' }).click()

    await expect(editDialog).toBeHidden()
    await expect(page.getByText(renamed).first()).toBeVisible({ timeout: 10_000 })
  })

  test('deletes a node after confirming', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/nodes')

    const name = `e2e-del-${stamp()}`
    await page.getByRole('button', { name: 'Add New Node' }).click()
    const createDialog = page.getByRole('dialog')
    await createDialog.locator('#name').fill(name)
    await createDialog.locator('#ip').fill('10.10.0.3')
    await createDialog.locator('#region').fill('e2e-test')
    await createDialog.getByRole('button', { name: 'Create Node' }).click()
    await expect(createDialog).toBeHidden()
    await expect(page.getByText(name).first()).toBeVisible({ timeout: 10_000 })

    const row = page.getByRole('row').filter({ hasText: name }).first()
    await row.getByRole('button', { name: /delete/i }).click()

    const confirm = page.getByRole('dialog').filter({ hasText: 'Delete Node' })
    await expect(confirm).toBeVisible()
    await confirm.getByRole('button', { name: 'Delete' }).click()

    await expect(confirm).toBeHidden()
    await expect(page.getByText(name)).toHaveCount(0, { timeout: 10_000 })
  })
})
