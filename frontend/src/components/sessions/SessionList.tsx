import { useState } from 'react'
import type { Session } from '../../types/auth'

interface SessionListProps {
  sessions: Session[]
  currentSessionId: string | null
  isLoading: boolean
  onRefresh: () => void
  onRevoke: (sessionId: string) => Promise<void>
}

/**
 * SessionList Component
 *
 * Displays a list of active sessions with:
 * - Current session badge
 * - Last used timestamp
 * - Revoke button with confirmation
 */
export default function SessionList({
  sessions,
  currentSessionId,
  isLoading,
  onRefresh,
  onRevoke,
}: SessionListProps) {
  const [revokingId, setRevokingId] = useState<string | null>(null)
  const [confirmRevoke, setConfirmRevoke] = useState<string | null>(null)

  const handleRevoke = async (sessionId: string, _isCurrent: boolean) => {
    if (confirmRevoke !== sessionId) {
      setConfirmRevoke(sessionId)
      return
    }

    setRevokingId(sessionId)
    try {
      await onRevoke(sessionId)
      setConfirmRevoke(null)
    } catch (error) {
      console.error('Failed to revoke session:', error)
    } finally {
      setRevokingId(null)
    }
  }

  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleString()
  }

  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins} minute${diffMins !== 1 ? 's' : ''} ago`
    if (diffHours < 24) return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`
    return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`
  }

  if (isLoading) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <div className="animate-pulse space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center space-x-4">
              <div className="h-4 bg-gray-200 rounded w-1/4"></div>
              <div className="h-4 bg-gray-200 rounded w-1/3"></div>
              <div className="h-4 bg-gray-200 rounded w-1/4"></div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
        <h3 className="text-lg font-medium text-gray-900">Active Sessions</h3>
        <button
          onClick={onRefresh}
          disabled={isLoading}
          className="inline-flex items-center px-3 py-1.5 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
        >
          <svg
            className={`-ml-1 mr-2 h-4 w-4 ${isLoading ? 'animate-spin' : ''}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
            />
          </svg>
          Refresh
        </button>
      </div>

      {sessions.length === 0 ? (
        <div className="px-6 py-8 text-center text-gray-500">
          No active sessions found.
        </div>
      ) : (
        <ul className="divide-y divide-gray-200">
          {sessions.map((session) => {
            const isCurrent = session.id === currentSessionId || session.is_current
            const isRevoking = revokingId === session.id
            const isConfirming = confirmRevoke === session.id

            return (
              <li key={session.id} className="px-6 py-4 hover:bg-gray-50">
                <div className="flex items-center justify-between">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {session.ip_address || 'Unknown IP'}
                      </p>
                      {isCurrent && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                          Current Session
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-gray-500 truncate">
                      {session.user_agent || 'Unknown device'}
                    </p>
                    <div className="mt-1 flex items-center space-x-4 text-xs text-gray-400">
                      <span>Created: {formatDateTime(session.created_at)}</span>
                      <span>Last used: {formatRelativeTime(session.last_used_at)}</span>
                    </div>
                  </div>

                  <div className="ml-4 flex items-center space-x-2">
                    {isConfirming ? (
                      <>
                        <span className="text-sm text-red-600">
                          {isCurrent ? 'This will log you out. ' : ''}Confirm?
                        </span>
                        <button
                          onClick={() => handleRevoke(session.id, isCurrent)}
                          disabled={isRevoking}
                          className="inline-flex items-center px-2 py-1 border border-red-300 rounded text-xs font-medium text-red-700 bg-red-50 hover:bg-red-100 focus:outline-none disabled:opacity-50"
                        >
                          {isRevoking ? 'Revoking...' : 'Yes, Revoke'}
                        </button>
                        <button
                          onClick={() => setConfirmRevoke(null)}
                          disabled={isRevoking}
                          className="inline-flex items-center px-2 py-1 border border-gray-300 rounded text-xs font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none disabled:opacity-50"
                        >
                          Cancel
                        </button>
                      </>
                    ) : (
                      <button
                        onClick={() => handleRevoke(session.id, isCurrent)}
                        disabled={isRevoking || isLoading}
                        className={`inline-flex items-center px-3 py-1.5 border rounded text-xs font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 ${
                          isCurrent
                            ? 'border-red-300 text-red-700 bg-red-50 hover:bg-red-100 focus:ring-red-500'
                            : 'border-gray-300 text-gray-700 bg-white hover:bg-gray-50 focus:ring-blue-500'
                        }`}
                      >
                        Revoke
                      </button>
                    )}
                  </div>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
