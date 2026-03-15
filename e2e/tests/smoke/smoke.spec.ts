/**
 * Smoke Tests - Core Functionality
 * 
 * Quick tests to verify the most critical application functions.
 * Should run in under 2 minutes and catch major regressions.
 * 
 * Test scope:
 * - Authentication (login/logout)
 * - Dashboard loads
 * - Basic navigation
 * - Critical API endpoints
 */
import { test, expect } from '../../fixtures/auth.fixture'
import { LoginPage, DashboardPage, NodesPage } from '../../pages'

const ADMIN_USERNAME = process.env.TEST_ADMIN_USER || 'admin'
const ADMIN_PASSWORD = process.env.TEST_ADMIN_PASS || 'Admin123'

test.describe('Smoke Tests - Authentication', () => {
  test('SMOKE-001: admin can login', async ({ page }) => {
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.fillForm(ADMIN_USERNAME, ADMIN_PASSWORD)
    await loginPage.submit()
    await loginPage.expectRedirectToDashboard()
  })

  test('SMOKE-002: login page loads correctly', async ({ page }) => {
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.expectFormVisible()
    await expect(loginPage.usernameInput).toBeVisible()
    await expect(loginPage.passwordInput).toBeVisible()
    await expect(loginPage.submitButton).toBeVisible()
  })

  test('SMOKE-003: invalid credentials show error', async ({ page }) => {
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.fillForm('invalid', 'Invalid123')
    await loginPage.submit()
    await loginPage.waitForReady()
    // Should show error or stay on login page
    await expect(page).toHaveURL(/.*login/)
  })
})

test.describe('Smoke Tests - Dashboard', () => {
  test('SMOKE-004: dashboard loads for admin', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage)
    await dashboardPage.goto()
    await dashboardPage.expectMetricsVisible()
    await dashboardPage.expectTitle()
  })

  test('SMOKE-005: dashboard shows navigation', async ({ adminPage }) => {
    await adminPage.goto('/dashboard')
    await adminPage.waitForLoadState('networkidle')
    
    // Check for navigation elements
    const navLocator = adminPage.locator('aside nav, aside')
    await expect(navLocator.first()).toBeVisible()
  })

  test('SMOKE-006: can navigate to nodes page', async ({ adminPage }) => {
    await adminPage.goto('/nodes')
    await adminPage.waitForLoadState('networkidle')
    await expect(adminPage).toHaveURL(/.*nodes/)
  })
})

test.describe('Smoke Tests - API Health', () => {
  test('SMOKE-007: health endpoint responds', async ({ request }) => {
    const response = await request.get('/api/v1/health')
    expect(response.ok()).toBeTruthy()
    expect(response.status()).toBe(200)
  })

  test('SMOKE-008: metrics endpoint responds', async ({ request }) => {
    const response = await request.get('/metrics')
    expect(response.ok()).toBeTruthy()
    expect(response.status()).toBe(200)
  })

  test('SMOKE-009: authenticated user can access nodes API', async ({ request }) => {
    const loginResponse = await request.post('/api/v1/auth/login', {
      data: { username: ADMIN_USERNAME, password: ADMIN_PASSWORD }
    })
    expect(loginResponse.ok()).toBeTruthy()
    const loginData = await loginResponse.json() as { data?: { access_token?: string } }
    const accessToken = loginData.data?.access_token
    expect(accessToken).toBeTruthy()

    const response = await request.get('/api/v1/nodes', {
      headers: { Authorization: `Bearer ${accessToken}` }
    })
    expect(response.ok()).toBeTruthy()
    expect(response.status()).toBe(200)
  })
})

test.describe('Smoke Tests - Core Pages', () => {
  test('SMOKE-010: alerts page loads', async ({ adminPage }) => {
    await adminPage.goto('/alerts/rules')
    await adminPage.waitForLoadState('networkidle')
    await expect(adminPage).toHaveURL(/.*alerts/)
  })

  test('SMOKE-011: webhooks page loads', async ({ adminPage }) => {
    await adminPage.goto('/integrations/webhooks')
    await adminPage.waitForLoadState('networkidle')
    await expect(adminPage).toHaveURL(/.*webhooks/)
  })

  test('SMOKE-012: sessions page loads', async ({ adminPage }) => {
    await adminPage.goto('/settings/sessions')
    await adminPage.waitForLoadState('networkidle')
    
    // Should have a table or empty state
    const table = adminPage.locator('table')
    const emptyState = adminPage.locator('.text-center')
    await expect(table.or(emptyState).first()).toBeVisible()
  })
})

test.describe('Smoke Tests - Critical User Journey', () => {
  test('SMOKE-013: complete login to dashboard flow', async ({ page }) => {
    const loginPage = new LoginPage(page)
    const dashboardPage = new DashboardPage(page)
    
    // Login
    await loginPage.goto()
    await loginPage.fillForm(ADMIN_USERNAME, ADMIN_PASSWORD)
    await loginPage.submit()
    await loginPage.expectRedirectToDashboard()
    
    // Verify dashboard
    await dashboardPage.expectMetricsVisible()
    await dashboardPage.expectNodesVisible()
  })

  test('SMOKE-014: can logout successfully', async ({ page }) => {
    const loginPage = new LoginPage(page)
    
    // Login first
    await loginPage.goto()
    await loginPage.fillForm(ADMIN_USERNAME, ADMIN_PASSWORD)
    await loginPage.submit()
    await loginPage.expectRedirectToDashboard()
    
    // Logout
    await page.goto('/dashboard')
    await page.waitForLoadState('networkidle')
    await page.getByRole('button', { name: /admin/i }).click()
    await page.getByRole('button', { name: /logout/i }).click()
    await page.waitForURL(/.*login/, { timeout: 10000 })
    
    // Verify on login page
    await expect(page).toHaveURL(/.*login/)
  })
})
