/**
 * ExportStatusCard Component
 *
 * Displays the status of an export task with progress indicators,
 * task details, and download action when completed.
 */

import { useMemo } from 'react'
import type { ExportTask } from '../../types/export'

interface ExportStatusCardProps {
  exportTask: ExportTask
  onDownload?: (exportId: string) => void
  onDismiss?: () => void
}

export function ExportStatusCard({ exportTask, onDownload, onDismiss }: ExportStatusCardProps) {
  /**
   * Format file size for display
   */
  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
  }

  /**
   * Format date for display
   */
  const formatDate = (dateString: string): string => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })
  }

  /**
   * Calculate duration between two dates
   */
  const calculateDuration = (start: string, end: string): string => {
    const startDate = new Date(start)
    const endDate = new Date(end)
    const diffMs = endDate.getTime() - startDate.getTime()
    const diffMins = Math.round(diffMs / 60000)

    if (diffMins < 60) {
      return `${diffMins} min`
    }
    const hours = Math.floor(diffMins / 60)
    const mins = diffMins % 60
    return `${hours}h ${mins}m`
  }

  /**
   * Calculate estimated completion time
   */
  const estimatedCompletion = useMemo(() => {
    if (exportTask.status === 'completed' || exportTask.status === 'failed') {
      return null
    }

    const created = new Date(exportTask.created_at)
    const now = new Date()
    const elapsedMins = Math.round((now.getTime() - created.getTime()) / 60000)

    // Estimate based on elapsed time (assume 5-10 min total)
    const avgExportTime = 5 // 5 minutes average
    const remainingMins = Math.max(0, avgExportTime - elapsedMins)

    if (remainingMins === 0) {
      return 'Completing soon...'
    }

    return `~${remainingMins} min remaining`
  }, [exportTask.status, exportTask.created_at])

  /**
   * Get status badge styling
   */
  const getStatusBadge = () => {
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
        className={`px-2 py-1 text-xs font-semibold rounded-full ${statusStyles[exportTask.status]}`}
      >
        {statusLabels[exportTask.status]}
      </span>
    )
  }

  return (
    <div className="bg-background rounded-lg shadow-md p-6 border border-border">
      {/* Header with status and dismiss */}
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-center space-x-2">
          <h3 className="text-lg font-semibold text-foreground">Export Task</h3>
          {getStatusBadge()}
        </div>
        {onDismiss && (
          <button
            type="button"
            onClick={onDismiss}
            className="text-muted-foreground hover:text-muted-foreground transition-colors"
            aria-label="Dismiss"
          >
            <svg
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        )}
      </div>

      {/* Progress indicator for pending/processing */}
      {(exportTask.status === 'pending' || exportTask.status === 'processing') && (
        <div className="mb-4">
          <div className="flex items-center space-x-2">
            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-primary"></div>
            <span className="text-sm text-muted-foreground">
              {exportTask.status === 'pending' ? 'Initializing export...' : 'Exporting data...'}
            </span>
          </div>
          {estimatedCompletion && (
            <p className="text-xs text-muted-foreground mt-1 ml-7">{estimatedCompletion}</p>
          )}
        </div>
      )}

      {/* Error message for failed exports */}
      {exportTask.status === 'failed' && exportTask.error && (
        <div className="mb-4 bg-destructive/10 border-l-4 border-destructive p-3 rounded">
          <p className="text-sm text-destructive">{exportTask.error}</p>
        </div>
      )}

      {/* Export Details */}
      <div className="space-y-2 text-sm text-foreground/80">
        <div className="flex items-center justify-between">
          <span className="font-medium">Nodes:</span>
          <span>{exportTask.node_ids.length} selected</span>
        </div>

        <div className="flex items-center justify-between">
          <span className="font-medium">Time Range:</span>
          <span>
            {formatDate(exportTask.start_time)} - {formatDate(exportTask.end_time)}
          </span>
        </div>

        <div className="flex items-center justify-between">
          <span className="font-medium">Metrics:</span>
          <span className="flex flex-wrap gap-1 justify-end">
            {exportTask.metrics.map((metric) => (
              <span
                key={metric}
                className="px-2 py-0.5 text-xs bg-muted rounded"
              >
                {metric === 'latency' && 'Latency'}
                {metric === 'packet_loss_rate' && 'Packet Loss'}
                {metric === 'jitter' && 'Jitter'}
              </span>
            ))}
          </span>
        </div>

        <div className="flex items-center justify-between">
          <span className="font-medium">Format:</span>
          <span className="uppercase">{exportTask.format}</span>
        </div>

        {/* Completion info when completed */}
        {exportTask.status === 'completed' && (
          <>
            {exportTask.record_count !== undefined && (
              <div className="flex items-center justify-between">
                <span className="font-medium">Records:</span>
                <span>{exportTask.record_count.toLocaleString()}</span>
              </div>
            )}
            {exportTask.file_size !== undefined && (
              <div className="flex items-center justify-between">
                <span className="font-medium">File Size:</span>
                <span>{formatFileSize(exportTask.file_size)}</span>
              </div>
            )}
            {exportTask.completed_at && (
              <div className="flex items-center justify-between">
                <span className="font-medium">Duration:</span>
                <span>{calculateDuration(exportTask.created_at, exportTask.completed_at)}</span>
              </div>
            )}
          </>
        )}
      </div>

      {/* Download button when completed */}
      {exportTask.status === 'completed' && onDownload && (
        <div className="mt-4 pt-4 border-t border-border">
          <button
            type="button"
            onClick={() => onDownload(exportTask.id)}
            className="w-full bg-primary hover:bg-primary/85 text-white font-medium py-2 px-4 rounded-md transition-colors duration-150 flex items-center justify-center space-x-2"
          >
            <svg
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
              />
            </svg>
            <span>Download CSV</span>
          </button>
        </div>
      )}
    </div>
  )
}
