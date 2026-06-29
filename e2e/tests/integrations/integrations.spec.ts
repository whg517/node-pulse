import { test, expect } from '../fixtures/test'
import { expectPageTitle, gotoRoute } from '../support/selectors'

const stamp = () => Date.now().toString().slice(-6)

test.describe('integrations', () => {
  test('shows webhook management', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/integrations/webhooks')
    await expectPageTitle(page, 'Webhooks')
    await expect(page.getByRole('button', { name: 'Add Webhook' })).toBeVisible()
  })

  test('shows system health checks', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/integrations/health')
    await expectPageTitle(page, 'System Health')
    await expect(page.getByRole('button', { name: /Refresh/ })).toBeVisible()
  })
})

test.describe('webhooks CRUD', () => {
  test('creates a webhook', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/integrations/webhooks')
    await expectPageTitle(page, 'Webhooks')

    const url = `https://example.test/hook-${stamp()}`
    await page.getByRole('button', { name: 'Add Webhook' }).click()
    const dialog = page.getByRole('dialog')
    await expect(dialog).toBeVisible()
    await dialog.locator('#webhook-url').fill(url)
    await dialog.getByRole('button', { name: 'Add Webhook' }).click()

    await expect(dialog).toBeHidden()
    await expect(page.getByText(url).first()).toBeVisible({ timeout: 10_000 })
  })

  test('rejects a non-HTTPS webhook URL', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/integrations/webhooks')
    await page.getByRole('button', { name: 'Add Webhook' }).click()

    const dialog = page.getByRole('dialog')
    await dialog.locator('#webhook-url').fill('http://insecure.example.test/hook')
    await dialog.getByRole('button', { name: 'Add Webhook' }).click()

    await expect(dialog).toBeVisible()
    await expect(dialog.getByText(/https/i)).toBeVisible()
  })

  test('previews the webhook payload', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/integrations/webhooks')
    await page.getByRole('button', { name: 'Add Webhook' }).click()

    const dialog = page.getByRole('dialog')
    // Default URL is empty; fill a valid one so preview can run.
    await dialog.locator('#webhook-url').fill(`https://example.test/preview-${stamp()}`)
    await dialog.getByRole('button', { name: 'Preview Payload' }).click()

    // A <pre> with the rendered JSON appears.
    await expect(dialog.locator('pre')).toBeVisible({ timeout: 10_000 })
  })

  test('deletes a webhook after confirming', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/integrations/webhooks')

    // Create a throwaway webhook.
    const url = `https://example.test/del-${stamp()}`
    await page.getByRole('button', { name: 'Add Webhook' }).click()
    const createDialog = page.getByRole('dialog')
    await createDialog.locator('#webhook-url').fill(url)
    await createDialog.getByRole('button', { name: 'Add Webhook' }).click()
    await expect(createDialog).toBeHidden()
    await expect(page.getByText(url).first()).toBeVisible({ timeout: 10_000 })

    const row = page.getByRole('row').filter({ hasText: url }).first()
    await row.getByRole('button', { name: /delete/i }).click()

    const confirm = page.getByRole('dialog').filter({ hasText: 'Delete Webhook' })
    await expect(confirm).toBeVisible()
    await confirm.getByRole('button', { name: 'Delete' }).click()

    await expect(confirm).toBeHidden()
    await expect(page.getByText(url)).toHaveCount(0, { timeout: 10_000 })
  })

  test('shows feedback when testing a webhook against an unreachable endpoint', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/integrations/webhooks')

    // Create a webhook pointing at a host that will not accept the delivery.
    const url = `https://10.255.255.1/hook-${stamp()}`
    await page.getByRole('button', { name: 'Add Webhook' }).click()
    const dialog = page.getByRole('dialog')
    await dialog.locator('#webhook-url').fill(url)
    await dialog.getByRole('button', { name: 'Add Webhook' }).click()
    await expect(dialog).toBeHidden()
    await expect(page.getByText(url).first()).toBeVisible({ timeout: 10_000 })

    // Trigger a test delivery.
    const row = page.getByRole('row').filter({ hasText: url }).first()
    await row.getByRole('button', { name: /test webhook/i }).click()

    // The page shows a feedback notice (success or failure text appears).
    await expect(page.getByText(/delivered successfully|delivery failed|test failed/i)).toBeVisible({ timeout: 30_000 })
  })
})
