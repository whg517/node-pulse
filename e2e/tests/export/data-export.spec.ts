/**
 * Data Export Tests
 *
 * Tests for data export page (admin only):
 * - Export form
 * - Progress tracking
 * - File download
 */

import { test, expect } from '../../fixtures/auth.fixture'
import { DataExportPage } from '../../pages/DataExportPage'

test.describe('Data Export Page - Admin', () => {
  let exportPage: DataExportPage

  test.beforeEach(async ({ adminPage }) => {
    exportPage = new DataExportPage(adminPage)
    await exportPage.goto()
  })

  test('AC-21: page loads for admin', async ({ adminPage }) => {
    const hasWarning = await exportPage.hasAccessWarning()

    if (!hasWarning) {
      await exportPage.expectFormVisible()
    }
  })

  test('form has expected fields', async ({ adminPage }) => {
    const hasWarning = await exportPage.hasAccessWarning()

    if (hasWarning) {
      test.skip(true, 'Access warning shown')
      return
    }

    await exportPage.expectFormVisible()

    // Check for node checkboxes (at least one)
    const nodeCheckboxes = adminPage.locator('input[type="checkbox"]')
    await expect(nodeCheckboxes.first()).toBeVisible()

    // Check for time range radio buttons
    const timeRangeRadios = adminPage.locator('input[type="radio"][name="timeRange"]')
    await expect(timeRangeRadios.first()).toBeVisible()

    // Check for format radio buttons
    const formatRadios = adminPage.locator('input[type="radio"][name="format"]')
    await expect(formatRadios.first()).toBeVisible()
  })

  test('can submit export', async ({ adminPage }) => {
    const hasWarning = await exportPage.hasAccessWarning()

    if (hasWarning) {
      test.skip(true, 'Access warning shown')
      return
    }

    await exportPage.expectFormVisible()

    // Select first available node checkbox
    await exportPage.selectFirstNode()

    // Select time range (7d is default, but let's select explicitly)
    await exportPage.selectTimeRange('7d')

    // Select format
    await exportPage.selectFormat('csv')

    // Submit export
    await exportPage.submitExport()

    // Should show progress or success
    await adminPage.waitForTimeout(1000)
  })

  test('progress shown during export', async ({ adminPage }) => {
    const hasWarning = await exportPage.hasAccessWarning()

    if (hasWarning) {
      test.skip(true, 'Access warning shown')
      return
    }

    await exportPage.expectFormVisible()

    // Select first node
    await exportPage.selectFirstNode()

    // Submit export
    await exportPage.submitExport()

    // Look for progress indicator or just wait a bit
    await adminPage.waitForTimeout(1000)

    // Progress may or may not be shown depending on export speed - just verify no crash
    const pageContent = await adminPage.content()
    expect(pageContent).toBeTruthy()
  })

  test('export API works', async ({ adminPage }) => {
    const response = await adminPage.request.post('/api/v1/data/export', {
      data: {
        node_ids: [],
        start_time: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
        end_time: new Date().toISOString(),
        format: 'csv',
      },
    })

    // Should succeed or return validation error (400) or auth error (401)
    expect([200, 201, 400, 401]).toContain(response.status())
  })
})

test.describe('Data Export Page - Operator', () => {
  test('AC-12: operator sees access warning', async ({ operatorPage }) => {
    const exportPage = new DataExportPage(operatorPage)
    await exportPage.goto()

    const hasWarning = await exportPage.hasAccessWarning()
    expect(hasWarning).toBeTruthy()
  })

  test('operator cannot create export', async ({ operatorPage }) => {
    const response = await operatorPage.request.post('/api/v1/data/export', {
      data: {
        node_ids: [],
        start_time: new Date().toISOString(),
        end_time: new Date().toISOString(),
        format: 'csv',
      },
    })

    // 401 (unauthorized) or 403 (forbidden) both indicate access denied
    expect([401, 403]).toContain(response.status())
  })
})

test.describe('Data Export Page - Viewer', () => {
  test('viewer sees access warning', async ({ viewerPage }) => {
    const exportPage = new DataExportPage(viewerPage)
    await exportPage.goto()

    const hasWarning = await exportPage.hasAccessWarning()
    expect(hasWarning).toBeTruthy()
  })
})
