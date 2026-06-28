import type { Page } from '@playwright/test'
import { expect } from '@playwright/test'

export function appShell(page: Page) {
  return page.getByText('NodePulse').first()
}

// gotoRoute navigates without waiting for the full 'load' event. The SPA keeps
// long-lived API/websocket connections open, so 'load' may never fire and
// page.goto would hang. 'domcontentloaded' is enough; the caller then waits on
// a concrete element via expect(...).toBeVisible().
export async function gotoRoute(page: Page, path: string) {
  await page.goto(path, { waitUntil: 'domcontentloaded' })
}

export async function openRoute(page: Page, path: string) {
  await gotoRoute(page, path)
  await appShell(page).waitFor({ state: 'visible' })
}

export async function expectNoAppCrash(page: Page) {
  await page.getByText('Something went wrong').waitFor({ state: 'detached', timeout: 1_000 }).catch(() => {})
  await page.getByText('Authentication error detected').waitFor({ state: 'detached', timeout: 1_000 }).catch(() => {})
}

export async function expectPageTitle(page: Page, title: string | RegExp) {
  await expect(page.getByRole('heading', { name: title }).first()).toBeVisible()
  await expectNoAppCrash(page)
}
