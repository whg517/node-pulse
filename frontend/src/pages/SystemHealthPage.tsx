import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { fetchSystemHealth } from '@/api/health'
import type { HealthCheckResponse } from '@/api/health'

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
  return [
    { name: 'Database (PostgreSQL)', key: 'database', status: mapCheckStatus(data.checks.database), detail: data.checks.database },
    { name: 'Alert Engine', key: 'alert_engine', status: mapCheckStatus(data.checks.alert_engine), detail: data.checks.alert_engine },
    { name: 'Webhook Delivery', key: 'webhook_delivery', status: mapCheckStatus(data.checks.webhook_delivery), detail: data.checks.webhook_delivery },
    { name: 'Alert Suppression', key: 'alert_suppression', status: mapCheckStatus(data.checks.alert_suppression), detail: data.checks.alert_suppression },
  ]
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
      if (mountedRef.current) setError(err instanceof Error ? err.message : String(err))
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
          if (mountedRef.current) { setHealthData(data); setServices(parseHealthResponse(data)); setLastRefresh(new Date()) }
        } catch { /* ignore polling errors */ }
        if (mountedRef.current) poll()
      }, 15_000)
    }
    poll()
    return () => { mountedRef.current = false; if (pollingRef.current) clearTimeout(pollingRef.current) }
  }, [loadHealth])

  const statusVariant = (status: ServiceStatus): 'default' | 'secondary' | 'destructive' => {
    if (status === 'healthy') return 'default'
    if (status === 'degraded') return 'secondary'
    return 'destructive'
  }

  const overallStatus: ServiceStatus = services.some((s) => s.status === 'down' || s.status === 'unhealthy')
    ? 'down' : services.some((s) => s.status === 'degraded') ? 'degraded' : 'healthy'

  const overallClasses = overallStatus === 'healthy'
    ? 'bg-green-50 border-green-200 text-green-700 dark:bg-green-950 dark:border-green-800 dark:text-green-400'
    : overallStatus === 'degraded'
    ? 'bg-yellow-50 border-yellow-200 text-yellow-700 dark:bg-yellow-950 dark:border-yellow-800 dark:text-yellow-400'
    : 'bg-red-50 border-red-200 text-red-700 dark:bg-red-950 dark:border-red-800 dark:text-red-400'

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('integrations.systemHealth')}
        subtitle={t('integrations.systemHealthDescription')}
        actions={
          <div className="flex items-center gap-3">
            <span className="text-sm text-muted-foreground">{t('integrations.lastRefresh')}: {lastRefresh.toLocaleTimeString()}</span>
            <Button variant="outline" size="sm" onClick={() => void loadHealth()} disabled={isLoading}>
              {isLoading ? t('common.refreshing') : t('common.refresh')}
            </Button>
          </div>
        }
      />

      {error && <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

      {/* Overall Status */}
      <div className={`p-4 rounded-lg border ${overallClasses}`}>
        <div className="flex items-center gap-3">
          <div className={`w-4 h-4 rounded-full ${
            overallStatus === 'healthy' ? 'bg-green-500' : overallStatus === 'degraded' ? 'bg-yellow-500 animate-pulse' : 'bg-red-500'
          }`} />
          <span className="text-lg font-semibold">
            {overallStatus === 'healthy' ? t('integrations.allSystemsOperational')
              : overallStatus === 'degraded' ? t('integrations.someSystemsDegraded')
              : t('integrations.someSystemsDown')}
          </span>
          {healthData && (
            <span className="ml-auto text-xs text-muted-foreground">{new Date(healthData.timestamp).toLocaleString()}</span>
          )}
        </div>
      </div>

      {/* Service Health Grid */}
      {isLoading && !healthData ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {services.map((service) => (
            <Card key={service.key}>
              <CardContent className="p-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-sm font-semibold">{service.name}</h3>
                  <Badge variant={statusVariant(service.status)}>{service.status.toUpperCase()}</Badge>
                </div>
                <div className="text-sm text-muted-foreground">{service.detail}</div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Alert System Details */}
      {healthData?.alert_system && (
        <Card>
          <CardContent className="p-0">
            <div className="px-4 py-3 border-b">
              <h3 className="text-sm font-semibold">{t('integrations.alertSystemDetails')}</h3>
            </div>
            <div className="p-4 grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
              <div>
                <div className="font-medium mb-1">Alert Engine</div>
                <div className="text-muted-foreground">
                  Status: {healthData.alert_system.alert_engine.status} · Cached rules: {healthData.alert_system.alert_engine.cached_rules}
                </div>
                <div className="text-xs text-muted-foreground mt-1">
                  Channel: {healthData.alert_system.alert_engine.metric_channel_depth}/{healthData.alert_system.alert_engine.metric_channel_capacity}
                </div>
              </div>
              <div>
                <div className="font-medium mb-1">Webhook Delivery</div>
                <div className="text-muted-foreground">
                  Status: {healthData.alert_system.webhook_delivery.status} · Success rate: {healthData.alert_system.webhook_delivery.success_rate.toFixed(1)}%
                </div>
                <div className="text-xs text-muted-foreground mt-1">
                  Total: {healthData.alert_system.webhook_delivery.total_count} · Success: {healthData.alert_system.webhook_delivery.success_count}
                </div>
              </div>
              <div>
                <div className="font-medium mb-1">Alert Suppression</div>
                <div className="text-muted-foreground">
                  Status: {healthData.alert_system.alert_suppression.status} · Active suppressions: {healthData.alert_system.alert_suppression.active_suppression_count}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Scheduler Tasks */}
      {healthData?.scheduler && (
        <Card>
          <CardContent className="p-0">
            <div className="px-4 py-3 border-b">
              <h3 className="text-sm font-semibold">
                {t('integrations.schedulerTasks')}
                <span className={`ml-2 text-xs ${healthData.scheduler.running ? 'text-green-600' : 'text-destructive'}`}>
                  {healthData.scheduler.running ? t('integrations.running') : t('integrations.stopped')}
                </span>
              </h3>
            </div>
            <div className="divide-y">
              {Object.entries(healthData.scheduler.tasks).map(([name, task]) => (
                <div key={name} className="p-4">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">{name}</span>
                    <span className={`text-xs ${task.is_running ? 'text-green-600' : 'text-muted-foreground'}`}>
                      {task.is_running ? t('integrations.running') : t('integrations.idle')} · {task.run_count}
                    </span>
                  </div>
                  <div className="mt-1 text-xs text-muted-foreground">
                    {task.last_run ? new Date(task.last_run).toLocaleString() : t('reports.never')}
                    {task.last_error && <span className="ml-2 text-destructive">Error: {task.last_error}</span>}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
