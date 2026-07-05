/**
 * Export API endpoints
 *
 * Provides typed functions for creating export tasks,
 * checking export status, and downloading exported files.
 */

import { apiClient } from './client'
import { API_BASE_URL } from '../config/constants'
import type {
  ExportTask,
  CreateExportRequest,
  CreateExportResponse,
  GetExportStatusResponse,
} from '../types/export'

// Export types for use in components and stores
export type {
  ExportTask,
  CreateExportRequest,
  CreateExportResponse,
  GetExportStatusResponse,
}
export type { ExportStatus, ExportMetric, ExportFormat } from '../types/export'

/**
 * Create a new export task
 *
 * Initiates an async export job for the specified data range.
 * Returns the created export task with status "pending".
 *
 * @param request - Export request parameters
 * @returns Created export task with ID
 * @throws ValidationError if export parameters are invalid
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const { data } = await createExport({
 *   node_ids: ['node-1', 'node-2'],
 *   start_time: '2024-01-01T00:00:00Z',
 *   end_time: '2024-01-07T23:59:59Z',
 *   metrics: ['latency', 'packet_loss_rate'],
 *   format: 'csv'
 * })
 *
 * console.log('Export task ID:', data.id)
 * console.log('Status:', data.status)
 */
export async function createExport(
  request: CreateExportRequest
): Promise<CreateExportResponse> {
  const params = new URLSearchParams()

  // Add node IDs
  request.node_ids.forEach((id) => params.append('node_ids', id))

  // Add time range
  params.append('start_time', request.start_time)
  params.append('end_time', request.end_time)

  // Add metrics
  request.metrics.forEach((m) => params.append('metrics', m))

  // Add format (defaults to csv)
  params.append('format', request.format || 'csv')

  return apiClient<CreateExportResponse>(`/api/v1/data/export?${params}`, {
    method: 'POST',
  })
}

/**
 * Get export task status
 *
 * Retrieves the current status of an export task.
 * Use this to poll for task completion.
 *
 * @param exportId - Export task ID
 * @returns Current export task status
 * @throws NotFoundError if export task doesn't exist
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const { data } = await getExportStatus('task-uuid')
 *
 * if (data.status === 'completed') {
 *   console.log('Export completed!')
 *   console.log('File:', data.file_path)
 *   console.log('Records:', data.record_count)
 * } else if (data.status === 'failed') {
 *   console.error('Export failed:', data.error)
 * }
 */
export async function getExportStatus(
  exportId: string
): Promise<GetExportStatusResponse> {
  return apiClient<GetExportStatusResponse>(`/api/v1/data/export/${exportId}`)
}

/**
 * List recent export tasks for the current admin user.
 *
 * @param limit - Maximum number of tasks to return
 * @returns Recent export tasks, newest first
 */
export async function listExports(
  limit = 50
): Promise<{ data: ExportTask[]; message: string; timestamp: string }> {
  const params = new URLSearchParams({ limit: String(limit) })
  return apiClient(`/api/v1/data/export?${params}`)
}

/**
 * Download exported file
 *
 * Downloads the exported CSV file as a blob.
 * Use this when export status is "completed".
 *
 * @param exportId - Export task ID
 * @returns File blob for download
 * @throws NotFoundError if export file doesn't exist
 * @throws AuthenticationError if user is not authenticated
 *
 * @example
 * const blob = await downloadExport('task-uuid')
 *
 * // Create download link
 * const url = window.URL.createObjectURL(blob)
 * const a = document.createElement('a')
 * a.href = url
 * a.download = 'metrics_export.csv'
 * document.body.appendChild(a)
 * a.click()
 * window.URL.revokeObjectURL(url)
 * document.body.removeChild(a)
 */
export async function downloadExport(exportId: string): Promise<Blob> {
  const url = `${API_BASE_URL}/api/v1/data/export/${exportId}/download`

  const response = await fetch(url, {
    method: 'GET',
    credentials: 'include', // Send Session Cookie for authentication
    headers: {
      'Content-Type': 'text/csv; charset=utf-8',
    },
  })

  if (!response.ok) {
    // Handle error responses
    let errorMessage = 'Failed to download export'

    try {
      const errorData = await response.json()
      errorMessage =
        typeof errorData === 'object' && 'message' in errorData
          ? String(errorData.message)
          : errorMessage
    } catch {
      // If response is not JSON, use status text
      errorMessage = response.statusText || errorMessage
    }

    throw new Error(errorMessage)
  }

  return response.blob()
}

/**
 * Delete an export task and its generated file.
 *
 * Removes both the export record and the file on disk. Admin only.
 * Returns void on success (HTTP 204).
 *
 * @param exportId - Export task ID
 * @throws AuthorizationError if user is not admin
 * @throws NotFoundError if export task doesn't exist
 */
export async function deleteExport(exportId: string): Promise<void> {
  await apiClient<void>(`/api/v1/data/export/${exportId}`, {
    method: 'DELETE',
  })
}
