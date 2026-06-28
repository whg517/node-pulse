import { test, expect } from '../fixtures/test'
import { expectNoAppCrash } from '../support/selectors'

test.describe('critical path smoke', () => {
  test('loads dashboard after login', async ({ authenticatedPage: page }) => {
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible()
    await expect(page.getByRole('link', { name: /Nodes/ })).toBeVisible()
    await expectNoAppCrash(page)
  })

  test('opens nodes from the sidebar', async ({ authenticatedPage: page }) => {
    await page.getByRole('link', { name: /^Nodes$/ }).click()

    await expect(page).toHaveURL(/\/nodes/)
    await expect(page.getByRole('heading', { name: 'Node Management' })).toBeVisible()
    await expectNoAppCrash(page)
  })

  test('opens alerts from the sidebar', async ({ authenticatedPage: page }) => {
    await page.getByRole('link', { name: /^Alerts$/ }).click()

    await expect(page).toHaveURL(/\/alerts\/rules/)
    await expect(page.getByRole('heading', { name: 'Alert Rules' })).toBeVisible()
    await expectNoAppCrash(page)
  })
})
