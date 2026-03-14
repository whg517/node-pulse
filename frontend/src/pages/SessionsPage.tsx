import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import SessionList from '../components/sessions/SessionList'
import { getSessions, deleteSession, getSessionInfo } from '../api/sessions'
import type { Session } from '../types/auth'
import { ToastNotification, type ToastProps } from '../components/ToastNotification'
import { PageContainer, ErrorBanner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'

/**
 * SessionsPage Component
 *
 * Displays and manages active user sessions.
 */
export default function SessionsPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
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
      const [sessionsResponse, infoResponse] = await Promise.all([
        getSessions(),
        getSessionInfo().catch(() => null),
      ])

      setSessions(Array.isArray(sessionsResponse) ? sessionsResponse : [])
      setCurrentSessionId(infoResponse?.session_id || null)
    } catch (err) {
      const error = err as { message?: string }
      setError(error.message || t('errors.failedToLoad'))
      showToast('error', t('common.error'), t('errors.failedToLoad'))
    } finally {
      setIsLoading(false)
    }
  }, [t])

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  const handleRevoke = async (sessionId: string) => {
    try {
      await deleteSession(sessionId)
      showToast('success', t('sessions.revoked'), t('sessions.revokedDescription'))

      if (sessionId === currentSessionId) {
        setTimeout(() => {
          navigate('/login', { replace: true })
        }, 1500)
      } else {
        await loadSessions()
      }
    } catch (err) {
      const error = err as { message?: string }
      showToast('error', t('common.error'), error.message || t('errors.failedToLoad'))
      throw error
    }
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('sessions.title')}
        subtitle={t('sessions.description')}
        showBreadcrumb
      />

      {error && (
        <ErrorBanner error={error} onRetry={loadSessions} className="mb-6" />
      )}

      <SessionList
        sessions={sessions}
        currentSessionId={currentSessionId}
        isLoading={isLoading}
        onRefresh={loadSessions}
        onRevoke={handleRevoke}
      />

      <div className="mt-6 rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-4">
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
            <h3 className="text-sm font-medium text-blue-800 dark:text-blue-300">{t('sessions.securityTips')}</h3>
            <div className="mt-2 text-sm text-blue-700 dark:text-blue-400">
              <ul className="list-disc pl-5 space-y-1">
                <li>{t('sessions.tip1')}</li>
                <li>{t('sessions.tip2')}</li>
                <li>{t('sessions.tip3')}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      {toasts.map((toast) => (
        <ToastNotification key={toast.id} {...toast} onClose={handleToastClose} />
      ))}
    </PageContainer>
  )
}
