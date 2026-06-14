import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { usePerformanceData } from '@/hooks/usePerformanceData'
import { PerformanceMetricCard } from '@/components/common/PerformanceMetricCard'
import { SystemHealthIndicator } from '@/components/common/SystemHealthIndicator'
import { PerformanceTrendChart } from '@/components/dashboard/PerformanceTrendChart'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'

export default function PerformanceDashboard() {
  const { t } = useTranslation()

  const { data, isLoading, error, refetch, isPolling } = usePerformanceData({
    pollingInterval: 60000, enablePolling: true, timeRange: '24h',
  })

  const [toast, setToast] = useState<{
    show: boolean; type: 'success' | 'error' | 'warning' | 'info'; title: string; message?: string
  }>({ show: false, type: 'success', title: '' })

  const showToast = (type: 'success' | 'error' | 'warning' | 'info', title: string, message?: string) => {
    setToast({ show: true, type, title, message })
    setTimeout(() => setToast((prev) => ({ ...prev, show: false })), 5000)
  }

  const handleRefresh = async () => {
    await refetch()
    showToast('success', t('performance.dataRefreshed'))
  }

  useEffect(() => {
    let timeoutId: number | undefined
    if (data?.anomalies?.length) {
      const critical = data.anomalies.filter((a) => a.severity === 'P0')
      const warning = data.anomalies.filter((a) => a.severity === 'P1')
      if (critical.length > 0) timeoutId = window.setTimeout(() => showToast('error', t('performance.criticalAnomalies', { count: critical.length }), critical[0].message), 0)
      else if (warning.length > 0) timeoutId = window.setTimeout(() => showToast('warning', t('performance.warningAnomalies', { count: warning.length }), warning[0].message), 0)
    }
    return () => { if (timeoutId !== undefined) window.clearTimeout(timeoutId) }
  }, [data, t])

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('performance.title')}
        subtitle={isPolling ? t('performance.realtimeUpdating') : t('performance.updatesPaused')}
        actions={
          <div className="flex items-center gap-4">
            {data && <SystemHealthIndicator health={data.system_health} />}
            <Button variant="outline" size="sm" onClick={handleRefresh} disabled={isLoading}>
              {isLoading ? t('common.refreshing') : t('common.refresh')}
            </Button>
          </div>
        }
      />

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error.message}
          <Button variant="link" size="sm" onClick={refetch}>Retry</Button>
        </div>
      )}

      {isLoading && !data && (
        <div className="flex justify-center py-12">
          <div className="text-center">
            <div className="h-12 w-12 animate-spin rounded-full border-2 border-primary border-t-transparent mx-auto" />
            <p className="mt-4 text-muted-foreground">{t('performance.loadingData')}</p>
          </div>
        </div>
      )}

      {data && (
        <div className="space-y-8">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {data.metrics.map((metric) => (
              <PerformanceMetricCard key={metric.metric_name} metric={metric} />
            ))}
          </div>

          <Card>
            <CardContent className="p-6">
              <h3 className="text-lg font-semibold mb-4">{t('performance.trendChart')}</h3>
              <PerformanceTrendChart trendData={data.trend_data} targetP99={data.metrics[0]?.target_p99} targetP95={data.metrics[0]?.target_p95} height="400px" isLoading={isLoading} />
            </CardContent>
          </Card>

          {data.anomalies && data.anomalies.length > 0 && (
            <Card>
              <CardContent className="p-6">
                <h3 className="text-lg font-semibold mb-4">{t('performance.anomalies')} ({data.anomalies.length})</h3>
                <div className="space-y-3">
                  {data.anomalies.map((anomaly, index) => (
                    <div key={index} className={`p-4 rounded-md border-l-4 ${anomaly.severity === 'P0' ? 'bg-destructive/10 border-l-destructive' : 'bg-yellow-50 border-l-yellow-500 dark:bg-yellow-950'}`}>
                      <div className="flex items-center gap-2">
                        <Badge variant={anomaly.severity === 'P0' ? 'destructive' : 'secondary'}>{anomaly.severity}</Badge>
                        <span className="text-sm font-medium">{anomaly.metric_name}</span>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">{anomaly.message}</p>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {data.summary && (
            <Card>
              <CardContent className="p-6">
                <h3 className="text-lg font-semibold mb-4">{t('performance.summary')}</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-primary">{data.summary.total_requests.toLocaleString()}</div>
                    <div className="text-sm text-muted-foreground mt-1">{t('performance.totalRequests')}</div>
                  </div>
                  <div className="text-center">
                    <div className="text-3xl font-bold text-green-600">{data.summary.avg_response_time.toFixed(2)} ms</div>
                    <div className="text-sm text-muted-foreground mt-1">{t('performance.avgResponseTime')}</div>
                  </div>
                  <div className="text-center">
                    <div className="text-3xl font-bold text-purple-600">{data.summary.max_response_time.toFixed(2)} ms</div>
                    <div className="text-sm text-muted-foreground mt-1">{t('performance.maxResponseTime')}</div>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* Toast */}
      {toast.show && (
        <div className={`fixed top-4 right-4 z-50 p-4 rounded-lg border shadow-lg ${
          toast.type === 'success' ? 'bg-green-50 border-green-200 text-green-700 dark:bg-green-950 dark:text-green-400'
          : toast.type === 'error' ? 'bg-destructive/10 border-destructive/30 text-destructive'
          : toast.type === 'warning' ? 'bg-yellow-50 border-yellow-200 text-yellow-700 dark:bg-yellow-950 dark:text-yellow-400'
          : 'bg-muted border text-foreground'
        }`}>
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className="font-semibold text-sm">{toast.title}</h3>
              {toast.message && <p className="text-sm mt-1">{toast.message}</p>}
            </div>
            <button onClick={() => setToast((p) => ({ ...p, show: false }))} className="opacity-60 hover:opacity-100">✕</button>
          </div>
        </div>
      )}
    </div>
  )
}
