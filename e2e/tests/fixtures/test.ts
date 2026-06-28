import { test as base, expect, type Page } from '@playwright/test'
import { signIn } from '../support/auth'

type NodePulseFixtures = {
  authenticatedPage: Page
}

export const test = base.extend<NodePulseFixtures>({
  authenticatedPage: async ({ page }, use) => {
    await signIn(page)
    await use(page)
  },
})

export { expect }
