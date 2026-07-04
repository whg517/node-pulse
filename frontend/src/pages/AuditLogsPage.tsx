import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/authStore'
import { PageHeader } from '@/components/layout/PageHeader'
import { getAuditLogs, type AuditLogDTO, type AuditLogQuery } from '@/api/auditLogs'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'

const PAGE_SIZE = 50

// Audit event types emitted by the backend (internal/auth audit logger).
const EVENT_TYPES = [
  'login',
  'login_failed',
  'login_locked',
  'logout',
  'refresh_token',
  'session_created',
  'session_revoked',
  'password_changed',
  'password_change_failed',
  'password_reset_requested',
  'password_reset_completed',
  'admin_revoke_all',
]

export default function AuditLogsPage() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const isAdmin = user?.role === 'admin'

  const [logs, setLogs] = useState<AuditLogDTO[]>([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Filters (committed on Apply)
  const [fEventType, setFEventType] = useState('')
  const [fUserId, setFUserId] = useState('')
  const [fStart, setFStart] = useState('')
  const [fEnd, setFEnd] = useState('')
  const [applied, setApplied] = useState<AuditLogQuery>({})

  const loadLogs = useCallback(async (query: AuditLogQuery) => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await getAuditLogs(query)
      setLogs(res.logs || [])
      setTotal(res.total_count || 0)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    if (isAdmin) void loadLogs({ ...applied, limit: PAGE_SIZE, offset })
  }, [isAdmin, applied, offset, loadLogs])

  const applyFilters = () => {
    const q: AuditLogQuery = { limit: PAGE_SIZE, offset: 0 }
    if (fEventType) q.event_type = fEventType
    if (fUserId.trim()) q.user_id = fUserId.trim()
    if (fStart) q.start_time = new Date(fStart).toISOString()
    if (fEnd) q.end_time = new Date(fEnd).toISOString()
    setApplied(q)
    setOffset(0)
  }

  const clearFilters = () => {
    setFEventType('')
    setFUserId('')
    setFStart('')
    setFEnd('')
    setApplied({})
    setOffset(0)
  }

  if (!isAdmin) {
    return (
      <div className="space-y-6">
        <PageHeader title={t('settings.auditLogs', 'Audit Logs')} subtitle={t('settings.auditLogsDescription', 'Security event audit trail.')} />
        <Card>
          <CardContent className="py-12 text-center text-sm text-muted-foreground">
            {t('settings.adminOnly', 'Access denied — administrator role required.')}
          </CardContent>
        </Card>
      </div>
    )
  }

  const variantFor = (eventType: string): 'destructive' | 'secondary' | 'default' => {
    if (eventType.includes('failed') || eventType.includes('locked') || eventType.includes('revoke')) return 'destructive'
    if (eventType.startsWith('password')) return 'secondary'
    return 'default'
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('settings.auditLogs', 'Audit Logs')}
        subtitle={t('settings.auditLogsDescription', 'Security event audit trail.')}
        actions={<Button variant="outline" size="sm" onClick={() => void loadLogs({ ...applied, limit: PAGE_SIZE, offset })}>{t('common.refresh', 'Refresh')}</Button>}
      />

      {error && <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

      <Card>
        <CardContent className="space-y-3 p-4">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <div className="space-y-1">
              <Label htmlFor="filter-event">{t('settings.eventType', 'Event type')}</Label>
              <select
                id="filter-event"
                value={fEventType}
                onChange={(e) => setFEventType(e.target.value)}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="">{t('settings.allEvents', 'All events')}</option>
                {EVENT_TYPES.map((ev) => (
                  <option key={ev} value={ev}>{ev}</option>
                ))}
              </select>
            </div>
            <div className="space-y-1">
              <Label htmlFor="filter-user">User ID</Label>
              <Input id="filter-user" value={fUserId} onChange={(e) => setFUserId(e.target.value)} placeholder="uuid" />
            </div>
            <div className="space-y-1">
              <Label htmlFor="filter-start">{t('settings.startTime', 'Start time')}</Label>
              <Input id="filter-start" type="datetime-local" value={fStart} onChange={(e) => setFStart(e.target.value)} />
            </div>
            <div className="space-y-1">
              <Label htmlFor="filter-end">{t('settings.endTime', 'End time')}</Label>
              <Input id="filter-end" type="datetime-local" value={fEnd} onChange={(e) => setFEnd(e.target.value)} />
            </div>
          </div>
          <div className="flex gap-2">
            <Button size="sm" onClick={applyFilters}>{t('common.apply', 'Apply')}</Button>
            <Button size="sm" variant="outline" onClick={clearFilters}>{t('common.clear', 'Clear')}</Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
          ) : logs.length === 0 ? (
            <div className="py-12 text-center text-sm text-muted-foreground">{t('settings.noAuditLogs', 'No audit log entries match the filters.')}</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-border">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.timestamp', 'Timestamp')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.event', 'Event')}</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">User</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">IP</th>
                    <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('settings.details', 'Details')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {logs.map((log) => (
                    <tr key={log.id} className="hover:bg-muted/50 align-top">
                      <td className="whitespace-nowrap px-6 py-4 text-sm text-muted-foreground">
                        {log.created_at ? new Date(log.created_at).toLocaleString() : '—'}
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm">
                        <Badge variant={variantFor(log.event_type)}>{log.event_type}</Badge>
                      </td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm font-mono text-muted-foreground">{log.user_id || '—'}</td>
                      <td className="whitespace-nowrap px-6 py-4 text-sm text-muted-foreground">{log.ip_address || '—'}</td>
                      <td className="px-6 py-4 text-xs text-muted-foreground">
                        {log.details ? (
                          <pre className="max-w-md overflow-x-auto whitespace-pre-wrap break-all rounded bg-muted/40 p-2">{JSON.stringify(log.details)}</pre>
                        ) : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {total > PAGE_SIZE && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {t('common.showing', 'Showing')} {offset + 1}–{Math.min(offset + PAGE_SIZE, total)} {t('common.of', 'of')} {total}
          </p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}>
              {t('common.previous', 'Previous')}
            </Button>
            <Button variant="outline" size="sm" disabled={offset + PAGE_SIZE >= total} onClick={() => setOffset(offset + PAGE_SIZE)}>
              {t('common.next', 'Next')}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
