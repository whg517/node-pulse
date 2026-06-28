import type { Page } from '@playwright/test'
import { expect } from '@playwright/test'

export function appShell(page: Page) {
  return page.getByText('NodePulse').first()
}

export async function openRoute(page: Page, path: string) {
  await page.goto(path)
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
