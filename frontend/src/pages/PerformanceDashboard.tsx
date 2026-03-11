import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { usePerformanceData } from '../hooks/usePerformanceData'
import { PerformanceMetricCard } from '../components/common/PerformanceMetricCard'
import { SystemHealthIndicator } from '../components/common/SystemHealthIndicator'
import { PerformanceTrendChart } from '../components/dashboard/PerformanceTrendChart'
import { ToastNotification } from '../components/ToastNotification'
import { PageContainer, ActionButton, ErrorBanner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'

export default function PerformanceDashboard() {
  const { t } = useTranslation()

  // Performance data state
  const {
    data,
    isLoading,
    error,
    refetch,
    isPolling,
  } = usePerformanceData({
    pollingInterval: 60000, // 60 seconds
    enablePolling: true,
    timeRange: '24h',
  })

  // Toast notification state
  const [toast, setToast] = useState<{
    show: boolean
    id: string
    type: 'success' | 'error' | 'warning' | 'info'
    title: string
    message?: string
  }>({
    show: false,
    id: '',
    type: 'success',
    title: '',
  })

  // Show toast notification
  const showToast = (
    type: 'success' | 'error' | 'warning' | 'info',
    title: string,
    message?: string
  ) => {
    const id = Date.now().toString()
    setToast({ show: true, id, type, title, message })
    setTimeout(() => {
      setToast((prev) => ({ ...prev, show: false }))
    }, 5000) // Show for 5 seconds
  }

  // Handle manual refresh
  const handleRefresh = async () => {
    await refetch()
    showToast('success', t('performance.dataRefreshed'))
  }

  // Handle toast close
  const handleToastClose = (_id: string) => {
    setToast((prev) => ({ ...prev, show: false }))
  }

  // Show toast when anomalies are detected
  useEffect(() => {
    if (data && data.anomalies && data.anomalies.length > 0) {
      // Find P0 (critical) anomalies first
      const criticalAnomalies = data.anomalies.filter((a) => a.severity === 'P0')
      const warningAnomalies = data.anomalies.filter((a) => a.severity === 'P1')

      if (criticalAnomalies.length > 0) {
        showToast(
          'error',
          t('performance.criticalAnomalies', { count: criticalAnomalies.length }),
          criticalAnomalies[0].message
        )
      } else if (warningAnomalies.length > 0) {
        showToast(
          'warning',
          t('performance.warningAnomalies', { count: warningAnomalies.length }),
          warningAnomalies[0].message
        )
      }
    }
  }, [data])

  return (
    <PageContainer>
      <PageHeader
        title={t('performance.title')}
        subtitle={isPolling ? t('performance.realtimeUpdating') : t('performance.updatesPaused')}
        showBreadcrumb
        actions={
          <div className="flex items-center space-x-6">
            {/* System Health Indicator */}
            {data && (
              <SystemHealthIndicator health={data.system_health} />
            )}

            {/* Refresh Button */}
            <ActionButton onClick={handleRefresh} disabled={isLoading}>
              {isLoading ? t('common.refreshing') : t('common.refresh')}
            </ActionButton>
          </div>
        }
      />

      {/* Error State */}
      {error && (
        <ErrorBanner error={error} onRetry={refetch} />
      )}

      {/* Loading State */}
      {isLoading && !data && (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">{t('performance.loadingData')}</p>
        </div>
      )}

      {/* Performance Content */}
      {data && (
        <div className="space-y-8">
          {/* Performance Metrics Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {data.metrics.map((metric) => (
              <PerformanceMetricCard key={metric.metric_name} metric={metric} />
            ))}
          </div>

          {/* Trend Chart */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-gray-200 dark:border-gray-700">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{t('performance.trendChart')}</h3>
            <PerformanceTrendChart
              trendData={data.trend_data}
              targetP99={data.metrics[0]?.target_p99}
              targetP95={data.metrics[0]?.target_p95}
              height="400px"
              isLoading={isLoading}
            />
          </div>

          {/* Anomalies Section */}
          {data.anomalies && data.anomalies.length > 0 && (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-gray-200 dark:border-gray-700">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                {t('performance.anomalies')} ({data.anomalies.length})
              </h3>
              <div className="space-y-3">
                {data.anomalies.map((anomaly, index) => (
                  <div
                    key={index}
                    className={`p-4 rounded-md border-l-4 ${
                      anomaly.severity === 'P0'
                        ? 'bg-red-50 dark:bg-red-900/20 border-red-500'
                        : 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-500'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div>
                        <div className="flex items-center space-x-2">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
                              anomaly.severity === 'P0'
                                ? 'bg-red-100 dark:bg-red-900/40 text-red-800 dark:text-red-300'
                                : 'bg-yellow-100 dark:bg-yellow-900/40 text-yellow-800 dark:text-yellow-300'
                            }`}
                          >
                            {anomaly.severity}
                          </span>
                          <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                            {anomaly.metric_name}
                          </span>
                        </div>
                        <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">{anomaly.message}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Summary Statistics */}
          {data.summary && (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-gray-200 dark:border-gray-700">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{t('performance.summary')}</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="text-center">
                  <div className="text-3xl font-bold text-blue-600">
                    {data.summary.total_requests.toLocaleString()}
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">{t('performance.totalRequests')}</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold text-green-600">
                    {data.summary.avg_response_time.toFixed(2)} ms
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">{t('performance.avgResponseTime')}</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold text-purple-600">
                    {data.summary.max_response_time.toFixed(2)} ms
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">{t('performance.maxResponseTime')}</div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Toast Notification */}
      {toast.show && (
        <ToastNotification
          id={toast.id}
          type={toast.type}
          title={toast.title}
          message={toast.message}
          onClose={handleToastClose}
        />
      )}
    </PageContainer>
  )
}
