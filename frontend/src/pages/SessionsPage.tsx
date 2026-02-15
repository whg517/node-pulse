import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import SessionList from '../components/sessions/SessionList'
import { getSessions, deleteSession, getSessionInfo } from '../api/sessions'
import type { Session } from '../types/auth'
import { ToastNotification, type ToastProps } from '../components/ToastNotification'

/**
 * SessionsPage Component
 *
 * Displays and manages active user sessions.
 */
export default function SessionsPage() {
  const navigate = useNavigate()
  const [sessions, setSessions] = useState<Session[]>([])
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [toasts, setToasts] = useState<ToastProps[]>([])

  const showToast = (type: ToastProps['type'], title: string, message?: string) => {
    const id = Date.now().toString()
    setToasts((prev) => [...prev, { id, type, title, message, onClose: handleToastClose }])
  }

  const handleToastClose = (id: string) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id))
  }

  const loadSessions = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      // Load sessions and session info in parallel
      const [sessionsResponse, infoResponse] = await Promise.all([
        getSessions(),
        getSessionInfo().catch(() => null), // Info endpoint may not exist
      ])

      setSessions(sessionsResponse.data.sessions)
      setCurrentSessionId(infoResponse?.data?.current_session_id || null)
    } catch (err) {
      const error = err as { message?: string }
      setError(error.message || 'Failed to load sessions')
      showToast('error', 'Error', 'Failed to load sessions')
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  const handleRevoke = async (sessionId: string) => {
    try {
      await deleteSession(sessionId)
      showToast('success', 'Session Revoked', 'The session has been terminated')

      // Check if we revoked our current session
      if (sessionId === currentSessionId) {
        // Redirect to login after a brief delay
        setTimeout(() => {
          navigate('/login', { replace: true })
        }, 1500)
      } else {
        // Refresh the session list
        await loadSessions()
      }
    } catch (err) {
      const error = err as { message?: string }
      showToast('error', 'Revoke Failed', error.message || 'Failed to revoke session')
      throw error
    }
  }

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Session Management</h1>
        <p className="mt-1 text-sm text-gray-500">
          View and manage your active sessions across all devices.
        </p>
      </div>

      {error && (
        <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md">
          <p className="text-sm">{error}</p>
          <button
            onClick={loadSessions}
            className="mt-2 text-sm font-medium text-red-700 hover:text-red-800 underline"
          >
            Try again
          </button>
        </div>
      )}

      <SessionList
        sessions={sessions}
        currentSessionId={currentSessionId}
        isLoading={isLoading}
        onRefresh={loadSessions}
        onRevoke={handleRevoke}
      />

      <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
              <path
                fillRule="evenodd"
                d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                clipRule="evenodd"
              />
            </svg>
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-blue-800">Security Tips</h3>
            <div className="mt-2 text-sm text-blue-700">
              <ul className="list-disc pl-5 space-y-1">
                <li>Revoke any sessions you don't recognize immediately</li>
                <li>Always log out from public or shared computers</li>
                <li>Review your sessions regularly for suspicious activity</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      {toasts.map((toast) => (
        <ToastNotification key={toast.id} {...toast} onClose={handleToastClose} />
      ))}
    </div>
  )
}
