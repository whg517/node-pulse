import { test, expect } from '../fixtures/test'
import { expectNoAppCrash, expectPageTitle } from '../support/selectors'

test.describe('alerts', () => {
  test('shows alert rules and creation actions', async ({ authenticatedPage: page }) => {
    await page.goto('/alerts/rules')

    await expectPageTitle(page, 'Alert Rules')
    await expect(page.getByRole('button', { name: 'Create Alert Rule' })).toBeVisible()
  })

  test('shows alert records page', async ({ authenticatedPage: page }) => {
    await page.goto('/alerts/records')

    await expectPageTitle(page, 'Alert History')
    await expect(page.getByRole('button', { name: /Export CSV/ })).toBeVisible()
  })

  test('shows alert history filters', async ({ authenticatedPage: page }) => {
    await page.goto('/alerts/history')

    await expectPageTitle(page, 'Alert History')
    await expect(page.getByText('Filters')).toBeVisible()
    await expectNoAppCrash(page)
  })
})
