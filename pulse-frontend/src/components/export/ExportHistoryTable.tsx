/**
 * ExportHistoryTable Component
 *
 * Displays export history with pagination, filtering,
 * and action buttons for download and delete.
 */

import { useState } from 'react'
import type { ExportTask, ExportStatus } from '../../types/export'

interface ExportHistoryTableProps {
  exports: ExportTask[]
  onDownload: (exportId: string) => void
  onDelete: (exportId: string) => void
  loading?: boolean
}

type FilterType = 'all' | ExportStatus

export function ExportHistoryTable({
  exports,
  onDownload,
  onDelete,
  loading = false,
}: ExportHistoryTableProps) {
  // Pagination state
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 20

  // Filter state
  const [statusFilter, setStatusFilter] = useState<FilterType>('all')

  /**
   * Filter exports by status
   */
  const filteredExports = exports.filter((exp) => {
    if (statusFilter === 'all') return true
    return exp.status === statusFilter
  })

  /**
   * Paginate filtered exports
   */
  const totalPages = Math.ceil(filteredExports.length / itemsPerPage)
  const startIndex = (currentPage - 1) * itemsPerPage
  const paginatedExports = filteredExports.slice(startIndex, startIndex + itemsPerPage)

  /**
   * Reset to page 1 when filter changes
   */
  const handleFilterChange = (filter: FilterType) => {
    setStatusFilter(filter)
    setCurrentPage(1)
  }

  /**
   * Format date for display
   */
  const formatDate = (dateString: string): string => {
    const date = new Date(dateString)
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  /**
   * Format file size for display
   */
  const formatFileSize = (bytes?: number): string => {
    if (!bytes) return '-'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  }

  /**
   * Get status badge styling
   */
  const getStatusBadge = (status: ExportStatus) => {
    const statusStyles = {
      pending: 'bg-yellow-100 text-yellow-800',
      processing: 'bg-blue-100 text-blue-800',
      completed: 'bg-green-100 text-green-800',
      failed: 'bg-red-100 text-red-800',
    }

    const statusLabels = {
      pending: 'Pending',
      processing: 'Processing',
      completed: 'Completed',
      failed: 'Failed',
    }

    return (
      <span
        className={`px-2 py-1 text-xs font-semibold rounded-full ${statusStyles[status]}`}
      >
        {statusLabels[status]}
      </span>
    )
  }

  // Loading state
  if (loading) {
    return (
      <div className="flex justify-center items-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <span className="ml-2 text-gray-600">Loading export history...</span>
      </div>
    )
  }

  // Empty state
  if (exports.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        <svg
          className="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
        <p className="mt-2">No export history</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Header with filter */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-gray-900">Export History</h3>
        <div className="flex space-x-2">
          <button
            type="button"
            onClick={() => handleFilterChange('all')}
            className={`px-3 py-1 text-sm rounded-md transition-colors ${
              statusFilter === 'all'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            All ({exports.length})
          </button>
          <button
            type="button"
            onClick={() => handleFilterChange('completed')}
            className={`px-3 py-1 text-sm rounded-md transition-colors ${
              statusFilter === 'completed'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            Completed
          </button>
          <button
            type="button"
            onClick={() => handleFilterChange('failed')}
            className={`px-3 py-1 text-sm rounded-md transition-colors ${
              statusFilter === 'failed'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            Failed
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto border border-gray-200 rounded-lg">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Date
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Nodes
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Metrics
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Size
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {paginatedExports.map((exp) => (
              <tr key={exp.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 text-sm text-gray-900">
                  {formatDate(exp.created_at)}
                </td>
                <td className="px-4 py-3 text-sm text-gray-600">
                  {exp.node_ids.length} node{exp.node_ids.length !== 1 ? 's' : ''}
                </td>
                <td className="px-4 py-3 text-sm text-gray-600">
                  <div className="flex flex-wrap gap-1">
                    {exp.metrics.map((metric) => (
                      <span
                        key={metric}
                        className="px-2 py-0.5 text-xs bg-gray-100 rounded"
                      >
                        {metric === 'latency' && 'Latency'}
                        {metric === 'packet_loss_rate' && 'Packet Loss'}
                        {metric === 'jitter' && 'Jitter'}
                      </span>
                    ))}
                  </div>
                </td>
                <td className="px-4 py-3 text-sm">{getStatusBadge(exp.status)}</td>
                <td className="px-4 py-3 text-sm text-gray-600">
                  {formatFileSize(exp.file_size)}
                </td>
                <td className="px-4 py-3 text-sm text-right space-x-2">
                  {exp.status === 'completed' && (
                    <button
                      type="button"
                      onClick={() => onDownload(exp.id)}
                      className="text-blue-600 hover:text-blue-800 font-medium"
                      title="Download"
                    >
                      Download
                    </button>
                  )}
                  <button
                    type="button"
                    onClick={() => onDelete(exp.id)}
                    className="text-red-600 hover:text-red-800 font-medium"
                    title="Delete"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <nav
          className="flex items-center justify-between border-t border-gray-200 pt-4"
          aria-label="Pagination"
        >
          <div className="text-sm text-gray-700">
            Showing {startIndex + 1} to {Math.min(startIndex + itemsPerPage, filteredExports.length)} of{' '}
            {filteredExports.length} results
          </div>
          <div className="flex space-x-2">
            <button
              type="button"
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              disabled={currentPage === 1}
              className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <span className="px-3 py-1 text-sm text-gray-700">
              Page {currentPage} of {totalPages}
            </span>
            <button
              type="button"
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
              className="px-3 py-1 text-sm border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </nav>
      )}
    </div>
  )
}
