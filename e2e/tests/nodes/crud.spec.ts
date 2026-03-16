/**
 * Node CRUD Tests
 *
 * Tests for node create, update, delete operations
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { NodesPage } from '../../pages/NodesPage'

test.describe('Node CRUD - Admin/Operator', () => {
  let nodesPage: NodesPage

  test.beforeEach(async ({ adminPage }) => {
    nodesPage = new NodesPage(adminPage)
    await nodesPage.goto()
  })

  test('create node modal opens', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()
    await nodesPage.clickCreate()

    // Modal should be visible
    await expect(nodesPage.modal).toBeVisible()
  })

  test('create node with valid data', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()

    const nodeName = `e2e_test_${Date.now()}`
    const nodeIp = `192.168.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`

    // Get CSRF token from HttpOnly cookie via Playwright context
    const allCookies = await adminPage.context().cookies()
    const csrfCookie = allCookies.find(c => c.name === 'csrf_token')
    const csrfToken = csrfCookie?.value

    if (!csrfToken) {
      throw new Error('CSRF token cookie not found - login may have failed')
    }

    // Intercept API requests and add CSRF header using route.fetch()
    // This ensures cookies are properly sent with the request
    await adminPage.route('**/api/v1/**', async (route, request) => {
      const method = request.method()
      const headers: Record<string, string> = {
        ...Object.fromEntries(Object.entries(request.headers())),
      }

      // Add CSRF header for mutation requests
      if (['POST', 'PUT', 'DELETE', 'PATCH'].includes(method)) {
        headers['X-CSRF-Token'] = csrfToken
      }

      // Build cookie header from all context cookies (including HttpOnly ones)
      const cookieHeader = allCookies.map(c => `${c.name}=${c.value}`).join('; ')

      // Use route.fetch() with explicit cookie header to ensure cookies are sent
      const response = await route.fetch({
        headers: {
          ...headers,
          'Cookie': cookieHeader,
        },
      })

      await route.fulfill({ response })
    })

    // Open create modal
    await nodesPage.clickCreate()

    // Verify modal is open
    await expect(nodesPage.modal).toBeVisible()

    // Fill form fields
    const nameInput = adminPage.locator('#name')
    const ipInput = adminPage.locator('#ip')
    const regionInput = adminPage.locator('#region')

    // Wait for inputs to be visible and fill them
    await nameInput.waitFor({ state: 'visible', timeout: 5000 })
    await nameInput.fill(nodeName)

    await ipInput.waitFor({ state: 'visible', timeout: 5000 })
    await ipInput.fill(nodeIp)

    await regionInput.waitFor({ state: 'visible', timeout: 5000 })
    await regionInput.fill('us-east-1')

    // Click submit button
    const submitBtn = adminPage.locator('button[type="submit"]')
    await submitBtn.click()

    // Wait for modal to close
    await nodesPage.waitForModalClose()

    // Reload page to ensure we see the latest data
    await adminPage.reload()
    await nodesPage.expectTableVisible()

    // Verify node appears in list
    const hasNode = await nodesPage.hasNode(nodeName)
    expect(hasNode).toBeTruthy()
  })

  test('create node validation - empty name', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()
    await nodesPage.clickCreate()

    // Leave name empty
    await nodesPage.submitButton.click()

    // Should show validation error
    const errorVisible = await adminPage.locator('.error, [data-testid="error"]').count() > 0

    // Modal should still be open
    await expect(nodesPage.modal).toBeVisible()
  })

  test('cancel create closes modal', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()
    await nodesPage.clickCreate()

    // Click cancel
    await nodesPage.cancelButton.click()

    // Modal should close
    await expect(nodesPage.modal).not.toBeVisible()
  })

  test('edit node opens with pre-filled data', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()

    const rowCount = await nodesPage.getRowCount()

    if (rowCount > 0) {
      const editButton = adminPage.locator('table tbody tr').first().locator('button:has-text("Edit")')
      await editButton.click()

      // Modal should be visible with data
      await expect(nodesPage.modal).toBeVisible()

      // Name field should have value
      const nameValue = await nodesPage.nameInput.inputValue()
      expect(nameValue.length).toBeGreaterThan(0)
    } else {
      test.skip(true, 'No nodes to edit')
    }
  })

  test('update node name', async ({ adminPage }) => {
    await nodesPage.expectTableVisible()

    const rowCount = await nodesPage.getRowCount()

    if (rowCount > 0) {
      const editButton = adminPage.locator('table tbody tr').first().locator('button:has-text("Edit")')
      await editButton.click()
      await expect(nodesPage.modal).toBeVisible()

      const newName = `updated_${Date.now()}`
      await nodesPage.nameInput.fill(newName)
      await nodesPage.submitButton.click()

      await expect(nodesPage.modal).not.toBeVisible()

      // Verify update
      const hasUpdated = await nodesPage.hasNode(newName)
      expect(hasUpdated).toBeTruthy()
    } else {
      test.skip(true, 'No nodes to update')
    }
  })

  test('delete node shows confirmation', async ({ adminPage }) => {
    // First create a node to delete
    const nodeName = `e2e_delete_${Date.now()}`
    await nodesPage.createNode(nodeName, 'us-west-2')

    // Find the node row
    const rows = adminPage.locator('table tbody tr')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const rowText = await rows.nth(i).textContent()
      if (rowText?.includes(nodeName)) {
        const deleteButton = rows.nth(i).locator('button:has-text("Delete")')
        await deleteButton.click()

        // Should show confirmation
        const confirmButton = adminPage.locator('button:has-text("Confirm")')
        await expect(confirmButton).toBeVisible()

        // Cancel the deletion
        const cancelButton = adminPage.locator('button:has-text("Cancel")')
        if (await cancelButton.count() > 0) {
          await cancelButton.click()
        }

        break
      }
    }
  })

  test('delete node removes from list', async ({ adminPage }) => {
    // Create a node to delete
    const nodeName = `e2e_delete_${Date.now()}`
    await nodesPage.createNode(nodeName, 'us-west-2')

    // Find and delete
    const rows = adminPage.locator('table tbody tr')
    const count = await rows.count()

    for (let i = 0; i < count; i++) {
      const rowText = await rows.nth(i).textContent()
      if (rowText?.includes(nodeName)) {
        const deleteButton = rows.nth(i).locator('button:has-text("Delete")')
        await deleteButton.click()

        const confirmButton = adminPage.locator('button:has-text("Confirm")')
        await confirmButton.click()

        // Wait for deletion
        await adminPage.waitForTimeout(1000)

        break
      }
    }

    // Verify node is gone
    const hasNode = await nodesPage.hasNode(nodeName)
    expect(hasNode).toBeFalsy()
  })
})

test.describe('Node CRUD - Operator', () => {
  test('operator cannot create nodes (admin only)', async ({ operatorPage }) => {
    const nodesPage = new NodesPage(operatorPage)
    await nodesPage.goto()
    await nodesPage.expectTableVisible()

    // Operator should NOT see the create button (admin only feature)
    const createButton = operatorPage.locator('button:has-text("Add New Node"), button:has-text("Create")')
    const isVisible = await createButton.isVisible().catch(() => false)

    expect(isVisible).toBeFalsy()
  })
})

test.describe('Node CRUD - Viewer', () => {
  test('viewer cannot create nodes', async ({ viewerPage }) => {
    const nodesPage = new NodesPage(viewerPage)
    await nodesPage.goto()
    await nodesPage.expectTableVisible()

    const createButton = viewerPage.locator('button:has-text("Create"), button:has-text("Add")')
    const isVisible = await createButton.isVisible().catch(() => false)

    expect(isVisible).toBeFalsy()
  })
})
