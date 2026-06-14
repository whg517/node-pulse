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
      pending: 'bg-warning-bg text-warning-text',
      processing: 'bg-primary/10 text-primary',
      completed: 'bg-healthy-bg text-healthy-text',
      failed: 'bg-destructive/10 text-destructive',
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
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <span className="ml-2 text-muted-foreground">Loading export history...</span>
      </div>
    )
  }

  // Empty state
  if (exports.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <svg
          className="mx-auto h-12 w-12 text-muted-foreground"
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
        <h3 className="text-lg font-semibold text-foreground">Export History</h3>
        <div className="flex space-x-2">
          <button
            type="button"
            onClick={() => handleFilterChange('all')}
            className={`px-3 py-1 text-sm rounded-md transition-colors ${
              statusFilter === 'all'
                ? 'bg-primary text-white'
                : 'bg-muted text-foreground/80 hover:bg-muted-foreground/30'
            }`}
          >
            All ({exports.length})
          </button>
          <button
            type="button"
            onClick={() => handleFilterChange('completed')}
            className={`px-3 py-1 text-sm rounded-md transition-colors ${
              statusFilter === 'completed'
                ? 'bg-primary text-white'
                : 'bg-muted text-foreground/80 hover:bg-muted-foreground/30'
            }`}
          >
            Completed
          </button>
          <button
            type="button"
            onClick={() => handleFilterChange('failed')}
            className={`px-3 py-1 text-sm rounded-md transition-colors ${
              statusFilter === 'failed'
                ? 'bg-primary text-white'
                : 'bg-muted text-foreground/80 hover:bg-muted-foreground/30'
            }`}
          >
            Failed
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto border border-border rounded-lg">
        <table className="min-w-full divide-y divide-border">
          <thead className="bg-muted/50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Date
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Nodes
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Metrics
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Status
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Size
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-background divide-y divide-border">
            {paginatedExports.map((exp) => (
              <tr key={exp.id} className="hover:bg-muted/50">
                <td className="px-4 py-3 text-sm text-foreground">
                  {formatDate(exp.created_at)}
                </td>
                <td className="px-4 py-3 text-sm text-muted-foreground">
                  {exp.node_ids.length} node{exp.node_ids.length !== 1 ? 's' : ''}
                </td>
                <td className="px-4 py-3 text-sm text-muted-foreground">
                  <div className="flex flex-wrap gap-1">
                    {exp.metrics.map((metric) => (
                      <span
                        key={metric}
                        className="px-2 py-0.5 text-xs bg-muted rounded"
                      >
                        {metric === 'latency' && 'Latency'}
                        {metric === 'packet_loss_rate' && 'Packet Loss'}
                        {metric === 'jitter' && 'Jitter'}
                      </span>
                    ))}
                  </div>
                </td>
                <td className="px-4 py-3 text-sm">{getStatusBadge(exp.status)}</td>
                <td className="px-4 py-3 text-sm text-muted-foreground">
                  {formatFileSize(exp.file_size)}
                </td>
                <td className="px-4 py-3 text-sm text-right space-x-2">
                  {exp.status === 'completed' && (
                    <button
                      type="button"
                      onClick={() => onDownload(exp.id)}
                      className="text-primary hover:text-primary font-medium"
                      title="Download"
                    >
                      Download
                    </button>
                  )}
                  <button
                    type="button"
                    onClick={() => onDelete(exp.id)}
                    className="text-destructive hover:text-destructive font-medium"
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
          className="flex items-center justify-between border-t border-border pt-4"
          aria-label="Pagination"
        >
          <div className="text-sm text-foreground/80">
            Showing {startIndex + 1} to {Math.min(startIndex + itemsPerPage, filteredExports.length)} of{' '}
            {filteredExports.length} results
          </div>
          <div className="flex space-x-2">
            <button
              type="button"
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              disabled={currentPage === 1}
              className="px-3 py-1 text-sm border border-border rounded-md hover:bg-muted/50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <span className="px-3 py-1 text-sm text-foreground/80">
              Page {currentPage} of {totalPages}
            </span>
            <button
              type="button"
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
              className="px-3 py-1 text-sm border border-border rounded-md hover:bg-muted/50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </nav>
      )}
    </div>
  )
}
