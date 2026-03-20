import { useNavigate } from 'react-router-dom'
import { memo, useMemo, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useAlertsStore } from '../../stores/alertsStore'
import { memoCompare } from '../../utils/deepEqual'
import * as NotificationService from '../../services/NotificationService'
import * as WebSocketService from '../../services/WebSocketService'

interface AlertStreamProps { maxItems?: number; className?: string; isLoading?: boolean }
type AlertLevel = 'P0' | 'P1' | 'P2'

function getSeverityStyles(level: AlertLevel): string {
  const styles: Record<AlertLevel, string> = {
    P0: 'bg-[var(--color-critical-bg)] border-[var(--color-critical-bg)] hover:opacity-80',
    P1: 'bg-[var(--color-warning-bg)] border-[var(--color-warning-bg)] hover:opacity-80',
    P2: 'bg-[var(--color-brand-muted)] border-[var(--color-brand-muted)] hover:opacity-80',
  }
  return styles[level]
}

function getLevelBadgeStyles(level: AlertLevel): string {
  const styles: Record<AlertLevel, string> = {
    P0: 'bg-[var(--color-critical-bg)] text-[var(--color-critical-text)]',
    P1: 'bg-[var(--color-warning-bg)] text-[var(--color-warning-text)]',
    P2: 'bg-[var(--color-brand-muted)] text-[var(--color-brand)]',
  }
  return styles[level]
}

function formatTimeAgo(timestamp: string, t: (key: string, options?: Record<string, unknown>) => string): string {
  const diffMs = new Date().getTime() - new Date(timestamp).getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)
  if (diffMins < 1) return t('time.justNow')
  if (diffMins < 60) return t('time.minutesAgo', { count: diffMins })
  if (diffHours < 24) return t('time.hoursAgo', { count: diffHours })
  return t('time.daysAgo', { count: diffDays })
}

function formatUTCTime(timestamp: string): string {
  return new Date(timestamp).toLocaleString('en-US', { timeZone: 'UTC', year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false }) + ' UTC'
}

function getMetricDisplay(metric: string): string {
  const displays: Record<string, string> = { latency: 'Latency', packet_loss_rate: 'Packet Loss', jitter: 'Jitter' }
  return displays[metric] || metric
}

export const AlertStream = memo(function AlertStream({ maxItems = 10, className = '', isLoading = false }: AlertStreamProps) {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const alertRecords = useAlertsStore((state) => state.alertRecords)

  useEffect(() => {
    const handleNotificationClick = (alertId: string) => navigate(`/alerts?highlight=${alertId}`)
    NotificationService.initialize(handleNotificationClick)
    return () => NotificationService.destroy()
  }, [navigate])

  useEffect(() => {
    const handleMessage = (message: WebSocketService.WebSocketMessage<unknown>) => {
      if (message.type === 'alert:new') {
        const payload = message.payload as { id: string; level: string; node_id: string; metric: string; threshold: string }
        const nodeName = `Node ${payload.node_id.slice(0, 8)}`
        NotificationService.showAlertNotification(payload.id, payload.level, nodeName, payload.metric, String(payload.threshold))
      }
    }
    WebSocketService.initialize(handleMessage)
    WebSocketService.connect()
    return () => WebSocketService.disconnect()
  }, [alertRecords, navigate])

  const activeAlerts = useMemo(() => {
    const safeRecords = Array.isArray(alertRecords) ? alertRecords : []
    return safeRecords.filter((record) => record.status !== 'resolved').sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()).slice(0, maxItems)
  }, [alertRecords, maxItems])

  const handleAlertClick = (alertId: string) => navigate(`/alerts?highlight=${alertId}`)

  if (isLoading) {
    return (<div className={`rounded-lg border border-[var(--color-border)] shadow-sm overflow-hidden bg-[var(--color-bg-surface)] ${className}`}><div className="px-4 py-3 border-b border-[var(--color-border)]"><div className="animate-pulse"><div className="h-5 rounded w-32 bg-[var(--color-bg-muted)]"></div></div></div><div className="p-3 space-y-2 max-h-96 overflow-y-auto">{[...Array(5)].map((_, i) => (<div key={i} className="h-16 rounded animate-pulse bg-[var(--color-bg-muted)]"></div>))}</div></div>)
  }

  if (activeAlerts.length === 0) {
    return (<div className={`rounded-lg border border-[var(--color-border)] shadow-sm overflow-hidden bg-[var(--color-bg-surface)] ${className}`}><div className="px-4 py-3 border-b border-[var(--color-border)]"><h3 className="text-sm font-semibold text-[var(--color-text-primary)]">{t('alerts.activeAlerts')}</h3></div><div className="text-center py-8"><svg className="mx-auto h-10 w-10 text-[var(--color-healthy)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg><p className="mt-2 text-sm font-medium text-[var(--color-text-primary)]">{t('alerts.noActiveAlerts')}</p><p className="mt-1 text-xs text-[var(--color-text-muted)]">{t('alerts.allSystemsNormal')}</p></div></div>)
  }

  return (
    <div className={`rounded-lg border border-[var(--color-border)] shadow-sm overflow-hidden bg-[var(--color-bg-surface)] ${className}`}>
      <div className="px-4 py-3 border-b border-[var(--color-border)]"><div className="flex items-center justify-between"><h3 className="text-sm font-semibold text-[var(--color-text-primary)]">{t('alerts.activeAlerts')}</h3><span className="px-2 py-0.5 text-xs font-medium rounded-full bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)]">{activeAlerts.length}</span></div></div>
      <ul className="p-3 space-y-2 max-h-96 overflow-y-auto">
        {activeAlerts.map((alert) => {
          const level = (alert.level || 'P2').toUpperCase() as AlertLevel
          const severityStyles = getSeverityStyles(level)
          const badgeStyles = getLevelBadgeStyles(level)
          return (
            <li key={alert.id} onClick={() => handleAlertClick(alert.id)} className={`cursor-pointer rounded-lg border p-3 transition-all duration-150 ${severityStyles}`} role="button" tabIndex={0} aria-label={`Alert: ${getMetricDisplay(alert.metric)} on node ${alert.nodeId.slice(0, 8)}`} onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') handleAlertClick(alert.id) }}>
              <div className="flex items-start justify-between gap-2">
                <div className="flex items-start gap-2 min-w-0 flex-1">
                  <span className={`inline-flex items-center px-2 py-0.5 text-xs font-semibold rounded shrink-0 ${badgeStyles}`}>{level}</span>
                  <div className="min-w-0"><p className="text-sm font-medium truncate text-[var(--color-text-primary)]">{getMetricDisplay(alert.metric)}</p><p className="text-xs truncate text-[var(--color-text-muted)]">Node: {alert.nodeId.slice(0, 8)}...</p></div>
                </div>
                <div className="text-right"><span className="text-xs shrink-0 block text-[var(--color-text-muted)]">{formatTimeAgo(alert.timestamp, t)}</span><span className="text-[10px] shrink-0 block mt-0.5 text-[var(--color-text-placeholder)]" title={formatUTCTime(alert.timestamp)}>{formatUTCTime(alert.timestamp)}</span></div>
              </div>
            </li>
          )
        })}
      </ul>
      <div className="px-4 py-2 border-t border-[var(--color-border)] text-center"><button onClick={() => navigate('/alerts')} className="text-xs font-medium hover:underline text-[var(--color-brand)] hover:text-[var(--color-brand-hover)]">{t('alerts.viewAllAlerts')}</button></div>
    </div>
  )
}, memoCompare)
