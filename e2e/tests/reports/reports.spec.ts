import { test, expect } from '../fixtures/test'
import { expectPageTitle, gotoRoute } from '../support/selectors'

test.describe('reports and performance', () => {
  test('shows report generation', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/reports')

    await expectPageTitle(page, 'Reports')
    await expect(page.getByRole('heading', { name: 'Generate Report' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Export History' }).first()).toBeVisible()
  })

  test('shows export workflow', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/reports/history')

    await expectPageTitle(page, 'Data Export')
    await expect(page.getByRole('heading', { name: 'Create New Export' })).toBeVisible()
  })

  test('shows performance dashboard', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/performance')

    await expectPageTitle(page, 'Performance Dashboard')
    await expect(page.getByRole('button', { name: /Refresh/ })).toBeVisible()
  })
})
