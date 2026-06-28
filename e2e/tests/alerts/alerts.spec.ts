import { test, expect } from '../fixtures/test'
import { expectNoAppCrash, expectPageTitle, gotoRoute } from '../support/selectors'

const stamp = () => Date.now().toString().slice(-6)

test.describe('alerts', () => {
  test('shows alert rules and creation actions', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/alerts/rules')
    await expectPageTitle(page, 'Alert Rules')
    await expect(page.getByRole('button', { name: 'Create Alert Rule' })).toBeVisible()
  })

  test('shows alert records page', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/alerts/records')
    await expectPageTitle(page, 'Alert History')
    await expect(page.getByRole('button', { name: /Export CSV/ })).toBeVisible()
  })

  test('shows alert history filters', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/alerts/history')
    await expectPageTitle(page, 'Alert History')
    await expect(page.getByText('Filters')).toBeVisible()
    await expectNoAppCrash(page)
  })
})

test.describe('alert rules CRUD', () => {
  test('creates a latency alert rule', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/alerts/rules')
    await expectPageTitle(page, 'Alert Rules')

    await page.getByRole('button', { name: 'Create Alert Rule' }).click()
    const dialog = page.getByRole('dialog')
    await expect(dialog).toBeVisible()

    // Defaults are usually fine (Latency / Global); just set a threshold.
    await dialog.locator('#threshold').fill('250')
    await dialog.getByRole('button', { name: 'Create Alert Rule' }).click()

    await expect(dialog).toBeHidden()
    // A global latency rule with threshold 250 should appear.
    await expect(page.getByText('250').first()).toBeVisible({ timeout: 10_000 })
  })

  test('rejects a non-positive threshold', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/alerts/rules')
    await page.getByRole('button', { name: 'Create Alert Rule' }).click()

    const dialog = page.getByRole('dialog')
    await dialog.locator('#threshold').fill('0')
    await dialog.getByRole('button', { name: 'Create Alert Rule' }).click()

    await expect(dialog).toBeVisible()
    await expect(dialog.getByText(/greater than 0/i)).toBeVisible()
  })

  test('toggles a rule enabled state', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/alerts/rules')

    // Create a rule first so there's at least one switch to toggle.
    await page.getByRole('button', { name: 'Create Alert Rule' }).click()
    const createDialog = page.getByRole('dialog')
    await createDialog.locator('#threshold').fill(`3${stamp().slice(-2)}`)
    await createDialog.getByRole('button', { name: 'Create Alert Rule' }).click()
    await expect(createDialog).toBeHidden()

    // The first Switch in the rules table toggles enabled.
    const switches = page.locator('table button[role="switch"]')
    const firstSwitch = switches.first()
    const before = await firstSwitch.getAttribute('aria-checked')
    await firstSwitch.click()
    // The state flips; allow the API round-trip.
    await expect(firstSwitch).not.toHaveAttribute('aria-checked', before ?? '', { timeout: 10_000 })
  })

  test('deletes a rule after confirming', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/alerts/rules')

    // Create a throwaway rule.
    await page.getByRole('button', { name: 'Create Alert Rule' }).click()
    const createDialog = page.getByRole('dialog')
    const threshold = `4${stamp().slice(-2)}`
    await createDialog.locator('#threshold').fill(threshold)
    await createDialog.getByRole('button', { name: 'Create Alert Rule' }).click()
    await expect(createDialog).toBeHidden()
    await expect(page.getByText(threshold).first()).toBeVisible({ timeout: 10_000 })

    const row = page.getByRole('row').filter({ hasText: threshold }).first()
    await row.getByRole('button', { name: /delete/i }).click()

    const confirm = page.getByRole('dialog').filter({ hasText: 'Delete Alert Rule' })
    await expect(confirm).toBeVisible()
    await confirm.getByRole('button', { name: 'Delete' }).click()

    await expect(confirm).toBeHidden()
    await expect(page.getByText(threshold)).toHaveCount(0, { timeout: 10_000 })
  })
})
