/**
 * System Health Page
 *
 * System health monitoring and integration status.
 * Route: /integrations/health
 */

import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PageContainer, ActionButton } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
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
    lastCheck: '2026-01-01T00:00:00Z',
  },
  {
    name: 'Database (PostgreSQL)',
    status: 'healthy',
    responseTime: 12,
    uptime: '99.99%',
    lastCheck: '2026-01-01T00:00:00Z',
  },
  {
    name: 'Cache Layer',
    status: 'healthy',
    responseTime: 2,
    uptime: '100%',
    lastCheck: '2026-01-01T00:00:00Z',
  },
  {
    name: 'Alert Engine',
    status: 'healthy',
    responseTime: 8,
    uptime: '99.95%',
    lastCheck: '2026-01-01T00:00:00Z',
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
    timestamp: '2026-01-01T00:00:00Z',
  },
  {
    id: '2',
    type: 'warning',
    message: 'High memory usage detected on API server (85%)',
    timestamp: '2025-12-31T23:00:00Z',
  },
  {
    id: '3',
    type: 'info',
    message: 'Database backup completed',
    timestamp: '2025-12-31T22:00:00Z',
  },
]

export default function SystemHealthPage() {
  const { t } = useTranslation()
  const [healthData] = useState<HealthStatus[]>(mockHealthData)
  const [events] = useState<SystemEvent[]>(mockEvents)
  const [isLoading, setIsLoading] = useState(false)
  const [lastRefresh, setLastRefresh] = useState(new Date('2026-01-01T00:00:00Z'))

  const handleRefresh = async () => {
    setIsLoading(true)
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1000))
    setLastRefresh(new Date('2026-01-01T00:00:00Z'))
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
    <PageContainer>
      <PageHeader
        title={t('integrations.systemHealth')}
        subtitle={t('integrations.systemHealthDescription')}
        showBreadcrumb
        actions={
          <div className="flex flex-shrink-0 items-center gap-3">
            <span className="text-sm text-gray-500 dark:text-gray-400">
              {t('integrations.lastRefresh')}: {lastRefresh.toLocaleTimeString()}
            </span>
            <ActionButton onClick={handleRefresh} disabled={isLoading}>
              {isLoading ? t('common.refreshing') : t('common.refresh')}
            </ActionButton>
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
              className="rounded-lg border p-4 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700"
            >
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-gray-900 dark:text-white">
                  {service.name}
                </h3>
                <span className={`inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium ${statusStyles.bg} ${statusStyles.text}`}>
                  <span className={`w-2 h-2 rounded-full ${statusStyles.dot}`} />
                  {service.status.toUpperCase()}
                </span>
              </div>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-gray-500 dark:text-gray-400">{t('integrations.responseTime')}:</span>
                  <span className="ml-2 font-medium text-gray-900 dark:text-gray-200">
                    {service.responseTime}ms
                  </span>
                </div>
                <div>
                  <span className="text-gray-500 dark:text-gray-400">{t('integrations.uptime')}:</span>
                  <span className="ml-2 font-medium text-gray-900 dark:text-gray-200">
                    {service.uptime}
                  </span>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Recent Events */}
      <div className="rounded-lg border bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
        <div className="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
          <h3 className="text-sm font-semibold text-gray-900 dark:text-white">
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
                <p className="text-sm text-gray-800 dark:text-gray-200">
                  {event.message}
                </p>
                <span className="text-xs text-gray-400 dark:text-gray-500">
                  {new Date(event.timestamp).toLocaleString()}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </PageContainer>
  )
}
