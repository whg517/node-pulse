import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { PageHeader } from '@/components/layout/PageHeader'
import { getSessions, deleteSession, getSessionInfo, revokeAllSessions } from '@/api/sessions'
import type { Session } from '@/types/auth'

export default function SessionsPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [sessions, setSessions] = useState<Session[]>([])
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isRevokingAll, setIsRevokingAll] = useState(false)

  const loadSessions = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [sessionsRes, infoRes] = await Promise.all([
        getSessions(),
        getSessionInfo().catch(() => null),
      ])
      setSessions(Array.isArray(sessionsRes) ? sessionsRes : [])
      setCurrentSessionId(infoRes?.session_id || null)
    } catch (err) {
      setError((err as { message?: string }).message || t('errors.failedToLoad'))
    } finally {
      setIsLoading(false)
    }
  }, [t])

  useEffect(() => { loadSessions() }, [loadSessions])

  const handleRevoke = async (sessionId: string) => {
    try {
      await deleteSession(sessionId)
      if (sessionId === currentSessionId) {
        setTimeout(() => navigate('/login', { replace: true }), 1500)
      } else {
        await loadSessions()
      }
    } catch (err) {
      setError((err as { message?: string }).message || t('errors.failedToLoad'))
    }
  }

  const handleRevokeAll = async () => {
    setIsRevokingAll(true)
    setError(null)
    try {
      await revokeAllSessions()
      // Revoking all sessions includes the current one; sign the user out.
      setTimeout(() => navigate('/login', { replace: true }), 1500)
    } catch (err) {
      setError((err as { message?: string }).message || t('errors.failedToLoad'))
      setIsRevokingAll(false)
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('sessions.title')}
        subtitle={t('sessions.description')}
        actions={
          sessions.length > 0 ? (
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="destructive" size="sm" disabled={isRevokingAll}>
                  {isRevokingAll ? t('sessions.revokingAll', 'Signing out...') : t('sessions.revokeAll', 'Sign out all sessions')}
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>{t('sessions.confirmRevokeAllTitle', 'Sign out all sessions?')}</AlertDialogTitle>
                  <AlertDialogDescription>
                    {t('sessions.confirmRevokeAllDesc', 'This will sign out every device including this one. You will need to log in again.')}
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
                  <AlertDialogAction variant="destructive" onClick={() => void handleRevokeAll()}>
                    {t('sessions.revokeAll', 'Sign out all')}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          ) : undefined
        }
      />

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
          <Button variant="link" size="sm" onClick={loadSessions} className="ml-2">
            {t('common.retry')}
          </Button>
        </div>
      )}

      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex justify-center py-12 text-muted-foreground">Loading...</div>
          ) : sessions.length === 0 ? (
            <div className="py-12 text-center text-sm text-muted-foreground">
              {t('sessions.noSessions')}
            </div>
          ) : (
            <div className="divide-y">
              {sessions.map((session) => (
                <div key={session.session_id} className="flex items-center justify-between px-6 py-4">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium">{session.ip_address}</span>
                      {session.session_id === currentSessionId && (
                        <Badge variant="secondary">Current</Badge>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {session.user_agent} · {new Date(session.created_at).toLocaleString()}
                    </p>
                  </div>
                  {session.session_id !== currentSessionId && (
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => handleRevoke(session.session_id)}
                    >
                      {t('sessions.revoke')}
                    </Button>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-primary/20 bg-primary/5">
        <CardHeader>
          <CardTitle className="text-sm">{t('sessions.securityTips')}</CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="list-disc space-y-1 pl-5 text-sm text-muted-foreground">
            <li>{t('sessions.tip1')}</li>
            <li>{t('sessions.tip2')}</li>
            <li>{t('sessions.tip3')}</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  )
}
