import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { usePerformanceData } from '../hooks/usePerformanceData'
import { PerformanceMetricCard } from '../components/common/PerformanceMetricCard'
import { SystemHealthIndicator } from '../components/common/SystemHealthIndicator'
import { PerformanceTrendChart } from '../components/dashboard/PerformanceTrendChart'
import { ToastNotification } from '../components/ToastNotification'

export default function PerformanceDashboard() {
  const navigate = useNavigate()
  const { user, logout: storeLogout, clearAuth } = useAuthStore()

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

  // Handle logout
  const handleLogout = async () => {
    try {
      await storeLogout()
      clearAuth()
      navigate('/login')
    } catch (error) {
      console.error('Logout failed:', error)
      showToast('error', '登出失败', '请重试')
    }
  }

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
    showToast('success', '数据已刷新')
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
          `检测到 ${criticalAnomalies.length} 个严重异常`,
          criticalAnomalies[0].message
        )
      } else if (warningAnomalies.length > 0) {
        showToast(
          'warning',
          `检测到 ${warningAnomalies.length} 个警告`,
          warningAnomalies[0].message
        )
      }
    }
  }, [data])

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <button
                onClick={() => navigate('/dashboard')}
                className="text-blue-600 hover:text-blue-800 mr-4"
              >
                ← 返回
              </button>
              <h1 className="text-xl font-bold text-gray-900">性能监控仪表盘</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-700">
                欢迎, {user?.username || 'Guest'}
              </span>
              <button
                type="button"
                onClick={handleLogout}
                className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors duration-150"
              >
                登出
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header with System Health and Refresh */}
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">系统性能概览</h2>
            <p className="text-sm text-gray-600 mt-1">
              {isPolling ? '实时更新中 (每60秒)' : '已暂停更新'}
            </p>
          </div>

          <div className="flex items-center space-x-6">
            {/* System Health Indicator */}
            {data && (
              <SystemHealthIndicator health={data.system_health} />
            )}

            {/* Refresh Button */}
            <button
              onClick={handleRefresh}
              disabled={isLoading}
              className={`flex items-center space-x-2 px-4 py-2 rounded-md text-sm font-medium transition-colors duration-150 ${
                isLoading
                  ? 'bg-gray-300 cursor-not-allowed'
                  : 'bg-blue-600 hover:bg-blue-700 text-white'
              }`}
            >
              <svg
                className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
              <span>{isLoading ? '刷新中...' : '刷新'}</span>
            </button>
          </div>
        </div>

        {/* Error State */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-400"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">加载失败</h3>
                <div className="mt-2 text-sm text-red-700">{error.message}</div>
              </div>
            </div>
          </div>
        )}

        {/* Loading State */}
        {isLoading && !data && (
          <div className="text-center py-12">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
            <p className="mt-4 text-gray-600">加载性能数据...</p>
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
            <div className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">性能趋势图</h3>
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
              <div className="bg-white rounded-lg shadow p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">
                  性能异常 ({data.anomalies.length})
                </h3>
                <div className="space-y-3">
                  {data.anomalies.map((anomaly, index) => (
                    <div
                      key={index}
                      className={`p-4 rounded-md border-l-4 ${
                        anomaly.severity === 'P0'
                          ? 'bg-red-50 border-red-500'
                          : 'bg-yellow-50 border-yellow-500'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div>
                          <div className="flex items-center space-x-2">
                            <span
                              className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
                                anomaly.severity === 'P0'
                                  ? 'bg-red-100 text-red-800'
                                  : 'bg-yellow-100 text-yellow-800'
                              }`}
                            >
                              {anomaly.severity}
                            </span>
                            <span className="text-sm font-medium text-gray-900">
                              {anomaly.metric_name}
                            </span>
                          </div>
                          <p className="text-sm text-gray-700 mt-1">{anomaly.message}</p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Summary Statistics */}
            {data.summary && (
              <div className="bg-white rounded-lg shadow p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">统计摘要</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="text-center">
                    <div className="text-3xl font-bold text-blue-600">
                      {data.summary.total_requests.toLocaleString()}
                    </div>
                    <div className="text-sm text-gray-600 mt-1">总请求数</div>
                  </div>
                  <div className="text-center">
                    <div className="text-3xl font-bold text-green-600">
                      {data.summary.avg_response_time.toFixed(2)} ms
                    </div>
                    <div className="text-sm text-gray-600 mt-1">平均响应时间</div>
                  </div>
                  <div className="text-center">
                    <div className="text-3xl font-bold text-purple-600">
                      {data.summary.max_response_time.toFixed(2)} ms
                    </div>
                    <div className="text-sm text-gray-600 mt-1">最大响应时间</div>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </main>

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
    </div>
  )
}
