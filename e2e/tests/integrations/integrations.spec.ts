import { test, expect } from '../fixtures/test'
import { expectPageTitle } from '../support/selectors'

test.describe('integrations', () => {
  test('shows webhook management', async ({ authenticatedPage: page }) => {
    await page.goto('/integrations/webhooks')

    await expectPageTitle(page, 'Webhooks')
    await expect(page.getByRole('button', { name: 'Add Webhook' })).toBeVisible()
  })

  test('shows system health checks', async ({ authenticatedPage: page }) => {
    await page.goto('/integrations/health')

    await expectPageTitle(page, 'System Health')
    await expect(page.getByRole('button', { name: /Refresh/ })).toBeVisible()
  })
})
