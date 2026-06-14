/**
 * AlertDetailMobile Component
 * 
 * Mobile-responsive alert detail view with status updates and note-taking capability.
 * Optimized for small screens with touch-friendly interactions.
 * 
 * Features:
 * - Full-screen mobile modal design
 * - Status update actions (acknowledge, resolve)
 * - Note-taking with timestamps
 * - UTC time display with timezone conversion
 * - Cross-team collaboration timeline
 * - Accessibility compliant (WCAG 2.1 AA)
 * 
 * @packageDocumentation
 */

import { useState, useCallback } from 'react'
import type { AlertRecordDTO, AlertRecordStatus } from '../../api/alertRecords'
import type { NodeDTO } from '../../api/types'
import { useTimezoneUtils } from '../../utils/timezone'

// ============== Types ==============

export interface AlertNote {
  id: string
  alertId: string
  userId: string
  userName: string
  content: string
  createdAt: string
}

export interface AlertDetailMobileProps {
  record: AlertRecordDTO
  nodes: NodeDTO[]
  canEdit: boolean
  notes?: AlertNote[]
  onClose: () => void
  onStatusUpdate: (id: string, status: AlertRecordStatus, note?: string) => Promise<void>
}

// ============== Helper Functions ==============

/**
 * Get status display name with i18n support
 */
function getStatusDisplayName(status: string): string {
  const displayNames: Record<string, string> = {
    pending: 'Pending',
    in_progress: 'In Progress',
    resolved: 'Resolved',
  }
  return displayNames[status] || status
}

/**
 * Get status badge color classes (Tailwind dark: prefix, no JS branching)
 */
function getStatusBadgeColor(status: string): string {
  const colors: Record<string, string> = {
    pending: 'bg-destructive/10 text-destructive border-destructive/10',
    in_progress: 'bg-warning-bg text-warning-text border-warning-bg',
    resolved: 'bg-healthy-bg text-healthy-text border-healthy-bg',
  }
  return colors[status] || colors.pending
}

/**
 * Get level badge color classes (Tailwind dark: prefix, no JS branching)
 */
function getLevelBadgeColor(level: string): string {
  const colors: Record<string, string> = {
    P0: 'bg-destructive/10 text-destructive border-destructive/10',
    P1: 'bg-warning-bg text-warning-text border-warning-bg',
    P2: 'bg-warning-bg text-warning-text border-warning-bg',
  }
  return colors[level] || colors.P2
}

/**
 * Get metric display name
 */
function getMetricDisplayName(metric: string): string {
  const displayNames: Record<string, string> = {
    latency: 'Latency',
    packet_loss_rate: 'Packet Loss Rate',
    jitter: 'Jitter',
  }
  return displayNames[metric] || metric
}

// ============== Components ==============

/**
 * Status Badge Component
 */
function StatusBadge({ status }: { status: string }) {
  return (
    <span className={`px-2.5 py-1 text-xs font-semibold rounded-full border ${getStatusBadgeColor(status)}`}>
      {getStatusDisplayName(status)}
    </span>
  )
}

/**
 * Level Badge Component
 */
function LevelBadge({ level }: { level: string }) {
  return (
    <span className={`px-2.5 py-1 text-xs font-semibold rounded-full border ${getLevelBadgeColor(level)}`}>
      {level}
    </span>
  )
}

/**
 * Timeline Event Component for collaboration
 */
function TimelineEvent({
  title,
  time,
  description,
  isLast,
}: {
  title: string
  time: string
  description?: string
  isLast: boolean
}) {
  return (
    <div className="flex gap-3">
      {/* Timeline line */}
      <div className="flex flex-col items-center">
        <div className="w-2 h-2 rounded-full bg-primary" />
        {!isLast && (
          <div className="w-0.5 flex-1 min-h-[24px] bg-muted dark:bg-accent" />
        )}
      </div>
      
      {/* Content */}
      <div className="flex-1 pb-4">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-sm font-medium text-foreground">
            {title}
          </span>
          <span className="text-xs text-muted-foreground">
            {time}
          </span>
        </div>
        {description && (
          <p className="mt-1 text-sm text-muted-foreground">
            {description}
          </p>
        )}
      </div>
    </div>
  )
}

/**
 * Note Input Component
 */
function NoteInput({
  onSubmit,
  isSubmitting,
  placeholder = "Add a note...",
}: {
  onSubmit: (note: string) => void
  isSubmitting: boolean
  placeholder?: string
}) {
  const [note, setNote] = useState('')
  
  const handleSubmit = useCallback(() => {
    if (note.trim() && !isSubmitting) {
      onSubmit(note.trim())
      setNote('')
    }
  }, [note, isSubmitting, onSubmit])
  
  const handleKeyDown = useCallback((e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleSubmit()
    }
  }, [handleSubmit])
  
  return (
    <div className="mt-4">
      <label className="block text-sm font-medium mb-2 text-muted-foreground">
        Add Note
      </label>
      <textarea
        value={note}
        onChange={(e) => setNote(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        rows={3}
        className="w-full px-3 py-2 rounded-lg border resize-none focus:outline-none focus:ring-2 focus:ring-primary bg-background border-input text-foreground placeholder:text-muted-foreground"
        disabled={isSubmitting}
        aria-label="Add a note (press Ctrl+Enter or Cmd+Enter to submit)"
      />
      <div className="mt-2 flex justify-between items-center">
        <p className="text-xs text-muted-foreground">
          Press Ctrl+Enter or Cmd+Enter to submit
        </p>
        <button
          onClick={handleSubmit}
          disabled={!note.trim() || isSubmitting}
          className="px-4 py-2 bg-primary text-white text-sm font-medium rounded-lg hover:bg-primary/85 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? 'Adding...' : 'Add Note'}
        </button>
      </div>
    </div>
  )
}

// ============== Main Component ==============

/**
 * AlertDetailMobile - Mobile-responsive alert detail view
 */
export function AlertDetailMobile({
  record,
  nodes,
  canEdit,
  notes = [],
  onClose,
  onStatusUpdate,
}: AlertDetailMobileProps) {
  const timezoneUtils = useTimezoneUtils()
  const [isUpdating, setIsUpdating] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [showNoteInput, setShowNoteInput] = useState(false)
  
  // Get node information
  const node = nodes.find((n) => n.id === record.node_id)
  
  // Format timestamps
  const createdAt = timezoneUtils.getFormattedTime(record.created_at)
  const updatedAt = record.updated_at !== record.created_at
    ? timezoneUtils.getFormattedTime(record.updated_at)
    : null
  
  // Build timeline events
  const timelineEvents = [
    {
      title: 'Alert Created',
      time: createdAt.relative,
      description: `Alert triggered at ${createdAt.utc}`,
      timestamp: record.created_at,
    },
    ...(updatedAt
      ? [
          {
            title: 'Alert Updated',
            time: updatedAt.relative,
            description: `Last updated at ${updatedAt.utc}`,
            timestamp: record.updated_at,
          },
        ]
      : []),
    ...notes.map((note) => ({
      title: `Note by ${note.userName}`,
      time: timezoneUtils.formatRelative(note.createdAt),
      description: note.content,
      timestamp: note.createdAt,
    })),
  ].sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
  
  // Handle status update
  const handleStatusUpdate = useCallback(async (newStatus: AlertRecordStatus, noteContent?: string) => {
    setIsUpdating(true)
    setError(null)
    try {
      await onStatusUpdate(record.id, newStatus, noteContent)
      if (!noteContent) {
        onClose()
      } else {
        setShowNoteInput(false)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update status')
    } finally {
      setIsUpdating(false)
    }
  }, [record.id, onStatusUpdate, onClose])
  
  // Handle view node details
  const handleViewNodeDetails = useCallback(() => {
    onClose()
    window.location.href = `/nodes/${record.node_id}`
  }, [record.node_id, onClose])
  
  return (
    <div
      className="fixed inset-0 z-50 flex flex-col"
      role="dialog"
      aria-modal="true"
      aria-labelledby="alert-detail-title"
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={onClose}
        aria-hidden="true"
      />
      
      {/* Modal Content */}
      <div className="relative flex flex-col h-full w-full bg-card">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-border bg-card">
          <h2
            id="alert-detail-title"
            className="text-lg font-semibold text-foreground"
          >
            Alert Details
          </h2>
          <button
            onClick={onClose}
            className="p-2 rounded-lg transition-colors text-muted-foreground hover:bg-accent/10"
            aria-label="Close alert details"
          >
            <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto">
          {/* Error Message */}
          {error && (
            <div className="mx-4 mt-4 bg-destructive/10 border-l-4 border-destructive p-4 rounded-md" role="alert">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}
          
          <div className="p-4 space-y-6">
            {/* Alert ID */}
            <section aria-labelledby="alert-id-label">
              <h3 id="alert-id-label" className="text-xs font-medium uppercase tracking-wider mb-1 text-muted-foreground">
                Alert ID
              </h3>
              <p className="text-sm font-mono break-all text-muted-foreground">
                {record.id}
              </p>
            </section>
            
            {/* Status and Level */}
            <section aria-labelledby="status-level-label" className="flex gap-3">
              <div>
                <h3 id="status-level-label" className="text-xs font-medium uppercase tracking-wider mb-2 text-muted-foreground">
                  Status
                </h3>
                <StatusBadge status={record.status} />
              </div>
              <div>
                <h3 className="text-xs font-medium uppercase tracking-wider mb-2 text-muted-foreground">
                  Level
                </h3>
                <LevelBadge level={record.level} />
              </div>
            </section>
            
            {/* Node Information */}
            <section aria-labelledby="node-label">
              <h3 id="node-label" className="text-xs font-medium uppercase tracking-wider mb-2 text-muted-foreground">
                Node
              </h3>
              <div className="p-3 rounded-lg border border-border bg-muted">
                <p className="text-sm font-medium text-foreground">
                  {node?.name || record.node_id}
                </p>
                {node && (
                  <>
                    <p className="text-xs mt-1 text-muted-foreground">
                      IP: {node.ip}
                    </p>
                    <button
                      onClick={handleViewNodeDetails}
                      className="mt-2 text-sm text-primary hover:text-primary font-medium"
                    >
                      View Node Details →
                    </button>
                  </>
                )}
              </div>
            </section>
            
            {/* Metric Information */}
            <section aria-labelledby="metric-label">
              <h3 id="metric-label" className="text-xs font-medium uppercase tracking-wider mb-2 text-muted-foreground">
                Metric
              </h3>
              <div className="p-3 rounded-lg border border-border bg-muted">
                <p className="text-sm font-medium text-foreground">
                  {getMetricDisplayName(record.metric)}
                </p>
              </div>
            </section>
            
            {/* Timeline - Shared UTC View */}
            <section aria-labelledby="timeline-label">
              <h3 id="timeline-label" className="text-xs font-medium uppercase tracking-wider mb-3 text-muted-foreground">
                Timeline (UTC)
              </h3>
              <div className="p-4 rounded-lg border border-border bg-muted">
                <div className="space-y-0">
                  {timelineEvents.map((event, index) => (
                    <TimelineEvent
                      key={event.timestamp + index}
                      title={event.title}
                      time={event.time}
                      description={event.description}
                      isLast={index === timelineEvents.length - 1}
                    />
                  ))}
                </div>
              </div>
            </section>
            
            {/* Notes Section */}
            {notes.length > 0 && (
              <section aria-labelledby="notes-label">
                <h3 id="notes-label" className="text-xs font-medium uppercase tracking-wider mb-3 text-muted-foreground">
                  Notes ({notes.length})
                </h3>
                <div className="space-y-3">
                  {notes.map((note) => (
                    <div
                      key={note.id}
                      className="p-3 rounded-lg border border-border bg-card"
                    >
                      <div className="flex items-center justify-between mb-1">
                        <span className="text-sm font-medium text-foreground">
                          {note.userName}
                        </span>
                        <span className="text-xs text-muted-foreground">
                          {timezoneUtils.formatRelative(note.createdAt)}
                        </span>
                      </div>
                      <p className="text-sm whitespace-pre-wrap text-muted-foreground">
                        {note.content}
                      </p>
                    </div>
                  ))}
                </div>
              </section>
            )}
            
            {/* Status Update Actions */}
            {canEdit && record.status !== 'resolved' && (
              <section aria-labelledby="actions-label" className="pt-4 border-t border-border">
                <h3 id="actions-label" className="text-sm font-medium mb-3 text-foreground">
                  Update Status
                </h3>
                
                <div className="space-y-3">
                  {/* Acknowledge Button */}
                  {record.status === 'pending' && (
                    <button
                      onClick={() => handleStatusUpdate('in_progress')}
                      disabled={isUpdating}
                      className="w-full py-3 px-4 rounded-lg font-medium transition-colors bg-warning text-white hover:bg-warning-hover disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {isUpdating ? 'Updating...' : 'Acknowledge Alert'}
                    </button>
                  )}
                  
                  {/* Add Note and Update Status */}
                  <button
                    onClick={() => setShowNoteInput(!showNoteInput)}
                    disabled={isUpdating}
                    className="w-full py-3 px-4 rounded-lg font-medium transition-colors border border-foreground/20 bg-card text-muted-foreground hover:bg-accent/10 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {showNoteInput ? 'Cancel' : 'Add Note'}
                  </button>
                  
                  {showNoteInput && (
                    <NoteInput
                      onSubmit={(note) => {
                        if (record.status === 'pending') {
                          handleStatusUpdate('in_progress', note)
                        } else {
                          handleStatusUpdate(record.status, note)
                        }
                      }}
                      isSubmitting={isUpdating}
                      placeholder="Add a note about this alert..."
                    />
                  )}
                  
                  {/* Resolve Button */}
                  <button
                    onClick={() => handleStatusUpdate('resolved')}
                    disabled={isUpdating}
                    className="w-full py-3 px-4 rounded-lg font-medium transition-colors bg-healthy text-white hover:bg-healthy-hover disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isUpdating ? 'Updating...' : 'Resolve Alert'}
                  </button>
                </div>
              </section>
            )}
          </div>
        </div>
        
        {/* Bottom padding for safe area on mobile */}
        <div className="h-safe-area-inset-bottom" />
      </div>
    </div>
  )
}

export default AlertDetailMobile
