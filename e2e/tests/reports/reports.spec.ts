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

test.describe('export workflow', () => {
  test('submits an export and shows it as active', async ({ authenticatedPage: page }) => {
    // Ensure at least one node exists for the export selection.
    await gotoRoute(page, '/nodes')
    const stamp = Date.now().toString().slice(-6)
    const nodeName = `e2e-export-${stamp}`
    await page.getByRole('button', { name: 'Add New Node' }).click()
    const nodeDialog = page.getByRole('dialog')
    await nodeDialog.locator('#name').fill(nodeName)
    await nodeDialog.locator('#ip').fill('10.30.0.1')
    await nodeDialog.locator('#region').fill('e2e-test')
    await nodeDialog.getByRole('button', { name: 'Create Node' }).click()
    await expect(nodeDialog).toBeHidden()

    // Go to the export page and submit the form.
    await gotoRoute(page, '/reports/history')
    await expectPageTitle(page, 'Data Export')

    // Select the node we just created.
    await page.getByRole('checkbox', { name: new RegExp(nodeName) }).check()
    // CSV is the only enabled format (Excel shows "Planned").
    await page.getByLabel('CSV').check()
    await page.getByRole('button', { name: 'Export' }).click()

    // An active export card appears.
    await expect(page.getByText(/active exports|export task|processing|pending/i).first()).toBeVisible({ timeout: 15_000 })
  })

  test('shows Excel format as planned/unavailable', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/reports/history')
    // The Excel option is visibly disabled with a "Planned" tag.
    await expect(page.getByText('Planned')).toBeVisible()
  })

  test('requires a node selection before exporting', async ({ authenticatedPage: page }) => {
    await gotoRoute(page, '/reports/history')
    // Submit without selecting any node.
    await page.getByRole('button', { name: 'Export' }).click()
    await expect(page.getByText(/select.*node|at least one node/i)).toBeVisible({ timeout: 10_000 })
  })
})
