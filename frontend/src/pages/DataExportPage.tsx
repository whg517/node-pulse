import { useCallback, useEffect, useState } from 'react'
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

  const loadNodes = useCallback(async () => {
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
  }, [])

  useEffect(() => {
    void loadNodes()
    void fetchExportHistory()

    return () => {
      useExportStore.getState().stopAllPolling()
    }
  }, [fetchExportHistory, loadNodes])

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
      />

      {/* Error State */}
      {error && (
        <ErrorBanner error={error} onRetry={loadNodes} />
      )}

      {/* Access Warning for Non-Admin Users */}
      {!isLoading && !error && user?.role !== 'admin' && (
        <div className="access-warning bg-[var(--color-warning-bg)] border-l-4 border-[var(--color-warning)] p-4 rounded-md mb-6">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg
                className="h-5 w-5 text-[var(--color-warning)]"
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
              <h3 className="text-sm font-medium text-[var(--color-warning-text)]">{t('dataExport.adminOnly')}</h3>
              <p className="text-sm text-[var(--color-warning-text)] mt-1">
                {t('dataExport.adminOnlyDescription')}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Loading State */}
      {isLoading && !error && (
        <div className="flex justify-center items-center py-12" data-testid="loading-spinner">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--color-brand)]"></div>
        </div>
      )}

      {/* Content - Export Form and Current Exports (Admin Only) */}
      {!isLoading && !error && user?.role === 'admin' && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Export Form */}
          <div className="bg-[var(--color-bg-surface)] rounded-lg shadow-md p-6 border border-[var(--color-border)]">
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">{t('dataExport.createExport')}</h3>
            <ExportForm
              nodes={nodes}
              onSubmit={handleExportSubmit}
              loading={exportLoading}
            />
          </div>

          {/* Current Exports */}
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)]">{t('dataExport.activeExports')}</h3>
            {currentExports.length === 0 ? (
              <div className="bg-[var(--color-bg-surface)] rounded-lg shadow-md p-6 text-center text-[var(--color-text-muted)] border border-[var(--color-border)]">
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
        <div className="mt-8 bg-[var(--color-bg-surface)] rounded-lg shadow-md p-6 border border-[var(--color-border)]">
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
