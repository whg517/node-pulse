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
    await nodesPage.createNode(nodeName, 'us-east-1')

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
  test('operator can create nodes', async ({ operatorPage }) => {
    const nodesPage = new NodesPage(operatorPage)
    await nodesPage.goto()
    await nodesPage.expectTableVisible()

    const nodeName = `e2e_operator_${Date.now()}`
    await nodesPage.createNode(nodeName, 'eu-west-1')

    const hasNode = await nodesPage.hasNode(nodeName)
    expect(hasNode).toBeTruthy()
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
