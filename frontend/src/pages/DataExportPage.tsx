import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../stores/authStore'
import { useExportStore } from '../stores/exportStore'
import { fetchNodes } from '../api/nodes'
import { ExportForm, ExportStatusCard, ExportHistoryTable } from '../components/export'
import type { NodeDTO } from '../api/types'
import type { CreateExportRequest } from '../types/export'
import { PageContainer, ErrorBanner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'

export default function DataExportPage() {
  const { t } = useTranslation()
  const { user } = useAuthStore()
  const {
    createExport,
    currentExports,
    exportHistory,
    downloadExport,
    fetchExportHistory,
    isLoading: exportLoading,
  } = useExportStore()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [nodes, setNodes] = useState<NodeDTO[]>([])

  useEffect(() => {
    loadNodes()
    fetchExportHistory()

    // Cleanup function - stop all polling when component unmounts
    return () => {
      useExportStore.getState().stopAllPolling()
    }
  }, [])

  const loadNodes = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await fetchNodes()
      setNodes(response.data.nodes ?? [])
    } catch (err) {
      setError(err as Error)
      console.error('Failed to load nodes:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleExportSubmit = async (request: CreateExportRequest) => {
    try {
      await createExport(request)
    } catch (error) {
      console.error('Failed to create export:', error)
      throw error
    }
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('dataExport.title')}
        subtitle={t('dataExport.description')}
        showBreadcrumb
      />

      {/* Error State */}
      {error && (
        <ErrorBanner error={error} onRetry={loadNodes} />
      )}

      {/* Access Warning for Non-Admin Users */}
      {!isLoading && !error && user?.role !== 'admin' && (
        <div className="access-warning bg-yellow-50 dark:bg-yellow-900/20 border-l-4 border-yellow-400 dark:border-yellow-600 p-4 rounded-md mb-6">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg
                className="h-5 w-5 text-yellow-400"
                fill="currentColor"
                viewBox="0 0 20 20"
                aria-hidden="true"
              >
                <path
                  fillRule="evenodd"
                  d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-yellow-800 dark:text-yellow-300">{t('dataExport.adminOnly')}</h3>
              <p className="text-sm text-yellow-700 dark:text-yellow-400 mt-1">
                {t('dataExport.adminOnlyDescription')}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Loading State */}
      {isLoading && !error && (
        <div className="flex justify-center items-center py-12" data-testid="loading-spinner">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      )}

      {/* Content - Export Form and Current Exports (Admin Only) */}
      {!isLoading && !error && user?.role === 'admin' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Export Form */}
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">{t('dataExport.createExport')}</h3>
            <ExportForm
              nodes={nodes}
              onSubmit={handleExportSubmit}
              loading={exportLoading}
            />
          </div>

          {/* Current Exports */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">{t('dataExport.activeExports')}</h3>
            {currentExports.length === 0 ? (
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 text-center text-gray-500 dark:text-gray-400 border border-gray-200 dark:border-gray-700">
                <p>{t('dataExport.noActiveExports')}</p>
              </div>
            ) : (
              currentExports.map((exportTask) => (
                <ExportStatusCard
                  key={exportTask.id}
                  exportTask={exportTask}
                  onDownload={downloadExport}
                />
              ))
            )}
          </div>
        </div>
      )}

      {/* Export History */}
      {!isLoading && !error && exportHistory.length > 0 && (
        <div className="mt-8 bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border border-gray-200 dark:border-gray-700">
          <ExportHistoryTable
            exports={exportHistory}
            onDownload={downloadExport}
            onDelete={() => {
              /* Delete not implemented in MVP */
            }}
          />
        </div>
      )}
    </PageContainer>
  )
}
