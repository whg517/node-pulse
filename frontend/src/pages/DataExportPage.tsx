import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/authStore'
import { useExportStore } from '@/stores/exportStore'
import { fetchNodes } from '@/api/nodes'
import { ExportForm, ExportStatusCard, ExportHistoryTable } from '@/components/export'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import type { NodeDTO } from '@/api/types'
import type { CreateExportRequest } from '@/types/export'

export default function DataExportPage() {
  const { t } = useTranslation()
  const { user } = useAuthStore()
  const { createExport, currentExports, exportHistory, downloadExport, fetchExportHistory, isLoading: exportLoading } = useExportStore()
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
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadNodes()
    void fetchExportHistory()
    return () => { useExportStore.getState().stopAllPolling() }
  }, [fetchExportHistory, loadNodes])

  const handleExportSubmit = async (request: CreateExportRequest) => {
    await createExport(request)
  }

  return (
    <div className="space-y-6">
      <PageHeader title={t('dataExport.title')} subtitle={t('dataExport.description')} />

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error.message}
          <Button variant="link" size="sm" onClick={loadNodes}>{t('common.retry')}</Button>
        </div>
      )}

      {!isLoading && !error && user?.role !== 'admin' && (
        <div className="rounded-md border-l-4 border-yellow-500 bg-yellow-50 p-4 dark:bg-yellow-950">
          <p className="text-sm font-medium">{t('dataExport.adminOnly')}</p>
          <p className="text-sm">{t('dataExport.adminOnlyDescription')}</p>
        </div>
      )}

      {isLoading && !error && (
        <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
      )}

      {!isLoading && !error && user?.role === 'admin' && (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <div className="rounded-lg border bg-card p-6">
            <h3 className="mb-4 text-lg font-semibold">{t('dataExport.createExport')}</h3>
            <ExportForm nodes={nodes} onSubmit={handleExportSubmit} loading={exportLoading} />
          </div>
          <div className="space-y-4">
            <h3 className="text-lg font-semibold">{t('dataExport.activeExports')}</h3>
            {currentExports.length === 0 ? (
              <div className="rounded-lg border bg-card p-6 text-center text-muted-foreground">{t('dataExport.noActiveExports')}</div>
            ) : (
              currentExports.map((exportTask) => <ExportStatusCard key={exportTask.id} exportTask={exportTask} onDownload={downloadExport} />)
            )}
          </div>
        </div>
      )}

      {!isLoading && !error && exportHistory.length > 0 && (
        <div className="mt-8 rounded-lg border bg-card p-6">
          <ExportHistoryTable exports={exportHistory} onDownload={downloadExport} onDelete={() => {}} />
        </div>
      )}
    </div>
  )
}
