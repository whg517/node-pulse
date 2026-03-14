import { useState } from 'react'
import { useTranslation } from 'react-i18next'
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
  const { t } = useTranslation()
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

  // Extract browser/device info from user agent
  const parseUserAgent = (userAgent: string) => {
    const ua = userAgent.toLowerCase()
    let browser = 'Unknown'
    let device = 'Unknown'

    // Detect browser
    if (ua.includes('firefox')) browser = 'Firefox'
    else if (ua.includes('edg')) browser = 'Edge'
    else if (ua.includes('chrome')) browser = 'Chrome'
    else if (ua.includes('safari')) browser = 'Safari'
    else if (ua.includes('opera') || ua.includes('opr')) browser = 'Opera'

    // Detect device
    if (ua.includes('mobile') || ua.includes('android') || ua.includes('iphone')) {
      device = 'Mobile'
    } else if (ua.includes('tablet') || ua.includes('ipad')) {
      device = 'Tablet'
    } else {
      device = 'Desktop'
    }

    return { browser, device }
  }

  if (isLoading) {
    return (
      <div className="bg-white dark:bg-slate-900 shadow rounded-lg p-6">
        <div className="animate-pulse space-y-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center space-x-4">
              <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-1/4"></div>
              <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-1/3"></div>
              <div className="h-4 bg-slate-200 dark:bg-slate-700 rounded w-1/4"></div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
        <h3 className="text-lg font-medium text-gray-900">{t('sessions.activeSessions')}</h3>
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
          {t('common.refresh')}
        </button>
      </div>

      {sessions.length === 0 ? (
        <div className="px-6 py-8 text-center text-gray-500">
          {t('sessions.noSessions')}
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('sessions.device')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('sessions.ipAddress')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('sessions.created')}
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('sessions.expires')}
                </th>
                <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('common.actions')}
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {sessions.map((session) => {
                const isCurrent = session.session_id === currentSessionId
                const isRevoking = revokingId === session.session_id
                const isConfirming = confirmRevoke === session.session_id
                const { browser, device } = parseUserAgent(session.user_agent || '')

                return (
                  <tr key={session.session_id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className="flex-shrink-0 h-8 w-8 rounded-full bg-gray-100 flex items-center justify-center">
                          <svg
                            className="h-4 w-4 text-gray-600"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                            />
                          </svg>
                        </div>
                        <div className="ml-4">
                          <div className="flex items-center space-x-2">
                            <span className="text-sm font-medium text-gray-900">
                              {device}
                            </span>
                            {isCurrent && (
                              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800" data-testid="current-session">
                                {t('sessions.currentSession')}
                              </span>
                            )}
                          </div>
                          <span className="text-xs text-gray-500">{browser}</span>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {session.ip_address || 'Unknown IP'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatDateTime(session.created_at)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatDateTime(session.expires_at)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      {isConfirming ? (
                        <div className="flex items-center justify-end space-x-2">
                          <span className="text-sm text-red-600">
                            {isCurrent ? t('sessions.logoutWarning') + ' ' : ''}{t('sessions.confirmRevoke')}
                          </span>
                          <button
                            onClick={() => handleRevoke(session.session_id, isCurrent)}
                            disabled={isRevoking}
                            className="inline-flex items-center px-2 py-1 border border-red-300 rounded text-xs font-medium text-red-700 bg-red-50 hover:bg-red-100 focus:outline-none disabled:opacity-50"
                          >
                            {isRevoking ? t('sessions.revoking') : t('sessions.yesRevoke')}
                          </button>
                          <button
                            onClick={() => setConfirmRevoke(null)}
                            disabled={isRevoking}
                            className="inline-flex items-center px-2 py-1 border border-gray-300 rounded text-xs font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none disabled:opacity-50"
                          >
                            {t('common.cancel')}
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={() => handleRevoke(session.session_id, isCurrent)}
                          disabled={isRevoking || isLoading}
                          className={`inline-flex items-center px-3 py-1.5 border rounded text-xs font-medium focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 ${
                            isCurrent
                              ? 'border-red-300 text-red-700 bg-red-50 hover:bg-red-100 focus:ring-red-500'
                              : 'border-gray-300 text-gray-700 bg-white hover:bg-gray-50 focus:ring-blue-500'
                          }`}
                        >
                          {t('sessions.revoke')}
                        </button>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
