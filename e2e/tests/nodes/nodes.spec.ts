import { test, expect } from '../fixtures/test'
import { expectPageTitle } from '../support/selectors'

test.describe('nodes', () => {
  test('shows node management controls', async ({ authenticatedPage: page }) => {
    await page.goto('/nodes')

    await expectPageTitle(page, 'Node Management')
    await expect(page.getByRole('button', { name: 'Add New Node' })).toBeVisible()
  })

  test('shows probe management', async ({ authenticatedPage: page }) => {
    await page.goto('/nodes/probes')

    await expectPageTitle(page, 'Probe Management')
    await expect(page.getByRole('button', { name: 'Add Probe' })).toBeVisible()
  })

  test('shows beacon configuration', async ({ authenticatedPage: page }) => {
    await page.goto('/beacons/config')

    await expectPageTitle(page, 'Beacon Configuration')
    await expect(page.getByText('Select Node')).toBeVisible()
  })
})
