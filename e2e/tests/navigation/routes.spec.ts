import { test, expect } from '../fixtures/test'
import { expectPageTitle } from '../support/selectors'

const protectedRoutes = [
  { path: '/dashboard', title: 'Dashboard' },
  { path: '/nodes', title: 'Node Management' },
  { path: '/nodes/probes', title: 'Probe Management' },
  { path: '/beacons/config', title: 'Beacon Configuration' },
  { path: '/alerts/rules', title: 'Alert Rules' },
  { path: '/alerts/records', title: 'Alert History' },
  { path: '/alerts/history', title: 'Alert History' },
  { path: '/performance', title: 'Performance Dashboard' },
  { path: '/reports', title: 'Reports' },
  { path: '/reports/history', title: 'Data Export' },
  { path: '/integrations/webhooks', title: 'Webhooks' },
  { path: '/integrations/health', title: 'System Health' },
  { path: '/settings/preferences', title: 'Preferences' },
  { path: '/settings/sessions', title: 'Session Management' },
  { path: '/settings/users', title: 'Users' },
]

test.describe('protected route inventory', () => {
  for (const route of protectedRoutes) {
    test(`renders ${route.path}`, async ({ authenticatedPage: page }) => {
      await page.goto(route.path)

      await expect(page).toHaveURL(new RegExp(route.path.replace(/\//g, '\\/')))
      await expectPageTitle(page, route.title)
    })
  }

  test('supports legacy aliases', async ({ authenticatedPage: page }) => {
    await page.goto('/webhooks')
    await expect(page).toHaveURL(/\/integrations\/webhooks/)

    await page.goto('/sessions')
    await expect(page).toHaveURL(/\/settings\/sessions/)

    await page.goto('/comparison')
    await expect(page).toHaveURL(/\/nodes\/comparison/)
  })
})
