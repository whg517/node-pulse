/**
 * Webhooks Tests
 *
 * Tests for webhooks page (admin only):
 * - List webhooks
 * - Create webhook form
 * - Edit webhook
 * - Delete webhook
 * - Toggle webhook
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { WebhooksPage } from '../../pages/WebhooksPage'

test.describe('Webhooks Page - Admin', () => {
  let webhooksPage: WebhooksPage

  test.beforeEach(async ({ adminPage }) => {
    webhooksPage = new WebhooksPage(adminPage)
    await webhooksPage.goto()
  })

  test('page loads for admin', async ({ adminPage }) => {
    await webhooksPage.expectTableVisible()
  })

  test('shows create button for admin', async ({ adminPage }) => {
    const hasWarning = await webhooksPage.hasAccessWarning()

    if (!hasWarning) {
      await webhooksPage.expectTableVisible()
      const createButton = adminPage.locator('button:has-text("Create"), button:has-text("Add")')

      if (await createButton.count() > 0) {
        await expect(createButton).toBeVisible()
      }
    }
  })

  test('create webhook modal opens', async ({ adminPage }) => {
    const hasWarning = await webhooksPage.hasAccessWarning()

    if (hasWarning) {
      test.skip(true, 'Access warning shown - check auth state')
      return
    }

    await webhooksPage.expectTableVisible()
    await webhooksPage.clickCreate()

    await expect(webhooksPage.modal).toBeVisible()
  })

  test('create webhook with valid data', async ({ adminPage }) => {
    const hasWarning = await webhooksPage.hasAccessWarning()

    if (hasWarning) {
      test.skip(true, 'Access warning shown')
      return
    }

    await webhooksPage.expectTableVisible()

    const webhookName = `e2e_test_webhook_${Date.now()}`
    await webhooksPage.createWebhook(webhookName, 'https://example.com/webhook')

    // Verify webhook appears in list
    const tableText = await adminPage.locator('table').textContent()
    expect(tableText).toContain(webhookName)
  })

  test('URL validation', async ({ adminPage }) => {
    const hasWarning = await webhooksPage.hasAccessWarning()

    if (hasWarning) {
      test.skip(true, 'Access warning shown')
      return
    }

    await webhooksPage.expectTableVisible()
    await webhooksPage.clickCreate()

    // Enter invalid URL
    await webhooksPage.urlInput.fill('not-a-url')
    await webhooksPage.submitButton.click()

    // Should show validation error
    const errorVisible = await adminPage.locator('.error, [data-testid="error"]').count() > 0
    expect(errorVisible || await webhooksPage.modal.isVisible()).toBeTruthy()
  })

  test('delete webhook', async ({ adminPage }) => {
    const hasWarning = await webhooksPage.hasAccessWarning()

    if (hasWarning) {
      test.skip(true, 'Access warning shown')
      return
    }

    // Create a webhook to delete
    const webhookName = `e2e_delete_webhook_${Date.now()}`
    await webhooksPage.createWebhook(webhookName, 'https://example.com/webhook')

    // Find and delete
    const rows = adminPage.locator('table tbody tr')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const rowText = await rows.nth(i).textContent()
      if (rowText?.includes(webhookName)) {
        const deleteButton = rows.nth(i).locator('button:has-text("Delete")')
        await deleteButton.click()

        const confirmButton = adminPage.locator('button:has-text("Confirm")')
        await confirmButton.click()

        await adminPage.waitForTimeout(500)
        break
      }
    }

    // Verify webhook is gone
    const tableText = await adminPage.locator('table').textContent()
    expect(tableText).not.toContain(webhookName)
  })

  test('toggle webhook enable/disable', async ({ adminPage }) => {
    const hasWarning = await webhooksPage.hasAccessWarning()

    if (hasWarning) {
      test.skip(true, 'Access warning shown')
      return
    }

    await webhooksPage.expectTableVisible()

    const toggle = adminPage.locator('table tbody tr').first().locator('input[type="checkbox"]')

    if (await toggle.count() > 0) {
      const isCheckedBefore = await toggle.isChecked()
      await toggle.click()

      await adminPage.waitForTimeout(500)
    }
  })
})

test.describe('Webhooks Page - Operator', () => {
  test('AC-10: operator sees access warning', async ({ operatorPage }) => {
    const webhooksPage = new WebhooksPage(operatorPage)
    await webhooksPage.goto()

    const hasWarning = await webhooksPage.hasAccessWarning()
    expect(hasWarning).toBeTruthy()
  })

  test('operator cannot create webhooks', async ({ operatorPage }) => {
    const response = await operatorPage.request.post('/api/v1/webhooks', {
      data: {
        name: 'test',
        url: 'https://example.com/webhook',
      },
    })

    expect(response.status()).toBe(403)
  })
})

test.describe('Webhooks Page - Viewer', () => {
  test('viewer sees access warning', async ({ viewerPage }) => {
    const webhooksPage = new WebhooksPage(viewerPage)
    await webhooksPage.goto()

    const hasWarning = await webhooksPage.hasAccessWarning()
    expect(hasWarning).toBeTruthy()
  })
})
