import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PageContainer, ActionButton, LoadingSpinner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { fetchSystemHealth } from '../api/health'
import type { HealthCheckResponse } from '../api/health'

type ServiceStatus = 'healthy' | 'degraded' | 'unhealthy' | 'down'

interface ServiceHealth {
  name: string
  key: string
  status: ServiceStatus
  detail: string
}

function mapCheckStatus(checkValue: string): ServiceStatus {
  const v = checkValue.toLowerCase()
  if (v === 'ok' || v === 'healthy') return 'healthy'
  if (v.startsWith('error') || v === 'unhealthy') return 'down'
  if (v === 'degraded' || v === 'stale' || v === 'full') return 'degraded'
  if (v === 'disabled' || v === 'nodata') return 'degraded'
  return 'healthy'
}

function parseHealthResponse(data: HealthCheckResponse): ServiceHealth[] {
  const services: ServiceHealth[] = []

  services.push({
    name: 'Database (PostgreSQL)',
    key: 'database',
    status: mapCheckStatus(data.checks.database),
    detail: data.checks.database,
  })

  services.push({
    name: 'Alert Engine',
    key: 'alert_engine',
    status: mapCheckStatus(data.checks.alert_engine),
    detail: data.checks.alert_engine,
  })

  services.push({
    name: 'Webhook Delivery',
    key: 'webhook_delivery',
    status: mapCheckStatus(data.checks.webhook_delivery),
    detail: data.checks.webhook_delivery,
  })

  services.push({
    name: 'Alert Suppression',
    key: 'alert_suppression',
    status: mapCheckStatus(data.checks.alert_suppression),
    detail: data.checks.alert_suppression,
  })

  return services
}

export default function SystemHealthPage() {
  const { t } = useTranslation()
  const [healthData, setHealthData] = useState<HealthCheckResponse | null>(null)
  const [services, setServices] = useState<ServiceHealth[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date())
  const mountedRef = useRef(true)
  const pollingRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  const loadHealth = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await fetchSystemHealth()
      if (!mountedRef.current) return
      setHealthData(data)
      setServices(parseHealthResponse(data))
      setLastRefresh(new Date())
    } catch (err) {
      if (mountedRef.current) {
        setError(err instanceof Error ? err.message : String(err))
      }
    } finally {
      if (mountedRef.current) setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    mountedRef.current = true
    void loadHealth()

    const poll = () => {
      pollingRef.current = setTimeout(async () => {
        try {
          const data = await fetchSystemHealth()
          if (mountedRef.current) {
            setHealthData(data)
            setServices(parseHealthResponse(data))
            setLastRefresh(new Date())
          }
        } catch { /* ignore polling errors */ }
        if (mountedRef.current) poll()
      }, 15_000)
    }
    poll()

    return () => {
      mountedRef.current = false
      if (pollingRef.current) clearTimeout(pollingRef.current)
    }
  }, [loadHealth])

  const getStatusColor = (status: ServiceStatus) => {
    switch (status) {
      case 'healthy':
        return { bg: 'bg-[var(--color-healthy-bg)]', text: 'text-[var(--color-healthy-text)]', dot: 'bg-[var(--color-healthy)]' }
      case 'degraded':
        return { bg: 'bg-[var(--color-warning-bg)]', text: 'text-[var(--color-warning-text)]', dot: 'bg-[var(--color-warning)]' }
      case 'down':
      case 'unhealthy':
        return { bg: 'bg-[var(--color-critical-bg)]', text: 'text-[var(--color-critical-text)]', dot: 'bg-[var(--color-critical)]' }
    }
  }

  const overallStatus: ServiceStatus = services.some((s) => s.status === 'down' || s.status === 'unhealthy')
    ? 'down'
    : services.some((s) => s.status === 'degraded')
      ? 'degraded'
      : 'healthy'

  const overallColor = getStatusColor(overallStatus)

  return (
    <PageContainer>
      <PageHeader
        title={t('integrations.systemHealth')}
        subtitle={t('integrations.systemHealthDescription')}
        actions={
          <div className="flex flex-shrink-0 items-center gap-3">
            <span className="text-sm text-[var(--color-text-secondary)]">
              {t('integrations.lastRefresh')}: {lastRefresh.toLocaleTimeString()}
            </span>
            <ActionButton onClick={() => void loadHealth()} disabled={isLoading}>
              {isLoading ? t('common.refreshing') : t('common.refresh')}
            </ActionButton>
          </div>
        }
      />

      {error && (
        <div className="mb-6 rounded-lg border border-[var(--color-critical)] bg-[var(--color-critical-bg)] px-4 py-3 text-sm text-[var(--color-critical)]">
          {error}
        </div>
      )}

      {/* Overall Status */}
      <div className={`mb-6 p-4 rounded-lg border ${overallColor.bg} border-current`}>
        <div className="flex items-center gap-3">
          <div className={`w-4 h-4 rounded-full ${overallColor.dot} ${overallStatus === 'degraded' ? 'animate-pulse' : ''}`} />
          <span className={`text-lg font-semibold ${overallColor.text}`}>
            {overallStatus === 'healthy'
              ? t('integrations.allSystemsOperational')
              : overallStatus === 'degraded'
                ? t('integrations.someSystemsDegraded')
                : t('integrations.someSystemsDown')}
          </span>
          {healthData && (
            <span className="ml-auto text-xs text-[var(--color-text-muted)]">
              {new Date(healthData.timestamp).toLocaleString()}
            </span>
          )}
        </div>
      </div>

      {/* Service Health Grid */}
      {isLoading && !healthData ? (
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
          {services.map((service) => {
            const color = getStatusColor(service.status)
            return (
              <div
                key={service.key}
                className="rounded-lg border border-[var(--color-border)] p-4 bg-[var(--color-bg-surface)]"
              >
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
                    {service.name}
                  </h3>
                  <span className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium ${color.bg} ${color.text}`}>
                    <span className={`w-2 h-2 rounded-full ${color.dot}`} />
                    {service.status.toUpperCase()}
                  </span>
                </div>
                <div className="text-sm text-[var(--color-text-secondary)]">
                  {service.detail}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* {t('integrations.alertSystemDetails')} */}
      {healthData?.alert_system && (
        <div className="mb-6 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
              {t('integrations.alertSystemDetails')}
            </h3>
          </div>
          <div className="p-4 grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
            {/* Alert Engine */}
            <div>
              <div className="font-medium text-[var(--color-text-primary)] mb-1">Alert Engine</div>
              <div className="text-[var(--color-text-secondary)]">
                Status: {healthData.alert_system.alert_engine.status} · Cached rules: {healthData.alert_system.alert_engine.cached_rules}
              </div>
              <div className="text-xs text-[var(--color-text-muted)] mt-1">
                Channel: {healthData.alert_system.alert_engine.metric_channel_depth}/{healthData.alert_system.alert_engine.metric_channel_capacity}
              </div>
            </div>
            {/* Webhook Delivery */}
            <div>
              <div className="font-medium text-[var(--color-text-primary)] mb-1">Webhook Delivery</div>
              <div className="text-[var(--color-text-secondary)]">
                Status: {healthData.alert_system.webhook_delivery.status} · Success rate: {healthData.alert_system.webhook_delivery.success_rate.toFixed(1)}%
              </div>
              <div className="text-xs text-[var(--color-text-muted)] mt-1">
                Total: {healthData.alert_system.webhook_delivery.total_count} · Success: {healthData.alert_system.webhook_delivery.success_count}
              </div>
            </div>
            {/* Alert Suppression */}
            <div>
              <div className="font-medium text-[var(--color-text-primary)] mb-1">Alert Suppression</div>
              <div className="text-[var(--color-text-secondary)]">
                Status: {healthData.alert_system.alert_suppression.status} · Active suppressions: {healthData.alert_system.alert_suppression.active_suppression_count}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* {t('integrations.schedulerTasks')} */}
      {healthData?.scheduler && (
        <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
          <div className="px-4 py-3 border-b border-[var(--color-border)]">
            <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
              {t('integrations.schedulerTasks')}
              <span className={`ml-2 text-xs ${healthData.scheduler.running ? 'text-[var(--color-healthy)]' : 'text-[var(--color-critical)]'}`}>
                {healthData.scheduler.running ? t('integrations.running') : t('integrations.stopped')}
              </span>
            </h3>
          </div>
          <div className="divide-y divide-[var(--color-border)]">
            {Object.entries(healthData.scheduler.tasks).map(([name, task]) => (
              <div key={name} className="p-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-[var(--color-text-primary)]">{name}</span>
                  <span className={`text-xs ${task.is_running ? 'text-[var(--color-healthy)]' : 'text-[var(--color-text-muted)]'}`}>
                    {task.is_running ? t('integrations.running') : t('integrations.idle')} · {task.run_count}
                  </span>
                </div>
                <div className="mt-1 text-xs text-[var(--color-text-muted)]">
                  {task.last_run ? new Date(task.last_run).toLocaleString() : t('reports.never')}
                  {task.last_error && (
                    <span className="ml-2 text-[var(--color-critical)]">Error: {task.last_error}</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </PageContainer>
  )
}
