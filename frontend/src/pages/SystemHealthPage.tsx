/**
 * System Health Page
 *
 * System health monitoring and integration status.
 * Route: /integrations/health
 */

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useTheme } from '../hooks/useTheme'
import { PageHeader } from '../components/layout'

interface HealthStatus {
  name: string
  status: 'healthy' | 'degraded' | 'down'
  responseTime?: number
  uptime?: string
  lastCheck: string
  message?: string
}

// TODO: Replace mock data with real API calls to /api/v1/health/services
// Mock health data (placeholder until backend API is implemented)
const mockHealthData: HealthStatus[] = [
  {
    name: 'API Server',
    status: 'healthy',
    responseTime: 45,
    uptime: '99.98%',
    lastCheck: new Date().toISOString(),
  },
  {
    name: 'Database (PostgreSQL)',
    status: 'healthy',
    responseTime: 12,
    uptime: '99.99%',
    lastCheck: new Date().toISOString(),
  },
  {
    name: 'Cache Layer',
    status: 'healthy',
    responseTime: 2,
    uptime: '100%',
    lastCheck: new Date().toISOString(),
  },
  {
    name: 'Alert Engine',
    status: 'healthy',
    responseTime: 8,
    uptime: '99.95%',
    lastCheck: new Date().toISOString(),
  },
]

interface SystemEvent {
  id: string
  type: 'info' | 'warning' | 'error'
  message: string
  timestamp: string
}

const mockEvents: SystemEvent[] = [
  {
    id: '1',
    type: 'info',
    message: 'System health check completed successfully',
    timestamp: new Date().toISOString(),
  },
  {
    id: '2',
    type: 'warning',
    message: 'High memory usage detected on API server (85%)',
    timestamp: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: '3',
    type: 'info',
    message: 'Database backup completed',
    timestamp: new Date(Date.now() - 7200000).toISOString(),
  },
]

export default function SystemHealthPage() {
  const { t } = useTranslation()
  const { isDark } = useTheme()
  const [healthData] = useState<HealthStatus[]>(mockHealthData)
  const [events] = useState<SystemEvent[]>(mockEvents)
  const [isLoading, setIsLoading] = useState(false)
  const [lastRefresh, setLastRefresh] = useState(new Date())

  const handleRefresh = async () => {
    setIsLoading(true)
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1000))
    setLastRefresh(new Date())
    setIsLoading(false)
  }

  const getStatusStyles = (status: HealthStatus['status']): { bg: string; text: string; dot: string } => {
    const styles = {
      healthy: {
        bg: 'bg-green-100 dark:bg-green-900/30',
        text: 'text-green-800 dark:text-green-400',
        dot: 'bg-green-500',
      },
      degraded: {
        bg: 'bg-yellow-100 dark:bg-yellow-900/30',
        text: 'text-yellow-800 dark:text-yellow-400',
        dot: 'bg-yellow-500',
      },
      down: {
        bg: 'bg-red-100 dark:bg-red-900/30',
        text: 'text-red-800 dark:text-red-400',
        dot: 'bg-red-500',
      },
    }
    return styles[status]
  }

  const getEventStyles = (type: SystemEvent['type']): string => {
    const styles = {
      info: 'border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20',
      warning: 'border-yellow-200 dark:border-yellow-800 bg-yellow-50 dark:bg-yellow-900/20',
      error: 'border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20',
    }
    return styles[type]
  }

  const overallStatus = healthData.every((h) => h.status === 'healthy')
    ? 'healthy'
    : healthData.some((h) => h.status === 'down')
    ? 'down'
    : 'degraded'

  return (
    <div className="max-w-6xl mx-auto">
      <PageHeader
        title={t('integrations.systemHealth')}
        subtitle={t('integrations.systemHealthDescription')}
        actions={
          <div className="flex items-center gap-3">
            <span className={`text-sm ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
              {t('integrations.lastRefresh')}: {lastRefresh.toLocaleTimeString()}
            </span>
            <button
              type="button"
              onClick={handleRefresh}
              disabled={isLoading}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white text-sm font-medium rounded-lg transition-colors"
            >
              {isLoading ? t('common.refreshing') : t('common.refresh')}
            </button>
          </div>
        }
      />

      {/* Overall Status */}
      <div className={`mb-6 p-4 rounded-lg border ${getStatusStyles(overallStatus).bg} border-current`}>
        <div className="flex items-center gap-3">
          <div className={`w-4 h-4 rounded-full ${getStatusStyles(overallStatus).dot} ${overallStatus === 'degraded' ? 'animate-pulse' : ''}`} />
          <span className={`text-lg font-semibold ${getStatusStyles(overallStatus).text}`}>
            {overallStatus === 'healthy' ? t('integrations.allSystemsOperational') : overallStatus === 'degraded' ? t('integrations.someSystemsDegraded') : t('integrations.someSystemsDown')}
          </span>
        </div>
      </div>

      {/* Health Status Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        {healthData.map((service) => {
          const statusStyles = getStatusStyles(service.status)
          return (
            <div
              key={service.name}
              className={`rounded-lg border p-4 ${isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}
            >
              <div className="flex items-center justify-between mb-3">
                <h3 className={`text-sm font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}>
                  {service.name}
                </h3>
                <span className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium ${statusStyles.bg} ${statusStyles.text}`}>
                  <span className={`w-2 h-2 rounded-full ${statusStyles.dot}`} />
                  {service.status.toUpperCase()}
                </span>
              </div>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className={isDark ? 'text-gray-400' : 'text-gray-500'}>{t('integrations.responseTime')}:</span>
                  <span className={`ml-2 font-medium ${isDark ? 'text-gray-200' : 'text-gray-900'}`}>
                    {service.responseTime}ms
                  </span>
                </div>
                <div>
                  <span className={isDark ? 'text-gray-400' : 'text-gray-500'}>{t('integrations.uptime')}:</span>
                  <span className={`ml-2 font-medium ${isDark ? 'text-gray-200' : 'text-gray-900'}`}>
                    {service.uptime}
                  </span>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Recent Events */}
      <div className={`rounded-lg border ${isDark ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'}`}>
        <div className={`px-4 py-3 border-b ${isDark ? 'border-gray-700' : 'border-gray-200'}`}>
          <h3 className={`text-sm font-semibold ${isDark ? 'text-white' : 'text-gray-900'}`}>
            {t('integrations.recentEvents')}
          </h3>
        </div>
        <div className="divide-y dark:divide-gray-700">
          {events.map((event) => (
            <div
              key={event.id}
              className={`p-4 border-l-4 ${getEventStyles(event.type)}`}
            >
              <div className="flex items-center justify-between">
                <p className={`text-sm ${isDark ? 'text-gray-200' : 'text-gray-800'}`}>
                  {event.message}
                </p>
                <span className={`text-xs ${isDark ? 'text-gray-500' : 'text-gray-400'}`}>
                  {new Date(event.timestamp).toLocaleString()}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
